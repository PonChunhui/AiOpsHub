package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

var sshWebSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SSHWebSocketSession struct {
	HostID     string
	UserID     string
	SessionID  string
	Conn       *websocket.Conn
	SSHClient  *ssh.Client
	SSHSession *ssh.Session
	StartTime  time.Time
	IPAddress  string
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func HandleSSHWebSocket(c *gin.Context) {
	hostID := c.Param("host_id")
	if hostID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少主机ID"})
		return
	}

	host, err := hostService.GetHostByID(hostID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	userID := getUserID(c)
	sessionID := model.NewSSHSessionLog().SessionID
	ipAddress := c.ClientIP()

	conn, err := sshWebSocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("WebSocket升级失败: %v", err))
		return
	}

	session := &SSHWebSocketSession{
		HostID:    hostID,
		UserID:    userID,
		SessionID: sessionID,
		Conn:      conn,
		StartTime: time.Now(),
		IPAddress: ipAddress,
	}

	ctx, cancel := context.WithCancel(context.Background())
	session.ctx = ctx
	session.cancel = cancel

	_, err = hostService.CreateSSHSessionLog(hostID, userID, "connect", sessionID, ipAddress)
	if err != nil {
		logger.Error(fmt.Sprintf("创建SSH会话日志失败: %v", err))
	}

	logger.Info(fmt.Sprintf("SSH WebSocket连接建立: 主机=%s, 用户=%s, 会话=%s", host.Name, userID, sessionID))

	defer session.Close()

	sshConfig := &ssh.ClientConfig{
		User:            host.Username,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	if host.AuthType == "password" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.Password(host.Password))
	} else if host.AuthType == "key" && host.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(host.PrivateKey))
		if err != nil {
			session.SendMessage("error", "私钥解析失败: "+err.Error())
			return
		}
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	}

	sshAddr := fmt.Sprintf("%s:%d", host.IP, host.Port)
	sshClient, err := ssh.Dial("tcp", sshAddr, sshConfig)
	if err != nil {
		session.SendMessage("error", "SSH连接失败: "+err.Error())
		return
	}
	session.SSHClient = sshClient

	sshSession, err := sshClient.NewSession()
	if err != nil {
		session.SendMessage("error", "创建SSH会话失败: "+err.Error())
		return
	}
	session.SSHSession = sshSession

	sshSession.Stdin = session
	sshSession.Stdout = session
	sshSession.Stderr = session

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	err = sshSession.RequestPty("xterm", 80, 40, modes)
	if err != nil {
		session.SendMessage("error", "请求PTY失败: "+err.Error())
		return
	}

	err = sshSession.Shell()
	if err != nil {
		session.SendMessage("error", "启动Shell失败: "+err.Error())
		return
	}

	session.SendMessage("connected", "SSH连接成功")

	session.ReadFromWebSocket()
}

func (s *SSHWebSocketSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cancel()

	if s.SSHSession != nil {
		s.SSHSession.Close()
	}
	if s.SSHClient != nil {
		s.SSHClient.Close()
	}
	if s.Conn != nil {
		s.Conn.Close()
	}

	duration := int(time.Since(s.StartTime).Seconds())
	logger.Info(fmt.Sprintf("SSH WebSocket会话结束: 会话=%s, 持续时间=%d秒", s.SessionID, duration))
}

func (s *SSHWebSocketSession) SendMessage(messageType string, data string) {
	msg := map[string]interface{}{
		"type": messageType,
		"data": data,
		"time": time.Now().Unix(),
	}
	jsonData, _ := json.Marshal(msg)
	s.Conn.WriteMessage(websocket.TextMessage, jsonData)
}

func (s *SSHWebSocketSession) ReadFromWebSocket() {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			_, message, err := s.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Error(fmt.Sprintf("WebSocket读取错误: %v", err))
				}
				return
			}

			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}

			if msgType, ok := msg["type"].(string); ok {
				switch msgType {
				case "resize":
					cols, _ := msg["cols"].(float64)
					rows, _ := msg["rows"].(float64)
					if s.SSHSession != nil {
						s.SSHSession.WindowChange(int(rows), int(cols))
					}
				case "data":
					if data, ok := msg["data"].(string); ok {
						s.mu.Lock()
						if s.SSHSession != nil && s.SSHSession.Stdin != nil {
							if stdinWriter, ok := s.SSHSession.Stdin.(io.Writer); ok {
								io.WriteString(stdinWriter, data)
							}
						}
						s.mu.Unlock()
					}
				}
			}
		}
	}
}

func (s *SSHWebSocketSession) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (s *SSHWebSocketSession) Write(p []byte) (n int, err error) {
	s.SendMessage("data", string(p))
	return len(p), nil
}

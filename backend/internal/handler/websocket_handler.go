package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	Clients     map[string]*WebSocketClient
	ClientsLock sync.RWMutex
}

type WebSocketClient struct {
	ID            string
	Connection    *websocket.Conn
	Subscriptions map[string]bool
	SendChan      chan []byte
}

type WebSocketMessage struct {
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type SubscriptionRequest struct {
	Action    string `json:"action"`
	SessionID string `json:"session_id"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		Clients: make(map[string]*WebSocketClient),
	}
}

func (wsh *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("WebSocket upgrade failed: %v", err))
		return
	}

	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
	client := &WebSocketClient{
		ID:            clientID,
		Connection:    conn,
		Subscriptions: make(map[string]bool),
		SendChan:      make(chan []byte, 100),
	}

	wsh.ClientsLock.Lock()
	wsh.Clients[clientID] = client
	wsh.ClientsLock.Unlock()

	logger.Info(fmt.Sprintf("WebSocket client connected: %s", clientID))

	go wsh.writePump(client)
	go wsh.readPump(client)
}

func (wsh *WebSocketHandler) readPump(client *WebSocketClient) {
	defer func() {
		wsh.ClientsLock.Lock()
		delete(wsh.Clients, client.ID)
		wsh.ClientsLock.Unlock()
		client.Connection.Close()
		logger.Info(fmt.Sprintf("WebSocket client disconnected: %s", client.ID))
	}()

	for {
		_, message, err := client.Connection.ReadMessage()
		if err != nil {
			break
		}

		var req SubscriptionRequest
		err = json.Unmarshal(message, &req)
		if err != nil {
			logger.Error(fmt.Sprintf("Invalid WebSocket message: %v", err))
			continue
		}

		switch req.Action {
		case "subscribe":
			if req.SessionID != "" {
				client.Subscriptions[req.SessionID] = true
				logger.Info(fmt.Sprintf("Client %s subscribed to session %s", client.ID, req.SessionID))
			}

		case "unsubscribe":
			if req.SessionID != "" {
				delete(client.Subscriptions, req.SessionID)
			}
		}
	}
}

func (wsh *WebSocketHandler) writePump(client *WebSocketClient) {
	defer client.Connection.Close()

	for {
		message, ok := <-client.SendChan
		if !ok {
			return
		}

		err := client.Connection.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			logger.Error(fmt.Sprintf("WebSocket write error: %v", err))
			return
		}
	}
}

func (wsh *WebSocketHandler) BroadcastSessionUpdate(sessionID string, data map[string]interface{}) {
	message := WebSocketMessage{
		Type:      "session_update",
		Timestamp: time.Now().Unix(),
		Data:      data,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to marshal WebSocket message: %v", err))
		return
	}

	wsh.ClientsLock.RLock()
	defer wsh.ClientsLock.RUnlock()

	for _, client := range wsh.Clients {
		if client.Subscriptions[sessionID] {
			select {
			case client.SendChan <- messageBytes:
			default:
				logger.Warn(fmt.Sprintf("Client %s send channel full", client.ID))
			}
		}
	}
}

func (wsh *WebSocketHandler) BroadcastAgentStatus(agentID string, status string, progress int) {
	message := WebSocketMessage{
		Type:      "agent_status",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"agent_id": agentID,
			"status":   status,
			"progress": progress,
		},
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to marshal WebSocket message: %v", err))
		return
	}

	wsh.ClientsLock.RLock()
	defer wsh.ClientsLock.RUnlock()

	for _, client := range wsh.Clients {
		select {
		case client.SendChan <- messageBytes:
		default:
		}
	}
}

var GlobalWebSocketHandler *WebSocketHandler

func InitWebSocketHandler() {
	GlobalWebSocketHandler = NewWebSocketHandler()
	logger.Info("WebSocket handler initialized")
}

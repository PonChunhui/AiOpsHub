package eino_tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/crypto/ssh"
)

// SSHTool Eino标准的SSH工具实现
type SSHTool struct {
	tool   *model.Tool
	config map[string]interface{}
}

// SSHConfig SSH连接配置
type SSHConfig struct {
	Username   string        `json:"username"`
	Password   string        `json:"password"`
	PrivateKey string        `json:"private_key"`
	Port       int           `json:"port"`
	Timeout    time.Duration `json:"timeout"`
}

// NewSSHTool 创建SSH工具实例
func NewSSHTool(toolModel *model.Tool, configOverride map[string]interface{}) tool.InvokableTool {
	config := make(map[string]interface{})

	// 解析默认配置
	if toolModel.DefaultConfig != "" {
		json.Unmarshal([]byte(toolModel.DefaultConfig), &config)
	}

	// 应用配置覆盖
	for k, v := range configOverride {
		config[k] = v
	}

	return &SSHTool{
		tool:   toolModel,
		config: config,
	}
}

// Info 返回工具信息（Eino标准接口）
func (t *SSHTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	toolInfo := &schema.ToolInfo{
		Name: "ssh_exec",
		Desc: t.tool.Description,
	}

	// 设置参数（使用ParamsOneOf嵌入字段）
	toolInfo.ParamsOneOf = schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"host": {
			Type:     schema.String,
			Desc:     "服务器IP或主机名",
			Required: true,
		},
		"command": {
			Type:     schema.String,
			Desc:     "要执行的命令",
			Required: true,
		},
	})

	return toolInfo, nil
}

// InvokableRun 执行工具（Eino标准接口）
func (t *SSHTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	logger.Info(fmt.Sprintf("SSH工具开始执行: %s", argumentsInJSON))

	// 解析参数
	var args struct {
		Host    string `json:"host"`
		Command string `json:"command"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	// 验证参数
	if args.Host == "" {
		return "", fmt.Errorf("缺少host参数")
	}
	if args.Command == "" {
		return "", fmt.Errorf("缺少command参数")
	}

	// 检查命令白名单
	allowedCommands := []interface{}{}
	if ac, ok := t.config["allowed_commands"].([]interface{}); ok {
		allowedCommands = ac
	}

	commandAllowed := false
	for _, ac := range allowedCommands {
		if cmdPattern, ok := ac.(string); ok {
			if cmdPattern == "*" {
				commandAllowed = true
				break
			}
			if args.Command == cmdPattern {
				commandAllowed = true
				break
			}
		}
	}

	if !commandAllowed {
		return "", fmt.Errorf("命令 '%s' 不在白名单中，允许的命令: %v", args.Command, allowedCommands)
	}

	// 检查主机白名单
	allowedHosts := []interface{}{}
	if ah, ok := t.config["allowed_hosts"].([]interface{}); ok {
		allowedHosts = ah
	}

	hostAllowed := false
	for _, ah := range allowedHosts {
		if hostPattern, ok := ah.(string); ok {
			if hostPattern == "*" || args.Host == hostPattern {
				hostAllowed = true
				break
			}
		}
	}

	if !hostAllowed {
		return "", fmt.Errorf("主机 '%s' 不在白名单中", args.Host)
	}

	sshConfig := t.getSSHConfig()

	result, err := t.executeSSHCommand(ctx, args.Host, args.Command, sshConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("SSH命令执行失败: %v", err))
		return fmt.Sprintf("执行失败: %v", err), err
	}

	logger.Info(fmt.Sprintf("SSH工具执行成功: host=%s, command=%s, result_length=%d", args.Host, args.Command, len(result)))
	return result, nil
}

func (t *SSHTool) getSSHConfig() *SSHConfig {
	config := &SSHConfig{
		Username: "root",
		Password: "idc.linux66.CN",
		Port:     22,
		Timeout:  30 * time.Second,
	}

	if username, ok := t.config["username"].(string); ok && username != "" {
		config.Username = username
	} else {
		config.Username = "root"
	}

	if password, ok := t.config["password"].(string); ok && password != "" {
		config.Password = password
	}

	if privateKey, ok := t.config["private_key"].(string); ok && privateKey != "" {
		config.PrivateKey = privateKey
	}

	if port, ok := t.config["port"].(int); ok && port > 0 {
		config.Port = port
	}

	if timeout, ok := t.config["timeout"].(int); ok && timeout > 0 {
		config.Timeout = time.Duration(timeout) * time.Second
	}

	return config
}

func (t *SSHTool) executeSSHCommand(ctx context.Context, host, command string, sshConfig *SSHConfig) (string, error) {
	address := fmt.Sprintf("%s:%d", host, sshConfig.Port)

	var authMethods []ssh.AuthMethod

	if sshConfig.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(sshConfig.PrivateKey))
		if err != nil {
			return "", fmt.Errorf("解析私钥失败: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if sshConfig.Password != "" {
		authMethods = append(authMethods, ssh.Password(sshConfig.Password))
	}

	if len(authMethods) == 0 {
		return "", fmt.Errorf("缺少认证方式（密码或私钥）")
	}

	clientConfig := &ssh.ClientConfig{
		User:            sshConfig.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         sshConfig.Timeout,
	}

	connCtx, cancel := context.WithTimeout(ctx, sshConfig.Timeout)
	defer cancel()

	var client *ssh.Client
	var err error

	done := make(chan struct{})
	go func() {
		client, err = ssh.Dial("tcp", address, clientConfig)
		close(done)
	}()

	select {
	case <-done:
		if err != nil {
			return "", fmt.Errorf("SSH连接失败 (%s): %w", address, err)
		}
	case <-connCtx.Done():
		return "", fmt.Errorf("SSH连接超时 (%s)", address)
	}

	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)

	result := stdout.String()
	if stderr.Len() > 0 {
		result += "\n[stderr]:\n" + stderr.String()
	}

	if err != nil {
		return result, fmt.Errorf("命令执行失败: %w", err)
	}

	return result, nil
}

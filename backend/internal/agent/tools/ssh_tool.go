package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiops/AiOpsHub/backend/internal/agent"
	"github.com/aiops/AiOpsHub/backend/internal/model"
)

type SSHTool struct {
	metadata        agent.ToolMetadata
	allowedCommands []string
	allowedHosts    []string
	timeout         int
}

func init() {
	agent.RegisterToolFactory("ssh_exec", NewSSHToolFromConfig)
}

func NewSSHToolFromConfig(tool *model.Tool, overrideConfig string) (agent.Tool, error) {
	config := agent.ParseConfig(tool.DefaultConfig, overrideConfig)

	allowedCommands := GetStringSlice(config, "allowed_commands", []string{"ls", "top", "free", "df", "ps"})
	allowedHosts := GetStringSlice(config, "allowed_hosts", []string{"*"})
	timeout := GetInt(config, "timeout", 30)

	return &SSHTool{
		metadata: agent.ToolMetadata{
			ID:               tool.ID,
			Name:             tool.Name,
			Description:      tool.Description,
			Category:         tool.Category,
			RiskLevel:        tool.RiskLevel,
			ExecutionTimeout: timeout,
		},
		allowedCommands: allowedCommands,
		allowedHosts:    allowedHosts,
		timeout:         timeout,
	}, nil
}

func (t *SSHTool) Name() string {
	return t.metadata.Name
}

func (t *SSHTool) Description() string {
	return t.metadata.Description
}

func (t *SSHTool) ParametersSchema() string {
	return `{
		"type": "object",
		"properties": {
			"host": {"type": "string", "description": "服务器IP或主机名"},
			"command": {"type": "string", "description": "要执行的命令"}
		},
		"required": ["host", "command"]
	}`
}

func (t *SSHTool) Call(ctx context.Context, input map[string]interface{}) (string, error) {
	host, ok := input["host"].(string)
	if !ok {
		return "", fmt.Errorf("host parameter required")
	}

	command, ok := input["command"].(string)
	if !ok {
		return "", fmt.Errorf("command parameter required")
	}

	if err := ValidateCommand(command); err != nil {
		return "", err
	}

	if !t.IsCommandAllowed(command) {
		return "", fmt.Errorf("command not allowed: %s (allowed commands: %v)", command, t.allowedCommands)
	}

	if !t.IsHostAllowed(host) {
		return "", fmt.Errorf("host not allowed: %s (allowed hosts: %v)", host, t.allowedHosts)
	}

	result := fmt.Sprintf("Mock SSH execution on %s: %s\nOutput: [mock result]", host, command)

	return result, nil
}

func ValidateCommand(command string) error {
	dangerousChars := []string{";", "|", "&", "$", "`", ">", "<", ".."}
	for _, char := range dangerousChars {
		if strings.Contains(command, char) {
			return fmt.Errorf("command contains dangerous character: %s", char)
		}
	}
	return nil
}

func (t *SSHTool) IsCommandAllowed(command string) bool {
	for _, allowed := range t.allowedCommands {
		if allowed == "*" {
			continue
		}

		if !strings.Contains(allowed, " ") {
			parts := strings.Fields(command)
			if len(parts) > 0 && parts[0] == allowed {
				return true
			}
			continue
		}

		if strings.HasPrefix(allowed, "cat ") {
			path := strings.TrimPrefix(command, "cat ")
			if strings.Contains(path, "..") {
				return false
			}
			return strings.HasPrefix(path, strings.TrimPrefix(allowed, "cat "))
		}
	}
	return false
}

func (t *SSHTool) IsHostAllowed(host string) bool {
	for _, allowed := range t.allowedHosts {
		if allowed == "*" {
			return true
		}
		if host == allowed {
			return true
		}
		if strings.HasSuffix(allowed, "*") {
			pattern := strings.TrimSuffix(allowed, "*")
			return strings.HasPrefix(host, pattern)
		}
	}
	return false
}

func GetStringSlice(m map[string]interface{}, key string, defaultVal []string) []string {
	if val, ok := m[key].([]interface{}); ok {
		result := []string{}
		for _, v := range val {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return defaultVal
}

func GetInt(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key].(int); ok {
		return val
	}
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return defaultVal
}

func GetString(m map[string]interface{}, key string, defaultVal string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultVal
}

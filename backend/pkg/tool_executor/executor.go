package tool_executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type ToolExecutor interface {
	Execute(ctx context.Context, tool *model.Tool, configOverride map[string]interface{}, args map[string]interface{}) (string, error)
}

type DefaultToolExecutor struct{}

func NewDefaultToolExecutor() *DefaultToolExecutor {
	return &DefaultToolExecutor{}
}

func (e *DefaultToolExecutor) Execute(ctx context.Context, tool *model.Tool, configOverride map[string]interface{}, args map[string]interface{}) (string, error) {
	logger.Info(fmt.Sprintf("执行工具: %s (类型: %s)", tool.Name, tool.Type))

	timeout := time.Duration(tool.ExecutionTimeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	switch tool.Type {
	case "builtin":
		return e.executeBuiltinTool(ctxWithTimeout, tool, configOverride, args)
	case "mcp":
		return e.executeMCPTool(ctxWithTimeout, tool, configOverride, args)
	default:
		return "", fmt.Errorf("不支持的工具类型: %s", tool.Type)
	}
}

func (e *DefaultToolExecutor) executeBuiltinTool(ctx context.Context, tool *model.Tool, configOverride map[string]interface{}, args map[string]interface{}) (string, error) {
	switch tool.Name {
	case "ssh_exec":
		return e.executeSSH(ctx, tool, configOverride, args)
	case "prometheus_query":
		return e.executePrometheus(ctx, tool, configOverride, args)
	case "kubernetes_query":
		return e.executeKubernetes(ctx, tool, configOverride, args)
	case "log_query":
		return e.executeLogQuery(ctx, tool, configOverride, args)
	default:
		return fmt.Sprintf("工具 %s 的执行逻辑待实现", tool.Name), nil
	}
}

func (e *DefaultToolExecutor) executeMCPTool(ctx context.Context, tool *model.Tool, configOverride map[string]interface{}, args map[string]interface{}) (string, error) {
	return fmt.Sprintf("MCP工具 %s 的执行逻辑待实现", tool.Name), nil
}

func (e *DefaultToolExecutor) executeSSH(ctx context.Context, tool *model.Tool, configOverride map[string]interface{}, args map[string]interface{}) (string, error) {
	host, ok := args["host"].(string)
	if !ok {
		return "", fmt.Errorf("缺少host参数")
	}

	command, ok := args["command"].(string)
	if !ok {
		return "", fmt.Errorf("缺少command参数")
	}

	config := make(map[string]interface{})
	if tool.DefaultConfig != "" {
		json.Unmarshal([]byte(tool.DefaultConfig), &config)
	}

	for k, v := range configOverride {
		config[k] = v
	}

	allowedCommands := []interface{}{}
	if ac, ok := config["allowed_commands"].([]interface{}); ok {
		allowedCommands = ac
	}

	commandAllowed := false
	for _, ac := range allowedCommands {
		if cmdPattern, ok := ac.(string); ok {
			if command == cmdPattern || strings.HasPrefix(cmdPattern, command) {
				commandAllowed = true
				break
			}
		}
	}

	if !commandAllowed {
		return "", fmt.Errorf("命令 %s 不在白名单中", command)
	}

	return fmt.Sprintf("SSH执行模拟: host=%s, command=%s (待实现真实SSH客户端)", host, command), nil
}

func (e *DefaultToolExecutor) executePrometheus(ctx context.Context, tool *model.Tool, configOverride map[string]interface{}, args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("缺少query参数")
	}

	timeRange := ""
	if tr, ok := args["time_range"].(string); ok {
		timeRange = tr
	}

	return fmt.Sprintf("Prometheus查询模拟: query=%s, time_range=%s (待实现真实Prometheus客户端)", query, timeRange), nil
}

func (e *DefaultToolExecutor) executeKubernetes(ctx context.Context, tool *model.Tool, configOverride map[string]interface{}, args map[string]interface{}) (string, error) {
	resourceType := ""
	if rt, ok := args["resource_type"].(string); ok {
		resourceType = rt
	}

	namespace := ""
	if ns, ok := args["namespace"].(string); ok {
		namespace = ns
	}

	name := ""
	if n, ok := args["name"].(string); ok {
		name = n
	}

	return fmt.Sprintf("Kubernetes查询模拟: type=%s, namespace=%s, name=%s (待实现真实K8s客户端)", resourceType, namespace, name), nil
}

func (e *DefaultToolExecutor) executeLogQuery(ctx context.Context, tool *model.Tool, configOverride map[string]interface{}, args map[string]interface{}) (string, error) {
	service := ""
	if s, ok := args["service"].(string); ok {
		service = s
	}

	level := ""
	if l, ok := args["level"].(string); ok {
		level = l
	}

	timeRange := ""
	if tr, ok := args["time_range"].(string); ok {
		timeRange = tr
	}

	return fmt.Sprintf("日志查询模拟: service=%s, level=%s, time_range=%s (待实现真实日志客户端)", service, level, timeRange), nil
}

package tools

import (
	"context"
	"fmt"

	"github.com/aiops/AiOpsHub/backend/internal/agent"
	"github.com/aiops/AiOpsHub/backend/internal/model"
)

type LogQueryTool struct {
	metadata agent.ToolMetadata
	timeout  int
}

func init() {
	agent.RegisterToolFactory("log_query", NewLogQueryToolFromConfig)
}

func NewLogQueryToolFromConfig(tool *model.Tool, overrideConfig string) (agent.Tool, error) {
	config := agent.ParseConfig(tool.DefaultConfig, overrideConfig)

	timeout := GetInt(config, "timeout", 20)

	return &LogQueryTool{
		metadata: agent.ToolMetadata{
			ID:               tool.ID,
			Name:             tool.Name,
			Description:      tool.Description,
			Category:         tool.Category,
			RiskLevel:        tool.RiskLevel,
			ExecutionTimeout: timeout,
		},
		timeout: timeout,
	}, nil
}

func (t *LogQueryTool) Name() string {
	return t.metadata.Name
}

func (t *LogQueryTool) Description() string {
	return t.metadata.Description
}

func (t *LogQueryTool) ParametersSchema() string {
	return `{
		"type": "object",
		"properties": {
			"service": {"type": "string", "description": "服务名称"},
			"level": {"type": "string", "description": "日志级别，如error, warning"},
			"time_range": {"type": "string", "description": "时间范围"}
		}
	}`
}

func (t *LogQueryTool) Call(ctx context.Context, input map[string]interface{}) (string, error) {
	service, ok := input["service"].(string)
	if !ok {
		service = "default"
	}

	level, ok := input["level"].(string)
	if !ok {
		level = "error"
	}

	result := fmt.Sprintf("Mock Log query: service %s, level %s\nResult: [mock logs]", service, level)
	return result, nil
}

package tools

import (
	"context"
	"fmt"

	"github.com/aiops/AiOpsHub/backend/internal/agent"
	"github.com/aiops/AiOpsHub/backend/internal/model"
)

type PrometheusTool struct {
	metadata agent.ToolMetadata
	url      string
	timeout  int
}

func init() {
	agent.RegisterToolFactory("prometheus_query", NewPrometheusToolFromConfig)
}

func NewPrometheusToolFromConfig(tool *model.Tool, overrideConfig string) (agent.Tool, error) {
	config := agent.ParseConfig(tool.DefaultConfig, overrideConfig)

	url := GetString(config, "url", "http://localhost:9090")
	timeout := GetInt(config, "timeout", 10)

	return &PrometheusTool{
		metadata: agent.ToolMetadata{
			ID:               tool.ID,
			Name:             tool.Name,
			Description:      tool.Description,
			Category:         tool.Category,
			RiskLevel:        tool.RiskLevel,
			ExecutionTimeout: timeout,
		},
		url:     url,
		timeout: timeout,
	}, nil
}

func (t *PrometheusTool) Name() string {
	return t.metadata.Name
}

func (t *PrometheusTool) Description() string {
	return t.metadata.Description
}

func (t *PrometheusTool) ParametersSchema() string {
	return `{
		"type": "object",
		"properties": {
			"query": {"type": "string", "description": "PromQL查询语句"},
			"time_range": {"type": "string", "description": "时间范围，如-1h"}
		},
		"required": ["query"]
	}`
}

func (t *PrometheusTool) Call(ctx context.Context, input map[string]interface{}) (string, error) {
	query, ok := input["query"].(string)
	if !ok {
		return "", fmt.Errorf("query parameter required")
	}

	result := fmt.Sprintf("Mock Prometheus query: %s\nResult: [mock data]", query)
	return result, nil
}

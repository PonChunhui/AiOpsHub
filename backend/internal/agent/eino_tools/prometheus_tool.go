package eino_tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// PrometheusTool Eino标准的Prometheus工具实现
type PrometheusTool struct {
	tool   *model.Tool
	config map[string]interface{}
}

// NewPrometheusTool 创建Prometheus工具实例
func NewPrometheusTool(toolModel *model.Tool, configOverride map[string]interface{}) tool.InvokableTool {
	config := make(map[string]interface{})

	if toolModel.DefaultConfig != "" {
		json.Unmarshal([]byte(toolModel.DefaultConfig), &config)
	}

	for k, v := range configOverride {
		config[k] = v
	}

	return &PrometheusTool{
		tool:   toolModel,
		config: config,
	}
}

// Info 返回工具信息
func (t *PrometheusTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	toolInfo := &schema.ToolInfo{
		Name: "prometheus_query",
		Desc: t.tool.Description,
	}

	toolInfo.ParamsOneOf = schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"query": {
			Type:     schema.String,
			Desc:     "PromQL查询语句",
			Required: true,
		},
		"time_range": {
			Type:     schema.String,
			Desc:     "时间范围，如 1h, 24h, 7d",
			Required: false,
		},
	})

	return toolInfo, nil
}

// InvokableRun 执行工具
func (t *PrometheusTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	logger.Info(fmt.Sprintf("Prometheus工具开始执行: %s", argumentsInJSON))

	var args struct {
		Query     string `json:"query"`
		TimeRange string `json:"time_range"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if args.Query == "" {
		return "", fmt.Errorf("缺少query参数")
	}

	// 获取Prometheus URL
	promURL := "http://prometheus:9090"
	if url, ok := t.config["url"].(string); ok && url != "" {
		promURL = url
	}

	// 模拟执行
	result := fmt.Sprintf("Prometheus查询模拟: url=%s, query=%s, time_range=%s (待实现真实Prometheus客户端)",
		promURL, args.Query, args.TimeRange)

	logger.Info(fmt.Sprintf("Prometheus工具执行成功"))

	return result, nil
}

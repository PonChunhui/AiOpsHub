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

// LogQueryTool Eino标准的日志查询工具实现
type LogQueryTool struct {
	tool   *model.Tool
	config map[string]interface{}
}

// NewLogQueryTool 创建日志查询工具实例
func NewLogQueryTool(toolModel *model.Tool, configOverride map[string]interface{}) tool.InvokableTool {
	config := make(map[string]interface{})

	if toolModel.DefaultConfig != "" {
		json.Unmarshal([]byte(toolModel.DefaultConfig), &config)
	}

	for k, v := range configOverride {
		config[k] = v
	}

	return &LogQueryTool{
		tool:   toolModel,
		config: config,
	}
}

// Info 返回工具信息
func (t *LogQueryTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	toolInfo := &schema.ToolInfo{
		Name: "log_query",
		Desc: t.tool.Description,
	}

	toolInfo.ParamsOneOf = schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"service": {
			Type:     schema.String,
			Desc:     "服务名称",
			Required: true,
		},
		"level": {
			Type:     schema.String,
			Desc:     "日志级别，如 error, warn, info",
			Required: false,
		},
		"time_range": {
			Type:     schema.String,
			Desc:     "时间范围",
			Required: false,
		},
	})

	return toolInfo, nil
}

// InvokableRun 执行工具
func (t *LogQueryTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	logger.Info(fmt.Sprintf("日志查询工具开始执行: %s", argumentsInJSON))

	var args struct {
		Service   string `json:"service"`
		Level     string `json:"level"`
		TimeRange string `json:"time_range"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if args.Service == "" {
		return "", fmt.Errorf("缺少service参数")
	}

	// 模拟执行
	result := fmt.Sprintf("日志查询模拟: service=%s, level=%s, time_range=%s (待实现真实日志客户端)",
		args.Service, args.Level, args.TimeRange)

	logger.Info(fmt.Sprintf("日志查询工具执行成功"))

	return result, nil
}

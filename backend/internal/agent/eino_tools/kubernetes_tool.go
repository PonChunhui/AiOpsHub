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

// KubernetesTool Eino标准的Kubernetes工具实现
type KubernetesTool struct {
	tool   *model.Tool
	config map[string]interface{}
}

// NewKubernetesTool 创建Kubernetes工具实例
func NewKubernetesTool(toolModel *model.Tool, configOverride map[string]interface{}) tool.InvokableTool {
	config := make(map[string]interface{})

	if toolModel.DefaultConfig != "" {
		json.Unmarshal([]byte(toolModel.DefaultConfig), &config)
	}

	for k, v := range configOverride {
		config[k] = v
	}

	return &KubernetesTool{
		tool:   toolModel,
		config: config,
	}
}

// Info 返回工具信息
func (t *KubernetesTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	toolInfo := &schema.ToolInfo{
		Name: "kubernetes_query",
		Desc: t.tool.Description,
	}

	toolInfo.ParamsOneOf = schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
		"resource_type": {
			Type:     schema.String,
			Desc:     "资源类型，如 pod, service, deployment",
			Required: true,
		},
		"namespace": {
			Type:     schema.String,
			Desc:     "命名空间",
			Required: false,
		},
		"name": {
			Type:     schema.String,
			Desc:     "资源名称",
			Required: false,
		},
	})

	return toolInfo, nil
}

// InvokableRun 执行工具
func (t *KubernetesTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	logger.Info(fmt.Sprintf("Kubernetes工具开始执行: %s", argumentsInJSON))

	var args struct {
		ResourceType string `json:"resource_type"`
		Namespace    string `json:"namespace"`
		Name         string `json:"name"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if args.ResourceType == "" {
		return "", fmt.Errorf("缺少resource_type参数")
	}

	// 模拟执行
	result := fmt.Sprintf("Kubernetes查询模拟: type=%s, namespace=%s, name=%s (待实现真实K8s客户端)",
		args.ResourceType, args.Namespace, args.Name)

	logger.Info(fmt.Sprintf("Kubernetes工具执行成功"))

	return result, nil
}

package tools

import (
	"context"
	"fmt"

	"github.com/aiops/AiOpsHub/backend/internal/agent"
	"github.com/aiops/AiOpsHub/backend/internal/model"
)

type KubernetesTool struct {
	metadata agent.ToolMetadata
	timeout  int
}

func init() {
	agent.RegisterToolFactory("kubernetes_query", NewKubernetesToolFromConfig)
}

func NewKubernetesToolFromConfig(tool *model.Tool, overrideConfig string) (agent.Tool, error) {
	config := agent.ParseConfig(tool.DefaultConfig, overrideConfig)

	timeout := GetInt(config, "timeout", 15)

	return &KubernetesTool{
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

func (t *KubernetesTool) Name() string {
	return t.metadata.Name
}

func (t *KubernetesTool) Description() string {
	return t.metadata.Description
}

func (t *KubernetesTool) ParametersSchema() string {
	return `{
		"type": "object",
		"properties": {
			"resource_type": {"type": "string", "description": "资源类型，如pods, services"},
			"namespace": {"type": "string", "description": "命名空间"},
			"name": {"type": "string", "description": "资源名称"}
		}
	}`
}

func (t *KubernetesTool) Call(ctx context.Context, input map[string]interface{}) (string, error) {
	resourceType, ok := input["resource_type"].(string)
	if !ok {
		resourceType = "pods"
	}

	namespace, ok := input["namespace"].(string)
	if !ok {
		namespace = "default"
	}

	result := fmt.Sprintf("Mock Kubernetes query: %s in namespace %s\nResult: [mock data]", resourceType, namespace)
	return result, nil
}

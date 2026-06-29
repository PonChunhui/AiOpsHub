package agent

import "context"

type Tool interface {
	Name() string
	Description() string
	ParametersSchema() string
	Call(ctx context.Context, input map[string]interface{}) (string, error)
}

type ToolMetadata struct {
	ID               string
	Name             string
	Description      string
	Category         string
	RiskLevel        string
	ExecutionTimeout int
}

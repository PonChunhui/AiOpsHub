package eino_tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/aiops/AiOpsHub/backend/pkg/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type MCPTool struct {
	serverID string
	toolDef  mcp.Tool
	mcpSvc   interface {
		CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]interface{}) (*mcp.ToolCallResult, error)
	}
}

func NewMCPTool(serverID string, toolDef mcp.Tool, mcpSvc interface {
	CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]interface{}) (*mcp.ToolCallResult, error)
}) tool.InvokableTool {
	return &MCPTool{
		serverID: serverID,
		toolDef:  toolDef,
		mcpSvc:   mcpSvc,
	}
}

func (t *MCPTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	toolInfo := &schema.ToolInfo{
		Name: t.toolDef.Name,
		Desc: t.toolDef.Description,
	}

	if t.toolDef.InputSchema != nil {
		params := make(map[string]*schema.ParameterInfo)

		if properties, ok := t.toolDef.InputSchema["properties"].(map[string]interface{}); ok {
			for paramName, propDef := range properties {
				propMap, ok := propDef.(map[string]interface{})
				if !ok {
					continue
				}

				paramInfo := &schema.ParameterInfo{
					Desc:     "",
					Required: false,
				}

				if desc, ok := propMap["description"].(string); ok {
					paramInfo.Desc = desc
				}

				if typeStr, ok := propMap["type"].(string); ok {
					switch typeStr {
					case "string":
						paramInfo.Type = schema.String
					case "number", "integer":
						paramInfo.Type = schema.Number
					case "boolean":
						paramInfo.Type = schema.Boolean
					case "array":
						paramInfo.Type = schema.Array
					case "object":
						paramInfo.Type = schema.Object
					default:
						paramInfo.Type = schema.String
					}
				}

				params[paramName] = paramInfo
			}
		}

		if required, ok := t.toolDef.InputSchema["required"].([]interface{}); ok {
			for _, reqParam := range required {
				if paramName, ok := reqParam.(string); ok {
					if paramInfo, exists := params[paramName]; exists {
						paramInfo.Required = true
					}
				}
			}
		}

		if len(params) > 0 {
			toolInfo.ParamsOneOf = schema.NewParamsOneOfByParams(params)
		}
	}

	logger.Info(fmt.Sprintf("MCP工具信息: name=%s, desc=%s", toolInfo.Name, toolInfo.Desc))

	return toolInfo, nil
}

func (t *MCPTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	logger.Info(fmt.Sprintf("MCP工具开始执行: server=%s, tool=%s, args=%s", t.serverID, t.toolDef.Name, argumentsInJSON))

	var arguments map[string]interface{}
	if err := json.Unmarshal([]byte(argumentsInJSON), &arguments); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	result, err := t.mcpSvc.CallTool(ctx, t.serverID, t.toolDef.Name, arguments)
	if err != nil {
		logger.Error(fmt.Sprintf("MCP工具调用失败: %v", err))
		return "", fmt.Errorf("MCP工具调用失败: %w", err)
	}

	resultText := mcp.ExtractTextContent(result)
	logger.Info(fmt.Sprintf("MCP工具执行成功: tool=%s, result_length=%d", t.toolDef.Name, len(resultText)))

	return resultText, nil
}

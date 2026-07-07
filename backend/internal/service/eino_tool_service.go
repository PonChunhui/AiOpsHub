package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aiops/AiOpsHub/backend/internal/agent/eino_tools"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/aiops/AiOpsHub/backend/pkg/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type ToolFactory struct {
	toolModels map[string]*model.Tool
	hostRepo   *repository.HostRepository
}

func NewToolFactory(hostRepo *repository.HostRepository) *ToolFactory {
	return &ToolFactory{
		toolModels: make(map[string]*model.Tool),
		hostRepo:   hostRepo,
	}
}

func (f *ToolFactory) RegisterTool(toolModel *model.Tool) {
	f.toolModels[toolModel.Name] = toolModel
	logger.Info(fmt.Sprintf("注册工具模型: %s", toolModel.Name))
}

func (f *ToolFactory) CreateTool(toolName string, configOverride map[string]interface{}) (tool.InvokableTool, error) {
	toolModel, ok := f.toolModels[toolName]
	if !ok {
		return nil, fmt.Errorf("工具未注册: %s", toolName)
	}

	switch toolName {
	case "ssh_exec":
		return eino_tools.NewSSHTool(toolModel, configOverride, f.hostRepo), nil
	case "prometheus_query":
		return eino_tools.NewPrometheusTool(toolModel, configOverride), nil
	case "kubernetes_query":
		return eino_tools.NewKubernetesTool(toolModel, configOverride), nil
	case "log_query":
		return eino_tools.NewLogQueryTool(toolModel, configOverride), nil
	default:
		return nil, fmt.Errorf("未知工具类型: %s", toolName)
	}
}

func (f *ToolFactory) CreateToolsForAgent(tools []model.Tool, bindings []model.AgentTool) ([]tool.BaseTool, error) {
	var baseTools []tool.BaseTool

	for i, toolModel := range tools {
		binding := bindings[i]

		// 只创建启用的工具
		if !binding.Enabled {
			logger.Info(fmt.Sprintf("工具 %s 已禁用，跳过", toolModel.Name))
			continue
		}

		// 解析配置覆盖
		configOverride := make(map[string]interface{})
		if binding.ConfigOverride != "" {
			if err := json.Unmarshal([]byte(binding.ConfigOverride), &configOverride); err != nil {
				logger.Error(fmt.Sprintf("解析工具配置失败: %s - %v", toolModel.Name, err))
				continue
			}
		}

		// 创建工具实例
		einoTool, err := f.CreateTool(toolModel.Name, configOverride)
		if err != nil {
			logger.Error(fmt.Sprintf("创建工具失败: %s - %v", toolModel.Name, err))
			continue
		}

		// InvokableTool继承BaseTool，可以直接赋值
		baseTools = append(baseTools, einoTool)
		logger.Info(fmt.Sprintf("创建Agent工具: %s (配置覆盖: %v)", toolModel.Name, configOverride))
	}

	return baseTools, nil
}

// EinoToolService Eino工具服务
type EinoToolService struct {
	factory *ToolFactory
}

func NewEinoToolService() *EinoToolService {
	hostRepo := repository.NewHostRepository()
	return &EinoToolService{
		factory: NewToolFactory(hostRepo),
	}
}

func (s *EinoToolService) LoadAgentTools(ctx context.Context, tools []model.Tool, bindings []model.AgentTool) ([]tool.BaseTool, error) {
	// 注册所有工具模型
	for _, toolModel := range tools {
		s.factory.RegisterTool(&toolModel)
	}

	// 创建Eino工具实例列表
	baseTools, err := s.factory.CreateToolsForAgent(tools, bindings)
	if err != nil {
		return nil, fmt.Errorf("创建工具列表失败: %w", err)
	}

	logger.Info(fmt.Sprintf("成功加载 %d 个Eino工具", len(baseTools)))

	return baseTools, nil
}

// CreateToolsNode 创建Eino工具执行节点
func (s *EinoToolService) CreateToolsNode(ctx context.Context, baseTools []tool.BaseTool) (*compose.AgenticToolsNode, error) {
	if len(baseTools) == 0 {
		return nil, fmt.Errorf("工具列表为空")
	}

	// 创建工具节点配置
	config := &compose.ToolsNodeConfig{
		Tools: baseTools,
	}

	// 创建AgenticToolsNode
	toolsNode, err := compose.NewAgenticToolsNode(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建工具节点失败: %w", err)
	}

	logger.Info(fmt.Sprintf("成功创建工具节点，包含 %d 个工具", len(baseTools)))

	return toolsNode, nil
}

// ExecuteToolCall 执行单个工具调用
func (s *EinoToolService) ExecuteToolCall(ctx context.Context, einoTool tool.InvokableTool, arguments string) (string, error) {
	// 获取工具信息
	toolInfo, err := einoTool.Info(ctx)
	if err != nil {
		return "", fmt.Errorf("获取工具信息失败: %w", err)
	}

	logger.Info(fmt.Sprintf("开始执行工具: %s", toolInfo.Name))
	logger.Debug(fmt.Sprintf("工具参数: %s", arguments))

	// 执行工具
	result, err := einoTool.InvokableRun(ctx, arguments)
	if err != nil {
		return "", fmt.Errorf("工具执行失败: %w", err)
	}

	logger.Info(fmt.Sprintf("工具 %s 执行成功，返回结果长度: %d", toolInfo.Name, len(result)))

	return result, nil
}

// ExecuteToolsBatch 批量执行工具调用
func (s *EinoToolService) ExecuteToolsBatch(ctx context.Context, toolsNode *compose.AgenticToolsNode, toolCalls []*schema.AgenticMessage) ([]*schema.AgenticMessage, error) {
	logger.Info(fmt.Sprintf("批量执行 %d 个工具调用", len(toolCalls)))

	// 使用工具节点执行
	results, err := toolsNode.Invoke(ctx, toolCalls[0]) // 注意：AgenticToolsNode.Invoke接受单个消息
	if err != nil {
		return nil, fmt.Errorf("工具批量执行失败: %w", err)
	}

	logger.Info(fmt.Sprintf("批量执行成功，返回 %d 个结果", len(results)))

	return results, nil
}

type MCPServiceInterface interface {
	GetTools(ctx context.Context, serverID string) ([]mcp.Tool, error)
	CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]interface{}) (*mcp.ToolCallResult, error)
}

func (s *EinoToolService) LoadMCPToolsByServerIDs(ctx context.Context, mcpSvc MCPServiceInterface, serverIDs []string) ([]tool.BaseTool, error) {
	var allTools []tool.BaseTool

	for _, serverID := range serverIDs {
		tools, err := mcpSvc.GetTools(ctx, serverID)
		if err != nil {
			logger.Error(fmt.Sprintf("获取MCP Server %s的工具失败: %v", serverID, err))
			continue
		}

		for _, toolDef := range tools {
			mcpTool := eino_tools.NewMCPTool(serverID, toolDef, mcpSvc)
			allTools = append(allTools, mcpTool)
			logger.Info(fmt.Sprintf("加载MCP工具: %s (Server: %s)", toolDef.Name, serverID))
		}
	}

	logger.Info(fmt.Sprintf("成功加载 %d 个MCP工具", len(allTools)))
	return allTools, nil
}

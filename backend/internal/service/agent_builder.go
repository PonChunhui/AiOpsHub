package service

import (
	"context"
	"fmt"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/pkg/llm"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

type AgentBuilder struct {
	agentSvc    *AgentService
	toolSvc     *ToolService
	einoToolSvc *EinoToolService
	llm         *llm.EinoLLM
}

func NewAgentBuilder(agentSvc *AgentService, toolSvc *ToolService, einoToolSvc *EinoToolService, llmInstance *llm.EinoLLM) *AgentBuilder {
	return &AgentBuilder{
		agentSvc:    agentSvc,
		toolSvc:     toolSvc,
		einoToolSvc: einoToolSvc,
		llm:         llmInstance,
	}
}

func (b *AgentBuilder) BuildAgent(ctx context.Context, agentID string) (adk.Agent, error) {
	agentModel, err := b.agentSvc.GetByID(agentID)
	if err != nil {
		return nil, fmt.Errorf("获取Agent失败: %w", err)
	}

	tools, bindings, err := b.toolSvc.GetAgentTools(agentID)
	if err != nil {
		logger.Error(fmt.Sprintf("获取Agent工具失败: %v", err))
	}

	var einoTools []tool.BaseTool
	if len(tools) > 0 {
		einoTools, err = b.einoToolSvc.LoadAgentTools(ctx, tools, bindings)
		if err != nil {
			logger.Error(fmt.Sprintf("加载Eino工具失败: %v", err))
		}
		logger.Info(fmt.Sprintf("Agent %s 加载了 %d 个工具", agentModel.Name, len(einoTools)))
	}

	agentConfig := &adk.ChatModelAgentConfig{
		Model: b.llm.GetChatModel(),
		Name:  agentModel.Name,
	}

	if agentModel.SystemPrompt != "" {
		agentConfig.Instruction = agentModel.SystemPrompt
	}

	if len(einoTools) > 0 {
		agentConfig.ToolsConfig = adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: einoTools,
			},
		}
		logger.Info(fmt.Sprintf("Agent %s 配置了 ToolsConfig，包含 %d 个工具", agentModel.Name, len(einoTools)))
	}

	agent, err := adk.NewChatModelAgent(ctx, agentConfig)
	if err != nil {
		return nil, fmt.Errorf("创建Eino Agent失败: %w", err)
	}

	logger.Info(fmt.Sprintf("成功构建Agent: %s (%s)", agentModel.Name, agentID))
	return agent, nil
}

func (b *AgentBuilder) BuildDefaultAgent(ctx context.Context) (adk.Agent, error) {
	agentConfig := &adk.ChatModelAgentConfig{
		Model:       b.llm.GetChatModel(),
		Name:        "default",
		Instruction: "你是一个智能运维助手，帮助用户解决运维问题。",
	}

	agent, err := adk.NewChatModelAgent(ctx, agentConfig)
	if err != nil {
		return nil, fmt.Errorf("创建默认Agent失败: %w", err)
	}

	logger.Info("成功构建默认Agent")
	return agent, nil
}

func (b *AgentBuilder) BuildPresetAgent(ctx context.Context, presetAgentID string) (adk.Agent, error) {
	presetAgents := GetPresetAgents()
	for _, pa := range presetAgents {
		if pa.ID == presetAgentID {
			return b.BuildAgentFromModel(ctx, &pa)
		}
	}
	return nil, fmt.Errorf("未找到预设Agent: %s", presetAgentID)
}

func (b *AgentBuilder) BuildAgentFromModel(ctx context.Context, agentModel *model.Agent) (adk.Agent, error) {
	tools, bindings, err := b.toolSvc.GetAgentTools(agentModel.ID)
	if err != nil {
		logger.Error(fmt.Sprintf("获取Agent工具失败: %v", err))
	}

	var einoTools []tool.BaseTool
	if len(tools) > 0 {
		einoTools, err = b.einoToolSvc.LoadAgentTools(ctx, tools, bindings)
		if err != nil {
			logger.Error(fmt.Sprintf("加载Eino工具失败: %v", err))
		}
		logger.Info(fmt.Sprintf("Agent %s 加载了 %d 个工具", agentModel.Name, len(einoTools)))
	}

	agentConfig := &adk.ChatModelAgentConfig{
		Model:       b.llm.GetChatModel(),
		Name:        agentModel.Name,
		Instruction: agentModel.SystemPrompt,
	}

	if len(einoTools) > 0 {
		agentConfig.ToolsConfig = adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: einoTools,
			},
		}
		logger.Info(fmt.Sprintf("Agent %s 配置了 ToolsConfig，包含 %d 个工具", agentModel.Name, len(einoTools)))
	}

	agent, err := adk.NewChatModelAgent(ctx, agentConfig)
	if err != nil {
		return nil, fmt.Errorf("创建Eino Agent失败: %w", err)
	}

	logger.Info(fmt.Sprintf("成功构建Agent: %s (%s)", agentModel.Name, agentModel.ID))
	return agent, nil
}

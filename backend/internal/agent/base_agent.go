package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aiops/AiOpsHub/backend/pkg/llm"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type BaseAgent struct {
	ID          string
	Name        string
	Type        string
	Description string
	Config      AgentConfig
	llm         *llm.EinoLLM
}

type AgentConfig struct {
	Model              string  `json:"model"`
	Temperature        float64 `json:"temperature"`
	MaxTokens          int     `json:"max_tokens"`
	Provider           string  `json:"provider"`
	APIKey             string  `json:"api_key"`
	BaseURL            string  `json:"base_url"`
	SystemPrompt       string  `json:"system_prompt"`
	UserPromptTemplate string  `json:"user_prompt_template"`
}

type AgentInput struct {
	TaskType     string                 `json:"task_type"`
	Input        map[string]interface{} `json:"input"`
	Conversation string                 `json:"conversation"`
}

type AgentOutput struct {
	Result     map[string]interface{} `json:"result"`
	Status     string                 `json:"status"`
	Message    string                 `json:"message"`
	TokensUsed int                    `json:"tokens_used"`
}

func NewBaseAgent(id, name, agentType, description string, config AgentConfig) (*BaseAgent, error) {
	agent := &BaseAgent{
		ID:          id,
		Name:        name,
		Type:        agentType,
		Description: description,
		Config:      config,
	}

	llm, err := agent.createLLM(config)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create LLM: %v", err))
		return nil, fmt.Errorf("failed to create LLM: %w", err)
	}

	agent.llm = llm
	logger.Info(fmt.Sprintf("Created agent: %s (%s) with provider: %s", name, id, config.Provider))

	return agent, nil
}

func (a *BaseAgent) createLLM(config AgentConfig) (*llm.EinoLLM, error) {
	llmConfig := llm.EinoLLMConfig{
		Model:       config.Model,
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
		Provider:    config.Provider,
		APIKey:      config.APIKey,
		BaseURL:     config.BaseURL,
	}

	einoLLM, err := llm.NewEinoLLM(llmConfig)
	if err != nil {
		return nil, err
	}

	return einoLLM, nil
}

func (a *BaseAgent) Execute(ctx context.Context, input AgentInput) (AgentOutput, error) {
	logger.Info(fmt.Sprintf("Agent %s executing task: %s", a.Name, input.TaskType))

	prompt := a.buildPrompt(input)

	resp, err := a.llm.Generate(ctx, prompt)
	if err != nil {
		logger.Error(fmt.Sprintf("Agent execution failed: %v", err))
		return AgentOutput{
			Status:  "error",
			Message: err.Error(),
		}, err
	}

	output := AgentOutput{
		Result: map[string]interface{}{
			"response": resp,
			"task":     input.TaskType,
		},
		Status:     "completed",
		Message:    "Agent executed successfully",
		TokensUsed: 0,
	}

	logger.Info(fmt.Sprintf("Agent %s completed task: %s", a.Name, input.TaskType))

	return output, nil
}

func (a *BaseAgent) buildPrompt(input AgentInput) string {
	if a.Config.UserPromptTemplate != "" {
		prompt := a.Config.UserPromptTemplate
		if input.Input != nil {
			inputJSON, _ := json.Marshal(input.Input)
			prompt = strings.ReplaceAll(prompt, "{{input}}", string(inputJSON))
		}
		if a.Config.SystemPrompt != "" {
			return a.Config.SystemPrompt + "\n\n" + prompt
		}
		return prompt
	}

	if a.Config.SystemPrompt != "" {
		inputJSON, _ := json.Marshal(input.Input)
		return a.Config.SystemPrompt + "\n\n输入数据: " + string(inputJSON)
	}

	switch input.TaskType {
	case "alert_analysis":
		return fmt.Sprintf("分析以下告警并提供诊断建议:\n告警内容: %v\n\n请提供:\n1. 告警严重性评估\n2. 可能的根本原因\n3. 推荐的处理步骤", input.Input)
	case "incident_diagnosis":
		return fmt.Sprintf("诊断以下故障事件:\n事件描述: %v\n\n请提供:\n1. 根本原因分析\n2. 影响范围评估\n3. 修复建议", input.Input)
	case "auto_remediation":
		return fmt.Sprintf("针对以下问题，建议自动化修复方案:\n问题描述: %v\n\n请提供:\n1. 可执行的修复步骤\n2. 风险评估\n3. 验证方法", input.Input)
	default:
		return fmt.Sprintf("处理以下任务:\n任务: %s\n输入: %v\n\n请提供详细的处理建议和步骤。", input.TaskType, input.Input)
	}
}

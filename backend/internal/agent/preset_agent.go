package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/service"
	"github.com/aiops/AiOpsHub/backend/pkg/llm"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type PresetAgentExecutor struct {
	agent  *model.Agent
	llm    *llm.EinoLLM
	mcpSvc *service.MCPService
	ragSvc *service.RAGService
}

func NewPresetAgentExecutor(agent *model.Agent, mcpSvc *service.MCPService, ragSvc *service.RAGService) (*PresetAgentExecutor, error) {
	llmConfig := llm.EinoLLMConfig{
		Model:       agent.Model,
		Temperature: agent.Temperature,
	}

	llm, err := llm.NewEinoLLM(llmConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 LLM 失败: %w", err)
	}

	return &PresetAgentExecutor{
		agent:  agent,
		llm:    llm,
		mcpSvc: mcpSvc,
		ragSvc: ragSvc,
	}, nil
}

func (e *PresetAgentExecutor) Execute(ctx context.Context, userInput string, additionalContext map[string]interface{}) (string, error) {
	logger.Debug(fmt.Sprintf("执行预设 Agent: %s (%s)", e.agent.Name, e.agent.Avatar))

	// 构建完整 prompt
	fullPrompt := e.buildPrompt(userInput, additionalContext)

	// 调用 LLM
	response, err := e.llm.Generate(ctx, fullPrompt)
	if err != nil {
		return "", fmt.Errorf("LLM 生成失败: %w", err)
	}

	// 检查是否需要工具调用
	if e.mcpSvc != nil && e.containsToolCall(response) {
		logger.Debug(fmt.Sprintf("Agent %s 需要调用工具", e.agent.Name))

		// 执行工具调用循环
		finalResponse, err := e.executeToolLoop(ctx, response, userInput)
		if err != nil {
			return "", err
		}

		return finalResponse, nil
	}

	return response, nil
}

func (e *PresetAgentExecutor) buildPrompt(userInput string, additionalContext map[string]interface{}) string {
	promptBuilder := fmt.Sprintf(`
你是 %s（%s），角色：%s

系统提示词：
%s

`, e.agent.Avatar, e.agent.Name, e.agent.Role, e.agent.SystemPrompt)

	// 添加 MCP 工具信息
	if e.mcpSvc != nil {
		toolsInfo := e.getAvailableToolsInfo()
		if toolsInfo != "" {
			promptBuilder += fmt.Sprintf("\n可用工具：\n%s\n", toolsInfo)
			promptBuilder += "\n如果需要使用工具，请按以下格式调用：\n```tool_call\n{\"tool\": \"工具名称\", \"server\": \"服务器名称\", \"arguments\": {参数}}\n```\n"
		}
	}

	// 添加额外上下文
	if additionalContext != nil {
		if contextStr, ok := additionalContext["context"].(string); ok && contextStr != "" {
			promptBuilder += fmt.Sprintf("\n上下文信息：\n%s\n", contextStr)
		}
	}

	promptBuilder += fmt.Sprintf("\n用户输入：\n%s\n", userInput)

	return promptBuilder
}

func (e *PresetAgentExecutor) getAvailableToolsInfo() string {
	if e.mcpSvc == nil {
		return ""
	}

	// 获取所有可用工具
	allTools, err := e.mcpSvc.GetAllActiveTools(context.Background())
	if err != nil || len(allTools) == 0 {
		return ""
	}

	toolsInfo := ""
	for serverName, tools := range allTools {
		toolsInfo += fmt.Sprintf("\n### %s:\n", serverName)
		for _, tool := range tools {
			toolsInfo += fmt.Sprintf("- %s: %s\n", tool.Name, tool.Description)
		}
	}

	return toolsInfo
}

func (e *PresetAgentExecutor) containsToolCall(response string) bool {
	return strings.Contains(response, "```tool_call")
}

func (e *PresetAgentExecutor) executeToolLoop(ctx context.Context, initialResponse string, originalInput string) (string, error) {
	// 解析工具调用
	toolCalls := parseToolCalls(initialResponse)

	var toolResults []string
	for _, tc := range toolCalls {
		toolName, ok := tc["tool"].(string)
		if !ok {
			continue
		}

		serverName, ok := tc["server"].(string)
		if !ok {
			serverName = "Unknown"
		}

		arguments, ok := tc["arguments"].(map[string]interface{})
		if !ok {
			arguments = make(map[string]interface{})
		}

		logger.Debug(fmt.Sprintf("Agent %s 执行工具: %s.%s", e.agent.Name, serverName, toolName))

		// 调用 MCP 工具
		result, err := e.callMCPTool(ctx, serverName, toolName, arguments)
		if err != nil {
			toolResults = append(toolResults, fmt.Sprintf("工具 %s 执行失败: %v", toolName, err))
		} else {
			// 截取过大的结果
			if len(result) > 2000 {
				result = result[:2000] + "\n...(数据过大，已截取)"
			}
			toolResults = append(toolResults, fmt.Sprintf("工具 %s 执行结果:\n%s", toolName, result))
		}
	}

	// 如果有工具执行结果，让 LLM 处理
	if len(toolResults) > 0 {
		toolContext := strings.Join(toolResults, "\n\n")

		processPrompt := fmt.Sprintf(`
作为 %s，你刚才调用了工具来帮助用户。

工具执行结果：
%s

请基于这些结果，用专业且易懂的语言回答用户的问题："%s"

回答时请：
1. 理解并解释数据的含义
2. 提取关键信息和发现的问题
3. 给出专业的处理建议或下一步操作
4. 保持角色的专业性和语气

开始回答：
`, e.agent.Name, toolContext, originalInput)

		finalResponse, err := e.llm.Generate(ctx, processPrompt)
		if err != nil {
			return fmt.Sprintf("工具已执行成功，但处理结果时出错:\n%s", toolContext), nil
		}

		return finalResponse, nil
	}

	return initialResponse, nil
}

func (e *PresetAgentExecutor) callMCPTool(ctx context.Context, serverName, toolName string, arguments map[string]interface{}) (string, error) {
	if e.mcpSvc == nil {
		return "", fmt.Errorf("MCP 服务未启用")
	}

	serverID, err := e.mcpSvc.FindServerByToolName(ctx, toolName)
	if err != nil {
		return "", fmt.Errorf("找不到工具 %s", toolName)
	}

	result, err := e.mcpSvc.CallTool(ctx, serverID, toolName, arguments)
	if err != nil {
		return "", err
	}

	return extractTextContent(result), nil
}

func parseToolCalls(text string) []map[string]interface{} {
	// 复用 ChatService 的解析逻辑
	var calls []map[string]interface{}
	start := 0
	for {
		idx := indexOf(text[start:], "```tool_call\n")
		if idx == -1 {
			break
		}
		start += idx + len("```tool_call\n")

		endIdx := indexOf(text[start:], "```")
		if endIdx == -1 {
			break
		}

		jsonStr := text[start : start+endIdx]
		start += endIdx + 3

		var tc map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &tc); err != nil {
			continue
		}

		calls = append(calls, tc)
	}
	return calls
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func extractTextContent(result interface{}) string {
	// 提取工具返回的文本内容
	// 简化实现，实际需要根据 MCP ToolCallResult 结构提取
	return fmt.Sprintf("%v", result)
}

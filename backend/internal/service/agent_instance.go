package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/pkg/llm"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/spf13/viper"
)

type AgentInstance struct {
	AgentModel     *model.Agent
	AvailableTools []model.Tool
	toolRegistry   *ToolRegistry
	llm            *llm.EinoLLM
	maxToolCalls   int
	callHistory    []ToolCallRecord
	agentID        string
}

type ToolCallRecord struct {
	ToolName  string
	Arguments map[string]interface{}
	Result    string
	Success   bool
	Duration  int
	Timestamp time.Time
}

type ToolCall struct {
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Execute 执行Agent任务（非流式版本）
func (a *AgentInstance) Execute(ctx context.Context, userMessage string, history []model.ChatMessage) (string, []ToolCallRecord, error) {
	prompt := a.buildExecutionPrompt(userMessage, history)

	response, err := a.llm.Generate(ctx, prompt)
	if err != nil {
		return "", nil, fmt.Errorf("LLM生成失败: %w", err)
	}

	toolCalls := a.parseToolCalls(response)

	if len(toolCalls) == 0 {
		return response, nil, nil
	}

	callCount := 0
	currentResponse := response

	for len(toolCalls) > 0 && callCount < a.maxToolCalls {
		logger.Info(fmt.Sprintf("Agent %s 第%d轮工具调用，共%d个工具", a.AgentModel.Name, callCount+1, len(toolCalls)))

		toolResults := a.executeTools(ctx, toolCalls)
		a.callHistory = append(a.callHistory, toolResults...)

		processPrompt := a.buildToolResultPrompt(userMessage, toolResults, currentResponse)

		currentResponse, err = a.llm.Generate(ctx, processPrompt)
		if err != nil {
			logger.Error(fmt.Sprintf("LLM处理工具结果失败: %v", err))
			return currentResponse, a.callHistory, err
		}

		toolCalls = a.parseToolCalls(currentResponse)
		callCount++
	}

	if callCount >= a.maxToolCalls && len(toolCalls) > 0 {
		currentResponse += "\n\n⚠️ 已达到最大工具调用次数限制。"
	}

	return currentResponse, a.callHistory, nil
}

// ExecuteStream 执行Agent任务（全流程流式版本）
// 返回AgentEvent流，包含thinking、content、tool_call、tool_result等事件
func (a *AgentInstance) ExecuteStream(ctx context.Context, userMessage string, history []model.ChatMessage) (<-chan *model.AgentEvent, []ToolCallRecord, error) {
	eventChan := make(chan *model.AgentEvent, 100)
	agentName := a.AgentModel.Name

	go func() {
		defer close(eventChan)

		// 1. 初始LLM调用（流式）- 发送thinking和content事件
		prompt := a.buildExecutionPrompt(userMessage, history)
		logger.Info(fmt.Sprintf("Agent %s 开始流式执行，prompt长度: %d", agentName, len(prompt)))

		streamChan, err := a.llm.StreamGenerateWithReasoning(ctx, prompt)
		if err != nil {
			logger.Error(fmt.Sprintf("LLM流式生成失败: %v", err))
			eventChan <- model.NewErrorEvent(agentName, fmt.Sprintf("LLM生成失败: %v", err), 500)
			return
		}

		// 收集完整响应（用于解析工具调用）
		var fullResponse strings.Builder
		var fullReasoning strings.Builder

		for result := range streamChan {
			// 发送thinking事件
			if result.ReasoningContent != "" {
				fullReasoning.WriteString(result.ReasoningContent)
				eventChan <- model.NewThinkingEvent(agentName, result.ReasoningContent)
			}

			// 发送content事件
			if result.Content != "" {
				fullResponse.WriteString(result.Content)
				eventChan <- model.NewContentChunkEvent(agentName, result.Content)
			}
		}

		response := fullResponse.String()
		logger.Info(fmt.Sprintf("Agent %s 初始响应完成，长度: %d, reasoning长度: %d",
			agentName, len(response), fullReasoning.Len()))

		// 2. 解析工具调用
		toolCalls := a.parseToolCalls(response)

		if len(toolCalls) == 0 {
			logger.Info(fmt.Sprintf("Agent %s 无工具调用，执行完成", agentName))
			eventChan <- model.NewDoneEvent(agentName, nil)
			return
		}

		// 3. 工具调用循环（每轮都流式输出）
		callCount := 0

		for len(toolCalls) > 0 && callCount < a.maxToolCalls {
			logger.Info(fmt.Sprintf("Agent %s 第%d轮工具调用，共%d个工具", agentName, callCount+1, len(toolCalls)))

			// 执行所有工具并发送事件
			toolResults := make([]ToolCallRecord, 0, len(toolCalls))

			for i, call := range toolCalls {
				toolID := fmt.Sprintf("tool_%d_%d", time.Now().UnixNano(), callCount*10+i)

				// 发送tool_call事件（工具调用开始）
				argsJSON := toJSON(call.Arguments)
				eventChan <- model.NewToolCallEvent(agentName, toolID, call.Tool, argsJSON)

				logger.Info(fmt.Sprintf("开始执行工具: %s, 参数: %s", call.Tool, argsJSON))

				// 执行工具（阻塞等待结果）
				startTime := time.Now()
				result, err := a.toolRegistry.ExecuteTool(ctx, a.agentID, call.Tool, call.Arguments)
				duration := int(time.Since(startTime).Milliseconds())

				// 记录工具调用
				record := ToolCallRecord{
					ToolName:  call.Tool,
					Arguments: call.Arguments,
					Result:    result,
					Success:   err == nil,
					Duration:  duration,
					Timestamp: startTime,
				}
				toolResults = append(toolResults, record)
				a.callHistory = append(a.callHistory, record)

				// 发送tool_result事件（工具调用完成）
				if err != nil {
					record.Result = fmt.Sprintf("错误: %v", err)
					eventChan <- model.NewToolResultEvent(agentName, toolID, call.Tool, record.Result, false)
					logger.Error(fmt.Sprintf("工具 %s 执行失败: %v", call.Tool, err))
				} else {
					if len(result) > 2000 {
						result = result[:2000] + "\n...(数据过大，已截取)"
					}
					eventChan <- model.NewToolResultEvent(agentName, toolID, call.Tool, result, true)
					logger.Info(fmt.Sprintf("工具 %s 执行成功，耗时%dms", call.Tool, duration))
				}
			}

			// 4. LLM处理工具结果（流式）- 发送thinking和content事件
			processPrompt := a.buildToolResultPrompt(userMessage, toolResults, response)
			logger.Info(fmt.Sprintf("Agent %s 开始处理工具结果，prompt长度: %d", agentName, len(processPrompt)))

			streamChan2, err := a.llm.StreamGenerateWithReasoning(ctx, processPrompt)
			if err != nil {
				logger.Error(fmt.Sprintf("LLM处理工具结果失败: %v", err))
				eventChan <- model.NewErrorEvent(agentName, fmt.Sprintf("处理工具结果失败: %v", err), 500)
				return
			}

			// 收集处理后的响应
			var processedResponse strings.Builder

			for result := range streamChan2 {
				// 发送thinking事件（第二轮thinking）
				if result.ReasoningContent != "" {
					eventChan <- model.NewThinkingEvent(agentName, result.ReasoningContent)
				}

				// 发送content事件（第二轮content）
				if result.Content != "" {
					processedResponse.WriteString(result.Content)
					eventChan <- model.NewContentChunkEvent(agentName, result.Content)
				}
			}

			response = processedResponse.String()
			logger.Info(fmt.Sprintf("Agent %s 工具结果处理完成，新响应长度: %d", agentName, len(response)))

			// 5. 解析下一轮工具调用
			toolCalls = a.parseToolCalls(response)
			callCount++
		}

		// 6. 达到最大调用次数限制
		if callCount >= a.maxToolCalls && len(toolCalls) > 0 {
			finalMessage := "\n\n⚠️ 已达到最大工具调用次数限制。"
			eventChan <- model.NewContentChunkEvent(agentName, finalMessage)
		}

		// 7. 发送完成事件
		logger.Info(fmt.Sprintf("Agent %s 执行完成，共%d轮工具调用", agentName, callCount))
		eventChan <- model.NewDoneEvent(agentName, nil)
	}()

	return eventChan, a.callHistory, nil
}

func (a *AgentInstance) buildExecutionPrompt(userMsg string, history []model.ChatMessage) string {
	var psb strings.Builder

	if a.AgentModel.SystemPrompt != "" {
		psb.WriteString(a.AgentModel.SystemPrompt)
		psb.WriteString("\n\n")
	}

	psb.WriteString("## 你可以使用的工具：\n\n")
	for i, tool := range a.AvailableTools {
		psb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, tool.Name))
		psb.WriteString(fmt.Sprintf("   - 描述: %s\n", tool.Description))
		psb.WriteString(fmt.Sprintf("   - 类别: %s\n", tool.Category))
		psb.WriteString(fmt.Sprintf("   - 风险等级: %s\n\n", tool.RiskLevel))
	}

	psb.WriteString(`## 工具调用方式：
当你需要使用工具时，请在回复中嵌入以下格式：

` + "```tool_call\n" + `{
  "tool": "工具名称",
  "arguments": {
    "参数名": "参数值"
  }
}
` + "```\n\n")

	if len(history) > 0 {
		psb.WriteString("## 历史对话：\n\n")
		for _, msg := range history {
			if msg.Role == "user" {
				psb.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
			} else {
				psb.WriteString(fmt.Sprintf("助手: %s\n", msg.Content))
			}
		}
		psb.WriteString("\n")
	}

	psb.WriteString(fmt.Sprintf("## 用户问题：\n%s\n\n请回答：", userMsg))

	return psb.String()
}

func (a *AgentInstance) buildToolResultPrompt(userMsg string, results []ToolCallRecord, previousResponse string) string {
	var psb strings.Builder

	psb.WriteString(fmt.Sprintf("你之前回答：%s\n\n", previousResponse))
	psb.WriteString("你调用的工具已经执行，结果如下：\n\n")

	for i, result := range results {
		psb.WriteString(fmt.Sprintf("### 工具调用 %d: %s\n", i+1, result.ToolName))
		psb.WriteString(fmt.Sprintf("执行时间: %d ms\n", result.Duration))
		if result.Success {
			psb.WriteString(fmt.Sprintf("结果: %s\n\n", result.Result))
		} else {
			psb.WriteString(fmt.Sprintf("错误: %s\n\n", result.Result))
		}
	}

	psb.WriteString(fmt.Sprintf("原始用户问题：%s\n\n", userMsg))
	psb.WriteString("请基于工具返回的实际数据，继续回答用户问题。如果还需要更多信息，可以继续调用工具。")

	return psb.String()
}

func (a *AgentInstance) executeTools(ctx context.Context, calls []ToolCall) []ToolCallRecord {
	var records []ToolCallRecord

	for _, call := range calls {
		startTime := time.Now()

		record := ToolCallRecord{
			ToolName:  call.Tool,
			Arguments: call.Arguments,
			Timestamp: startTime,
		}

		result, err := a.toolRegistry.ExecuteTool(ctx, a.agentID, call.Tool, call.Arguments)

		record.Duration = int(time.Since(startTime).Milliseconds())

		if err != nil {
			record.Result = fmt.Sprintf("错误: %v", err)
			record.Success = false
			logger.Error(fmt.Sprintf("工具 %s 执行失败: %v", call.Tool, err))
		} else {
			if len(result) > 2000 {
				result = result[:2000] + "\n...(数据过大，已截取)"
			}
			record.Result = result
			record.Success = true
			logger.Info(fmt.Sprintf("工具 %s 执行成功，耗时%dms", call.Tool, record.Duration))
		}

		records = append(records, record)
	}

	return records
}

func (a *AgentInstance) parseToolCalls(text string) []ToolCall {
	var calls []ToolCall

	start := 0
	for {
		idx := strings.Index(text[start:], "```tool_call\n")
		if idx == -1 {
			break
		}
		start += idx + len("```tool_call\n")

		endIdx := strings.Index(text[start:], "```")
		if endIdx == -1 {
			break
		}

		jsonStr := text[start : start+endIdx]
		start += endIdx + 3

		var call ToolCall
		if err := json.Unmarshal([]byte(jsonStr), &call); err != nil {
			logger.Error(fmt.Sprintf("解析工具调用失败: %v", err))
			continue
		}

		calls = append(calls, call)
	}

	return calls
}

// ApplyRuntimeConfig 应用运行时配置到Agent实例
// 根据配置动态创建LLM实例并过滤可用工具
func (a *AgentInstance) ApplyRuntimeConfig(ctx context.Context, config RuntimeConfig, sessionModel string) error {
	logger.Info(fmt.Sprintf("ApplyRuntimeConfig: model=%s, temp=%.2f, maxTokens=%d, tools=%v",
		config.Model, config.Temperature, config.MaxTokens, config.EnabledTools))

	// 1. 根据配置创建新的LLM实例
	if config.Model != "" {
		newLLM, err := a.createDynamicLLM(config, sessionModel)
		if err != nil {
			logger.Error(fmt.Sprintf("创建动态LLM失败: %v", err))
			return err
		}
		a.llm = newLLM
		logger.Info(fmt.Sprintf("成功应用动态LLM配置: model=%s", config.Model))
	}

	// 2. 根据配置过滤可用工具
	if len(config.EnabledTools) > 0 {
		a.filterAvailableTools(config.EnabledTools)
		logger.Info(fmt.Sprintf("成功过滤可用工具，从%d个减少到%d个",
			len(a.AgentModel.MCPServerIDs), len(a.AvailableTools)))
	}

	return nil
}

// createDynamicLLM 创建动态配置的LLM实例
func (a *AgentInstance) createDynamicLLM(config RuntimeConfig, defaultModel string) (*llm.EinoLLM, error) {
	model := config.Model
	if model == "" {
		model = defaultModel
		if model == "" {
			model = viper.GetString("llm.model")
			if model == "" {
				model = "qwen-turbo"
			}
		}
	}

	temperature := config.Temperature
	if temperature == 0 {
		temperature = viper.GetFloat64("llm.temperature")
		if temperature == 0 {
			temperature = 0.7
		}
	}

	maxTokens := config.MaxTokens
	if maxTokens == 0 {
		maxTokens = viper.GetInt("llm.max_tokens")
		if maxTokens == 0 {
			maxTokens = 4096
		}
	}

	// 从全局配置中获取provider、APIKey和BaseURL
	provider := viper.GetString("llm.provider")
	if provider == "" {
		provider = "aliyun_bailian"
	}

	apiKey := viper.GetString("llm.api_key")
	baseURL := viper.GetString("llm.base_url")

	llmConfig := llm.EinoLLMConfig{
		Model:       model,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Provider:    provider,
		APIKey:      apiKey,
		BaseURL:     baseURL,
	}

	newLLM, err := llm.NewEinoLLM(llmConfig)
	if err != nil {
		return nil, fmt.Errorf("创建LLM实例失败: %w", err)
	}

	logger.Info(fmt.Sprintf("创建动态LLM: model=%s, temp=%.2f, maxTokens=%d, provider=%s",
		model, temperature, maxTokens, provider))

	return newLLM, nil
}

// filterAvailableTools 根据配置过滤可用工具
func (a *AgentInstance) filterAvailableTools(enabledToolIDs []string) {
	if len(enabledToolIDs) == 0 {
		return
	}

	filteredTools := []model.Tool{}
	for _, tool := range a.AvailableTools {
		enabled := false
		for _, toolID := range enabledToolIDs {
			if tool.ID == toolID {
				enabled = true
				break
			}
		}

		if enabled {
			filteredTools = append(filteredTools, tool)
			logger.Debug(fmt.Sprintf("工具 %s (ID=%s) 已启用", tool.Name, tool.ID))
		} else {
			logger.Debug(fmt.Sprintf("工具 %s (ID=%s) 已禁用", tool.Name, tool.ID))
		}
	}

	a.AvailableTools = filteredTools
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/aiops/AiOpsHub/backend/pkg/llm"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/aiops/AiOpsHub/backend/pkg/mcp"
)

func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}

type ChatService struct {
	repo        *repository.ChatRepository
	llm         *llm.EinoLLM
	ragSvc      *RAGService
	mcpSvc      *MCPService
	agentSvc    *AgentService
	agentRouter *AgentRouter
	tokenSvc    *TokenService
	maxCtx      int
	enableRAG   bool
}

func NewChatService(llmConfig llm.EinoLLMConfig, ragSvc *RAGService, mcpSvc *MCPService, agentSvc *AgentService, tokenSvc *TokenService) (*ChatService, error) {
	einoLLM, err := llm.NewEinoLLM(llmConfig)
	if err != nil {
		return nil, fmt.Errorf("创建LLM失败: %w", err)
	}

	logger.Debug(fmt.Sprintf("NewChatService: ragSvc=%v, mcpSvc=%v, agentSvc=%v, tokenSvc=%v", ragSvc != nil, mcpSvc != nil, agentSvc != nil, tokenSvc != nil))

	agentRouter := NewAgentRouter(agentSvc)

	return &ChatService{
		repo:        repository.NewChatRepository(),
		llm:         einoLLM,
		ragSvc:      ragSvc,
		mcpSvc:      mcpSvc,
		agentSvc:    agentSvc,
		agentRouter: agentRouter,
		tokenSvc:    tokenSvc,
		maxCtx:      10,
		enableRAG:   ragSvc != nil,
	}, nil
}

// CreateSession 创建新的对话会话
// 参数：userID - 用户ID，title - 会话标题，modelName - 使用的模型名称
// 返回：会话对象和错误信息
func (s *ChatService) CreateSession(userID, title, modelName string) (*model.ChatSession, error) {
	session := &model.ChatSession{
		UserID: userID,
		Title:  title,
		Model:  modelName,
		Status: "active",
	}

	err := s.repo.CreateSession(session)
	if err != nil {
		logger.Error(fmt.Sprintf("创建会话失败: %v", err))
		return nil, err
	}

	logger.Debug(fmt.Sprintf("创建会话成功: %s", session.ID))
	return session, nil
}

// SendMessage 发送消息并获取AI回复（非流式）
// 参数：ctx - 上下文，sessionID - 会话ID，content - 用户消息内容
// 返回：AI回复内容、用户消息对象、AI消息对象和错误信息
func (s *ChatService) SendMessage(ctx context.Context, sessionID, content string) (string, *model.ChatMessage, *model.ChatMessage, []map[string]interface{}, error) {
	logger.Debug(fmt.Sprintf("=== SendMessage START: session=%s ===", sessionID))

	selectedAgent, err := s.agentRouter.RouteAgent(ctx, content)
	if err != nil {
		logger.Error(fmt.Sprintf("Agent 路由失败: %v", err))
	} else if selectedAgent != nil {
		logger.Info(fmt.Sprintf("✅ 智能路由选择 Agent: %s (%s)", selectedAgent.Name, selectedAgent.ID))
	}

	var ragReferences []map[string]interface{}
	var ragRefsJSON string

	_, err = s.repo.GetSessionByID(sessionID)
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("获取会话失败: %w", err)
	}

	userMessage := &model.ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   content,
	}
	if err := s.repo.CreateMessage(userMessage); err != nil {
		return "", nil, nil, nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	historyMessages, err := s.repo.GetRecentMessages(sessionID, s.maxCtx)
	if err != nil {
		logger.Error(fmt.Sprintf("获取历史消息失败: %v", err))
	}

	var promptBuilder strings.Builder

	if selectedAgent != nil && selectedAgent.SystemPrompt != "" {
		logger.Info(fmt.Sprintf("使用 Agent SystemPrompt: %s", selectedAgent.Name))
		promptBuilder.WriteString(selectedAgent.SystemPrompt)
		promptBuilder.WriteString("\n\n")
	}

	if s.enableRAG && s.ragSvc != nil {
		logger.Info(fmt.Sprintf("RAG已启用,正在检索相关知识: query=%s", content))
		searchResults, err := s.ragSvc.SearchKnowledge(ctx, content, 3)
		if err != nil {
			logger.Error(fmt.Sprintf("RAG检索失败: %v", err))
		} else if len(searchResults) > 0 {
			logger.Info(fmt.Sprintf("RAG检索成功,找到%d个相关文档", len(searchResults)))

			for _, result := range searchResults {
				ragReferences = append(ragReferences, map[string]interface{}{
					"id":        result.Document.ID,
					"title":     result.Document.Title,
					"doc_type":  result.Document.DocType,
					"component": result.Document.Component,
					"score":     result.Score,
					"snippet":   truncateContent(result.Document.Content, 100),
				})
			}

			if len(ragReferences) > 0 {
				refsBytes, err := json.Marshal(ragReferences)
				if err != nil {
					logger.Error(fmt.Sprintf("序列化RAG引用失败: %v", err))
				} else {
					ragRefsJSON = string(refsBytes)
				}
			}

			knowledgeContext, err := s.ragSvc.GetContextForQuery(ctx, content, 1000)
			if err != nil {
				logger.Error(fmt.Sprintf("构建知识上下文失败: %v", err))
			} else if knowledgeContext != "" {
				logger.Info(fmt.Sprintf("RAG检索成功,检索到%d个字符的上下文", len(knowledgeContext)))
				promptBuilder.WriteString(knowledgeContext)
				promptBuilder.WriteString("\n")
			}
		} else {
			logger.Debug("RAG检索完成,未找到相关知识")
		}
	} else {
		logger.Debug("RAG未启用或RAGService为nil")
	}

	if len(historyMessages) > 0 {
		promptBuilder.WriteString("以下是历史对话记录:\n\n")
		for _, msg := range historyMessages {
			if msg.Role == "user" {
				promptBuilder.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
			} else if msg.Role == "assistant" {
				promptBuilder.WriteString(fmt.Sprintf("助手: %s\n", msg.Content))
			}
		}

		if s.mcpSvc != nil {
			logger.Info("MCP已启用,正在加载工具...")
			mcpTools, err := s.mcpSvc.GetAllActiveTools(ctx)
			if err != nil {
				logger.Error(fmt.Sprintf("获取MCP工具失败: %v", err))
			} else if len(mcpTools) > 0 {
				totalTools := 0
				for _, tools := range mcpTools {
					totalTools += len(tools)
				}
				logger.Info(fmt.Sprintf("MCP加载成功,共%d个工具", totalTools))

				promptBuilder.WriteString("\n可用工具:\n")
				for serverName, tools := range mcpTools {
					promptBuilder.WriteString(fmt.Sprintf("### %s:\n", serverName))
					for _, tool := range tools {
						promptBuilder.WriteString(fmt.Sprintf("- %s: %s\n", tool.Name, tool.Description))
					}
				}
				promptBuilder.WriteString("\n如果需要使用工具,请按以下格式调用:\n")
				promptBuilder.WriteString("```tool_call\n{\"tool\": \"工具名称\", \"server\": \"服务器名称\", \"arguments\": {参数}}\n```\n")
			}
		}

		promptBuilder.WriteString("\n当前用户问题: ")
		promptBuilder.WriteString(content)
	} else {
		if s.mcpSvc != nil {
			logger.Debug("MCP已启用,正在加载工具...")
			mcpTools, err := s.mcpSvc.GetAllActiveTools(ctx)
			if err != nil {
				logger.Error(fmt.Sprintf("获取MCP工具失败: %v", err))
			} else if len(mcpTools) > 0 {
				totalTools := 0
				for _, tools := range mcpTools {
					totalTools += len(tools)
				}
				logger.Info(fmt.Sprintf("MCP加载成功,共%d个工具", totalTools))

				promptBuilder.WriteString("\n可用工具:\n")
				for serverName, tools := range mcpTools {
					promptBuilder.WriteString(fmt.Sprintf("### %s:\n", serverName))
					for _, tool := range tools {
						promptBuilder.WriteString(fmt.Sprintf("- %s: %s\n", tool.Name, tool.Description))
					}
				}
				promptBuilder.WriteString("\n如果需要使用工具,请按以下格式调用:\n")
				promptBuilder.WriteString("```tool_call\n{\"tool\": \"工具名称\", \"server\": \"服务器名称\", \"arguments\": {参数}}\n```\n")
			}
		}

		if s.enableRAG && s.ragSvc != nil || s.mcpSvc != nil {
			promptBuilder.WriteString("\n当前用户问题: ")
			promptBuilder.WriteString(content)
		} else {
			promptBuilder.WriteString(content)
		}
	}

	fullPrompt := promptBuilder.String()

	var agentID string
	if selectedAgent != nil {
		agentID = selectedAgent.ID
	} else {
		agentID = "default"
	}

	var aiResponse string

	if s.tokenSvc != nil {
		recorder := llm.NewTokenRecorder(sessionID, agentID, "qwen3.7-max", s.tokenSvc)
		handler := recorder.CreateCallbackHandler()
		aiResponse, _, err = s.llm.GenerateWithCallback(ctx, fullPrompt, handler)
	} else {
		aiResponse, err = s.llm.Generate(ctx, fullPrompt)
	}

	if err != nil {
		logger.Error(fmt.Sprintf("LLM生成回复失败: %v", err))
		return "", userMessage, nil, ragReferences, fmt.Errorf("生成回复失败: %w", err)
	}

	finalResponse := aiResponse
	if s.mcpSvc != nil && strings.Contains(aiResponse, "```tool_call") {
		logger.Debug("检测到工具调用请求，准备执行工具并让 LLM 处理结果")
		toolCalls := s.parseToolCalls(aiResponse)

		var toolResults []string
		for _, tc := range toolCalls {
			toolName, ok := tc["tool"].(string)
			if !ok {
				logger.Error("工具调用缺少 tool 字段")
				continue
			}

			serverName, ok := tc["server"].(string)
			if !ok {
				serverName = "Unknown"
				logger.Error("工具调用缺少 server 字段，使用默认值")
			}

			arguments, ok := tc["arguments"].(map[string]interface{})
			if !ok {
				arguments = make(map[string]interface{})
				logger.Error("工具调用缺少 arguments 字段，使用空参数")
			}

			logger.Info(fmt.Sprintf("执行工具: %s.%s", serverName, toolName))
			result, err := s.executeMCPTool(ctx, serverName, toolName, arguments)
			if err != nil {
				logger.Error(fmt.Sprintf("工具执行失败: %v", err))
				toolResults = append(toolResults, fmt.Sprintf("工具 %s 执行失败: %v", toolName, err))
			} else {
				logger.Info(fmt.Sprintf("工具执行成功: %s，结果长度: %d", toolName, len(result)))
				if len(result) > 2000 {
					logger.Debug(fmt.Sprintf("工具结果过大（%d 字符），截取前 2000 字符", len(result)))
					result = result[:2000]
				}
				toolResults = append(toolResults, result)
			}
		}

		if len(toolResults) > 0 {
			var toolContext strings.Builder
			toolContext.WriteString("\n\n以下是工具执行结果:\n")
			for i, result := range toolResults {
				toolContext.WriteString(fmt.Sprintf("### 工具结果 %d:\n%s\n", i+1, result))
			}

			processPrompt := fullPrompt + aiResponse + toolContext.String()

			logger.Info("让 LLM 处理工具结果并生成自然语言回复")

			var llmResponse string
			if s.tokenSvc != nil {
				recorder := llm.NewTokenRecorder(sessionID, agentID, "qwen3.7-max", s.tokenSvc)
				handler := recorder.CreateCallbackHandler()
				llmResponse, _, err = s.llm.GenerateWithCallback(ctx, processPrompt, handler)
			} else {
				llmResponse, err = s.llm.Generate(ctx, processPrompt)
			}

			if err != nil {
				logger.Error(fmt.Sprintf("LLM 处理工具结果失败: %v", err))
				finalResponse = aiResponse + "\n\n工具执行结果:\n" + strings.Join(toolResults, "\n")
			} else {
				finalResponse = llmResponse
			}
		}
	}

	aiMessage := &model.ChatMessage{
		SessionID:     sessionID,
		Role:          "assistant",
		Content:       finalResponse,
		RAGReferences: ragRefsJSON,
	}
	if err := s.repo.CreateMessage(aiMessage); err != nil {
		logger.Error(fmt.Sprintf("保存AI消息失败: %v", err))
		return finalResponse, userMessage, nil, ragReferences, fmt.Errorf("保存AI消息失败: %w", err)
	}

	logger.Debug(fmt.Sprintf("对话完成 - 会话: %s, 用户消息: %s, AI回复长度: %d",
		sessionID, userMessage.ID, len(finalResponse)))

	return finalResponse, userMessage, aiMessage, ragReferences, nil
}

func (s *ChatService) GetSessionHistory(sessionID string) (*model.ChatSessionWithMessagesResponse, error) {
	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取会话失败: %w", err)
	}

	// 获取会话的所有消息
	messages, err := s.repo.GetMessagesBySessionID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}

	// 转换为响应格式，并解析RAGReferences字段
	messageResponses := make([]model.ChatMessageResponse, len(messages))
	for i := range messages {
		messageResponses[i] = model.ChatMessageResponse{
			ID:        messages[i].ID,
			SessionID: messages[i].SessionID,
			Role:      messages[i].Role,
			Content:   messages[i].Content,
			Tokens:    messages[i].Tokens,
			CreatedAt: messages[i].CreatedAt,
		}

		// 解析RAGReferences JSON字符串为数组
		if messages[i].RAGReferences != "" {
			var ragRefs []map[string]interface{}
			if err := json.Unmarshal([]byte(messages[i].RAGReferences), &ragRefs); err != nil {
				logger.Error(fmt.Sprintf("解析消息RAG引用失败: %v", err))
				messageResponses[i].RAGReferences = []map[string]interface{}{}
			} else {
				messageResponses[i].RAGReferences = ragRefs
				if len(ragRefs) > 0 {
					logger.Debug(fmt.Sprintf("消息 %s 有 %d 个RAG引用", messages[i].ID, len(ragRefs)))
				}
			}
		} else {
			messageResponses[i].RAGReferences = []map[string]interface{}{}
		}
	}

	return &model.ChatSessionWithMessagesResponse{
		ChatSession: *session,
		Messages:    messageResponses,
	}, nil
}

// GetUserSessions 获取用户的所有对话会话列表
// 参数：userID - 用户ID，limit - 返回数量限制（0表示不限制）
// 返回：会话列表和错误信息
func (s *ChatService) GetUserSessions(userID string, limit int) ([]model.ChatSession, error) {
	sessions, err := s.repo.GetSessionsByUserID(userID, limit)
	if err != nil {
		logger.Error(fmt.Sprintf("获取用户会话列表失败: %v", err))
		return nil, err
	}
	return sessions, nil
}

// DeleteSession 删除会话及其所有消息
// 参数：sessionID - 会话ID
// 返回：错误信息
func (s *ChatService) DeleteSession(sessionID string) error {
	err := s.repo.DeleteSession(sessionID)
	if err != nil {
		logger.Error(fmt.Sprintf("删除会话失败: %v", err))
		return err
	}
	logger.Debug(fmt.Sprintf("删除会话成功: %s", sessionID))
	return nil
}

// generateSessionTitle 根据用户消息生成会话标题
// 参数：content - 用户消息内容
// 返回：生成的标题
func (s *ChatService) generateSessionTitle(content string) string {
	// 截取前30个字符作为标题，使用rune处理多字节字符
	runes := []rune(content)
	if len(runes) > 30 {
		return string(runes[:30]) + "..."
	}
	return content
}

// StreamSendMessage 发送消息并流式获取AI回复
// 参数：ctx - 上下文，sessionID - 会话ID，content - 用户消息内容
// 返回：流式响应channel、用户消息对象、RAG引用和错误信息
func (s *ChatService) StreamSendMessage(ctx context.Context, sessionID, content string) (<-chan string, *model.ChatMessage, []map[string]interface{}, error) {
	logger.Debug(fmt.Sprintf("=== StreamSendMessage START: session=%s ===", sessionID))

	var ragReferences []map[string]interface{}
	var toolCalls []map[string]interface{}

	// 智能路由选择 Agent
	selectedAgent, err := s.agentRouter.RouteAgent(ctx, content)
	if err != nil {
		logger.Error(fmt.Sprintf("Agent 路由失败: %v", err))
	} else if selectedAgent != nil {
		logger.Info(fmt.Sprintf("✅ 智能路由选择 Agent: %s (%s)", selectedAgent.Name, selectedAgent.ID))
	}

	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("获取会话失败: %w", err)
	}

	userMessage := &model.ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   content,
	}
	if err := s.repo.CreateMessage(userMessage); err != nil {
		return nil, nil, nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	historyMessages, err := s.repo.GetRecentMessages(sessionID, s.maxCtx)
	if err != nil {
		logger.Error(fmt.Sprintf("获取历史消息失败: %v", err))
	}

	var promptBuilder strings.Builder

	// 如果智能路由选择了 Agent，使用 Agent 的 SystemPrompt
	if selectedAgent != nil && selectedAgent.SystemPrompt != "" {
		logger.Debug(fmt.Sprintf("使用 Agent SystemPrompt: %s", selectedAgent.Name))
		promptBuilder.WriteString(selectedAgent.SystemPrompt)
		promptBuilder.WriteString("\n\n")
	}

	if s.enableRAG && s.ragSvc != nil {
		logger.Debug(fmt.Sprintf("RAG已启用,正在检索相关知识: query=%s", content))

		searchResults, err := s.ragSvc.SearchKnowledge(ctx, content, 3)
		if err != nil {
			logger.Error(fmt.Sprintf("RAG检索失败: %v", err))
		} else if len(searchResults) > 0 {
			logger.Info(fmt.Sprintf("RAG检索成功,找到%d个相关文档", len(searchResults)))

			for _, result := range searchResults {
				ragReferences = append(ragReferences, map[string]interface{}{
					"id":        result.Document.ID,
					"title":     result.Document.Title,
					"doc_type":  result.Document.DocType,
					"component": result.Document.Component,
					"score":     result.Score,
					"snippet":   truncateContent(result.Document.Content, 100),
				})
			}

			knowledgeContext, err := s.ragSvc.GetContextForQuery(ctx, content, 1000)
			if err != nil {
				logger.Error(fmt.Sprintf("构建知识上下文失败: %v", err))
			} else if knowledgeContext != "" {
				logger.Info(fmt.Sprintf("RAG检索成功,检索到%d个字符的上下文", len(knowledgeContext)))
				promptBuilder.WriteString(knowledgeContext)
				promptBuilder.WriteString("\n")
			}
		} else {
			logger.Debug("RAG检索完成,未找到相关知识")
		}
	} else {
		logger.Debug("RAG未启用或RAGService为nil")
	}

	// MCP 工具集成
	if s.mcpSvc != nil {
		logger.Debug("MCP已启用,正在加载工具...")
		mcpTools, err := s.mcpSvc.GetAllActiveTools(ctx)
		if err != nil {
			logger.Error(fmt.Sprintf("获取MCP工具失败: %v", err))
		} else if len(mcpTools) > 0 {
			containsToolCall := strings.Contains(content, "调用工具") ||
				strings.Contains(content, "执行命令") ||
				strings.Contains(content, "Jenkins") ||
				strings.Contains(content, "构建") ||
				strings.Contains(content, "部署到")

			if containsToolCall {
				totalTools := 0
				for _, tools := range mcpTools {
					totalTools += len(tools)
				}
				logger.Info(fmt.Sprintf("MCP加载成功,共%d个工具（检测到工具调用意图）", totalTools))

				promptBuilder.WriteString("\n可用工具:\n")
				for serverName, tools := range mcpTools {
					promptBuilder.WriteString(fmt.Sprintf("### %s:\n", serverName))
					for _, tool := range tools {
						promptBuilder.WriteString(fmt.Sprintf("- %s: %s\n", tool.Name, tool.Description))
					}
				}
				promptBuilder.WriteString("\n如果需要使用工具,请按以下格式调用:\n")
				promptBuilder.WriteString("```tool_call\n{\"tool\": \"工具名称\", \"server\": \"服务器名称\", \"arguments\": {参数}}\n```\n")
			} else {
				logger.Debug("用户问题不需要工具调用，跳过 MCP 工具加载")
			}
		}
	}

	if len(historyMessages) > 0 {
		promptBuilder.WriteString("以下是历史对话记录:\n\n")
		for _, msg := range historyMessages {
			if msg.Role == "user" {
				promptBuilder.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
			} else if msg.Role == "assistant" {
				promptBuilder.WriteString(fmt.Sprintf("助手: %s\n", msg.Content))
			}
		}
		promptBuilder.WriteString("\n当前用户问题: ")
		promptBuilder.WriteString(content)
	} else {
		if s.enableRAG && s.ragSvc != nil || s.mcpSvc != nil {
			promptBuilder.WriteString("\n当前用户问题: ")
			promptBuilder.WriteString(content)
		} else {
			promptBuilder.WriteString(content)
		}
	}

	fullPrompt := promptBuilder.String()

	// 创建输出channel，缓冲大小100，用于流式传输AI回复
	// 注意：不要在此等待goroutine完成，应立即返回channel让handler读取
	outputChan := make(chan string, 100)

	// 启动goroutine异步处理流式生成，避免阻塞主流程
	go func() {
		// goroutine结束时关闭channel，通知handler数据传输完成
		defer close(outputChan)

		logger.Debug("Starting LLM stream generation...")

		// 调用LLM流式生成接口，获取流式响应channel
		streamChan, err := s.llm.StreamGenerate(ctx, fullPrompt)
		if err != nil {
			logger.Error(fmt.Sprintf("LLM流式生成失败: %v", err))
			// 发送错误信息到channel，通知前端
			outputChan <- fmt.Sprintf("生成回复失败: %v", err)
			return
		}

		logger.Debug("LLM stream channel obtained, reading chunks...")

		// 用于累积完整响应内容，便于后续工具调用检测
		var fullResponse strings.Builder
		chunkCount := 0

		// 从LLM流式channel读取每个chunk并转发到输出channel
		for chunk := range streamChan {
			chunkCount++
			// 记录chunk信息（仅前20字符避免日志过长）
			logger.Debug(fmt.Sprintf("Received chunk #%d: %s", chunkCount, chunk[:min(20, len(chunk))]))
			// 累积完整响应
			fullResponse.WriteString(chunk)
			// 立即发送chunk到前端，实现实时流式输出
			outputChan <- chunk
		}

		logger.Debug(fmt.Sprintf("Stream completed, total chunks: %d, total length: %d", chunkCount, fullResponse.Len()))

		// 检查是否有工具调用（MCP工具集成）
		responseText := fullResponse.String()
		if strings.Contains(responseText, "```tool_call") {
			logger.Debug("检测到工具调用请求，准备执行工具并让 LLM 处理结果")
			// 解析工具调用请求
			toolCalls = s.parseToolCalls(responseText)

			// 执行所有工具并收集结果
			var toolResults []string
			for _, tc := range toolCalls {
				// 安全地提取工具参数
				toolName, ok := tc["tool"].(string)
				if !ok {
					logger.Error("工具调用缺少 tool 字段")
					continue
				}

				serverName, ok := tc["server"].(string)
				if !ok {
					serverName = "Unknown"
					logger.Error("工具调用缺少 server 字段，使用默认值")
				}

				arguments, ok := tc["arguments"].(map[string]interface{})
				if !ok {
					arguments = make(map[string]interface{})
					logger.Error("工具调用缺少 arguments 字段，使用空参数")
				}

				logger.Info(fmt.Sprintf("执行工具: %s.%s", serverName, toolName))
				// 发送工具执行状态到前端
				outputChan <- fmt.Sprintf("\n🔧 正在执行工具: %s...\n", toolName)

				// 执行MCP工具调用
				result, err := s.executeMCPTool(ctx, serverName, toolName, arguments)
				if err != nil {
					logger.Error(fmt.Sprintf("工具执行失败: %v", err))
					toolResults = append(toolResults, fmt.Sprintf("工具 %s 执行失败: %v", toolName, err))
					// 发送失败状态到前端
					outputChan <- fmt.Sprintf("❌ 工具 %s 执行失败\n", toolName)
				} else {
					logger.Info(fmt.Sprintf("工具执行成功: %s，结果长度: %d", toolName, len(result)))

					// 如果结果太大，截取前 2000 字符，避免传输过大数据
					if len(result) > 2000 {
						logger.Debug(fmt.Sprintf("工具结果过大（%d 字符），截取前 2000 字符", len(result)))
						result = result[:2000] + "\n...(数据过大，已截取部分内容)"
					}

					toolResults = append(toolResults, fmt.Sprintf("工具 %s 执行结果:\n%s", toolName, result))
					// 发送成功状态到前端
					outputChan <- fmt.Sprintf("✅ 工具 %s 执行成功，正在处理结果...\n\n", toolName)
				}
			}

			// 如果有工具执行结果，让 LLM 流式处理
			if len(toolResults) > 0 {
				toolContext := strings.Join(toolResults, "\n\n")

				processPrompt := fmt.Sprintf(`
之前的对话中，AI 助手决定调用工具来帮助用户。

工具执行结果如下：
%s

现在工具已经执行完毕，返回了上述结果。
请基于这些工具返回的实际数据，用自然、友好的语言回答用户的问题。
不要简单地重复工具返回的 JSON 数据，而是要：
1. 理解数据的含义
2. 提取关键信息
3. 用通俗易懂的语言向用户解释结果
4. 如果有错误，说明错误原因并建议解决方案

请开始回答：`, toolContext)

				logger.Debug("将工具结果发送给 LLM 进行流式处理和理解")

				// 再次调用 LLM 流式处理工具结果
				streamChan2, err := s.llm.StreamGenerate(ctx, processPrompt)
				if err != nil {
					logger.Error(fmt.Sprintf("LLM 处理工具结果失败: %v", err))
					outputChan <- fmt.Sprintf("\n处理结果时出错，原始工具结果:\n%s\n", toolContext)
				} else {
					for chunk := range streamChan2 {
						outputChan <- chunk
					}
				}
			}
		}
	}()

	session.Title = s.generateSessionTitle(content)
	if err := s.repo.UpdateSession(session); err != nil {
		logger.Error(fmt.Sprintf("更新会话失败: %v", err))
	}

	logger.Debug(fmt.Sprintf("流式对话开始 - 会话: %s, 用户消息: %s", sessionID, userMessage.ID))

	return outputChan, userMessage, ragReferences, nil
}

func (s *ChatService) parseToolCalls(text string) []map[string]interface{} {
	var calls []map[string]interface{}

	// 解析 ```tool_call 块
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

		var tc map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &tc); err != nil {
			logger.Error(fmt.Sprintf("解析工具调用失败: %v", err))
			continue
		}

		calls = append(calls, tc)
	}

	return calls
}

func (s *ChatService) executeMCPTool(ctx context.Context, serverName, toolName string, arguments map[string]interface{}) (string, error) {
	if s.mcpSvc == nil {
		return "", fmt.Errorf("MCP服务未启用")
	}

	serverID, err := s.mcpSvc.FindServerByToolName(ctx, toolName)
	if err != nil {
		return "", fmt.Errorf("找不到工具 %s: %v", toolName, err)
	}

	result, err := s.mcpSvc.CallTool(ctx, serverID, toolName, arguments)
	if err != nil {
		return "", err
	}

	return mcp.ExtractTextContent(result), nil
}

// SaveAIMessage 保存AI消息到数据库（流式完成后调用）
func (s *ChatService) SaveAIMessage(sessionID, content string, ragReferences []map[string]interface{}) (*model.ChatMessage, error) {
	ragRefsJSON := ""
	if len(ragReferences) > 0 {
		refsBytes, err := json.Marshal(ragReferences)
		if err != nil {
			logger.Error(fmt.Sprintf("序列化RAG引用失败: %v", err))
		} else {
			ragRefsJSON = string(refsBytes)
		}
	}

	aiMessage := &model.ChatMessage{
		SessionID:     sessionID,
		Role:          "assistant",
		Content:       content,
		RAGReferences: ragRefsJSON,
	}

	if err := s.repo.CreateMessage(aiMessage); err != nil {
		logger.Error(fmt.Sprintf("保存AI消息失败: %v", err))
		return nil, err
	}

	if s.tokenSvc != nil {
		history, err := s.repo.GetRecentMessages(sessionID, 2)
		if err == nil && len(history) >= 2 {
			userMsg := history[len(history)-2]
			inputTokens := len(userMsg.Content) / 4
			outputTokens := len(content) / 4
			totalTokens := inputTokens + outputTokens

			session, err := s.repo.GetSessionByID(sessionID)
			var agentID string
			if err == nil && session != nil {
				agentID = "default"
			}

			usage := TokenUsage{
				SessionID:    sessionID,
				AgentID:      agentID,
				Model:        "qwen-turbo",
				InputTokens:  inputTokens,
				OutputTokens: outputTokens,
				TotalTokens:  totalTokens,
			}

			if err := s.tokenSvc.RecordUsage(context.Background(), usage); err != nil {
				logger.Error(fmt.Sprintf("记录Token使用失败: %v", err))
			} else {
				logger.Debug(fmt.Sprintf("记录Token使用成功: session=%s, input=%d, output=%d, total=%d",
					sessionID, inputTokens, outputTokens, totalTokens))
			}
		}
	}

	logger.Debug(fmt.Sprintf("保存AI消息成功: %s", aiMessage.ID))
	return aiMessage, nil
}

func formatToolResult(result string) string {
	maxLen := 2000
	if len(result) <= maxLen {
		return result
	}
	return result[:maxLen] + "\n...(结果过长，已截取)"
}

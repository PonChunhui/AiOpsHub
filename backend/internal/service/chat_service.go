package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

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
	maxCtx      int
	enableRAG   bool
}

func NewChatService(llmConfig llm.EinoLLMConfig, ragSvc *RAGService, mcpSvc *MCPService, agentSvc *AgentService) (*ChatService, error) {
	einoLLM, err := llm.NewEinoLLM(llmConfig)
	if err != nil {
		return nil, fmt.Errorf("创建LLM失败: %w", err)
	}

	logger.Info(fmt.Sprintf("NewChatService: ragSvc=%v, mcpSvc=%v, agentSvc=%v", ragSvc != nil, mcpSvc != nil, agentSvc != nil))

	// 创建 Agent 路由器
	agentRouter := NewAgentRouter(agentSvc)

	return &ChatService{
		repo:        repository.NewChatRepository(),
		llm:         einoLLM,
		ragSvc:      ragSvc,
		mcpSvc:      mcpSvc,
		agentSvc:    agentSvc,
		agentRouter: agentRouter,
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

	logger.Info(fmt.Sprintf("创建会话成功: %s", session.ID))
	return session, nil
}

// SendMessage 发送消息并获取AI回复（非流式）
// 参数：ctx - 上下文，sessionID - 会话ID，content - 用户消息内容
// 返回：AI回复内容、用户消息对象、AI消息对象和错误信息
func (s *ChatService) SendMessage(ctx context.Context, sessionID, content string) (string, *model.ChatMessage, *model.ChatMessage, []map[string]interface{}, error) {
	logger.Info(fmt.Sprintf("=== SendMessage START: session=%s ===", sessionID))
	logger.Info(fmt.Sprintf("enableRAG=%v, ragSvc=%v", s.enableRAG, s.ragSvc != nil))

	// 智能路由选择 Agent
	selectedAgent, err := s.agentRouter.RouteAgent(ctx, content)
	if err != nil {
		logger.Error(fmt.Sprintf("Agent 路由失败: %v", err))
	} else if selectedAgent != nil {
		logger.Info(fmt.Sprintf("✅ 智能路由选择 Agent: %s (%s)", selectedAgent.Name, selectedAgent.ID))
	}
	logger.Info(fmt.Sprintf("enableRAG=%v, ragSvc=%v", s.enableRAG, s.ragSvc != nil))

	// RAG引用的知识库文档
	var ragReferences []map[string]interface{}

	// 获取会话信息
	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("获取会话失败: %w", err)
	}

	// 创建用户消息
	userMessage := &model.ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   content,
	}
	if err := s.repo.CreateMessage(userMessage); err != nil {
		return "", nil, nil, nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	// 获取历史消息构建上下文
	historyMessages, err := s.repo.GetRecentMessages(sessionID, s.maxCtx)
	if err != nil {
		logger.Error(fmt.Sprintf("获取历史消息失败: %v", err))
		// 如果获取失败，继续使用单轮对话
	}

	// 构建完整的对话提示（包含历史上下文和RAG知识）
	var promptBuilder strings.Builder

	// 如果智能路由选择了 Agent，使用 Agent 的 SystemPrompt
	if selectedAgent != nil && selectedAgent.SystemPrompt != "" {
		logger.Info(fmt.Sprintf("使用 Agent SystemPrompt: %s", selectedAgent.Name))
		promptBuilder.WriteString(selectedAgent.SystemPrompt)
		promptBuilder.WriteString("\n\n")
	}

	// 如果启用了RAG，先检索相关知识
	if s.enableRAG && s.ragSvc != nil {
		logger.Info(fmt.Sprintf("RAG已启用,正在检索相关知识: query=%s", content))

		// 检索知识库文档
		searchResults, err := s.ragSvc.SearchKnowledge(ctx, content, 3)
		if err != nil {
			logger.Error(fmt.Sprintf("RAG检索失败: %v", err))
		} else if len(searchResults) > 0 {
			logger.Info(fmt.Sprintf("RAG检索成功,找到%d个相关文档", len(searchResults)))

			// 构建RAG引用信息
			for _, result := range searchResults {
				ragReferences = append(ragReferences, map[string]interface{}{
					"id":       result.Document.ID,
					"title":    result.Document.Title,
					"category": result.Document.Category,
					"score":    result.Score,
					"snippet":  truncateContent(result.Document.Content, 100), // 截取前100字符
				})
			}

			// 构建知识上下文
			knowledgeContext, err := s.ragSvc.GetContextForQuery(ctx, content, 1000)
			if err != nil {
				logger.Error(fmt.Sprintf("构建知识上下文失败: %v", err))
			} else if knowledgeContext != "" {
				logger.Info(fmt.Sprintf("RAG检索成功,检索到%d个字符的上下文", len(knowledgeContext)))
				promptBuilder.WriteString(knowledgeContext)
				promptBuilder.WriteString("\n")
			}
		} else {
			logger.Info("RAG检索完成,未找到相关知识")
		}
	} else {
		logger.Info("RAG未启用或RAGService为nil")
	}

	// 添加历史对话记录
	if len(historyMessages) > 0 {
		promptBuilder.WriteString("以下是历史对话记录:\n\n")
		for _, msg := range historyMessages {
			if msg.Role == "user" {
				promptBuilder.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
			} else if msg.Role == "assistant" {
				promptBuilder.WriteString(fmt.Sprintf("助手: %s\n", msg.Content))
			}
		}

		// MCP 工具集成
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

				// 添加工具信息到 prompt
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
		// MCP 工具集成
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

				// 添加工具信息到 prompt
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

	// 调用LLM生成回复
	aiResponse, err := s.llm.Generate(ctx, fullPrompt)
	if err != nil {
		logger.Error(fmt.Sprintf("LLM生成回复失败: %v", err))
		return "", userMessage, nil, ragReferences, fmt.Errorf("生成回复失败: %w", err)
	}

	// 检查是否有工具调用
	finalResponse := aiResponse
	if s.mcpSvc != nil && strings.Contains(aiResponse, "```tool_call") {
		logger.Info("检测到工具调用请求，准备执行工具并让 LLM 处理结果")
		toolCalls := s.parseToolCalls(aiResponse)

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

			result, err := s.executeMCPTool(ctx, serverName, toolName, arguments)
			if err != nil {
				logger.Error(fmt.Sprintf("工具执行失败: %v", err))
				toolResults = append(toolResults, fmt.Sprintf("工具 %s 执行失败: %v", toolName, err))
			} else {
				logger.Info(fmt.Sprintf("工具执行成功: %s，结果长度: %d", toolName, len(result)))

				// 如果结果太大，只保留前 2000 字符
				if len(result) > 2000 {
					logger.Info(fmt.Sprintf("工具结果过大（%d 字符），截取前 2000 字符", len(result)))
					result = result[:2000] + "\n...(数据过大，已截取部分内容)"
				}

				toolResults = append(toolResults, fmt.Sprintf("工具 %s 执行结果:\n%s", toolName, result))
			}
		}

		// 如果有工具执行结果，构建新的 prompt 让 LLM 处理
		if len(toolResults) > 0 {
			toolContext := strings.Join(toolResults, "\n\n")

			// 构建让 LLM 理解工具结果的 prompt
			processPrompt := fmt.Sprintf(`
之前的对话中，AI 助手决定调用以下工具来帮助用户：

%s

现在工具已经执行完毕，返回了上述结果。
请基于这些工具返回的实际数据，用自然、友好的语言回答用户的问题。
不要简单地重复工具返回的 JSON 数据，而是要：
1. 理解数据的含义
2. 提取关键信息
3. 用通俗易懂的语言向用户解释结果
4. 如果有错误，说明错误原因并建议解决方案

请开始回答：`, toolContext)

			logger.Info("将工具结果发送给 LLM 进行理解和处理")

			// 再次调用 LLM 让它处理工具结果
			llmResponse, err := s.llm.Generate(ctx, processPrompt)
			if err != nil {
				logger.Error(fmt.Sprintf("LLM 处理工具结果失败: %v", err))
				// 如果 LLM 处理失败，至少返回原始结果
				finalResponse = fmt.Sprintf("工具已执行成功，但 AI 处理结果时出错:\n\n%s", toolContext)
			} else {
				logger.Info(fmt.Sprintf("LLM 成功处理工具结果，回复长度: %d", len(llmResponse)))
				finalResponse = llmResponse
			}
		}
	}

	// 创建AI消息
	aiMessage := &model.ChatMessage{
		SessionID: sessionID,
		Role:      "assistant",
		Content:   finalResponse,
	}
	if err := s.repo.CreateMessage(aiMessage); err != nil {
		logger.Error(fmt.Sprintf("保存AI消息失败: %v", err))
		return finalResponse, userMessage, nil, ragReferences, fmt.Errorf("保存AI消息失败: %w", err)
	}

	// 更新会话的更新时间
	session.Title = s.generateSessionTitle(content)
	if err := s.repo.UpdateSession(session); err != nil {
		logger.Error(fmt.Sprintf("更新会话失败: %v", err))
	}

	logger.Info(fmt.Sprintf("对话完成 - 会话: %s, 用户消息: %s, AI回复长度: %d",
		sessionID, userMessage.ID, len(finalResponse)))

	return finalResponse, userMessage, aiMessage, ragReferences, nil
}

// GetSessionHistory 获取会话的完整历史记录
// 参数：sessionID - 会话ID
// 返回：包含消息列表的会话对象和错误信息
func (s *ChatService) GetSessionHistory(sessionID string) (*model.ChatSessionWithMessages, error) {
	// 获取会话信息
	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取会话失败: %w", err)
	}

	// 获取会话的所有消息
	messages, err := s.repo.GetMessagesBySessionID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}

	// 将每条消息的RAGReferences JSON字符串解析为数组
	for i := range messages {
		if messages[i].RAGReferences != "" {
			var ragRefs []map[string]interface{}
			if err := json.Unmarshal([]byte(messages[i].RAGReferences), &ragRefs); err != nil {
				logger.Error(fmt.Sprintf("解析消息RAG引用失败: %v", err))
			} else {
				// 将解析后的数据添加到metadata中，以便前端访问
				if messages[i].RAGReferences != "" && len(ragRefs) > 0 {
					logger.Info(fmt.Sprintf("消息 %s 有 %d 个RAG引用", messages[i].ID, len(ragRefs)))
				}
			}
		}
	}

	return &model.ChatSessionWithMessages{
		ChatSession: *session,
		Messages:    messages,
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
	logger.Info(fmt.Sprintf("删除会话成功: %s", sessionID))
	return nil
}

// generateSessionTitle 根据用户消息生成会话标题
// 参数：content - 用户消息内容
// 返回：生成的标题
func (s *ChatService) generateSessionTitle(content string) string {
	// 简单截取前30个字符作为标题
	if len(content) > 30 {
		return content[:30] + "..."
	}
	return content
}

// StreamSendMessage 发送消息并流式获取AI回复
// 参数：ctx - 上下文，sessionID - 会话ID，content - 用户消息内容
// 返回：流式响应channel、用户消息对象、RAG引用和错误信息
func (s *ChatService) StreamSendMessage(ctx context.Context, sessionID, content string) (<-chan string, *model.ChatMessage, []map[string]interface{}, error) {
	logger.Info(fmt.Sprintf("=== StreamSendMessage START: session=%s ===", sessionID))

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
					"id":       result.Document.ID,
					"title":    result.Document.Title,
					"category": result.Document.Category,
					"score":    result.Score,
					"snippet":  truncateContent(result.Document.Content, 100),
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
			logger.Info("RAG检索完成,未找到相关知识")
		}
	} else {
		logger.Info("RAG未启用或RAGService为nil")
	}

	// MCP 工具集成
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

			// 添加工具信息到 prompt
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

	outputChan := make(chan string, 100)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(outputChan)

		streamChan, err := s.llm.StreamGenerate(ctx, fullPrompt)
		if err != nil {
			logger.Error(fmt.Sprintf("LLM流式生成失败: %v", err))
			outputChan <- fmt.Sprintf("生成回复失败: %v", err)
			return
		}

		var fullResponse strings.Builder
		for chunk := range streamChan {
			fullResponse.WriteString(chunk)
			outputChan <- chunk
		}

		// 检查是否有工具调用
		responseText := fullResponse.String()
		if strings.Contains(responseText, "```tool_call") {
			logger.Info("检测到工具调用请求，准备执行工具并让 LLM 处理结果")
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
				outputChan <- fmt.Sprintf("\n🔧 正在执行工具: %s...\n", toolName)

				result, err := s.executeMCPTool(ctx, serverName, toolName, arguments)
				if err != nil {
					logger.Error(fmt.Sprintf("工具执行失败: %v", err))
					toolResults = append(toolResults, fmt.Sprintf("工具 %s 执行失败: %v", toolName, err))
					outputChan <- fmt.Sprintf("❌ 工具 %s 执行失败\n", toolName)
				} else {
					logger.Info(fmt.Sprintf("工具执行成功: %s，结果长度: %d", toolName, len(result)))

					// 如果结果太大，截取前 2000 字符
					if len(result) > 2000 {
						logger.Info(fmt.Sprintf("工具结果过大（%d 字符），截取前 2000 字符", len(result)))
						result = result[:2000] + "\n...(数据过大，已截取部分内容)"
					}

					toolResults = append(toolResults, fmt.Sprintf("工具 %s 执行结果:\n%s", toolName, result))
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

				logger.Info("将工具结果发送给 LLM 进行流式处理和理解")

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

	logger.Info(fmt.Sprintf("流式对话开始 - 会话: %s, 用户消息: %s", sessionID, userMessage.ID))

	wg.Wait()
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
	// 将ragReferences转换为JSON字符串
	ragRefsJSON := ""
	if len(ragReferences) > 0 {
		ragRefsBytes, err := json.Marshal(ragReferences)
		if err != nil {
			logger.Error(fmt.Sprintf("序列化RAG引用失败: %v", err))
		} else {
			ragRefsJSON = string(ragRefsBytes)
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
	logger.Info(fmt.Sprintf("保存AI消息成功: %s, RAG引用数量: %d", aiMessage.ID, len(ragReferences)))
	return aiMessage, nil
}

// formatToolResult 格式化工具返回结果为易读的 markdown 格式

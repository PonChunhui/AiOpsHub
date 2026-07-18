package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}

func toJSON(v interface{}) string {
	arr, _ := json.Marshal(v)
	return string(arr)
}

type RuntimeConfig struct {
	Model          string
	Temperature    float64
	MaxTokens      int
	EnableRAG      bool
	RAGTopK        int
	RAGThreshold   float64
	EnabledTools   []string
	EnableThinking bool
}

type ChatService struct {
	repo            *repository.ChatRepository
	masterRouter    *MasterRouter
	agentRuntime    *AgentRuntime
	ragSvc          *RAGService
	tokenSvc        *TokenService
	routingLogRepo  *repository.RoutingLogRepository
	toolCallLogRepo *repository.ToolCallLogRepository
	maxCtx          int
	enableRAG       bool
}

func NewChatService(
	masterRouter *MasterRouter,
	agentRuntime *AgentRuntime,
	ragSvc *RAGService,
	tokenSvc *TokenService,
) (*ChatService, error) {
	logger.Info(fmt.Sprintf("NewChatService (New Architecture): ragSvc=%v", ragSvc != nil))

	return &ChatService{
		repo:            repository.NewChatRepository(),
		masterRouter:    masterRouter,
		agentRuntime:    agentRuntime,
		ragSvc:          ragSvc,
		tokenSvc:        tokenSvc,
		routingLogRepo:  repository.NewRoutingLogRepository(),
		toolCallLogRepo: repository.NewToolCallLogRepository(),
		maxCtx:          10,
		enableRAG:       ragSvc != nil,
	}, nil
}

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

func (s *ChatService) SendMessage(ctx context.Context, sessionID, content string) (string, *model.ChatMessage, *model.ChatMessage, []map[string]interface{}, error) {
	logger.Info("=== SendMessage (New Architecture) ===")

	_, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("获取会话失败: %w", err)
	}

	history, err := s.repo.GetRecentMessages(sessionID, s.maxCtx)
	if err != nil {
		logger.Error(fmt.Sprintf("获取历史消息失败: %v", err))
	}

	var ragReferences []map[string]interface{}
	var knowledgeContext string

	if s.enableRAG && s.ragSvc != nil {
		searchResults, err := s.ragSvc.SearchKnowledge(ctx, content, 3)
		if err != nil {
			logger.Error(fmt.Sprintf("RAG检索失败: %v", err))
		} else if len(searchResults) > 0 {
			logger.Info(fmt.Sprintf("RAG检索成功，找到%d个文档", len(searchResults)))

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

			knowledgeContext, err = s.ragSvc.GetContextForQuery(ctx, content, 1000)
			if err != nil {
				logger.Error(fmt.Sprintf("构建知识上下文失败: %v", err))
			}
		}
	}

	sessionContext := s.buildSessionContext(history, knowledgeContext)

	agentInstance, routingLog, err := s.masterRouter.Route(ctx, content, sessionContext)
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("路由失败: %w", err)
	}

	logger.Info(fmt.Sprintf("✅ 选中Agent: %s (置信度 %.2f)", routingLog.SelectedAgentID, routingLog.Confidence))

	if err := s.routingLogRepo.Create(routingLog); err != nil {
		logger.Error(fmt.Sprintf("保存路由日志失败: %v", err))
	}

	userMessage := &model.ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   content,
	}
	if err := s.repo.CreateMessage(userMessage); err != nil {
		return "", nil, nil, nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	response, toolCallRecords, err := agentInstance.Execute(ctx, content, history)
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("Agent执行失败: %w", err)
	}

	for _, record := range toolCallRecords {
		callLog := &model.ToolCallLog{
			ID:           uuid.New().String(),
			SessionID:    sessionID,
			AgentID:      routingLog.SelectedAgentID,
			ToolName:     record.ToolName,
			Arguments:    toJSON(record.Arguments),
			Result:       record.Result,
			Success:      record.Success,
			Duration:     record.Duration,
			ErrorMessage: "",
			CreatedAt:    time.Now(),
		}
		if !record.Success {
			callLog.ErrorMessage = record.Result
		}

		if err := s.toolCallLogRepo.Create(callLog); err != nil {
			logger.Error(fmt.Sprintf("保存工具调用日志失败: %v", err))
		}
	}

	aiMessage := &model.ChatMessage{
		SessionID:     sessionID,
		Role:          "assistant",
		Content:       response,
		AgentID:       routingLog.SelectedAgentID,
		RAGReferences: toJSON(ragReferences),
	}
	if err := s.repo.CreateMessage(aiMessage); err != nil {
		return response, userMessage, nil, ragReferences, fmt.Errorf("保存AI消息失败: %w", err)
	}

	logger.Info(fmt.Sprintf("=== SendMessage完成，响应长度: %d ===", len(response)))
	return response, userMessage, aiMessage, ragReferences, nil
}

// SendMessageStream 发送消息并返回流式事件（全流程流式版本）
func (s *ChatService) SendMessageStream(ctx context.Context, sessionID, content string) (<-chan *model.AgentEvent, *model.ChatMessage, []map[string]interface{}, error) {
	logger.Info(fmt.Sprintf("=== SendMessageStream START: session=%s ===", sessionID))

	eventChan := make(chan *model.AgentEvent, 100)

	go func() {
		defer close(eventChan)

		// 1. 获取会话
		_, err := s.repo.GetSessionByID(sessionID)
		if err != nil {
			eventChan <- model.NewErrorEvent("ChatService", fmt.Sprintf("获取会话失败: %v", err), 500)
			return
		}

		// 2. 获取历史消息
		history, err := s.repo.GetRecentMessages(sessionID, s.maxCtx)
		if err != nil {
			logger.Error(fmt.Sprintf("获取历史消息失败: %v", err))
		}

		// 3. RAG知识检索
		var ragReferences []map[string]interface{}
		var knowledgeContext string

		if s.enableRAG && s.ragSvc != nil {
			searchResults, err := s.ragSvc.SearchKnowledge(ctx, content, 3)
			if err == nil && len(searchResults) > 0 {
				logger.Info(fmt.Sprintf("RAG检索成功，找到%d个文档", len(searchResults)))

				for _, result := range searchResults {
					ragReferences = append(ragReferences, map[string]interface{}{
						"id":       result.Document.ID,
						"title":    result.Document.Title,
						"doc_type": result.Document.DocType,
						"score":    result.Score,
					})
				}

				knowledgeContext, _ = s.ragSvc.GetContextForQuery(ctx, content, 1000)

				if len(ragReferences) > 0 {
					eventChan <- model.NewRagReferencesEvent(ragReferences)
				}
			}
		}

		// 4. 创建用户消息
		userMessage := &model.ChatMessage{
			SessionID: sessionID,
			Role:      "user",
			Content:   content,
		}
		if err := s.repo.CreateMessage(userMessage); err != nil {
			eventChan <- model.NewErrorEvent("ChatService", fmt.Sprintf("保存用户消息失败: %v", err), 500)
			return
		}

		// 发送用户消息事件
		eventChan <- model.NewUserMessageEvent(sessionID, "user", content)

		// 5. Agent路由选择
		sessionContext := s.buildSessionContext(history, knowledgeContext)
		agentInstance, routingLog, err := s.masterRouter.Route(ctx, content, sessionContext)
		if err != nil {
			eventChan <- model.NewErrorEvent("ChatService", fmt.Sprintf("路由失败: %v", err), 500)
			return
		}

		logger.Info(fmt.Sprintf("✅ 选中Agent: %s (置信度 %.2f)", routingLog.SelectedAgentID, routingLog.Confidence))

		// 保存路由日志
		if err := s.routingLogRepo.Create(routingLog); err != nil {
			logger.Error(fmt.Sprintf("保存路由日志失败: %v", err))
		}

		// 6. Agent执行（全流程流式）
		agentEventChan, toolCallRecords, err := agentInstance.ExecuteStream(ctx, content, history)
		if err != nil {
			eventChan <- model.NewErrorEvent("ChatService", fmt.Sprintf("Agent执行失败: %v", err), 500)
			return
		}

		// 7. 转发Agent事件
		var fullContent strings.Builder
		for event := range agentEventChan {
			if event.Type == model.EventContentChunk {
				if data, ok := event.Data.(model.ContentChunkEventData); ok {
					fullContent.WriteString(data.Content)
				}
			}
			eventChan <- event
		}

		// 8. 保存工具调用日志
		for _, record := range toolCallRecords {
			callLog := &model.ToolCallLog{
				ID:        uuid.New().String(),
				SessionID: sessionID,
				ToolName:  record.ToolName,
				Arguments: toJSON(record.Arguments),
				Result:    record.Result,
				Success:   record.Success,
				Duration:  record.Duration,
			}
			if !record.Success {
				callLog.ErrorMessage = record.Result
			}
			if err := s.toolCallLogRepo.Create(callLog); err != nil {
				logger.Error(fmt.Sprintf("保存工具调用日志失败: %v", err))
			}
		}

		// 9. 保存AI消息
		aiMessage := &model.ChatMessage{
			SessionID:     sessionID,
			Role:          "assistant",
			Content:       fullContent.String(),
			RAGReferences: toJSON(ragReferences),
		}
		if err := s.repo.CreateMessage(aiMessage); err != nil {
			logger.Error(fmt.Sprintf("保存AI消息失败: %v", err))
		}

		logger.Info(fmt.Sprintf("=== SendMessageStream完成，总内容长度: %d ===", fullContent.Len()))
	}()

	userMsg := &model.ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   content,
	}

	return eventChan, userMsg, nil, nil
}

func (s *ChatService) DeleteSession(sessionID string) error {
	err := s.repo.DeleteSession(sessionID)
	if err != nil {
		logger.Error(fmt.Sprintf("删除会话失败: %v", err))
		return err
	}
	logger.Info(fmt.Sprintf("删除会话成功: %s", sessionID))
	return nil
}

func (s *ChatService) StreamSendMessage(ctx context.Context, sessionID, content string) (<-chan string, *model.ChatMessage, []map[string]interface{}, error) {
	response, userMsg, _, ragRefs, err := s.SendMessage(ctx, sessionID, content)
	if err != nil {
		return nil, nil, nil, err
	}

	outputChan := make(chan string, 1)
	go func() {
		defer close(outputChan)
		outputChan <- response
	}()

	return outputChan, userMsg, ragRefs, nil
}

func (s *ChatService) StreamSendMessageWithEvents(ctx context.Context, sessionID, content string, enableThinking bool) (<-chan *model.AgentEvent, *model.ChatMessage, []map[string]interface{}, error) {
	defaultConfig := RuntimeConfig{
		Model:          viper.GetString("llm.model"),
		Temperature:    viper.GetFloat64("llm.temperature"),
		MaxTokens:      viper.GetInt("llm.max_tokens"),
		EnableRAG:      s.enableRAG,
		RAGTopK:        3,
		RAGThreshold:   0.5,
		EnabledTools:   []string{},
		EnableThinking: enableThinking,
	}
	return s.SendMessageStreamWithConfig(ctx, sessionID, content, defaultConfig)
}

func (s *ChatService) SendMessageStreamWithConfig(ctx context.Context, sessionID, content string, config RuntimeConfig) (<-chan *model.AgentEvent, *model.ChatMessage, []map[string]interface{}, error) {
	logger.Info(fmt.Sprintf("=== SendMessageStreamWithConfig START: session=%s, model=%s, temp=%.2f ===", sessionID, config.Model, config.Temperature))

	eventChan := make(chan *model.AgentEvent, 100)
	var ragReferences []map[string]interface{}

	go func() {
		defer close(eventChan)

		// 1. 获取会话
		session, err := s.repo.GetSessionByID(sessionID)
		if err != nil {
			eventChan <- model.NewErrorEvent("ChatService", fmt.Sprintf("获取会话失败: %v", err), 500)
			return
		}

		// 2. 获取历史消息
		history, err := s.repo.GetRecentMessages(sessionID, s.maxCtx)
		if err != nil {
			logger.Error(fmt.Sprintf("获取历史消息失败: %v", err))
		}

		// 3. RAG知识检索（根据配置）
		var knowledgeContext string

		if config.EnableRAG && s.ragSvc != nil {
			searchResults, err := s.ragSvc.SearchKnowledgeWithConfig(ctx, content, config.RAGTopK, config.RAGThreshold)
			if err == nil && len(searchResults) > 0 {
				logger.Info(fmt.Sprintf("RAG检索成功，找到%d个文档 (TopK=%d, Threshold=%.2f)", len(searchResults), config.RAGTopK, config.RAGThreshold))

				for _, result := range searchResults {
					ragReferences = append(ragReferences, map[string]interface{}{
						"id":       result.Document.ID,
						"title":    result.Document.Title,
						"doc_type": result.Document.DocType,
						"score":    result.Score,
					})
				}

				knowledgeContext, _ = s.ragSvc.GetContextForQuery(ctx, content, 1000)

				if len(ragReferences) > 0 {
					eventChan <- model.NewRagReferencesEvent(ragReferences)
				}
			}
		}

		// 4. 创建用户消息
		userMessage := &model.ChatMessage{
			SessionID: sessionID,
			Role:      "user",
			Content:   content,
		}
		if err := s.repo.CreateMessage(userMessage); err != nil {
			eventChan <- model.NewErrorEvent("ChatService", fmt.Sprintf("保存用户消息失败: %v", err), 500)
			return
		}

		// 发送用户消息事件
		eventChan <- model.NewUserMessageEvent(sessionID, "user", content)

		// 5. Agent路由选择
		sessionContext := s.buildSessionContext(history, knowledgeContext)
		agentInstance, routingLog, err := s.masterRouter.Route(ctx, content, sessionContext)
		if err != nil {
			eventChan <- model.NewErrorEvent("ChatService", fmt.Sprintf("路由失败: %v", err), 500)
			return
		}

		logger.Info(fmt.Sprintf("✅ 选中Agent: %s (置信度 %.2f)", routingLog.SelectedAgentID, routingLog.Confidence))

		// 保存路由日志
		if err := s.routingLogRepo.Create(routingLog); err != nil {
			logger.Error(fmt.Sprintf("保存路由日志失败: %v", err))
		}

		// 6. 应用运行时配置到Agent实例
		if err := agentInstance.ApplyRuntimeConfig(ctx, config, session.Model); err != nil {
			logger.Error(fmt.Sprintf("应用运行时配置失败: %v", err))
			// 配置失败不影响继续执行，使用默认配置
		}

		// 7. Agent执行（全流程流式）
		agentEventChan, toolCallRecords, err := agentInstance.ExecuteStream(ctx, content, history)
		if err != nil {
			eventChan <- model.NewErrorEvent("ChatService", fmt.Sprintf("Agent执行失败: %v", err), 500)
			return
		}

		// 8. 转发Agent事件
		var fullContent strings.Builder
		for event := range agentEventChan {
			if event.Type == model.EventContentChunk {
				if data, ok := event.Data.(model.ContentChunkEventData); ok {
					fullContent.WriteString(data.Content)
				}
			}
			eventChan <- event
		}

		// 9. 保存工具调用日志
		for _, record := range toolCallRecords {
			callLog := &model.ToolCallLog{
				ID:        uuid.New().String(),
				SessionID: sessionID,
				ToolName:  record.ToolName,
				Arguments: toJSON(record.Arguments),
				Result:    record.Result,
				Success:   record.Success,
				Duration:  record.Duration,
			}
			if !record.Success {
				callLog.ErrorMessage = record.Result
			}
			if err := s.toolCallLogRepo.Create(callLog); err != nil {
				logger.Error(fmt.Sprintf("保存工具调用日志失败: %v", err))
			}
		}

		// 10. 保存AI消息
		aiMessage := &model.ChatMessage{
			SessionID:     sessionID,
			Role:          "assistant",
			Content:       fullContent.String(),
			AgentID:       routingLog.SelectedAgentID,
			RAGReferences: toJSON(ragReferences),
		}
		if err := s.repo.CreateMessage(aiMessage); err != nil {
			logger.Error(fmt.Sprintf("保存AI消息失败: %v", err))
		} else {
			logger.Info(fmt.Sprintf("AI消息已保存: ID=%s, ContentLen=%d", aiMessage.ID, fullContent.Len()))
		}

		logger.Info(fmt.Sprintf("=== SendMessageStreamWithConfig完成 ==="))
	}()

	return eventChan, nil, ragReferences, nil
}

func (s *ChatService) buildSessionContext(history []model.ChatMessage, knowledgeContext string) string {
	var contextBuilder strings.Builder

	if len(history) > 0 {
		contextBuilder.WriteString("## 历史对话：\n\n")
		for _, msg := range history {
			if msg.Role == "user" {
				contextBuilder.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
			} else {
				contextBuilder.WriteString(fmt.Sprintf("助手: %s\n", msg.Content))
			}
		}
		contextBuilder.WriteString("\n")
	}

	if knowledgeContext != "" {
		contextBuilder.WriteString("## 相关知识：\n\n")
		contextBuilder.WriteString(knowledgeContext)
		contextBuilder.WriteString("\n")
	}

	return contextBuilder.String()
}

func (s *ChatService) generateSessionTitle(content string) string {
	if len(content) > 30 {
		return content[:30] + "..."
	}
	return content
}

func (s *ChatService) SaveAIMessage(sessionID, content string, ragReferences []map[string]interface{}) (*model.ChatMessage, error) {
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

	return aiMessage, nil
}

func (s *ChatService) GetSessionHistory(sessionID string) (*model.ChatSessionWithMessages, error) {
	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return nil, err
	}

	messages, err := s.repo.GetMessagesBySessionID(sessionID)
	if err != nil {
		return nil, err
	}

	return &model.ChatSessionWithMessages{
		ChatSession: *session,
		Messages:    messages,
	}, nil
}

func (s *ChatService) GetUserSessions(userID string, limit int) ([]model.ChatSession, error) {
	sessions, err := s.repo.GetSessionsByUserID(userID, limit)
	if err != nil {
		logger.Error(fmt.Sprintf("获取用户会话列表失败: %v", err))
		return nil, err
	}
	return sessions, nil
}

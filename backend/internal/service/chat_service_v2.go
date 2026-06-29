package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/cloudwego/eino/adk"
)

func (s *ChatService) SendMessageV2(ctx context.Context, sessionID, content string) (string, *model.ChatMessage, *model.ChatMessage, []map[string]interface{}, error) {
	logger.Info(fmt.Sprintf("=== SendMessageV2 (Eino Agent): session=%s ===", sessionID))

	selectedAgent, err := s.agentRouter.RouteAgent(ctx, content)
	if err != nil {
		logger.Error(fmt.Sprintf("Agent路由失败: %v", err))
		selectedAgent = nil
	} else if selectedAgent != nil {
		logger.Info(fmt.Sprintf("智能路由选择 Agent: %s (%s)", selectedAgent.Name, selectedAgent.ID))
	}

	var ragReferences []map[string]interface{}

	if s.enableRAG && s.ragSvc != nil {
		searchResults, err := s.ragSvc.SearchKnowledge(ctx, content, 3)
		if err == nil && len(searchResults) > 0 {
			logger.Info(fmt.Sprintf("RAG检索成功，找到%d个相关文档", len(searchResults)))
			for _, result := range searchResults {
				ragReferences = append(ragReferences, map[string]interface{}{
					"id":    result.Document.ID,
					"title": result.Document.Title,
					"score": result.Score,
				})
			}
		}
	}

	userMessage := &model.ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   content,
	}
	if err := s.repo.CreateMessage(userMessage); err != nil {
		return "", nil, nil, nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	var agent adk.Agent
	var agentID string

	if selectedAgent != nil {
		agentID = selectedAgent.ID
		agent, err = s.agentBuilder.BuildAgentFromModel(ctx, selectedAgent)
		if err != nil {
			logger.Error(fmt.Sprintf("构建Agent失败: %v，使用默认Agent", err))
			agent, err = s.agentBuilder.BuildDefaultAgent(ctx)
			if err != nil {
				return "", nil, nil, nil, fmt.Errorf("构建默认Agent失败: %w", err)
			}
			agentID = "default"
		}
	} else {
		agentID = "default"
		agent, err = s.agentBuilder.BuildDefaultAgent(ctx)
		if err != nil {
			return "", nil, nil, nil, fmt.Errorf("构建默认Agent失败: %w", err)
		}
	}

	logger.Info(fmt.Sprintf("使用Agent: %s", agentID))

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
	})

	prompt := content
	if len(ragReferences) > 0 {
		knowledgeContext, err := s.ragSvc.GetContextForQuery(ctx, content, 1000)
		if err == nil && knowledgeContext != "" {
			prompt = knowledgeContext + "\n\n用户问题: " + content
		}
	}

	iter := runner.Query(ctx, prompt)

	var fullResponse string
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, err := event.Output.MessageOutput.GetMessage()
			if err == nil && msg != nil {
				fullResponse += msg.Content
				logger.Debug(fmt.Sprintf("收到消息: 长度=%d", len(msg.Content)))
			}
		}

		if event.Err != nil {
			logger.Error(fmt.Sprintf("Agent执行错误: %v", event.Err))
		}
	}

	logger.Info(fmt.Sprintf("Agent执行完成，响应长度: %d", len(fullResponse)))

	aiMessage := &model.ChatMessage{
		SessionID: sessionID,
		Role:      "assistant",
		Content:   fullResponse,
	}
	if err := s.repo.CreateMessage(aiMessage); err != nil {
		logger.Error(fmt.Sprintf("保存AI消息失败: %v", err))
	}

	return fullResponse, userMessage, aiMessage, ragReferences, nil
}

func (s *ChatService) StreamSendMessageV2(ctx context.Context, sessionID, content string) (<-chan string, *model.ChatMessage, []map[string]interface{}, error) {
	logger.Info(fmt.Sprintf("=== StreamSendMessageV2 (Eino Agent): session=%s ===", sessionID))

	outputChan := make(chan string, 100)

	selectedAgent, err := s.agentRouter.RouteAgent(ctx, content)
	if err != nil {
		logger.Error(fmt.Sprintf("Agent路由失败: %v", err))
		selectedAgent = nil
	} else if selectedAgent != nil {
		logger.Info(fmt.Sprintf("智能路由选择 Agent: %s (%s)", selectedAgent.Name, selectedAgent.ID))
	}

	var ragReferences []map[string]interface{}

	if s.enableRAG && s.ragSvc != nil {
		searchResults, err := s.ragSvc.SearchKnowledge(ctx, content, 3)
		if err == nil && len(searchResults) > 0 {
			logger.Info(fmt.Sprintf("RAG检索成功，找到%d个相关文档", len(searchResults)))
			for _, result := range searchResults {
				ragReferences = append(ragReferences, map[string]interface{}{
					"id":    result.Document.ID,
					"title": result.Document.Title,
					"score": result.Score,
				})
			}
		}
	}

	userMessage := &model.ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   content,
	}
	if err := s.repo.CreateMessage(userMessage); err != nil {
		close(outputChan)
		return nil, nil, nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	go func() {
		defer close(outputChan)

		var agent adk.Agent
		var agentID string

		if selectedAgent != nil {
			agentID = selectedAgent.ID
			agent, err = s.agentBuilder.BuildAgentFromModel(ctx, selectedAgent)
			if err != nil {
				logger.Error(fmt.Sprintf("构建Agent失败: %v，使用默认Agent", err))
				agent, _ = s.agentBuilder.BuildDefaultAgent(ctx)
				agentID = "default"
			}
		} else {
			agentID = "default"
			agent, _ = s.agentBuilder.BuildDefaultAgent(ctx)
		}

		logger.Info(fmt.Sprintf("使用Agent: %s (启用流式)", agentID))

		runner := adk.NewRunner(ctx, adk.RunnerConfig{
			Agent:           agent,
			EnableStreaming: true,
		})

		prompt := content
		if len(ragReferences) > 0 {
			knowledgeContext, err := s.ragSvc.GetContextForQuery(ctx, content, 1000)
			if err == nil && knowledgeContext != "" {
				prompt = knowledgeContext + "\n\n用户问题: " + content
			}
		}

		iter := runner.Query(ctx, prompt)

		var fullResponse strings.Builder
		for {
			event, ok := iter.Next()
			if !ok {
				break
			}

			if event.Output != nil && event.Output.MessageOutput != nil {
				if event.Output.MessageOutput.IsStreaming {
					stream := event.Output.MessageOutput.MessageStream
					if stream != nil {
						for {
							chunk, err := stream.Recv()
							if err != nil {
								break
							}
							if chunk.Content != "" {
								outputChan <- chunk.Content
								fullResponse.WriteString(chunk.Content)
							}
						}
					}
				} else {
					msg, err := event.Output.MessageOutput.GetMessage()
					if err == nil && msg != nil && msg.Content != "" {
						outputChan <- msg.Content
						fullResponse.WriteString(msg.Content)
					}
				}
			}

			if event.Err != nil {
				logger.Error(fmt.Sprintf("Agent执行错误: %v", event.Err))
				outputChan <- fmt.Sprintf("\n[错误: %v]", event.Err)
			}
		}

		logger.Info(fmt.Sprintf("Agent流式执行完成，总长度: %d", fullResponse.Len()))

		aiMessage := &model.ChatMessage{
			SessionID: sessionID,
			Role:      "assistant",
			Content:   fullResponse.String(),
		}
		if err := s.repo.CreateMessage(aiMessage); err != nil {
			logger.Error(fmt.Sprintf("保存AI消息失败: %v", err))
		}
	}()

	return outputChan, userMessage, ragReferences, nil
}

func (s *ChatService) SendMessage(ctx context.Context, sessionID, content string) (string, *model.ChatMessage, *model.ChatMessage, []map[string]interface{}, error) {
	return s.SendMessageV2(ctx, sessionID, content)
}

func (s *ChatService) StreamSendMessage(ctx context.Context, sessionID, content string) (<-chan string, *model.ChatMessage, []map[string]interface{}, error) {
	return s.StreamSendMessageV2(ctx, sessionID, content)
}

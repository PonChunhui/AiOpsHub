package main

import (
	"fmt"
	"strings"

	"github.com/aiops/AiOpsHub/backend/internal/model"
)

const maxHistoryTokens = 4000

func truncateHistoryByTokens(messages []model.ChatMessage, maxTokens int) []model.ChatMessage {
	if len(messages) == 0 {
		return messages
	}

	var result []model.ChatMessage
	totalTokens := 0

	for i := len(messages) - 1; i >= 0; i-- {
		estimatedTokens := len(messages[i].Content) / 2

		if totalTokens+estimatedTokens > maxTokens {
			fmt.Printf("[历史截断] 达到token限制(%d)，保留%d条消息\n", maxTokens, len(result))
			break
		}

		result = append([]model.ChatMessage{messages[i]}, result...)
		totalTokens += estimatedTokens
	}

	return result
}

func buildContextWithHistory(sessionID, currentPrompt string, allMessages []model.ChatMessage) string {
	if len(allMessages) == 0 {
		fmt.Printf("[历史上下文] 会话%s无历史消息\n", sessionID)
		return currentPrompt
	}

	historyMessages := truncateHistoryByTokens(allMessages, maxHistoryTokens)

	var contextBuilder strings.Builder

	contextBuilder.WriteString("以下是我们的对话历史：\n\n")
	for _, msg := range historyMessages {
		if msg.Role == "user" {
			contextBuilder.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
		} else {
			contextBuilder.WriteString(fmt.Sprintf("助手: %s\n", msg.Content))
		}
	}
	contextBuilder.WriteString("\n---\n\n")
	contextBuilder.WriteString(fmt.Sprintf("当前问题: %s", currentPrompt))

	prompt := contextBuilder.String()

	fmt.Printf("[历史上下文] 会话%s: 包含%d条历史消息，构建后prompt长度%d字符\n", sessionID, len(historyMessages), len(prompt))

	return prompt
}

func main() {
	fmt.Println("=== 测试1: 空历史消息 ===")
	sessionID := "test-session-1"
	currentPrompt := "你好"
	emptyMessages := []model.ChatMessage{}
	result := buildContextWithHistory(sessionID, currentPrompt, emptyMessages)
	fmt.Printf("结果: %s\n\n", result)

	fmt.Println("=== 测试2: 少量消息不超过限制 ===")
	sessionID = "test-session-2"
	currentPrompt = "请帮我分析这个问题"
	messages := []model.ChatMessage{
		{ID: "msg1", Role: "user", Content: "你好，我是张三"},
		{ID: "msg2", Role: "assistant", Content: "你好张三，有什么可以帮助您的？"},
		{ID: "msg3", Role: "user", Content: "我想了解系统监控"},
		{ID: "msg4", Role: "assistant", Content: "好的，我可以帮您了解系统监控的功能"},
	}
	result = buildContextWithHistory(sessionID, currentPrompt, messages)
	fmt.Printf("结果:\n%s\n\n", result)

	fmt.Println("=== 测试3: Token截断测试 ===")
	sessionID = "test-session-3"
	currentPrompt = "继续分析"
	longMessages := []model.ChatMessage{
		{ID: "msg1", Role: "user", Content: "这是第一条很长的消息，用来测试token截断功能"},
		{ID: "msg2", Role: "assistant", Content: "这是第一条很长的回复消息"},
		{ID: "msg3", Role: "user", Content: "这是第二条很长的消息"},
		{ID: "msg4", Role: "assistant", Content: "这是第二条很长的回复消息"},
		{ID: "msg5", Role: "user", Content: "这是第三条很长的消息"},
		{ID: "msg6", Role: "assistant", Content: "这是第三条很长的回复消息"},
		{ID: "msg7", Role: "user", Content: "这是第四条很长的消息"},
		{ID: "msg8", Role: "assistant", Content: "这是第四条很长的回复消息"},
		{ID: "msg9", Role: "user", Content: "这是第五条很长的消息"},
		{ID: "msg10", Role: "assistant", Content: "这是第五条很长的回复消息"},
	}

	truncated := truncateHistoryByTokens(longMessages, 100)
	fmt.Printf("截断前: %d条消息，截断后: %d条消息\n", len(longMessages), len(truncated))

	if len(truncated) > 0 {
		fmt.Printf("保留的最新消息ID: %s\n", truncated[len(truncated)-1].ID)
	}
	fmt.Println()

	fmt.Println("=== 测试4: 验证消息顺序 ===")
	sessionID = "test-session-4"
	orderedMessages := []model.ChatMessage{
		{ID: "1", Role: "user", Content: "消息1"},
		{ID: "2", Role: "assistant", Content: "回复1"},
		{ID: "3", Role: "user", Content: "消息2"},
		{ID: "4", Role: "assistant", Content: "回复2"},
	}
	result = buildContextWithHistory(sessionID, "测试顺序", orderedMessages)
	fmt.Printf("结果:\n%s\n\n", result)

	fmt.Println("✅ 所有测试完成")
}

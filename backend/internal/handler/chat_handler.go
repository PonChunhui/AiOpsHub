package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/service"
	"github.com/aiops/AiOpsHub/backend/pkg/llm"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// ChatHandler 对话处理器
// 处理AI对话相关的HTTP请求
type ChatHandler struct {
	chatService *service.ChatService
}

// NewChatHandler 创建对话处理器实例
func NewChatHandler(ragSvc *service.RAGService, mcpSvc *service.MCPService, agentSvc *service.AgentService, tokenSvc *service.TokenService) (*ChatHandler, error) {
	llmConfig := llm.EinoLLMConfig{
		Model:       viper.GetString("llm.model"),
		Temperature: viper.GetFloat64("llm.temperature"),
		MaxTokens:   viper.GetInt("llm.max_tokens"),
		Provider:    viper.GetString("llm.provider"),
		APIKey:      viper.GetString("llm.api_key"),
		BaseURL:     viper.GetString("llm.base_url"),
	}

	enableRAG := viper.GetBool("llm.enable_rag")

	if llmConfig.Model == "" {
		llmConfig.Model = "gpt-3.5-turbo"
	}
	if llmConfig.Provider == "" {
		llmConfig.Provider = "openai"
	}
	if llmConfig.Temperature == 0 {
		llmConfig.Temperature = 0.7
	}

	var ragServiceToUse *service.RAGService
	if enableRAG && ragSvc != nil {
		ragServiceToUse = ragSvc
		logger.Info("ChatHandler已启用RAG功能")
	} else {
		ragServiceToUse = nil
		logger.Info("ChatHandler未启用RAG功能")
	}

	chatService, err := service.NewChatService(llmConfig, ragServiceToUse, mcpSvc, agentSvc, tokenSvc)
	if err != nil {
		return nil, err
	}

	return &ChatHandler{
		chatService: chatService,
	}, nil
}

// CreateSessionRequest 创建会话请求结构
type CreateSessionRequest struct {
	Title string `json:"title" binding:"required"` // 会话标题
	Model string `json:"model"`                    // 使用的模型（可选）
}

// SendMessageRequest 发送消息请求结构
type SendMessageRequest struct {
	SessionID      string `json:"session_id" binding:"required"` // 会话ID
	Content        string `json:"content" binding:"required"`    // 消息内容
	EnableThinking bool   `json:"enable_thinking"`               // 是否启用思考过程显示（可选，默认false）
}

// CreateSession 创建新的对话会话
// POST /api/v1/chat/sessions
func (h *ChatHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 从上下文获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 设置默认模型
	model := req.Model
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	// 创建会话
	session, err := h.chatService.CreateSession(userID, req.Title, model)
	if err != nil {
		logger.Error("创建会话失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "会话创建成功",
		"data":    session,
	})
}

// SendMessage 发送消息并获取AI回复（非流式）
// POST /api/v1/chat/messages
func (h *ChatHandler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	aiResponse, userMsg, aiMsg, ragReferences, err := h.chatService.SendMessage(c.Request.Context(), req.SessionID, req.Content)
	if err != nil {
		logger.Error("发送消息失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "消息发送成功",
		"ai_response":    aiResponse,
		"user_message":   userMsg,
		"ai_message":     aiMsg,
		"rag_references": ragReferences,
	})
}

// SendMessageStream 发送消息并流式获取AI回复（SSE）
// POST /api/v1/chat/messages/stream
func (h *ChatHandler) SendMessageStream(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	streamChan, userMsg, ragReferences, err := h.chatService.StreamSendMessage(c.Request.Context(), req.SessionID, req.Content)
	if err != nil {
		logger.Error("流式发送消息失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Debug(fmt.Sprintf("SendMessageStream: ragReferences数量=%d", len(ragReferences)))
	for i, ref := range ragReferences {
		logger.Debug(fmt.Sprintf("RAG引用[%d]: title=%s, score=%.2f", i, ref["title"], ref["score"]))
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "不支持流式输出"})
		return
	}

	sendSSE(c, flusher, "user_message", gin.H{
		"id":      userMsg.ID,
		"role":    userMsg.Role,
		"content": userMsg.Content,
	})

	// 总是发送rag_references事件，即使为空（便于前端调试）
	sendSSE(c, flusher, "rag_references", ragReferences)
	logger.Debug(fmt.Sprintf("已发送rag_references事件，数量=%d", len(ragReferences)))

	fullContent := ""
	for chunk := range streamChan {
		fullContent += chunk
		sendSSE(c, flusher, "chunk", gin.H{"content": chunk})
	}

	aiMsg, err := h.chatService.SaveAIMessage(req.SessionID, fullContent, ragReferences)
	if err != nil {
		sendSSE(c, flusher, "error", gin.H{"message": "保存AI消息失败"})
	} else {
		sendSSE(c, flusher, "ai_message", gin.H{
			"id":      aiMsg.ID,
			"role":    aiMsg.Role,
			"content": fullContent,
		})
	}

	sendSSE(c, flusher, "done", gin.H{"message": "流式输出完成"})

	// SSE 注释行，用于关闭连接的额外信号
	c.Writer.WriteString(": connection closed\n\n")
	flusher.Flush()

	logger.Info("SSE stream completed, connection will close")
}

func (h *ChatHandler) SendMessageStreamWithEvents(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 调用Service层，传递enable_thinking参数
	eventChan, _, ragReferences, err := h.chatService.StreamSendMessageWithEvents(
		c.Request.Context(),
		req.SessionID,
		req.Content,
		req.EnableThinking, // 传递深度思考开关
	)
	if err != nil {
		logger.Error(fmt.Sprintf("流式发送消息失败: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "不支持流式输出"})
		return
	}

	// 初始化工具调用缓冲区（用于合并流式分片）
	toolCallsBuffer := model.NewToolCallsBuffer()
	fullContent := ""

	// 转换并发送OpenAI格式的流式响应
	for event := range eventChan {
		// 将AgentEvent转换为OpenAI ChatCompletionChunk格式
		openaiChunk, err := model.ConvertAgentEventToOpenAIChunk(event, toolCallsBuffer)
		if err != nil {
			logger.Error(fmt.Sprintf("转换OpenAI格式失败: %v", err))
			continue
		}

		// 只发送转换成功的事件（nil表示该事件不需要发送给前端）
		if openaiChunk != nil {
			// 发送OpenAI标准SSE格式
			c.Writer.WriteString(openaiChunk.ToSSE())
			flusher.Flush()

			// 同时保存完整内容（用于数据库存储）
			if event.Type == model.EventContentChunk {
				if data, ok := event.Data.(model.ContentChunkEventData); ok {
					fullContent += data.Content
				}
			}
		}
	}

	// 保存AI消息到数据库（包括RAG引用）
	aiMsg, err := h.chatService.SaveAIMessage(req.SessionID, fullContent, ragReferences)
	if err != nil {
		logger.Error(fmt.Sprintf("保存AI消息失败: %v", err))
	} else {
		logger.Info(fmt.Sprintf("AI消息已保存: ID=%s, ContentLen=%d", aiMsg.ID, len(fullContent)))
	}

	// 发送OpenAI标准的[DONE]标记（表示流结束）
	c.Writer.WriteString("data: [DONE]\n\n")
	flusher.Flush()

	logger.Info("OpenAI格式流式输出完成，连接将关闭")
}

func sendSSE(c *gin.Context, flusher http.Flusher, event string, data interface{}) {
	// 将数据转换为紧凑的JSON字符串（不含换行符）
	jsonData := toJson(data)

	// 构建SSE协议格式：event行 + data行 + 空行结束
	eventLine := fmt.Sprintf("event: %s\n", event)
	dataLine := fmt.Sprintf("data: %s\n\n", jsonData)

	// 写入事件类型行，失败则记录错误并返回
	if _, err := c.Writer.WriteString(eventLine); err != nil {
		logger.Error(fmt.Sprintf("Failed to write SSE event: %v", err))
		return
	}

	// 写入数据行，失败则记录错误并返回
	if _, err := c.Writer.WriteString(dataLine); err != nil {
		logger.Error(fmt.Sprintf("Failed to write SSE data: %v", err))
		return
	}

	// 立即刷新缓冲区，确保数据发送到客户端
	flusher.Flush()
	logger.Debug(fmt.Sprintf("SSE sent: event=%s, data_len=%d", event, len(jsonData)))
}

// toJson 将数据对象转换为紧凑的JSON字符串
// 参数：
//   - data: 要转换的数据对象
//
// 返回：紧凑的JSON字符串（不含换行符，符合SSE协议要求）
// 说明：使用json.Marshal确保JSON紧凑输出，字符串中的换行符会被转义为\n
func toJson(data interface{}) string {
	// 使用json.Marshal生成紧凑JSON，避免真实换行符破坏SSE格式
	result, err := json.Marshal(data)
	if err != nil {
		logger.Error(fmt.Sprintf("JSON marshal error: %v", err))
		return "{}"
	}
	return string(result)
}

// GetSessionHistory 获取会话历史记录
// GET /api/v1/chat/sessions/:id/history
func (h *ChatHandler) GetSessionHistory(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	// 获取会话历史
	history, err := h.chatService.GetSessionHistory(sessionID)
	if err != nil {
		logger.Error("获取会话历史失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取会话历史失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取成功",
		"data":    history,
	})
}

// GetUserSessions 获取用户的所有会话列表
// GET /api/v1/chat/sessions
func (h *ChatHandler) GetUserSessions(c *gin.Context) {
	// 从上下文获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取会话列表
	sessions, err := h.chatService.GetUserSessions(userID, 0)
	if err != nil {
		logger.Error("获取会话列表失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取会话列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取成功",
		"data":    sessions,
	})
}

// DeleteSession 删除会话
// DELETE /api/v1/chat/sessions/:id
func (h *ChatHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	// 删除会话
	err := h.chatService.DeleteSession(sessionID)
	if err != nil {
		logger.Error("删除会话失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// 全局ChatHandler实例
var GlobalChatHandler *ChatHandler

// InitChatHandler 初始化对话处理器
func InitChatHandler() {
	handler, err := NewChatHandler(ragService, mcpService, agentService, tokenService)
	if err != nil {
		logger.Error("初始化ChatHandler失败: " + err.Error())
		return
	}
	GlobalChatHandler = handler
	logger.Info("ChatHandler初始化成功(已启用RAG、MCP和Token统计功能)")
}

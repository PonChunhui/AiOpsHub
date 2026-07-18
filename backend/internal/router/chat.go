package router

import (
	"github.com/aiops/AiOpsHub/backend/internal/handler"
	"github.com/aiops/AiOpsHub/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// ChatRouter 聊天路由
type ChatRouter struct{}

func (r *ChatRouter) Register(engine *gin.Engine) {
	v1 := engine.Group("/api/v1")
	chat := v1.Group("/chat")
	chat.Use(middleware.Auth())
	{
		chat.POST("/sessions", func(c *gin.Context) {
			if handler.GlobalChatHandler != nil {
				handler.GlobalChatHandler.CreateSession(c)
			} else {
				c.JSON(500, gin.H{"error": "Chat handler not initialized"})
			}
		})
		chat.GET("/sessions", func(c *gin.Context) {
			if handler.GlobalChatHandler != nil {
				handler.GlobalChatHandler.GetUserSessions(c)
			} else {
				c.JSON(500, gin.H{"error": "Chat handler not initialized"})
			}
		})
		chat.GET("/sessions/:id/history", func(c *gin.Context) {
			if handler.GlobalChatHandler != nil {
				handler.GlobalChatHandler.GetSessionHistory(c)
			} else {
				c.JSON(500, gin.H{"error": "Chat handler not initialized"})
			}
		})
		chat.DELETE("/sessions/:id", func(c *gin.Context) {
			if handler.GlobalChatHandler != nil {
				handler.GlobalChatHandler.DeleteSession(c)
			} else {
				c.JSON(500, gin.H{"error": "Chat handler not initialized"})
			}
		})
		chat.POST("/messages", func(c *gin.Context) {
			if handler.GlobalChatHandler != nil {
				handler.GlobalChatHandler.SendMessage(c)
			} else {
				c.JSON(500, gin.H{"error": "Chat handler not initialized"})
			}
		})
		chat.POST("/messages/stream", func(c *gin.Context) {
			if handler.GlobalChatHandler != nil {
				handler.GlobalChatHandler.SendMessageStream(c)
			} else {
				c.JSON(500, gin.H{"error": "Chat handler not initialized"})
			}
		})
		chat.POST("/messages/stream/events", func(c *gin.Context) {
			if handler.GlobalChatHandler != nil {
				handler.GlobalChatHandler.SendMessageStreamWithEvents(c)
			} else {
				c.JSON(500, gin.H{"error": "Chat handler not initialized"})
			}
		})
	}
}

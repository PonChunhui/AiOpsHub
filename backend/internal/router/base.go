package router

import (
	"github.com/aiops/AiOpsHub/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

// BaseRouter 基础路由（健康检查、WebSocket）
type BaseRouter struct{}

func (r *BaseRouter) Register(engine *gin.Engine) {
	engine.GET("/health", handler.HealthHandler)
	engine.GET("/healthz", handler.LivenessHandler)
	engine.GET("/ready", handler.ReadinessHandler)

	engine.GET("/ws", func(c *gin.Context) {
		if handler.GlobalWebSocketHandler != nil {
			handler.GlobalWebSocketHandler.HandleWebSocket(c)
		} else {
			c.JSON(500, gin.H{"error": "WebSocket handler not initialized"})
		}
	})

	engine.GET("/ws/ssh/:host_id", handler.HandleSSHWebSocket)
}

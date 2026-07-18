package router

import (
	"github.com/aiops/AiOpsHub/backend/internal/handler"
	"github.com/aiops/AiOpsHub/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// TokensRouter Token统计路由
type TokensRouter struct{}

func (r *TokensRouter) Register(engine *gin.Engine) {
	v1 := engine.Group("/api/v1")
	tokens := v1.Group("/tokens")
	tokens.Use(middleware.Auth())
	{
		tokens.GET("/stats", handler.GetTokenUsageStats)
		tokens.GET("/cost", handler.GetTokenCost)
		tokens.GET("/session/:id", handler.GetSessionTokens)
		tokens.POST("/estimate", handler.EstimateCost)
	}
}

package router

import (
	"github.com/aiops/AiOpsHub/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

// AuthRouter 认证路由
type AuthRouter struct{}

func (r *AuthRouter) Register(engine *gin.Engine) {
	v1 := engine.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/login", handler.Login)
		auth.POST("/logout", handler.Logout)
		auth.POST("/register", handler.Register)
	}
}

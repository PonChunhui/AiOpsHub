package router

import (
	"github.com/aiops/AiOpsHub/backend/internal/handler"
	"github.com/aiops/AiOpsHub/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// UsersRouter 用户路由
type UsersRouter struct{}

func (r *UsersRouter) Register(engine *gin.Engine) {
	v1 := engine.Group("/api/v1")
	users := v1.Group("/users")
	users.Use(middleware.Auth())
	{
		users.GET("", handler.ListUsers)
		users.GET("/:id", handler.GetUser)
		users.PUT("/:id", handler.UpdateUser)
		users.DELETE("/:id", handler.DeleteUser)
	}
}

package router

import (
	"github.com/aiops/AiOpsHub/backend/internal/handler"
	"github.com/aiops/AiOpsHub/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// ToolsRouter 工具路由
type ToolsRouter struct{}

func (r *ToolsRouter) Register(engine *gin.Engine) {
	v1 := engine.Group("/api/v1")
	tools := v1.Group("/tools")
	tools.Use(middleware.Auth())
	{
		tools.GET("", handler.ListTools)
		tools.GET("/:id", handler.GetTool)
		tools.POST("", handler.CreateTool)
		tools.PUT("/:id", handler.UpdateTool)
		tools.DELETE("/:id", handler.DeleteTool)
		tools.POST("/:id/execute", handler.ExecuteTool)
		tools.POST("/init-presets", handler.InitPresets)
	}
}

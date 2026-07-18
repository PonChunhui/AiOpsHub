package router

import (
	"github.com/aiops/AiOpsHub/backend/internal/handler"
	"github.com/aiops/AiOpsHub/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// AgentsRouter Agent路由
type AgentsRouter struct{}

func (r *AgentsRouter) Register(engine *gin.Engine) {
	v1 := engine.Group("/api/v1")

	// Agent管理路由
	agents := v1.Group("/agents")
	agents.Use(middleware.Auth())
	{
		agents.GET("", handler.ListAgents)
		agents.GET("/:id", handler.GetAgent)
		agents.POST("", handler.CreateAgent)
		agents.PUT("/:id", handler.UpdateAgent)
		agents.DELETE("/:id", handler.DeleteAgent)
	}

	// Agent工具关联路由
	agentTools := v1.Group("/agents/:id/tools")
	agentTools.Use(middleware.Auth())
	{
		agentTools.GET("", handler.GetAgentTools)
		agentTools.POST("/:tool_id", handler.BindToolToAgent)
		agentTools.DELETE("/:tool_id", handler.UnbindToolFromAgent)
		agentTools.PUT("/:tool_id/config", handler.UpdateAgentToolConfig)
		agentTools.POST("/:tool_id/toggle", handler.ToggleAgentToolEnabled)
	}
}

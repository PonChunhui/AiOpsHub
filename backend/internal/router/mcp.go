package router

import (
	"github.com/aiops/AiOpsHub/backend/internal/handler"
	"github.com/aiops/AiOpsHub/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// MCPRouter MCP服务路由
type MCPRouter struct{}

func (r *MCPRouter) Register(engine *gin.Engine) {
	v1 := engine.Group("/api/v1")
	mcpGroup := v1.Group("/mcp")
	mcpGroup.Use(middleware.Auth())
	{
		mcpGroup.GET("/servers", handler.ListMCPServers)
		mcpGroup.GET("/servers/:id", handler.GetMCPServer)
		mcpGroup.POST("/servers", handler.CreateMCPServer)
		mcpGroup.PUT("/servers/:id", handler.UpdateMCPServer)
		mcpGroup.DELETE("/servers/:id", handler.DeleteMCPServer)
		mcpGroup.POST("/servers/:id/test", handler.TestMCPServer)
		mcpGroup.GET("/servers/:id/tools", handler.GetMCPServerTools)
	}
}

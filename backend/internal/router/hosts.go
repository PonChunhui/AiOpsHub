package router

import (
	"github.com/aiops/AiOpsHub/backend/internal/handler"
	"github.com/aiops/AiOpsHub/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// HostsRouter 主机路由（包含主机组和主机）
type HostsRouter struct{}

func (r *HostsRouter) Register(engine *gin.Engine) {
	v1 := engine.Group("/api/v1")

	// 主机组路由
	hostGroups := v1.Group("/host-groups")
	hostGroups.Use(middleware.Auth())
	{
		hostGroups.GET("", handler.GetGroupTree)
		hostGroups.GET("/:id", handler.GetGroupByID)
		hostGroups.POST("", handler.CreateGroup)
		hostGroups.PUT("/:id", handler.UpdateGroup)
		hostGroups.DELETE("/:id", handler.DeleteGroup)
		hostGroups.GET("/:id/check-cascade", handler.CheckGroupCascade)
	}

	// 主机路由
	hosts := v1.Group("/hosts")
	hosts.Use(middleware.Auth())
	{
		hosts.GET("", handler.ListHosts)
		hosts.GET("/:id", handler.GetHostByID)
		hosts.POST("", handler.CreateHost)
		hosts.PUT("/:id", handler.UpdateHost)
		hosts.DELETE("/:id", handler.DeleteHost)
		hosts.POST("/batch-import", handler.BatchImportHosts)
		hosts.POST("/batch-delete", handler.BatchDeleteHosts)
		hosts.POST("/:id/test-connection", handler.TestHostConnection)
		hosts.GET("/:id/files", handler.ListFiles)
		hosts.GET("/:id/files/info", handler.GetFileInfo)
		hosts.POST("/:id/files/upload", handler.UploadFile)
		hosts.GET("/:id/files/download", handler.DownloadFile)
	}
}

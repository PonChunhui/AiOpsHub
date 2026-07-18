package router

import (
	"github.com/gin-gonic/gin"
)

// ModuleRouter 定义路由模块接口
type ModuleRouter interface {
	Register(r *gin.Engine)
}

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine) {
	modules := []ModuleRouter{
		&BaseRouter{},
		&AuthRouter{},
		&AlertsRouter{},
		&AgentsRouter{},
		&ToolsRouter{},
		&RAGRouter{},
		&MCPRouter{},
		&TokensRouter{},
		&UsersRouter{},
		&HostsRouter{},
		&ChatRouter{},
	}

	for _, module := range modules {
		module.Register(r)
	}
}

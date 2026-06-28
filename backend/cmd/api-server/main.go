package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/config"
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/handler"
	"github.com/aiops/AiOpsHub/backend/internal/middleware"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	redisutil "github.com/aiops/AiOpsHub/backend/pkg/redis"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 初始化日志
	logger.Init()

	// 初始化数据库
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化Redis
	if err := redisutil.Init(); err != nil {
		log.Printf("Warning: Failed to initialize Redis: %v (token validation will fallback to JWT only)", err)
	} else {
		logger.Info("Redis connected successfully")
	}

	// Auto migrate new models
	if err := database.DB.AutoMigrate(&model.AlertAnalysisResult{}); err != nil {
		log.Printf("Warning: Failed to migrate alert_analysis_results table: %v", err)
	}
	if err := database.DB.AutoMigrate(&model.Agent{}); err != nil {
		log.Printf("Warning: Failed to migrate agents table: %v", err)
	}
	if err := database.DB.AutoMigrate(&model.MCPServer{}); err != nil {
		log.Printf("Warning: Failed to migrate mcp_servers table: %v", err)
	}

	// 初始化WebSocket Handler
	handler.InitWebSocketHandler()

	// 初始化服务(包括RAGService)
	handler.InitServices()
	handler.InitInfraServices()

	// 初始化Chat Handler(需要在InitServices之后,因为依赖RAGService)
	handler.InitChatHandler()

	// 设置Gin模式
	if viper.GetString("app.mode") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.New()

	// 中间件
	r.Use(
		gin.Logger(),
		gin.Recovery(),
		middleware.CORS(),
		middleware.RequestID(),
		middleware.Logger(),
		middleware.ErrorHandler(),
	)

	// 注册路由
	registerRoutes(r)

	// 404处理
	r.NoRoute(middleware.NotFoundHandler())

	// 405处理
	r.NoMethod(middleware.MethodNotAllowedHandler())

	// 启动HTTP服务器
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.Info(fmt.Sprintf("API Server started on port %s", port))

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down API Server...")

	database.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("Server forced to shutdown: %v", err))
	}

	logger.Info("API Server exited")
}

func registerRoutes(r *gin.Engine) {
	// 基础健康检查
	r.GET("/health", handler.HealthHandler)
	r.GET("/healthz", handler.LivenessHandler)
	r.GET("/ready", handler.ReadinessHandler)
	r.GET("/metrics", handler.MetricsHandler)
	r.GET("/prometheus", handler.PrometheusMetricsHandler)

	// 测试端点
	r.GET("/test/db", handler.TestDB)
	r.POST("/test/user", handler.TestCreateUser)

	// WebSocket端点
	r.GET("/ws", func(c *gin.Context) {
		if handler.GlobalWebSocketHandler != nil {
			handler.GlobalWebSocketHandler.HandleWebSocket(c)
		} else {
			c.JSON(500, gin.H{"error": "WebSocket handler not initialized"})
		}
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// 用户认证
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handler.Login)
			auth.POST("/logout", handler.Logout)
			auth.POST("/register", handler.Register)
		}

		// Agent管理（需要认证）
		agents := v1.Group("/agents")
		agents.Use(middleware.Auth())
		{
			agents.GET("", handler.ListAgents)
			agents.GET("/presets", handler.ListPresetAgents)
			agents.GET("/enabled", handler.ListEnabledAgents)
			agents.GET("/:id", handler.GetAgent)
			agents.POST("", handler.CreateAgent)
			agents.POST("/:id/execute", handler.ExecuteAgent)
			agents.POST("/:id/toggle", handler.ToggleAgentEnabled)
			agents.PUT("/:id", handler.UpdateAgent)
			agents.DELETE("/:id", handler.DeleteAgent)
		}

		// 数据源管理
		datasources := v1.Group("/datasources")
		datasources.Use(middleware.Auth())
		{
			datasources.GET("", handler.ListDatasources)
			datasources.GET("/:id", handler.GetDatasource)
			datasources.POST("", handler.CreateDatasource)
			datasources.PUT("/:id", handler.UpdateDatasource)
			datasources.DELETE("/:id", handler.DeleteDatasource)
			datasources.POST("/:id/test", handler.TestDatasource)
		}

		// 告警管理
		alerts := v1.Group("/alerts")
		alerts.Use(middleware.Auth())
		{
			alerts.GET("", handler.ListAlerts)
			alerts.GET("/:id", handler.GetAlert)
			alerts.POST("", handler.CreateAlert)
			alerts.POST("/webhook", handler.AlertWebhook) // 告警Webhook接收
			alerts.GET("/rules", handler.ListAlertRules)
			alerts.POST("/rules", handler.CreateAlertRule)
			alerts.PUT("/rules/:id", handler.UpdateAlertRule)
			alerts.DELETE("/rules/:id", handler.DeleteAlertRule)

			// 告警分析结果（避免路由冲突）
			alerts.GET("/analysis/:alert_id", handler.GetAlertAnalysis)
			alerts.POST("/analysis", handler.SaveAlertAnalysis)
			alerts.GET("/analysis/list", handler.ListAlertAnalysis)
			alerts.DELETE("/analysis/:id", handler.DeleteAlertAnalysis)
		}

		// 工具管理
		tools := v1.Group("/tools")
		tools.Use(middleware.Auth())
		{
			tools.GET("", handler.ListTools)
			tools.GET("/:id", handler.GetTool)
			tools.POST("/:id/execute", handler.ExecuteTool)
		}

		// 工具管理
		monitor := v1.Group("/monitor")
		monitor.Use(middleware.Auth())
		{
			monitor.GET("/stats", handler.GetStats)
			monitor.GET("/tokens", handler.GetTokenUsageStats)
			monitor.GET("/cost", handler.GetTokenCost)
			monitor.GET("/performance", handler.GetPerformance)
		}

		// RAG知识库
		rag := v1.Group("/rag")
		rag.Use(middleware.Auth())
		{
			rag.POST("/search", handler.SearchRAGKnowledge)
			rag.GET("/context", handler.GetRAGContext)
			rag.GET("/documents", handler.ListRAGDocuments)
			rag.GET("/documents/:id", handler.GetRAGDocument)
			rag.POST("/documents", handler.AddRAGDocument)
			rag.PUT("/documents/:id", handler.UpdateRAGDocument)
			rag.DELETE("/documents/:id", handler.DeleteRAGDocument)
		}

		// MCP Server管理
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

		// Prometheus监控
		prom := v1.Group("/prometheus")
		prom.Use(middleware.Auth())
		{
			prom.GET("/query", handler.QueryPrometheus)
			prom.GET("/service/:service", handler.GetServiceMetricsHandler)
			prom.GET("/top", handler.GetTopServicesHandler)
			prom.GET("/alerts", handler.GetActiveAlerts)
		}

		// Token统计
		tokens := v1.Group("/tokens")
		tokens.Use(middleware.Auth())
		{
			tokens.GET("/stats", handler.GetTokenUsageStats)
			tokens.GET("/cost", handler.GetTokenCost)
			tokens.GET("/session/:id", handler.GetSessionTokens)
			tokens.POST("/estimate", handler.EstimateCost)
		}

		// Kubernetes
		k8s := v1.Group("/k8s")
		k8s.Use(middleware.Auth())
		{
			k8s.GET("/pods", handler.ListPods)
			k8s.GET("/pods/:namespace/:name", handler.GetPod)
			k8s.GET("/pods/:namespace/:name/logs", handler.GetPodLogs)
			k8s.GET("/deployments", handler.ListDeployments)
			k8s.GET("/deployments/:namespace/:name", handler.GetDeployment)
			k8s.POST("/deployments/:namespace/:name/scale", handler.ScaleDeployment)
			k8s.POST("/deployments/:namespace/:name/restart", handler.RestartDeployment)
			k8s.GET("/usage", handler.GetK8sResourceUsage)
		}

		// 日志查询
		logs := v1.Group("/logs")
		logs.Use(middleware.Auth())
		{
			logs.POST("/query", handler.QueryLogs)
			logs.GET("/stats", handler.GetLogStats)
			logs.GET("/service/:service", handler.GetServiceLogsHandler)
			logs.GET("/errors", handler.GetErrorLogs)
			logs.POST("/search", handler.SearchLogs)
			logs.GET("/recent", handler.GetRecentLogs)
			logs.GET("/export", handler.ExportLogs)
		}

		// 自动修复
		remediation := v1.Group("/remediation")
		remediation.Use(middleware.Auth())
		{
			remediation.POST("/plans", handler.CreateRemediationPlan)
			remediation.POST("/plans/:plan_id/execute", handler.ExecuteRemediationPlan)
			remediation.GET("/plans/:plan_id", handler.GetRemediationPlan)
			remediation.GET("/plans", handler.ListRemediationPlans)
			remediation.POST("/plans/:plan_id/cancel", handler.CancelRemediationPlan)
			remediation.POST("/actions/:action_id/approve", handler.ApproveRemediationAction)
			remediation.GET("/stats", handler.GetRemediationStats)
		}

		// 用户管理
		users := v1.Group("/users")
		users.Use(middleware.Auth(), middleware.AdminOnly())
		{
			users.GET("", handler.ListUsers)
			users.GET("/:id", handler.GetUser)
			users.PUT("/:id", handler.UpdateUser)
			users.DELETE("/:id", handler.DeleteUser)
		}

		// AI助手对话
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
		}
	}
}

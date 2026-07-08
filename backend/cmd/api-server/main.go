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
	"github.com/aiops/AiOpsHub/backend/internal/service"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	redisutil "github.com/aiops/AiOpsHub/backend/pkg/redis"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	logger.Init()

	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := redisutil.Init(); err != nil {
		log.Printf("Warning: Failed to initialize Redis: %v (token validation will fallback to JWT only)", err)
	} else {
		logger.Info("Redis connected successfully")
	}

	if err := database.DB.AutoMigrate(&model.AlertAnalysisResult{}); err != nil {
		log.Printf("Warning: Failed to migrate alert_analysis_results table: %v", err)
	}
	if err := database.DB.AutoMigrate(&model.Agent{}); err != nil {
		log.Printf("Warning: Failed to migrate agents table: %v", err)
	}
	if err := database.DB.AutoMigrate(&model.MCPServer{}); err != nil {
		log.Printf("Warning: Failed to migrate mcp_servers table: %v", err)
	}
	if err := database.DB.AutoMigrate(&model.TokenUsageRecord{}); err != nil {
		log.Printf("Warning: Failed to migrate token_usage_records table: %v", err)
	}
	if err := database.DB.AutoMigrate(&model.ChatSession{}); err != nil {
		log.Printf("Warning: Failed to migrate chat_sessions table: %v", err)
	}
	if err := database.DB.AutoMigrate(&model.ChatMessage{}); err != nil {
		log.Printf("Warning: Failed to migrate chat_messages table: %v", err)
	}
	if err := database.DB.AutoMigrate(&model.HostGroup{}); err != nil {
		log.Printf("Warning: Failed to migrate host_groups table: %v", err)
	}
	if err := database.DB.AutoMigrate(&model.Host{}); err != nil {
		log.Printf("Warning: Failed to migrate hosts table: %v", err)
	}
	if err := database.DB.AutoMigrate(&model.SSHSessionLog{}); err != nil {
		log.Printf("Warning: Failed to migrate ssh_session_logs table: %v", err)
	}
	if err := database.DB.AutoMigrate(&model.RoutingLog{}); err != nil {
		log.Printf("Warning: Failed to migrate routing_logs table: %v", err)
	}
	if err := database.DB.AutoMigrate(&model.ToolCallLog{}); err != nil {
		log.Printf("Warning: Failed to migrate tool_call_logs table: %v", err)
	}

	handler.InitWebSocketHandler()
	handler.InitServices()
	handler.InitChatHandler()
	handler.InitHostHandler()

	// 创建默认主机分组（如果不存在）
	hostService := service.NewHostService()
	groups, _ := hostService.GetGroupTree()
	if len(groups) == 0 {
		hostService.CreateGroup("默认分组", "", "系统默认分组", "system")
		logger.Info("Created default host group")
	}

	if viper.GetString("app.mode") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(
		gin.Logger(),
		gin.Recovery(),
		middleware.CORS(),
		middleware.RequestID(),
		middleware.Logger(),
		middleware.ErrorHandler(),
	)

	registerRoutes(r)

	r.NoRoute(middleware.NotFoundHandler())
	r.NoMethod(middleware.MethodNotAllowedHandler())

	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.Info(fmt.Sprintf("API Server started on port %s", port))

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
	r.GET("/health", handler.HealthHandler)
	r.GET("/healthz", handler.LivenessHandler)
	r.GET("/ready", handler.ReadinessHandler)

	r.GET("/ws", func(c *gin.Context) {
		if handler.GlobalWebSocketHandler != nil {
			handler.GlobalWebSocketHandler.HandleWebSocket(c)
		} else {
			c.JSON(500, gin.H{"error": "WebSocket handler not initialized"})
		}
	})

	r.GET("/ws/ssh/:host_id", handler.HandleSSHWebSocket)

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handler.Login)
			auth.POST("/logout", handler.Logout)
			auth.POST("/register", handler.Register)
		}

		alerts := v1.Group("/alerts")
		alerts.Use(middleware.Auth())
		{
			alerts.GET("", handler.ListAlerts)
			alerts.GET("/:id", handler.GetAlert)
			alerts.POST("", handler.CreateAlert)
			alerts.GET("/analysis/:alert_id", handler.GetAlertAnalysis)
			alerts.POST("/analysis", handler.SaveAlertAnalysis)
			alerts.GET("/analysis/list", handler.ListAlertAnalysis)
			alerts.DELETE("/analysis/:id", handler.DeleteAlertAnalysis)
		}

		agents := v1.Group("/agents")
		agents.Use(middleware.Auth())
		{
			agents.GET("", handler.ListAgents)
			agents.GET("/:id", handler.GetAgent)
			agents.POST("", handler.CreateAgent)
			agents.PUT("/:id", handler.UpdateAgent)
			agents.DELETE("/:id", handler.DeleteAgent)
		}

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

		agentTools := v1.Group("/agents/:id/tools")
		agentTools.Use(middleware.Auth())
		{
			agentTools.GET("", handler.GetAgentTools)
			agentTools.POST("/:tool_id", handler.BindToolToAgent)
			agentTools.DELETE("/:tool_id", handler.UnbindToolFromAgent)
			agentTools.PUT("/:tool_id/config", handler.UpdateAgentToolConfig)
			agentTools.POST("/:tool_id/toggle", handler.ToggleAgentToolEnabled)
		}

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

		tokens := v1.Group("/tokens")
		tokens.Use(middleware.Auth())
		{
			tokens.GET("/stats", handler.GetTokenUsageStats)
			tokens.GET("/cost", handler.GetTokenCost)
			tokens.GET("/session/:id", handler.GetSessionTokens)
			tokens.POST("/estimate", handler.EstimateCost)
		}

		users := v1.Group("/users")
		users.Use(middleware.Auth())
		{
			users.GET("", handler.ListUsers)
			users.GET("/:id", handler.GetUser)
			users.PUT("/:id", handler.UpdateUser)
			users.DELETE("/:id", handler.DeleteUser)
		}

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
		}

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
			chat.POST("/messages/stream/events", func(c *gin.Context) {
				if handler.GlobalChatHandler != nil {
					handler.GlobalChatHandler.SendMessageStreamWithEvents(c)
				} else {
					c.JSON(500, gin.H{"error": "Chat handler not initialized"})
				}
			})
		}
	}
}

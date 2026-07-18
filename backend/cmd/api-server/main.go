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
	"github.com/aiops/AiOpsHub/backend/internal/router"
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

	router.RegisterRoutes(r)

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

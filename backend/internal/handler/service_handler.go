package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var (
	ragService        *service.RAGService
	mcpService        *service.MCPService
	prometheusService *service.PrometheusService
	tokenService      *service.TokenService
	milvusService     *service.MilvusService
	embeddingService  *service.EmbeddingService
	agentService      *service.AgentService
)

func InitServices() {
	initRAGService()
	initMCPService()
	prometheusService = service.NewPrometheusService(viper.GetString("prometheus.url"))
	tokenService = service.NewTokenService()

	agentService = service.NewAgentService()
	if err := agentService.InitializePresets(); err != nil {
		fmt.Printf("Failed to initialize preset agents: %v\n", err)
	}
}

func initMCPService() {
	mcpService = service.NewMCPService()
	fmt.Println("MCP Service initialized")
}

func initRAGService() {
	milvusEnabled := viper.IsSet("milvus.host")

	if milvusEnabled {
		milvusHost := viper.GetString("milvus.host")
		milvusPort := viper.GetString("milvus.port")

		milvusSvc, err := service.NewMilvusService(milvusHost, milvusPort, "default")
		if err != nil {
			fmt.Printf("Failed to connect to Milvus: %v, falling back to memory mode\n", err)
			ragService = service.NewRAGService("aiops_knowledge")
			return
		}

		err = milvusSvc.CreateCollection(context.Background())
		if err != nil {
			fmt.Printf("Failed to create Milvus collection: %v\n", err)
		}

		err = milvusSvc.LoadCollection(context.Background())
		if err != nil {
			fmt.Printf("Failed to load Milvus collection: %v\n", err)
		}

		milvusService = milvusSvc

		embeddingProvider := viper.GetString("embedding.provider")
		embeddingModel := viper.GetString("embedding.model")
		embeddingAPIKey := viper.GetString("embedding.api_key")
		embeddingBaseURL := viper.GetString("embedding.base_url")

		embeddingSvc := service.NewEmbeddingService(embeddingProvider, embeddingModel, embeddingAPIKey, embeddingBaseURL)
		embeddingService = embeddingSvc

		ragService = service.NewRAGServiceWithMilvus("aiops_knowledge", milvusSvc, embeddingSvc)
		fmt.Println("RAG Service initialized with Milvus backend")
	} else {
		ragService = service.NewRAGService("aiops_knowledge")
		fmt.Println("RAG Service initialized with memory backend")
	}
}

// ==================== RAG知识库 ====================

func SearchRAGKnowledge(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
		TopK  int    `json:"top_k"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TopK == 0 {
		req.TopK = 5
	}

	results, err := ragService.SearchKnowledge(c.Request.Context(), req.Query, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   req.Query,
		"results": results,
		"count":   len(results),
	})
}

func GetRAGContext(c *gin.Context) {
	query := c.Query("query")
	maxTokens := c.Query("max_tokens")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}

	tokens := 500
	if maxTokens != "" {
		tokens, _ = strconv.Atoi(maxTokens)
	}

	context, err := ragService.GetContextForQuery(c.Request.Context(), query, tokens)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"context": context,
		"length":  len(context),
	})
}

func ListRAGDocuments(c *gin.Context) {
	category := c.Query("category")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	docs, total, err := ragService.ListDocuments(c.Request.Context(), category, search, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      200,
		"documents": docs,
		"total":     total,
		"page":      page,
		"pageSize":  pageSize,
	})
}

func GetRAGDocument(c *gin.Context) {
	id := c.Param("id")

	doc, err := ragService.GetDocument(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"document": doc,
	})
}

func AddRAGDocument(c *gin.Context) {
	var req struct {
		Title    string   `json:"title" binding:"required"`
		Content  string   `json:"content" binding:"required"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		username = "unknown"
	}

	now := time.Now().Format(time.RFC3339)

	doc := service.KnowledgeDocument{
		ID:       fmt.Sprintf("kb-%d", time.Now().Unix()),
		Title:    req.Title,
		Content:  req.Content,
		Category: req.Category,
		Tags:     req.Tags,
		Metadata: map[string]interface{}{
			"created_at": now,
			"updated_at": now,
			"created_by": username,
			"updated_by": username,
		},
	}

	err := ragService.AddDocument(c.Request.Context(), doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "文档添加成功",
		"document": doc,
	})
}

func UpdateRAGDocument(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Title    string   `json:"title"`
		Content  string   `json:"content"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		username = "unknown"
	}

	oldDoc, err := ragService.GetDocument(c.Request.Context(), id)
	metadata := map[string]interface{}{}
	if err == nil && oldDoc.Metadata != nil {
		if createdAt, ok := oldDoc.Metadata["created_at"]; ok {
			metadata["created_at"] = createdAt
		}
		if createdBy, ok := oldDoc.Metadata["created_by"]; ok {
			metadata["created_by"] = createdBy
		}
	}

	metadata["updated_at"] = time.Now().Format(time.RFC3339)
	metadata["updated_by"] = username

	doc, err := ragService.UpdateDocument(c.Request.Context(), id, req.Title, req.Content, req.Category, req.Tags, metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "文档更新成功",
		"document": doc,
	})
}

func DeleteRAGDocument(c *gin.Context) {
	id := c.Param("id")

	err := ragService.DeleteDocument(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "文档删除成功",
	})
}

// ==================== Prometheus监控 ====================

func QueryPrometheus(c *gin.Context) {
	query := c.Query("query")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}

	metrics, err := prometheusService.Query(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"metrics": metrics,
		"count":   len(metrics),
	})
}

func GetServiceMetricsHandler(c *gin.Context) {
	serviceName := c.Param("service")

	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name required"})
		return
	}

	metrics, err := prometheusService.GetServiceMetrics(c.Request.Context(), serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func GetTopServicesHandler(c *gin.Context) {
	metricName := c.Query("metric")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	services, err := prometheusService.GetTopServices(c.Request.Context(), metricName, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metric":   metricName,
		"services": services,
		"count":    len(services),
	})
}

func GetActiveAlerts(c *gin.Context) {
	alerts, err := prometheusService.GetAlerts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// ==================== Token统计 ====================

func GetTokenUsageStats(c *gin.Context) {
	stats, err := tokenService.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func GetTokenCost(c *gin.Context) {
	breakdown, err := tokenService.GetCostBreakdown(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, breakdown)
}

func GetSessionTokens(c *gin.Context) {
	sessionID := c.Param("session_id")

	usage, err := tokenService.GetSessionUsage(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"usage":      usage,
		"count":      len(usage),
	})
}

func EstimateCost(c *gin.Context) {
	var req struct {
		Model           string `json:"model" binding:"required"`
		EstimatedTokens int    `json:"estimated_tokens" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cost := tokenService.EstimateCost(req.Model, req.EstimatedTokens)

	c.JSON(http.StatusOK, gin.H{
		"model":            req.Model,
		"estimated_tokens": req.EstimatedTokens,
		"estimated_cost":   cost,
	})
}

package router

import (
	"github.com/aiops/AiOpsHub/backend/internal/handler"
	"github.com/aiops/AiOpsHub/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RAGRouter RAG知识检索路由
type RAGRouter struct{}

func (r *RAGRouter) Register(engine *gin.Engine) {
	v1 := engine.Group("/api/v1")
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
}

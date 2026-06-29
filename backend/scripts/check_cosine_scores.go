package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aiops/AiOpsHub/backend/internal/service"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/spf13/viper"
)

func main() {
	logger.Init()
	viper.SetConfigFile("../configs/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	milvusHost := viper.GetString("milvus.host")
	milvusPort := viper.GetString("milvus.port")

	fmt.Println("=========================================")
	fmt.Println("  检查COSINE相似度评分范围")
	fmt.Println("=========================================")

	milvusSvc, err := service.NewMilvusService(milvusHost, milvusPort, "default")
	if err != nil {
		log.Fatalf("Failed to connect to Milvus: %v", err)
	}
	defer milvusSvc.Close()

	embeddingProvider := viper.GetString("embedding.provider")
	embeddingModel := viper.GetString("embedding.model")
	embeddingAPIKey := viper.GetString("embedding.api_key")
	embeddingBaseURL := viper.GetString("embedding.base_url")

	embeddingSvc := service.NewEmbeddingService(embeddingProvider, embeddingModel, embeddingAPIKey, embeddingBaseURL)

	// 直接搜索（绕过阈值过滤）
	ctx := context.Background()
	query := "Docker镜像管理"

	fmt.Printf("\n查询: %s\n", query)
	fmt.Println("生成查询向量...")

	queryEmbedding, err := embeddingSvc.GetEmbedding(ctx, query)
	if err != nil {
		log.Fatalf("Failed to get embedding: %v", err)
	}

	fmt.Printf("向量维度: %d\n", len(queryEmbedding))
	fmt.Println("开始检索（不应用阈值过滤）...")

	results, err := milvusSvc.SearchDocuments(ctx, queryEmbedding, 3)
	if err != nil {
		log.Fatalf("Failed to search: %v", err)
	}

	fmt.Printf("\n找到 %d 个文档（未过滤）\n", len(results))
	fmt.Println("-----------------------------------------")

	for i, result := range results {
		fmt.Printf("\n结果 %d:\n", i+1)
		fmt.Printf("  标题: %s\n", result.Document.Title)
		fmt.Printf("  COSINE相似度: %.4f\n", result.Score)
		fmt.Printf("  距离: %.4f\n", result.Distance)
		fmt.Printf("  分类: %s\n", result.Document.Category)

		// 判断等级
		var level string
		if result.Score >= 0.90 {
			level = "high"
		} else if result.Score >= 0.80 {
			level = "medium"
		} else if result.Score >= 0.70 {
			level = "low"
		} else {
			level = "none"
		}
		fmt.Printf("  等级: %s\n", level)
	}

	fmt.Println()
	fmt.Println("=========================================")
	fmt.Println("阈值设置建议:")
	fmt.Println("  高相关度: >= 0.90")
	fmt.Println("  中等相关度: >= 0.80")
	fmt.Println("  边缘相关度: >= 0.70")
	fmt.Println("=========================================")
}

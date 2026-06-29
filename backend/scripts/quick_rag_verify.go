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
	fmt.Println("  RAG检索功能快速验证")
	fmt.Println("=========================================")
	fmt.Println()

	// 连接Milvus
	milvusSvc, err := service.NewMilvusService(milvusHost, milvusPort, "default")
	if err != nil {
		log.Fatalf("Failed to connect to Milvus: %v", err)
	}
	defer milvusSvc.Close()

	// 加载collection
	err = milvusSvc.LoadCollection(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to load collection: %v", err)
	}

	// 创建Embedding服务
	embeddingProvider := viper.GetString("embedding.provider")
	embeddingModel := viper.GetString("embedding.model")
	embeddingAPIKey := viper.GetString("embedding.api_key")
	embeddingBaseURL := viper.GetString("embedding.base_url")

	embeddingSvc := service.NewEmbeddingService(embeddingProvider, embeddingModel, embeddingAPIKey, embeddingBaseURL)

	// 创建RAG服务
	ragSvc := service.NewRAGServiceWithMilvus("aiops_knowledge", milvusSvc, embeddingSvc)

	// 测试检索
	testQueries := []string{
		"Docker镜像管理",
		"Rancher地址",
		"Kubernetes Pod",
		"MySQL安装",
		"Harbor配置",
	}

	ctx := context.Background()
	for i, query := range testQueries {
		fmt.Printf("\n【测试%d】查询: %s\n", i+1, query)
		fmt.Println("-----------------------------------------")

		results, err := ragSvc.SearchKnowledge(ctx, query, 2)
		if err != nil {
			fmt.Printf("✗ 检索失败: %v\n", err)
			continue
		}

		if len(results) == 0 {
			fmt.Println("✗ 未找到相关文档（所有结果被阈值过滤）")
			continue
		}

		fmt.Printf("✓ 找到 %d 个相关文档\n", len(results))
		for j, result := range results {
			fmt.Printf("  结果%d:\n", j+1)
			fmt.Printf("    标题: %s\n", result.Document.Title)
			fmt.Printf("    评分: %.4f\n", result.Score)
			fmt.Printf("    距离: %.4f\n", result.Distance)
			fmt.Printf("    等级: %s\n", result.RelevanceLevel)
			fmt.Printf("    分类: %s\n", result.Document.Category)
		}
	}

	fmt.Println()
	fmt.Println("=========================================")
	fmt.Println("  ✓ RAG检索验证完成")
	fmt.Println("=========================================")
	fmt.Println()
	fmt.Println("分级阈值标准:")
	fmt.Println("  高相关度: >= 0.75")
	fmt.Println("  中等相关度: >= 0.65")
	fmt.Println("  边缘相关度: >= 0.50")
	fmt.Println("  不相关: < 0.50")
	fmt.Println()
}

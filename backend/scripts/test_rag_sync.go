package main

import (
	"context"
	"log"

	"github.com/aiops/AiOpsHub/backend/internal/config"
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/service"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/spf13/viper"
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to init config: %v", err)
	}

	logger.Init()

	if err := database.Init(); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	milvusHost := viper.GetString("milvus.host")
	milvusPort := viper.GetString("milvus.port")

	if milvusHost == "" || milvusPort == "" {
		log.Fatal("Milvus configuration not found")
	}

	milvusSvc, err := service.NewMilvusService(milvusHost, milvusPort, "default")
	if err != nil {
		log.Fatalf("Failed to connect to Milvus: %v", err)
	}
	defer milvusSvc.Close()

	err = milvusSvc.CreateCollection(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to create collection: %v", err)
	}

	err = milvusSvc.LoadCollection(context.Background())
	if err != nil {
		log.Fatalf("Failed to load collection: %v", err)
	}

	embeddingProvider := viper.GetString("embedding.provider")
	embeddingModel := viper.GetString("embedding.model")
	embeddingAPIKey := viper.GetString("embedding.api_key")
	embeddingBaseURL := viper.GetString("embedding.base_url")

	embeddingSvc := service.NewEmbeddingService(embeddingProvider, embeddingModel, embeddingAPIKey, embeddingBaseURL)

	ragSvc := service.NewRAGServiceWithMilvus("aiops_knowledge", milvusSvc, embeddingSvc)

	// 只同步前5条文档进行测试
	testDocs := []service.KnowledgeDocument{
		{
			ID:       "kb-30",
			Title:    "Docker常用命令",
			Content:  "Docker基础操作：镜像管理、容器管理、网络管理、卷管理",
			Category: "docker",
			Tags:     []string{"docker", "容器"},
			Metadata: map[string]interface{}{
				"created_by": "system",
			},
		},
		{
			ID:       "kb-5",
			Title:    "MySQL数据库安装部署",
			Content:  "MySQL安装配置、CentOS安装、Docker部署、参数配置",
			Category: "database",
			Tags:     []string{"mysql", "数据库"},
			Metadata: map[string]interface{}{
				"created_by": "system",
			},
		},
		{
			ID:       "kb-4",
			Title:    "Kubernetes核心概念Pod",
			Content:  "Pod定义、生命周期、控制器、配置要点、管理命令",
			Category: "kubernetes",
			Tags:     []string{"k8s", "pod"},
			Metadata: map[string]interface{}{
				"created_by": "system",
			},
		},
		{
			ID:       "kb-rancher",
			Title:    "Rancher地址配置",
			Content:  "Rancher平台访问地址配置、环境管理、集群接入",
			Category: "rancher",
			Tags:     []string{"rancher", "平台"},
			Metadata: map[string]interface{}{
				"created_by": "system",
			},
		},
		{
			ID:       "kb-harbor",
			Title:    "Harbor镜像仓库",
			Content:  "Harbor镜像仓库配置、HTTPS证书、镜像推送拉取",
			Category: "harbor",
			Tags:     []string{"harbor", "镜像仓库"},
			Metadata: map[string]interface{}{
				"created_by": "system",
			},
		},
	}

	ctx := context.Background()
	for _, doc := range testDocs {
		err := ragSvc.AddDocument(ctx, doc)
		if err != nil {
			log.Printf("Failed to add document %s: %v", doc.ID, err)
		} else {
			log.Printf("Successfully added document: %s - %s", doc.ID, doc.Title)
		}
	}

	log.Println("测试数据同步完成！共5条文档")

	// 立即测试检索
	results, err := ragSvc.SearchKnowledge(ctx, "Docker镜像管理", 3)
	if err != nil {
		log.Printf("Search failed: %v", err)
	} else {
		log.Printf("Search test: Found %d results", len(results))
		for _, result := range results {
			log.Printf("  - Title: %s, Score: %.4f, Level: %s", result.Document.Title, result.Score, result.RelevanceLevel)
		}
	}
}

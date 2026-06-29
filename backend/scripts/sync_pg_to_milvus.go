package main

import (
	"context"
	"log"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/config"
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
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

	ctx := context.Background()

	if err := milvusSvc.CreateCollection(ctx); err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}

	if err := milvusSvc.LoadCollection(ctx); err != nil {
		log.Fatalf("Failed to load collection: %v", err)
	}

	embeddingProvider := viper.GetString("embedding.provider")
	embeddingModel := viper.GetString("embedding.model")
	embeddingAPIKey := viper.GetString("embedding.api_key")
	embeddingBaseURL := viper.GetString("embedding.base_url")

	embeddingSvc := service.NewEmbeddingService(embeddingProvider, embeddingModel, embeddingAPIKey, embeddingBaseURL)

	var docs []model.RAGDocument
	if err := database.DB.Order("id").Find(&docs).Error; err != nil {
		log.Fatalf("Failed to fetch documents from PostgreSQL: %v", err)
	}

	log.Printf("Found %d documents in PostgreSQL", len(docs))

	successCount := 0
	failCount := 0
	batchSize := 10

	for i := 0; i < len(docs); i += batchSize {
		end := i + batchSize
		if end > len(docs) {
			end = len(docs)
		}

		batch := docs[i:end]

		for _, doc := range batch {
			embedding, err := embeddingSvc.GetEmbedding(ctx, doc.Content)
			if err != nil {
				log.Printf("Failed to get embedding for %s: %v", doc.ID, err)
				failCount++
				continue
			}

			knowledgeDoc := service.KnowledgeDocument{
				ID:        doc.ID,
				Title:     doc.Title,
				Content:   doc.Content,
				DocType:   doc.DocType,
				Component: doc.Component,
				Tags:      []string{},
				Metadata: map[string]interface{}{
					"created_at": doc.CreatedAt.Format(time.RFC3339),
					"updated_at": doc.UpdatedAt.Format(time.RFC3339),
					"created_by": doc.CreatedBy,
					"updated_by": doc.UpdatedBy,
				},
			}

			if doc.Tags != "" {
				tags := []string{}
				for _, tag := range []string{doc.Tags} {
					if tag != "" {
						tags = append(tags, tag)
					}
				}
				knowledgeDoc.Tags = tags
			}

			err = milvusSvc.InsertDocument(ctx, knowledgeDoc, embedding)
			if err != nil {
				log.Printf("Failed to insert document %s to Milvus: %v", doc.ID, err)
				failCount++
			} else {
				successCount++
				if successCount%50 == 0 {
					log.Printf("Progress: %d/%d documents synced", successCount, len(docs))
				}
			}

			time.Sleep(100 * time.Millisecond)
		}

		log.Printf("Processed batch %d-%d", i+1, end)
	}

	log.Printf("Sync completed: %d success, %d failed", successCount, failCount)
}

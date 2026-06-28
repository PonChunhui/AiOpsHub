package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/config"
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/service"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func init() {
	logger.Init()
}

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to init config: %v", err)
	}

	log.Println("Logger initializing...")
	logger.Init()

	if err := database.Init(); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	milvusHost := viper.GetString("milvus.host")
	milvusPort := viper.GetString("milvus.port")

	if milvusHost == "" {
		log.Fatal("Milvus host not configured")
	}

	milvusSvc, err := service.NewMilvusService(milvusHost, milvusPort, "default")
	if err != nil {
		log.Fatalf("Failed to connect to Milvus: %v", err)
	}

	err = milvusSvc.LoadCollection(context.Background())
	if err != nil {
		log.Fatalf("Failed to load collection: %v", err)
	}

	log.Println("Fetching documents from Milvus...")
	docs, err := milvusSvc.ListDocuments(context.Background(), 10000)
	if err != nil {
		log.Fatalf("Failed to list documents: %v", err)
	}

	log.Printf("Found %d documents in Milvus", len(docs))

	synced := 0
	skipped := 0

	for _, doc := range docs {
		var existing model.RAGDocument
		err := database.DB.Where("id = ?", doc.ID).First(&existing).Error

		if err == nil {
			log.Printf("Document %s already exists in PostgreSQL, skipping", doc.ID)
			skipped++
			continue
		}

		if err != gorm.ErrRecordNotFound {
			log.Printf("Error checking document %s: %v", doc.ID, err)
			continue
		}

		tagsJSON, _ := json.Marshal(doc.Tags)

		var createdAt, updatedAt time.Time
		var createdBy, updatedBy string

		if doc.Metadata != nil {
			if v, ok := doc.Metadata["created_at"]; ok {
				if str, ok := v.(string); ok {
					t, err := time.Parse(time.RFC3339, str)
					if err == nil {
						createdAt = t
					}
				}
			}
			if v, ok := doc.Metadata["updated_at"]; ok {
				if str, ok := v.(string); ok {
					t, err := time.Parse(time.RFC3339, str)
					if err == nil {
						updatedAt = t
					}
				}
			}
			if v, ok := doc.Metadata["created_by"]; ok {
				createdBy = fmt.Sprintf("%v", v)
			}
			if v, ok := doc.Metadata["updated_by"]; ok {
				updatedBy = fmt.Sprintf("%v", v)
			}
		}

		if createdAt.IsZero() {
			createdAt = time.Now()
		}
		if updatedAt.IsZero() {
			updatedAt = createdAt
		}

		pgDoc := &model.RAGDocument{
			ID:        doc.ID,
			Title:     doc.Title,
			Content:   doc.Content,
			Category:  doc.Category,
			Tags:      string(tagsJSON),
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}

		if err := database.DB.Create(pgDoc).Error; err != nil {
			log.Printf("Failed to insert document %s: %v", doc.ID, err)
			continue
		}

		log.Printf("Synced document: %s - %s", doc.ID, doc.Title)
		synced++
	}

	log.Printf("Sync completed: %d synced, %d skipped, %d total", synced, skipped, len(docs))
}

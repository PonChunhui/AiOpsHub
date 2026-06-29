package main

import (
	"fmt"
	"log"

	"github.com/aiops/AiOpsHub/backend/internal/config"
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to init config: %v", err)
	}

	logger.Init()

	if err := database.Init(); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	var total int64
	database.DB.Model(&model.RAGDocument{}).Count(&total)

	var categories []struct {
		Category string
		Count    int
	}

	database.DB.Model(&model.RAGDocument{}).
		Select("category, count(*) as count").
		Group("category").
		Order("count DESC").
		Find(&categories)

	fmt.Printf("\n=== Database Statistics ===\n")
	fmt.Printf("Total documents: %d\n\n", total)
	fmt.Printf("Documents by category:\n")
	fmt.Printf("%-20s %s\n", "Category", "Count")
	fmt.Printf("%-20s %s\n", "--------", "-----")
	for _, cat := range categories {
		fmt.Printf("%-20s %d\n", cat.Category, cat.Count)
	}

	var recentDocs []model.RAGDocument
	database.DB.Order("created_at DESC").Limit(5).Find(&recentDocs)

	fmt.Printf("\nRecent documents:\n")
	for i, doc := range recentDocs {
		fmt.Printf("%d. %s [%s]\n", i+1, doc.Title, doc.Category)
	}
}

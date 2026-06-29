package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	docsPath := os.Getenv("DOCS_PATH")
	if docsPath == "" {
		docsPath = "/Users/pengchunhui/Documents/document"
	}

	successCount := 0
	failCount := 0
	skipCount := 0

	err := filepath.WalkDir(docsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Failed to read file %s: %v", path, err)
			failCount++
			return nil
		}

		relPath, err := filepath.Rel(docsPath, path)
		if err != nil {
			log.Printf("Failed to get relative path for %s: %v", path, err)
			failCount++
			return nil
		}

		parts := strings.Split(relPath, string(os.PathSeparator))
		category := "未分类"
		if len(parts) > 1 {
			category = parts[0]
		}

		filename := d.Name()
		title := strings.TrimSuffix(filename, filepath.Ext(filename))

		var existingDoc model.RAGDocument
		result := database.DB.Where("title = ?", title).First(&existingDoc)
		if result.Error == nil {
			log.Printf("Document '%s' already exists, skipping", title)
			skipCount++
			return nil
		}

		doc := model.RAGDocument{
			ID:        model.GenerateID(),
			Title:     title,
			Content:   string(content),
			Category:  category,
			Tags:      "",
			CreatedBy: "system",
			UpdatedBy: "system",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := database.DB.Create(&doc).Error; err != nil {
			log.Printf("Failed to insert document '%s': %v", title, err)
			failCount++
			return nil
		}

		successCount++
		log.Printf("Imported: %s [Category: %s]", title, category)
		return nil
	})

	if err != nil {
		log.Fatalf("Failed to walk directory: %v", err)
	}

	fmt.Printf("\n=== Import Summary ===\n")
	fmt.Printf("Success: %d\n", successCount)
	fmt.Printf("Skipped: %d\n", skipCount)
	fmt.Printf("Failed: %d\n", failCount)
	fmt.Printf("Total processed: %d\n", successCount+skipCount+failCount)
}

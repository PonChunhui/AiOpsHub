package main

import (
	"fmt"
	"log"

	"github.com/aiops/AiOpsHub/backend/internal/config"
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/service"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	logger.Init()

	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	db := database.DB

	presetTools := service.GetPresetTools()

	for _, tool := range presetTools {
		var existing model.Tool
		result := db.Where("id = ?", tool.ID).First(&existing)

		if result.Error != nil {
			if err := db.Create(&tool).Error; err != nil {
				log.Printf("Failed to create tool %s: %v", tool.ID, err)
			} else {
				fmt.Printf("Created preset tool: %s (%s)\n", tool.ID, tool.Name)
			}
		} else {
			fmt.Printf("Tool already exists: %s (%s)\n", existing.ID, existing.Name)
		}
	}

	var count int64
	db.Model(&model.Tool{}).Count(&count)
	fmt.Printf("\nTotal tools in database: %d\n", count)
}

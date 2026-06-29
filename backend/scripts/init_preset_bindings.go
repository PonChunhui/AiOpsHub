package main

import (
	"fmt"
	"log"
	"time"

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

	fmt.Println("\n=== 初始化预设 Agent ===")
	presetAgents := service.GetPresetAgents()
	for _, agent := range presetAgents {
		var existing model.Agent
		result := db.Where("id = ?", agent.ID).First(&existing)

		if result.Error != nil {
			if err := db.Create(&agent).Error; err != nil {
				log.Printf("Failed to create agent %s: %v", agent.ID, err)
			} else {
				fmt.Printf("Created preset agent: %s (%s)\n", agent.ID, agent.Name)
			}
		} else {
			fmt.Printf("Agent already exists: %s (%s)\n", existing.ID, existing.Name)
		}
	}

	fmt.Println("\n=== 初始化预设 Tool ===")
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

	fmt.Println("\n=== 初始化预设 Agent-Tool 绑定 ===")
	presetBindings := getPresetBindings()
	for _, binding := range presetBindings {
		var existing model.AgentTool
		result := db.Where("agent_id = ? AND tool_id = ?", binding.AgentID, binding.ToolID).First(&existing)

		if result.Error != nil {
			if err := db.Create(&binding).Error; err != nil {
				log.Printf("Failed to create binding %s->%s: %v", binding.AgentID, binding.ToolID, err)
			} else {
				fmt.Printf("Created binding: %s -> %s\n", binding.AgentID, binding.ToolID)
			}
		} else {
			fmt.Printf("Binding already exists: %s -> %s\n", existing.AgentID, existing.ToolID)
		}
	}

	var agentCount, toolCount, bindingCount int64
	db.Model(&model.Agent{}).Count(&agentCount)
	db.Model(&model.Tool{}).Count(&toolCount)
	db.Model(&model.AgentTool{}).Count(&bindingCount)
	fmt.Printf("\n统计: Agents=%d, Tools=%d, Bindings=%d\n", agentCount, toolCount, bindingCount)
}

func getPresetBindings() []model.AgentTool {
	now := time.Now()

	return []model.AgentTool{
		{
			ID:             model.GenerateID(),
			AgentID:        "preset-server-command",
			ToolID:         "tool-ssh-exec",
			ConfigOverride: `{"allowed_commands": ["df", "top", "free", "ps", "netstat", "ls", "cat"], "timeout": 30}`,
			Enabled:        true,
			Priority:       10,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             model.GenerateID(),
			AgentID:        "preset-system-inspection",
			ToolID:         "tool-ssh-exec",
			ConfigOverride: `{"allowed_commands": ["df", "top", "free", "ps", "uptime", "iostat"], "timeout": 30}`,
			Enabled:        true,
			Priority:       10,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             model.GenerateID(),
			AgentID:        "preset-auto-inspection",
			ToolID:         "tool-ssh-exec",
			ConfigOverride: `{"allowed_commands": ["df", "top", "free", "ps", "uptime"], "timeout": 30}`,
			Enabled:        true,
			Priority:       10,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             model.GenerateID(),
			AgentID:        "preset-fault-diagnosis",
			ToolID:         "tool-ssh-exec",
			ConfigOverride: `{"allowed_commands": ["df", "top", "free", "ps", "netstat", "tail", "grep"], "timeout": 30}`,
			Enabled:        true,
			Priority:       5,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             model.GenerateID(),
			AgentID:        "preset-alert-handler",
			ToolID:         "tool-prometheus-query",
			ConfigOverride: `{"url": "http://prometheus:9090", "timeout": 10}`,
			Enabled:        true,
			Priority:       10,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             model.GenerateID(),
			AgentID:        "preset-fault-diagnosis",
			ToolID:         "tool-prometheus-query",
			ConfigOverride: `{"url": "http://prometheus:9090", "timeout": 10}`,
			Enabled:        true,
			Priority:       10,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             model.GenerateID(),
			AgentID:        "preset-log-analyzer",
			ToolID:         "tool-log-query",
			ConfigOverride: `{"datasource": "elasticsearch", "timeout": 20}`,
			Enabled:        true,
			Priority:       10,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             model.GenerateID(),
			AgentID:        "preset-fault-diagnosis",
			ToolID:         "tool-kubernetes-query",
			ConfigOverride: `{"timeout": 15}`,
			Enabled:        true,
			Priority:       8,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}
}

package main

import (
	"fmt"
	"log"

	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/service"
)

func main() {
	// 初始化数据库连接
	if err := database.Init(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 创建AgentService
	agentSvc := service.NewAgentService()

	// 强制重置所有预设agents（包括新添加的流水线助手）
	if err := agentSvc.ForceResetPresets(); err != nil {
		log.Fatalf("强制重置预设agents失败: %v", err)
	}

	fmt.Println("✅ 成功重置所有预设agents，包括流水线助手")

	// 验证流水线助手是否已创建
	agent, err := agentSvc.GetByID("preset-pipeline-helper")
	if err != nil {
		log.Fatalf("验证失败: 无法找到流水线助手agent: %v", err)
	}

	fmt.Printf("✅ 流水线助手已创建:\n")
	fmt.Printf("  ID: %s\n", agent.ID)
	fmt.Printf("  Name: %s\n", agent.Name)
	fmt.Printf("  Category: %s\n", agent.Category)
	fmt.Printf("  Description: %s\n", agent.Description)
	fmt.Printf("  Enabled: %v\n", agent.Enabled)
}

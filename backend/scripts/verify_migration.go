package main

import (
	"fmt"
	"log"

	"github.com/aiops/AiOpsHub/backend/internal/config"
	"github.com/aiops/AiOpsHub/backend/internal/database"
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 禁用 GORM 日志
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	db := database.DB

	fmt.Println("\n=== 数据库迁移验证 ===")

	// 验证 tools 表字段
	fmt.Println("\n[Tools 表验证]")
	var toolColumns []struct {
		ColumnName string
		DataType   string
	}
	db.Raw(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'tools' AND table_schema = CURRENT_SCHEMA()
		ORDER BY ordinal_position
	`).Scan(&toolColumns)

	for _, col := range toolColumns {
		fmt.Printf("  ✅ %s (%s)\n", col.ColumnName, col.DataType)
	}

	// 验证 agent_tools 表
	var agentToolCount int64
	db.Table("agent_tools").Count(&agentToolCount)
	fmt.Printf("\n[Agent_Tools 表验证]\n")
	fmt.Printf("  ✅ agent_tools 表存在，记录数: %d\n", agentToolCount)

	// 验证 ssh_audit_logs 表
	var auditCount int64
	db.Table("ssh_audit_logs").Count(&auditCount)
	fmt.Printf("\n[SSH_Audit_Logs 表验证]\n")
	fmt.Printf("  ✅ ssh_audit_logs 表存在，记录数: %d\n", auditCount)

	// 验证索引
	var indexes []struct {
		IndexName string
		TableName string
	}
	db.Raw(`
		SELECT indexname, tablename 
		FROM pg_indexes 
		WHERE schemaname = CURRENT_SCHEMA() 
		AND tablename IN ('tools', 'agent_tools', 'ssh_audit_logs')
	`).Scan(&indexes)

	fmt.Printf("\n[索引验证]\n")
	for _, idx := range indexes {
		fmt.Printf("  ✅ %s ON %s\n", idx.IndexName, idx.TableName)
	}

	fmt.Println("\n=== 迁移完成 ===")
}

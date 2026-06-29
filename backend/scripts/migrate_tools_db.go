package main

import (
	"fmt"
	"log"

	"github.com/aiops/AiOpsHub/backend/internal/config"
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"gorm.io/gorm"
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	db := database.DB

	fmt.Println("=== 开始数据库迁移 ===")

	if err := executeMigration(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("=== 数据库迁移成功 ===")
}

func executeMigration(db *gorm.DB) error {
	fmt.Println("\n[1] 检查并增强 tools 表...")

	// 添加新字段
	fields := []struct {
		name string
		sql  string
	}{
		{"category", "ALTER TABLE tools ADD COLUMN IF NOT EXISTS category VARCHAR(100)"},
		{"icon", "ALTER TABLE tools ADD COLUMN IF NOT EXISTS icon VARCHAR(10)"},
		{"parameters_schema", "ALTER TABLE tools ADD COLUMN IF NOT EXISTS parameters_schema TEXT"},
		{"default_config", "ALTER TABLE tools ADD COLUMN IF NOT EXISTS default_config TEXT"},
		{"enabled", "ALTER TABLE tools ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT true"},
		{"is_preset", "ALTER TABLE tools ADD COLUMN IF NOT EXISTS is_preset BOOLEAN DEFAULT false"},
		{"created_by", "ALTER TABLE tools ADD COLUMN IF NOT EXISTS created_by VARCHAR(100)"},
		{"updated_by", "ALTER TABLE tools ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100)"},
		{"risk_level", "ALTER TABLE tools ADD COLUMN IF NOT EXISTS risk_level VARCHAR(20) DEFAULT 'low'"},
		{"execution_timeout", "ALTER TABLE tools ADD COLUMN IF NOT EXISTS execution_timeout INTEGER DEFAULT 60"},
	}

	for _, field := range fields {
		if err := db.Exec(field.sql).Error; err != nil {
			fmt.Printf("  ⚠️  添加字段 %s 失败: %v\n", field.name, err)
		} else {
			fmt.Printf("  ✅ 字段 %s 已添加\n", field.name)
		}
	}

	// 创建唯一索引
	fmt.Println("\n[2] 创建唯一索引...")
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tools_name ON tools(name)").Error; err != nil {
		fmt.Printf("  ⚠️  创建唯一索引失败: %v\n", err)
	} else {
		fmt.Println("  ✅ 唯一索引已创建")
	}

	fmt.Println("\n[3] 创建 agent_tools 关联表...")
	createAgentToolsSQL := `
CREATE TABLE IF NOT EXISTS agent_tools (
    id VARCHAR(36) PRIMARY KEY,
    agent_id VARCHAR(36) NOT NULL,
    tool_id VARCHAR(36) NOT NULL,
    config_override TEXT,
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(agent_id, tool_id)
)`

	if err := db.Exec(createAgentToolsSQL).Error; err != nil {
		return fmt.Errorf("创建 agent_tools 表失败: %w", err)
	}
	fmt.Println("  ✅ agent_tools 表已创建")

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_agent_tools_agent_id ON agent_tools(agent_id)",
		"CREATE INDEX IF NOT EXISTS idx_agent_tools_tool_id ON agent_tools(tool_id)",
		"CREATE INDEX IF NOT EXISTS idx_agent_tools_enabled ON agent_tools(enabled)",
	}

	for _, idxSQL := range indexes {
		if err := db.Exec(idxSQL).Error; err != nil {
			fmt.Printf("  ⚠️  创建索引失败: %v\n", err)
		}
	}
	fmt.Println("  ✅ agent_tools 索引已创建")

	fmt.Println("\n[4] 创建 ssh_audit_logs 表（可选）...")
	createAuditSQL := `
CREATE TABLE IF NOT EXISTS ssh_audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    agent_id VARCHAR(36),
    tool_id VARCHAR(36),
    host VARCHAR(100),
    command TEXT,
    allowed BOOLEAN,
    result TEXT,
    executed_at TIMESTAMP,
    executed_by VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`

	if err := db.Exec(createAuditSQL).Error; err != nil {
		fmt.Printf("  ⚠️  创建 ssh_audit_logs 表失败: %v\n", err)
	} else {
		fmt.Println("  ✅ ssh_audit_logs 表已创建")
	}

	// 创建审计日志索引
	auditIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_ssh_audit_agent_id ON ssh_audit_logs(agent_id)",
		"CREATE INDEX IF NOT EXISTS idx_ssh_audit_tool_id ON ssh_audit_logs(tool_id)",
		"CREATE INDEX IF NOT EXISTS idx_ssh_audit_executed_at ON ssh_audit_logs(executed_at)",
		"CREATE INDEX IF NOT EXISTS idx_ssh_audit_allowed ON ssh_audit_logs(allowed)",
	}

	for _, idxSQL := range auditIndexes {
		if err := db.Exec(idxSQL).Error; err != nil {
			fmt.Printf("  ⚠️  创建审计索引失败: %v\n", err)
		}
	}
	fmt.Println("  ✅ ssh_audit_logs 索引已创建")

	fmt.Println("\n[5] 验证表结构...")

	// 验证 tools 表
	var toolCount int64
	db.Table("tools").Count(&toolCount)
	fmt.Printf("  ✅ tools 表存在，当前记录数: %d\n", toolCount)

	// 验证 agent_tools 表
	var agentToolCount int64
	db.Table("agent_tools").Count(&agentToolCount)
	fmt.Printf("  ✅ agent_tools 表存在，当前记录数: %d\n", agentToolCount)

	// 验证 ssh_audit_logs 表
	var auditCount int64
	db.Table("ssh_audit_logs").Count(&auditCount)
	fmt.Printf("  ✅ ssh_audit_logs 表存在，当前记录数: %d\n", auditCount)

	return nil
}

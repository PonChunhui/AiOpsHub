-- Tool 管理系统数据库迁移
-- Version: 001
-- Description: 增强 tools 表，创建 agent_tools 关联表

-- 1. 增强 tools 表（添加新字段）
ALTER TABLE tools ADD COLUMN IF NOT EXISTS category VARCHAR(100);
ALTER TABLE tools ADD COLUMN IF NOT EXISTS icon VARCHAR(10);
ALTER TABLE tools ADD COLUMN IF NOT EXISTS parameters_schema TEXT;
ALTER TABLE tools ADD COLUMN IF NOT EXISTS default_config TEXT;
ALTER TABLE tools ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT true;
ALTER TABLE tools ADD COLUMN IF NOT EXISTS is_preset BOOLEAN DEFAULT false;
ALTER TABLE tools ADD COLUMN IF NOT EXISTS created_by VARCHAR(100);
ALTER TABLE tools ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100);
ALTER TABLE tools ADD COLUMN IF NOT EXISTS risk_level VARCHAR(20) DEFAULT 'low';
ALTER TABLE tools ADD COLUMN IF NOT EXISTS execution_timeout INTEGER DEFAULT 60;

-- 创建唯一索引（防止重复工具名称）
CREATE UNIQUE INDEX IF NOT EXISTS idx_tools_name ON tools(name);

-- 2. 创建 agent_tools 关联表（多对多关系）
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
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_agent_tools_agent_id ON agent_tools(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_tools_tool_id ON agent_tools(tool_id);
CREATE INDEX IF NOT EXISTS idx_agent_tools_enabled ON agent_tools(enabled);

-- 添加外键约束（可选，取决于是否需要级联删除）
-- ALTER TABLE agent_tools ADD CONSTRAINT fk_agent_tools_agent 
--     FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE;
-- ALTER TABLE agent_tools ADD CONSTRAINT fk_agent_tools_tool 
--     FOREIGN KEY (tool_id) REFERENCES tools(id) ON DELETE CASCADE;

-- 3. 创建 SSH 审计日志表（可选功能）
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
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_ssh_audit_agent_id ON ssh_audit_logs(agent_id);
CREATE INDEX IF NOT EXISTS idx_ssh_audit_tool_id ON ssh_audit_logs(tool_id);
CREATE INDEX IF NOT EXISTS idx_ssh_audit_executed_at ON ssh_audit_logs(executed_at);
CREATE INDEX IF NOT EXISTS idx_ssh_audit_allowed ON ssh_audit_logs(allowed);
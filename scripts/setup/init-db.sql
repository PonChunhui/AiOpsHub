-- AiOpsHub 数据库初始化脚本
-- PostgreSQL DDL

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- 用户和权限相关表
-- ============================================

-- 角色表
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    permissions JSONB NOT NULL DEFAULT '{}',
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 插入默认角色
INSERT INTO roles (name, permissions, description) VALUES
('admin', '{"users": ["read", "write", "delete"], "datasources": ["read", "write", "delete"], "alerts": ["read", "write", "delete"], "agents": ["read", "write", "delete"], "knowledge": ["read", "write", "delete"]}', '系统管理员，拥有所有权限'),
('operator', '{"users": ["read"], "datasources": ["read", "write"], "alerts": ["read", "write"], "agents": ["read", "write"], "knowledge": ["read", "write"]}', '运维工程师，可以操作和管理'),
('viewer', '{"users": ["read"], "datasources": ["read"], "alerts": ["read"], "agents": ["read"], "knowledge": ["read"]}', '只读用户，只能查看');

-- 用户表
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'deleted')),
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role_id);

-- 插入默认管理员用户（密码：admin123，实际部署时需要修改）
INSERT INTO users (username, email, password_hash, role_id) VALUES
('admin', 'admin@aiops.com', '$2a$10$YourHashedPasswordHere', (SELECT id FROM roles WHERE name = 'admin'));

-- ============================================
-- 数据源相关表
-- ============================================

-- 数据源表
CREATE TABLE datasources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('prometheus', 'zabbix', 'elk', 'k8s', 'jaeger', 'cloud', 'custom')),
    endpoint VARCHAR(255) NOT NULL,
    credentials JSONB,  -- 加密存储
    sync_config JSONB DEFAULT '{}',
    health_status VARCHAR(20) DEFAULT 'unknown' CHECK (health_status IN ('healthy', 'unhealthy', 'unknown')),
    last_sync_at TIMESTAMP,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_datasources_type ON datasources(type);
CREATE INDEX idx_datasources_health ON datasources(health_status);
CREATE INDEX idx_datasources_created ON datasources(created_by);

-- ============================================
-- Agent相关表
-- ============================================

-- Agent会话表
CREATE TABLE agent_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    intent TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'success', 'failed', 'cancelled', 'waiting')),
    current_agent VARCHAR(100),
    context JSONB DEFAULT '{}',
    result JSONB,
    workflow_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_sessions_user ON agent_sessions(user_id);
CREATE INDEX idx_sessions_status ON agent_sessions(status);
CREATE INDEX idx_sessions_created ON agent_sessions(created_at DESC);
CREATE INDEX idx_sessions_workflow ON agent_sessions(workflow_id);

-- Agent执行日志表
CREATE TABLE agent_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    agent_name VARCHAR(100) NOT NULL,
    action TEXT,
    tool_used VARCHAR(100),
    input JSONB,
    output JSONB,
    llm_model VARCHAR(100),
    tokens_used INTEGER DEFAULT 0,
    duration_ms INTEGER,
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_logs_session ON agent_logs(session_id);
CREATE INDEX idx_logs_agent ON agent_logs(agent_name);
CREATE INDEX idx_logs_created ON agent_logs(created_at DESC);
CREATE INDEX idx_logs_tool ON agent_logs(tool_used);

-- Agent记忆表
CREATE TABLE agent_memories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_name VARCHAR(100) NOT NULL,
    memory_type VARCHAR(50) NOT NULL CHECK (memory_type IN ('short_term', 'long_term', 'episodic')),
    content TEXT NOT NULL,
    embedding_id VARCHAR(255),  -- Milvus向量ID
    importance_score FLOAT DEFAULT 0.5,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_accessed TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_memories_agent ON agent_memories(agent_name);
CREATE INDEX idx_memories_type ON agent_memories(memory_type);
CREATE INDEX idx_memories_importance ON agent_memories(importance_score DESC);
CREATE INDEX idx_memories_embedding ON agent_memories(embedding_id);

-- 工具调用记录表
CREATE TABLE tool_calls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    agent_name VARCHAR(100) NOT NULL,
    tool_name VARCHAR(100) NOT NULL,
    parameters JSONB NOT NULL,
    result JSONB,
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    duration_ms INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tool_session ON tool_calls(session_id);
CREATE INDEX idx_tool_agent ON tool_calls(agent_name);
CREATE INDEX idx_tool_name ON tool_calls(tool_name);
CREATE INDEX idx_tool_created ON tool_calls(created_at DESC);

-- ============================================
-- 告警相关表
-- ============================================

-- 告警规则表
CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('deduplication', 'aggregation', 'inhibition', 'silence', 'threshold')),
    config JSONB NOT NULL,
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 100,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_rules_type ON alert_rules(type);
CREATE INDEX idx_rules_enabled ON alert_rules(enabled);
CREATE INDEX idx_rules_priority ON alert_rules(priority);

-- 告警历史表（主表）
CREATE TABLE alert_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fingerprint VARCHAR(255) NOT NULL,
    source VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('P0', 'P1', 'P2', 'P3', 'info')),
    status VARCHAR(50) NOT NULL DEFAULT 'firing' CHECK (status IN ('firing', 'resolved', 'silenced', 'suppressed', 'acknowledged')),
    title TEXT NOT NULL,
    description TEXT,
    labels JSONB DEFAULT '{}',
    annotations JSONB DEFAULT '{}',
    starts_at TIMESTAMP NOT NULL,
    ends_at TIMESTAMP,
    processed_by_agent BOOLEAN DEFAULT false,
    agent_session_id UUID REFERENCES agent_sessions(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_alerts_fingerprint ON alert_history(fingerprint);
CREATE INDEX idx_alerts_status ON alert_history(status);
CREATE INDEX idx_alerts_severity ON alert_history(severity);
CREATE INDEX idx_alerts_starts ON alert_history(starts_at DESC);
CREATE INDEX idx_alerts_source ON alert_history(source);
CREATE INDEX idx_alerts_session ON alert_history(agent_session_id);

-- 告警处理日志表
CREATE TABLE alert_processing_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_id UUID NOT NULL REFERENCES alert_history(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL CHECK (action IN ('created', 'deduplicated', 'aggregated', 'silenced', 'suppressed', 'escalated', 'resolved', 'acknowledged')),
    agent_name VARCHAR(100),
    details JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_alert_log_alert ON alert_processing_logs(alert_id);
CREATE INDEX idx_alert_log_action ON alert_processing_logs(action);
CREATE INDEX idx_alert_log_created ON alert_processing_logs(created_at DESC);

-- ============================================
-- 知识库相关表
-- ============================================

-- 知识条目表
CREATE TABLE knowledge_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(100) NOT NULL CHECK (category IN ('incident', 'manual', 'best_practice', 'architecture', 'tool')),
    embedding_id VARCHAR(255),  -- Milvus向量ID
    tags JSONB DEFAULT '{}',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_knowledge_category ON knowledge_items(category);
CREATE INDEX idx_knowledge_embedding ON knowledge_items(embedding_id);
CREATE INDEX idx_knowledge_created ON knowledge_items(created_by);
CREATE INDEX idx_knowledge_updated ON knowledge_items(updated_at DESC);

-- 知识图谱节点表（可选）
CREATE TABLE topology_nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('service', 'database', 'middleware', 'network', 'storage', 'application')),
    ip_address VARCHAR(50),
    port INTEGER,
    labels JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_topology_name ON topology_nodes(name);
CREATE INDEX idx_topology_type ON topology_nodes(type);

-- 知识图谱关系表（可选）
CREATE TABLE topology_relations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_node_id UUID NOT NULL REFERENCES topology_nodes(id) ON DELETE CASCADE,
    target_node_id UUID NOT NULL REFERENCES topology_nodes(id) ON DELETE CASCADE,
    relation_type VARCHAR(50) NOT NULL CHECK (relation_type IN ('depends', 'calls', 'connects', 'contains', 'replicates')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_topology_source ON topology_relations(source_node_id);
CREATE INDEX idx_topology_target ON topology_relations(target_node_id);
CREATE INDEX idx_topology_relation ON topology_relations(relation_type);

-- ============================================
-- 系统配置表
-- ============================================

-- Agent配置表
CREATE TABLE agent_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_name VARCHAR(100) UNIQUE NOT NULL,
    llm_model VARCHAR(100) NOT NULL,
    temperature FLOAT DEFAULT 0.7,
    max_tokens INTEGER DEFAULT 4000,
    tools JSONB DEFAULT '{}',
    prompt_template TEXT,
    memory_config JSONB DEFAULT '{}',
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_agent_config_name ON agent_configs(agent_name);
CREATE INDEX idx_agent_config_enabled ON agent_configs(enabled);

-- 插入默认Agent配置
INSERT INTO agent_configs (agent_name, llm_model, tools, prompt_template, memory_config) VALUES
('monitor_agent', 'gpt-3.5-turbo', '{"tools": ["prometheus_query", "kubernetes_api", "zabbix_api"]}', '你是监控采集Agent，负责从各数据源采集监控数据...', '{"short_term": {"ttl": 3600}, "long_term": {"enabled": true}}'),
('analysis_agent', 'gpt-4-turbo-preview', '{"tools": ["log_analysis", "metric_correlation", "root_cause_analysis", "topology_query"]}', '你是分析诊断Agent，负责分析异常和定位根因...', '{"short_term": {"ttl": 3600}, "long_term": {"enabled": true}}'),
('alert_agent', 'glm-4', '{"tools": ["alert_deduplication", "alert_aggregation", "severity_assessment", "notification_sender"]}', '你是告警处理Agent，负责告警降噪和分派...', '{"short_term": {"ttl": 3600}, "long_term": {"enabled": true}}'),
('decision_agent', 'gpt-4-turbo-preview', '{"tools": ["risk_assessment", "playbook_executor", "kubernetes_operator", "ssh_executor", "sql_executor"]}', '你是决策执行Agent，负责自动化决策和故障自愈...', '{"short_term": {"ttl": 3600}, "long_term": {"enabled": true}}'),
('learning_agent', 'qwen-72b', '{"tools": ["pattern_miner", "threshold_optimizer", "knowledge_updater", "model_trainer"]}', '你是学习优化Agent，负责持续学习和优化...', '{"short_term": {"ttl": 3600}, "long_term": {"enabled": true}}'),
('interaction_agent', 'gpt-4-turbo-preview', '{"tools": ["chat_manager", "intent_recognition", "report_generator", "visualization_builder"]}', '你是交互服务Agent，负责用户交互...', '{"short_term": {"ttl": 3600}, "long_term": {"enabled": true}}');

-- LLM配置表
CREATE TABLE llm_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider VARCHAR(50) NOT NULL CHECK (provider IN ('openai', 'zhipu', 'anthropic', 'local', 'custom')),
    model_name VARCHAR(100) NOT NULL,
    api_endpoint VARCHAR(255),
    api_key_encrypted TEXT,  -- 加密存储
    max_tokens INTEGER DEFAULT 4000,
    temperature FLOAT DEFAULT 0.7,
    cost_per_1k_tokens_input FLOAT DEFAULT 0.0,
    cost_per_1k_tokens_output FLOAT DEFAULT 0.0,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_llm_provider ON llm_configs(provider);
CREATE INDEX idx_llm_model ON llm_configs(model_name);
CREATE INDEX idx_llm_enabled ON llm_configs(enabled);

-- ============================================
-- 监控和审计表
-- ============================================

-- 操作审计日志表
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    details JSONB DEFAULT '{}',
    ip_address VARCHAR(50),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_audit_user ON audit_logs(user_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_created ON audit_logs(created_at DESC);

-- ============================================
-- 更新时间触发器
-- ============================================

-- 创建更新时间函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要自动更新时间的表创建触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_datasources_updated_at BEFORE UPDATE ON datasources FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sessions_updated_at BEFORE UPDATE ON agent_sessions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_alert_rules_updated_at BEFORE UPDATE ON alert_rules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_knowledge_updated_at BEFORE UPDATE ON knowledge_items FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_agent_configs_updated_at BEFORE UPDATE ON agent_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_llm_configs_updated_at BEFORE UPDATE ON llm_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_topology_nodes_updated_at BEFORE UPDATE ON topology_nodes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 视图（可选）
-- ============================================

-- Agent会话详情视图
CREATE VIEW agent_session_details AS
SELECT
    s.id,
    s.user_id,
    u.username,
    s.intent,
    s.status,
    s.current_agent,
    s.created_at,
    s.updated_at,
    COUNT(l.id) AS log_count,
    SUM(l.tokens_used) AS total_tokens,
    SUM(l.duration_ms) AS total_duration_ms
FROM agent_sessions s
LEFT JOIN users u ON s.user_id = u.id
LEFT JOIN agent_logs l ON s.id = l.session_id
GROUP BY s.id, s.user_id, u.username, s.intent, s.status, s.current_agent, s.created_at, s.updated_at;

-- 告警统计视图
CREATE VIEW alert_statistics AS
SELECT
    source,
    severity,
    status,
    COUNT(*) AS alert_count,
    MIN(starts_at) AS first_alert,
    MAX(starts_at) AS last_alert
FROM alert_history
WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY source, severity, status;

-- ============================================
-- 完成
-- ============================================

-- 打印完成信息
DO $$
BEGIN
    RAISE NOTICE 'AiOpsHub数据库初始化完成';
    RAISE NOTICE '创建表数量：%', (SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public');
    RAISE NOTICE '创建索引数量：%', (SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public');
END $$;
-- ============================================
-- Workflow相关表
-- ============================================

-- Workflow定义表
CREATE TABLE workflow_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    
    -- Workflow配置
    workflow_type VARCHAR(50) NOT NULL,        -- incident_handling/alert_dedup/root_cause_analysis
    nodes JSONB NOT NULL,                      -- DAG节点定义
    edges JSONB NOT NULL,                      -- DAG边定义
    start_node VARCHAR(100) NOT NULL,
    end_nodes JSONB DEFAULT '[]',
    
    -- 执行配置
    timeout_seconds INTEGER DEFAULT 3600,
    retry_policy JSONB DEFAULT '{"max_attempts": 3, "initial_interval": 10}',
    
    -- 状态
    enabled BOOLEAN DEFAULT true,
    version INTEGER DEFAULT 1,
    
    -- 元数据
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Workflow执行历史表
CREATE TABLE workflow_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_id UUID REFERENCES workflow_definitions(id),
    workflow_run_id VARCHAR(100) UNIQUE,
    
    -- 执行状态
    status VARCHAR(50) NOT NULL,               -- running/success/failed/cancelled
    input JSONB NOT NULL,
    output JSONB,
    error_message TEXT,
    
    -- 进度跟踪
    current_node VARCHAR(100),
    completed_nodes JSONB DEFAULT '[]',
    node_states JSONB DEFAULT '{}',            -- 各节点执行状态
    
    -- 性能指标
    duration_ms INTEGER,
    total_tokens INTEGER DEFAULT 0,
    total_cost_usd FLOAT DEFAULT 0.0,
    
    -- 时间
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

-- Agent配置表（适配langchaingo）
CREATE TABLE agent_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    
    -- LLM配置
    llm_provider VARCHAR(50) NOT NULL,         -- openai/zhipu/ollama/gemini
    llm_model VARCHAR(100) NOT NULL,           -- gpt-4-turbo/glm-4/llama2
    temperature FLOAT DEFAULT 0.7,
    max_tokens INTEGER DEFAULT 4000,
    
    -- Prompt配置
    system_prompt TEXT NOT NULL,
    user_prompt_template TEXT,
    
    -- 工具配置
    tools JSONB DEFAULT '[]',                  -- 工具列表
    tool_configs JSONB DEFAULT '{}',           -- 工具配置
    
    -- Memory配置
    memory_type VARCHAR(50) DEFAULT 'conversation_buffer',
    memory_config JSONB DEFAULT '{}',
    
    -- 执行配置
    max_iterations INTEGER DEFAULT 10,
    timeout_seconds INTEGER DEFAULT 300,
    
    -- 状态
    enabled BOOLEAN DEFAULT true,
    
    -- 元数据
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 工具注册表
CREATE TABLE tool_registry (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    
    -- 工具类型
    category VARCHAR(50) NOT NULL,             -- monitor/analysis/alert/execution/knowledge
    tool_type VARCHAR(50) NOT NULL,            -- prometheus/kubernetes/sql/ssh/http
    
    -- 参数定义
    parameters_schema JSONB NOT NULL,          -- 参数schema（JSON Schema格式）
    return_schema JSONB,                       -- 返回值schema
    
    -- 实现
    implementation VARCHAR(200),               -- Go函数名或路径
    config_template JSONB,                     -- 配置模板
    
    -- 状态
    enabled BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- Activity执行记录表
CREATE TABLE activity_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_execution_id UUID REFERENCES workflow_executions(id),
    
    -- Activity信息
    activity_name VARCHAR(100) NOT NULL,
    activity_type VARCHAR(50) NOT NULL,        -- agent/tool/chain
    
    -- 执行详情
    input JSONB NOT NULL,
    output JSONB,
    
    -- 状态
    status VARCHAR(50) NOT NULL,               -- running/success/failed
    error_message TEXT,
    
    -- 性能
    duration_ms INTEGER,
    tokens_used INTEGER DEFAULT 0,
    cost_usd FLOAT DEFAULT 0.0,
    
    -- 重试信息
    attempt_number INTEGER DEFAULT 1,
    retry_history JSONB DEFAULT '[]',
    
    -- 时间
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

-- 知识库表（更新）
ALTER TABLE knowledge_items ADD COLUMN IF NOT EXISTS embedding_model VARCHAR(100) DEFAULT 'text-embedding-3-large';
ALTER TABLE knowledge_items ADD COLUMN IF NOT EXISTS vector_status VARCHAR(50) DEFAULT 'pending';
ALTER TABLE knowledge_items ADD COLUMN IF NOT EXISTS upload_progress FLOAT DEFAULT 0.0;

-- ============================================
-- 索引创建
-- ============================================

CREATE INDEX idx_workflow_def_type ON workflow_definitions(workflow_type);
CREATE INDEX idx_workflow_def_enabled ON workflow_definitions(enabled);
CREATE INDEX idx_workflow_exec_status ON workflow_executions(status);
CREATE INDEX idx_workflow_exec_workflow ON workflow_executions(workflow_id);
CREATE INDEX idx_agent_config_llm ON agent_configs(llm_provider, llm_model);
CREATE INDEX idx_agent_config_enabled ON agent_configs(enabled);
CREATE INDEX idx_tool_registry_category ON tool_registry(category);
CREATE INDEX idx_tool_registry_enabled ON tool_registry(enabled);
CREATE INDEX idx_activity_exec_workflow ON activity_executions(workflow_execution_id);
CREATE INDEX idx_activity_exec_status ON activity_executions(status);

-- ============================================
-- 插入默认Agent配置
-- ============================================

INSERT INTO agent_configs (name, llm_provider, llm_model, system_prompt, tools, memory_type) VALUES
('monitor_agent', 'openai', 'gpt-3.5-turbo', 
 '你是监控采集Agent，负责从Prometheus、Kubernetes等数据源采集监控数据。可用工具：prometheus_query、kubernetes_api。请根据用户需求选择合适的工具采集数据。',
 '["prometheus_query", "kubernetes_api"]', 'conversation_buffer'),

('analysis_agent', 'openai', 'gpt-4-turbo-preview',
 '你是分析诊断Agent，负责分析异常和定位根因。可用工具：log_analysis、metric_correlation、root_cause_analysis、topology_query。请多维度分析，生成根因报告。',
 '["log_analysis", "metric_correlation", "root_cause_analysis"]', 'conversation_buffer'),

('alert_agent', 'zhipu', 'glm-4',
 '你是告警处理Agent，负责告警去重、聚合、分派。请分析告警语义，合并相似告警，评估严重性，推荐处理人。',
 '["alert_deduplication", "alert_aggregation", "severity_assessment"]', 'conversation_buffer'),

('decision_agent', 'openai', 'gpt-4-turbo-preview',
 '你是决策执行Agent，负责制定执行方案和自动化修复。可用工具：kubernetes_operator、ssh_executor、sql_executor。请评估风险，生成执行计划，高风险操作需人工确认。',
 '["kubernetes_operator", "ssh_executor", "sql_executor"]', 'conversation_buffer'),

('learning_agent', 'ollama', 'llama2',
 '你是学习优化Agent，负责持续学习和知识沉淀。可用工具：pattern_miner、threshold_optimizer、knowledge_updater。请分析历史数据，生成优化建议。',
 '["pattern_miner", "threshold_optimizer"]', 'conversation_buffer'),

('interaction_agent', 'openai', 'gpt-4-turbo-preview',
 '你是交互服务Agent，负责用户自然语言交互。请理解用户意图，调用其他Agent，生成人类可读的回复。',
 '["chat_manager", "intent_recognition", "report_generator"]', 'conversation_buffer_window');

-- ============================================
-- 插入默认Workflow定义
-- ============================================

INSERT INTO workflow_definitions (name, workflow_type, nodes, edges, start_node, description) VALUES
('incident_handling', 'incident_handling',
 '[{"id": "node-1", "name": "monitor_agent", "type": "agent"}, {"id": "node-2", "name": "analysis_agent", "type": "agent"}, {"id": "node-3", "name": "decision_agent", "type": "agent"}]',
 '[{"from": "node-1", "to": "node-2"}, {"from": "node-2", "to": "node-3"}]',
 'node-1', '故障处理工作流：监控采集 → 根因分析 → 决策执行'),

('alert_dedup', 'alert_dedup',
 '[{"id": "node-1", "name": "alert_agent", "type": "agent"}]',
 '[]',
 'node-1', '告警降噪工作流：去重 → 聚合 → 分派');

-- ============================================
-- 插入默认工具注册
-- ============================================

INSERT INTO tool_registry (name, category, tool_type, description, parameters_schema, implementation) VALUES
('prometheus_query', 'monitor', 'prometheus', '查询Prometheus监控指标',
 '{"type": "object", "properties": {"query": {"type": "string", "description": "PromQL查询语句"}}}', 'tools.PrometheusQuery'),

('kubernetes_api', 'monitor', 'kubernetes', '查询和操作Kubernetes资源',
 '{"type": "object", "properties": {"action": {"type": "string"}, "resource": {"type": "string"}}}', 'tools.KubernetesAPI'),

('log_analysis', 'analysis', 'log', '分析日志数据',
 '{"type": "object", "properties": {"logs": {"type": "array"}}}', 'tools.LogAnalysis'),

('kubernetes_operator', 'execution', 'kubernetes', '执行Kubernetes操作',
 '{"type": "object", "properties": {"operation": {"type": "string"}}}', 'tools.KubernetesOperator');


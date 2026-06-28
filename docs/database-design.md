# AiOpsHub 数据库设计文档

## 文档信息

| 项目 | 内容 |
|------|------|
| 文档名称 | 数据库设计 |
| 版本 | v1.0.0 |
| 创建日期 | 2026-06-24 |

## 一、数据库架构

### 1.1 数据库分层

```
┌─────────────────────────────────────────────┐
│          业务数据层 (PostgreSQL)             │
│  - 用户和权限                                │
│  - Agent配置和会话                          │
│  - 告警规则                                  │
│  - 数据源配置                                │
└─────────────────────────────────────────────┘
│
┌─────────────────────────────────────────────┐
│          时序数据层 (ClickHouse)             │
│  - 监控指标                                  │
│  - 日志数据                                  │
│  - 告警历史                                  │
│  - Agent执行记录                             │
└─────────────────────────────────────────────┘
│
┌─────────────────────────────────────────────┐
│         向量数据层 (Milvus)                  │
│  - 知识库向量                                │
│  - Agent记忆向量                            │
│  - 案例库向量                                │
└─────────────────────────────────────────────┘
│
┌─────────────────────────────────────────────┐
│        缓存和状态层 (Redis)                  │
│  - Agent状态                                 │
│  - 会话记忆                                  │
│  - 缓存数据                                  │
└─────────────────────────────────────────────┘
```

### 1.2 数据库选型理由

#### PostgreSQL

**选型理由**：
- 成熟稳定，适合业务数据
- 支持JSONB，灵活存储配置
- 支持复杂查询和事务
- 支持主从复制和备份

**使用场景**：
- 用户、角色、权限
- Agent会话和配置
- 告警规则
- 数据源配置
- 工具调用记录

#### ClickHouse

**选型理由**：
- 查询性能极高
- 列式存储，压缩率高
- 支持实时插入
- 适合时序数据和日志

**使用场景**：
- 监控指标（每秒百万级）
- 日志数据（每秒万级）
- 告警历史
- Agent执行指标

#### Milvus

**选型理由**：
- 高性能向量检索
- 支持多种索引
- 云原生架构
- 支持分布式部署

**使用场景**：
- 知识库向量存储
- Agent记忆向量
- 故障案例向量

#### Redis

**选型理由**：
- 高性能读写
- 支持多种数据结构
- 支持持久化
- 支持集群

**使用场景**：
- Agent状态共享
- 会话短期记忆
- 缓存
- 分布式锁

## 二、PostgreSQL数据模型

### 2.1 ER图

```
┌─────────────┐         ┌─────────────┐
│   users     │         │   roles     │
│             │         │             │
│ id (PK)     │──────┐  │ id (PK)     │
│ username    │      │  │ name        │
│ email       │      │  │ permissions │
│ password    │      │  │             │
│ role_id(FK) │◄─────┘  │             │
│ created_at  │         │             │
└─────────────┘         └─────────────┘
       │
       │
       ↓
┌─────────────┐         ┌─────────────┐
│datasources  │         │alert_rules  │
│             │         │             │
│ id (PK)     │         │ id (PK)     │
│ name        │         │ name        │
│ type        │         │ type        │
│ endpoint    │         │ config      │
│ credentials │         │ enabled     │
│ created_by  │         │ created_by  │
│ (FK→users)  │         │ (FK→users)  │
└─────────────┘         └─────────────┘

┌─────────────┐         ┌─────────────┐
│agent_sessions│        │agent_logs   │
│             │         │             │
│ id (PK)     │──────┐  │ id (PK)     │
│ user_id(FK) │      │  │session_id   │
│ intent      │      │  │ (FK→sessions)│
│ status      │      │  │ agent_name  │
│current_agent│      │  │ action      │
│ context     │      │  │ tool_used   │
│ result      │      │  │ input       │
│ created_at  │      │  │ output      │
│ updated_at  │      │  │ llm_model   │
└─────────────┘      │  │ tokens_used │
                     │  │ duration_ms │
                     └─►│ created_at  │
                        └─────────────┘
```

### 2.2 核心数据表

#### 2.2.1 用户和权限表

**users（用户表）**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| username | VARCHAR(50) | 用户名，唯一 |
| email | VARCHAR(100) | 邮箱，唯一 |
| password_hash | VARCHAR(255) | 密码哈希 |
| role_id | UUID | 外键，关联roles |
| status | VARCHAR(20) | 状态：active/inactive |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

**roles（角色表）**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| name | VARCHAR(50) | 角色名称：admin/operator/viewer |
| permissions | JSONB | 权限列表 |
| description | TEXT | 描述 |
| created_at | TIMESTAMP | 创建时间 |

#### 2.2.2 数据源表

**datasources（数据源表）**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| name | VARCHAR(100) | 数据源名称 |
| type | VARCHAR(50) | 类型：prometheus/zabbix/elk/k8s/custom |
| endpoint | VARCHAR(255) | API地址 |
| credentials | JSONB | 认证信息（加密存储） |
| sync_config | JSONB | 同步配置 |
| health_status | VARCHAR(20) | 健康状态：healthy/unhealthy |
| last_sync_at | TIMESTAMP | 最后同步时间 |
| created_by | UUID | 创建人 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

#### 2.2.3 Agent相关表

**agent_sessions（Agent会话表）**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| user_id | UUID | 外键，关联users |
| intent | TEXT | 用户意图描述 |
| status | VARCHAR(50) | 状态：running/success/failed/cancelled |
| current_agent | VARCHAR(100) | 当前执行的Agent |
| context | JSONB | 会话上下文（共享状态） |
| result | JSONB | 最终结果 |
| workflow_id | VARCHAR(100) | LangGraph工作流ID |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

**agent_logs（Agent执行日志表）**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| session_id | UUID | 外键，关联agent_sessions |
| agent_name | VARCHAR(100) | Agent名称 |
| action | TEXT | 执行的动作 |
| tool_used | VARCHAR(100) | 使用的工具 |
| input | JSONB | 输入参数 |
| output | JSONB | 输出结果 |
| llm_model | VARCHAR(100) | 使用的LLM模型 |
| tokens_used | INTEGER | Token消耗 |
| duration_ms | INTEGER | 执行耗时（毫秒） |
| success | BOOLEAN | 是否成功 |
| error_message | TEXT | 错误信息 |
| created_at | TIMESTAMP | 创建时间 |

**agent_memories（Agent记忆表）**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| agent_name | VARCHAR(100) | Agent名称 |
| memory_type | VARCHAR(50) | 类型：short_term/long_term/episodic |
| content | TEXT | 记忆内容 |
| embedding_id | VARCHAR(255) | Milvus中的向量ID |
| importance_score | FLOAT | 重要性评分 |
| metadata | JSONB | 元数据 |
| created_at | TIMESTAMP | 创建时间 |
| last_accessed | TIMESTAMP | 最后访问时间 |

**tool_calls（工具调用记录表）**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| session_id | UUID | 外键，关联agent_sessions |
| agent_name | VARCHAR(100) | Agent名称 |
| tool_name | VARCHAR(100) | 工具名称 |
| parameters | JSONB | 调用参数 |
| result | JSONB | 调用结果 |
| success | BOOLEAN | 是否成功 |
| error_message | TEXT | 错误信息 |
| duration_ms | INTEGER | 执行耗时 |
| created_at | TIMESTAMP | 创建时间 |

#### 2.2.4 告警相关表

**alert_rules（告警规则表）**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| name | VARCHAR(100) | 规则名称 |
| type | VARCHAR(50) | 类型：deduplication/aggregation/inhibition/silence |
| config | JSONB | 规则配置 |
| enabled | BOOLEAN | 是否启用 |
| priority | INTEGER | 优先级 |
| created_by | UUID | 创建人 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

**alert_history（告警历史表）**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| fingerprint | VARCHAR(255) | 告警指纹（用于去重） |
| source | VARCHAR(100) | 告警源：prometheus/zabbix/custom |
| severity | VARCHAR(20) | 严重性：P0/P1/P2/P3 |
| status | VARCHAR(50) | 状态：firing/resolved/silenced/suppressed |
| title | TEXT | 告警标题 |
| description | TEXT | 告警描述 |
| labels | JSONB | 标签 |
| annotations | JSONB | 注释 |
| starts_at | TIMESTAMP | 开始时间 |
| ends_at | TIMESTAMP | 结束时间 |
| processed_by_agent | BOOLEAN | 是否被Agent处理 |
| agent_session_id | UUID | 关联的Agent会话 |
| created_at | TIMESTAMP | 创建时间 |

#### 2.2.5 知识库表

**knowledge_items（知识条目表）**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| title | VARCHAR(255) | 标题 |
| content | TEXT | 内容 |
| category | VARCHAR(100) | 分类：incident/manual/best_practice |
| embedding_id | VARCHAR(255) | Milvus向量ID |
| tags | JSONB | 标签 |
| created_by | UUID | 创建人 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### 2.3 索引设计

**性能关键索引**：

```sql
-- users表
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

-- agent_sessions表
CREATE INDEX idx_sessions_user ON agent_sessions(user_id);
CREATE INDEX idx_sessions_status ON agent_sessions(status);
CREATE INDEX idx_sessions_created ON agent_sessions(created_at DESC);

-- agent_logs表
CREATE INDEX idx_logs_session ON agent_logs(session_id);
CREATE INDEX idx_logs_agent ON agent_logs(agent_name);
CREATE INDEX idx_logs_created ON agent_logs(created_at DESC);

-- alert_history表
CREATE INDEX idx_alerts_fingerprint ON alert_history(fingerprint);
CREATE INDEX idx_alerts_status ON alert_history(status);
CREATE INDEX idx_alerts_starts ON alert_history(starts_at DESC);
CREATE INDEX idx_alerts_source ON alert_history(source);

-- tool_calls表
CREATE INDEX idx_tool_session ON tool_calls(session_id);
CREATE INDEX idx_tool_created ON tool_calls(created_at DESC);
```

### 2.4 分区策略

**大表分区**：

```sql
-- agent_logs按月分区
CREATE TABLE agent_logs (
    id UUID,
    session_id UUID,
    created_at TIMESTAMP,
    ...
) PARTITION BY RANGE (created_at);

CREATE TABLE agent_logs_202601 PARTITION OF agent_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE agent_logs_202602 PARTITION OF agent_logs
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

-- alert_history按月分区
CREATE TABLE alert_history (
    id UUID,
    starts_at TIMESTAMP,
    ...
) PARTITION BY RANGE (starts_at);
```

## 三、ClickHouse数据模型

### 3.1 监控指标表

**metrics（监控指标表）**

```sql
CREATE TABLE metrics (
    timestamp DateTime,
    metric_name String,
    labels Map(String, String),
    value Float64,
    source String,
    service String,
    host String
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (metric_name, timestamp)
SETTINGS index_granularity = 8192;
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| timestamp | DateTime | 时间戳 |
| metric_name | String | 指标名称 |
| labels | Map(String, String) | 标签（键值对） |
| value | Float64 | 指标值 |
| source | String | 数据源 |
| service | String | 服务名称 |
| host | String | 主机名称 |

### 3.2 日志数据表

**logs（日志数据表）**

```sql
CREATE TABLE logs (
    timestamp DateTime,
    level String,
    message String,
    source String,
    service String,
    host String,
    trace_id String,
    labels Map(String, String),
    parsed_fields Map(String, String)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, source)
SETTINGS index_granularity = 8192;
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| timestamp | DateTime | 时间戳 |
| level | String | 日志级别 |
| message | String | 日志内容 |
| source | String | 日志源 |
| service | String | 服务名称 |
| host | String | 主机名称 |
| trace_id | String | 链路追踪ID |
| labels | Map | 标签 |
| parsed_fields | Map | 解析后的字段 |

### 3.3 Agent执行指标表

**agent_metrics（Agent性能指标表）**

```sql
CREATE TABLE agent_metrics (
    timestamp DateTime,
    agent_name String,
    action String,
    llm_model String,
    tokens_input UInt32,
    tokens_output UInt32,
    duration_ms UInt32,
    success UInt8,
    error_type String,
    cost_usd Float64
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (agent_name, timestamp)
SETTINGS index_granularity = 8192;
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| timestamp | DateTime | 时间戳 |
| agent_name | String | Agent名称 |
| action | String | 执行动作 |
| llm_model | String | LLM模型 |
| tokens_input | UInt32 | 输入Token数 |
| tokens_output | UInt32 | 输出Token数 |
| duration_ms | UInt32 | 执行耗时 |
| success | UInt8 | 是否成功（0/1） |
| error_type | String | 错误类型 |
| cost_usd | Float64 | 成本（美元） |

### 3.4 告警详细历史表

**alert_details（告警详细历史表）**

```sql
CREATE TABLE alert_details (
    timestamp DateTime,
    fingerprint String,
    severity String,
    title String,
    description String,
    labels Map(String, String),
    source String,
    service String,
    deduplicated UInt8,
    aggregated UInt8,
    agent_processed UInt8
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (fingerprint, timestamp)
SETTINGS index_granularity = 8192;
```

## 四、Milvus向量数据库设计

### 4.1 Collections设计

#### 4.1.1 故障案例Collection

**incident_cases**

```python
from pymilvus import Collection, FieldSchema, CollectionSchema, DataType

fields = [
    FieldSchema(name="id", dtype=DataType.INT64, is_primary=True, auto_id=True),
    FieldSchema(name="case_id", dtype=DataType.VARCHAR, max_length=100),
    FieldSchema(name="title", dtype=DataType.VARCHAR, max_length=500),
    FieldSchema(name="description", dtype=DataType.VARCHAR, max_length=2000),
    FieldSchema(name="root_cause", dtype=DataType.VARCHAR, max_length=1000),
    FieldSchema(name="solution", dtype=DataType.VARCHAR, max_length=2000),
    FieldSchema(name="embedding", dtype=DataType.FLOAT_VECTOR, dim=3072),
    FieldSchema(name="service", dtype=DataType.VARCHAR, max_length=100),
    FieldSchema(name="severity", dtype=DataType.VARCHAR, max_length=20),
    FieldSchema(name="timestamp", dtype=DataType.INT64),
    FieldSchema(name="metadata", dtype=DataType.JSON),
]

schema = CollectionSchema(fields, "故障案例库")
collection = Collection("incident_cases", schema)

# 创建索引
index_params = {
    "metric_type": "IP",
    "index_type": "IVF_FLAT",
    "params": {"nlist": 1024}
}
collection.create_index(field_name="embedding", index_params=index_params)
```

**向量维度**：3072（OpenAI text-embedding-3-large）

#### 4.1.2 运维手册Collection

**operation_manuals**

```python
fields = [
    FieldSchema(name="id", dtype=DataType.INT64, is_primary=True, auto_id=True),
    FieldSchema(name="manual_id", dtype=DataType.VARCHAR, max_length=100),
    FieldSchema(name="title", dtype=DataType.VARCHAR, max_length=500),
    FieldSchema(name="content", dtype=DataType.VARCHAR, max_length=5000),
    FieldSchema(name="category", dtype=DataType.VARCHAR, max_length=100),
    FieldSchema(name="embedding", dtype=DataType.FLOAT_VECTOR, dim=3072),
    FieldSchema(name="tags", dtype=DataType.JSON),
    FieldSchema(name="created_at", dtype=DataType.INT64),
]

schema = CollectionSchema(fields, "运维手册知识库")
collection = Collection("operation_manuals", schema)
```

#### 4.1.3 Agent记忆Collection

**agent_memories**

```python
fields = [
    FieldSchema(name="id", dtype=DataType.INT64, is_primary=True, auto_id=True),
    FieldSchema(name="memory_id", dtype=DataType.VARCHAR, max_length=100),
    FieldSchema(name="agent_name", dtype=DataType.VARCHAR, max_length=100),
    FieldSchema(name="content", dtype=DataType.VARCHAR, max_length=2000),
    FieldSchema(name="memory_type", dtype=DataType.VARCHAR, max_length=50),
    FieldSchema(name="embedding", dtype=DataType.FLOAT_VECTOR, dim=3072),
    FieldSchema(name="importance", dtype=DataType.FLOAT),
    FieldSchema(name="timestamp", dtype=DataType.INT64),
    FieldSchema(name="metadata", dtype=DataType.JSON),
]

schema = CollectionSchema(fields, "Agent记忆向量库")
collection = Collection("agent_memories", schema)
```

### 4.2 向量检索策略

**检索配置**：

```python
search_params = {
    "metric_type": "IP",  # Inner Product
    "params": {"nprobe": 10}
}

# 检索示例
results = collection.search(
    data=[query_embedding],
    anns_field="embedding",
    param=search_params,
    limit=5,
    expr='agent_name == "analysis_agent"',  # 过滤条件
    output_fields=["title", "description", "root_cause"]
)
```

**检索策略**：
- Top-K：返回5-10个最相似结果
- 阈值过滤：相似度 > 0.75
- 分类别检索：按service、severity过滤
- Rerank：使用Cross-Encoder重排序（可选）

## 五、Redis数据结构设计

### 5.1 Agent状态存储

**数据结构**：Hash

```redis
Key: agent_state:{session_id}

Fields:
{
    "session_id": "uuid",
    "current_agent": "analysis_agent",
    "status": "running",
    "context": "json_string",
    "last_updated": "timestamp"
}

TTL: 86400秒（24小时）
```

### 5.2 会话短期记忆

**数据结构**：Hash

```redis
Key: session:{session_id}:memory

Fields:
{
    "user_input": "订单服务响应慢",
    "intent": "故障处理",
    "collected_data": "json_string",
    "analysis_result": "json_string"
}

TTL: 3600秒（1小时）
```

### 5.3 Agent通信队列

**数据结构**：List

```redis
Key: agent_queue:{agent_name}

操作：
- LPUSH: 添加任务
- RPOP: 获取任务
- LLEN: 队列长度

用途：任务调度
```

### 5.4 告警去重缓存

**数据结构**：Set

```redis
Key: alert_fingerprints:{time_window}

Members: 告警指纹列表

TTL: 300秒（5分钟窗口）

用途：告警去重判断
```

### 5.5 用户会话缓存

**数据结构**：String

```redis
Key: user_session:{user_id}

Value: JWT Token或会话信息

TTL: 7200秒（2小时）
```

### 5.6 API调用计数

**数据结构**：String

```redis
Key: api_call_count:{user_id}:{date}

Value: 调用次数

TTL: 86400秒

用途：Rate Limiting
```

## 六、数据同步策略

### 6.1 PostgreSQL与ClickHouse同步

**同步场景**：
- 告警数据：从PostgreSQL同步到ClickHouse
- Agent日志：从PostgreSQL同步到ClickHouse

**同步方式**：
- 定时任务：每5分钟同步一次
- CDC（Change Data Capture）：实时同步（可选）

### 6.2 PostgreSQL与Milvus同步

**同步场景**：
- 知识条目：新增时向量化并存入Milvus
- Agent记忆：新增时向量化并存入Milvus

**同步方式**：
- 事件触发：插入PostgreSQL时，同时向量化并插入Milvus

### 6.3 Redis状态清理

**清理策略**：
- 定时任务：每小时清理过期状态
- TTL机制：Redis自动清理过期Key

## 七、数据备份策略

### 7.1 PostgreSQL备份

**备份方式**：
- 全量备份：每日一次
- 增量备份：每小时一次
- 保留时长：30天

**备份工具**：
- pg_dump（全量）
- WAL归档（增量）

### 7.2 ClickHouse备份

**备份方式**：
- 分区备份：每日备份新分区
- 冷数据迁移：30天前数据迁移到冷存储

**备份工具**：
- clickhouse-backup

### 7.3 Milvus备份

**备份方式**：
- 定时导出：每日导出向量数据
- Binlog备份：实时备份（可选）

### 7.4 Redis备份

**备份方式**：
- RDB快照：每6小时一次
- AOF日志：实时记录

## 八、数据安全

### 8.1 数据加密

**加密场景**：
- 密码：bcrypt哈希
- API Key：AES-256加密
- 数据源凭证：AES-256加密

**加密实现**：

```go
// Go加密示例
import "crypto/aes"

func EncryptCredential(plaintext string, key []byte) string {
    block, err := aes.NewCipher(key)
    if err != nil {
        panic(err)
    }
    // ... 加密逻辑
}
```

### 8.2 数据脱敏

**脱敏场景**：
- 日志中敏感信息
- Agent记忆中的敏感数据
- 导出的报告

**脱敏规则**：
- IP地址：192.168.1.100 → 192.168.1.xxx
- 用户名：张三 → 张*
- 手机号：13812345678 → 138****5678

### 8.3 访问控制

**访问策略**：
- 数据库用户权限分离
- 应用层权限控制（RBAC）
- API访问审计

## 九、数据迁移计划

### 9.1 迁移场景

**从现有系统迁移**：
- 历史告警数据
- 监控配置
- 用户数据

### 9.2 迁移步骤

1. **数据评估**：评估数据量和质量
2. **数据清洗**：清洗不完整数据
3. **数据转换**：转换为新系统格式
4. **数据导入**：批量导入新系统
5. **数据验证**：验证迁移完整性

### 9.3 迁移工具

```python
# 数据迁移脚本示例
import psycopg2
from pymilvus import Collection

def migrate_alerts():
    # 从旧系统读取告警
    old_alerts = read_old_alerts()
    
    # 转换格式
    new_alerts = transform_alerts(old_alerts)
    
    # 导入新系统
    insert_new_alerts(new_alerts)
    
    # 验证
    verify_migration(old_alerts, new_alerts)
```

## 十、性能优化建议

### 10.1 PostgreSQL优化

- 合理使用索引
- 大表分区
- 定期VACUUM
- 查询优化

### 10.2 ClickHouse优化

- 合理分区
- 索引优化
- 查询优化（避免全表扫描）
- 数据压缩

### 10.3 Milvus优化

- 选择合适的索引类型
- 批量插入
- 定期Compact
- 预热查询

### 10.4 Redis优化

- 合理设置TTL
- 避免大Key
- 使用Pipeline批量操作
- 监控内存使用

## 十一、参考资料

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [ClickHouse Documentation](https://clickhouse.com/docs/)
- [Milvus Documentation](https://milvus.io/docs)
- [Redis Documentation](https://redis.io/documentation)

## 十二、附录

### 12.1 数据量估算

| 数据类型 | 日增量 | 月增量 | 年增量 |
|---------|--------|--------|--------|
| 监控指标 | 1TB | 30TB | 360TB |
| 日志数据 | 500GB | 15TB | 180TB |
| 告警历史 | 100万条 | 3000万条 | 3.6亿条 |
| Agent日志 | 10万条 | 300万条 | 3600万条 |
| 向量数据 | 100条 | 3000条 | 3.6万条 |

### 12.2 更新记录

| 版本 | 日期 | 更新内容 |
|------|------|----------|
| v1.0.0 | 2026-06-24 | 初稿 |
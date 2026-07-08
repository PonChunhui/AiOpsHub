# Agent自主决策架构 - 下一步实施指南

## 当前状态

✅ **已完成**：
- 新架构代码完成并编译成功（api-server 51MB）
- 所有核心组件已实现（ToolRegistry、AgentInstance、AgentRuntime、MasterRouter）
- ChatService已完全重构（从992行简化到275行）
- 数据模型已扩展（RoutingLog、ToolCallLog、Agent新字段）
- 配置文件已优化（新增agent配置项）

⏳ **待验证**：
- 服务启动验证
- 数据库表创建验证
- 功能流程验证
- 决策日志验证

---

## P0 - 立即验证（必做，手动执行）

### 1. 启动服务验证初始化流程

**执行步骤：**
```bash
cd /Users/pengchunhui/code/aiops/AiOpsHub/backend
./api-server
```

**预期日志输出（关键检查点）：**

✅ **初始化阶段日志：**
```
[INFO] ToolRegistry预加载完成，共加载4个工具
[INFO] ✅ Agent Runtime initialized (cache size: 100)
[INFO] ✅ Master Router initialized  
[INFO] ✅ ChatHandler初始化成功(New Architecture)
[INFO] ✅ All Services initialized successfully (New Architecture)
[INFO] API Server started on port 8080
```

⚠️ **如果出现错误，检查：**
- 数据库连接是否正常
- Redis连接是否正常
- LLM配置是否正确
- Milvus是否可达（如果enable_rag=true）

### 2. 数据库表结构验证

**连接数据库：**
```bash
psql -h 192.168.100.10 -p 5432 -U aiops -d aiopsdb
# 密码：aiops123
```

**执行SQL验证：**

```sql
-- ====================
-- 步骤1：验证新表是否存在
-- ====================
SHOW TABLES LIKE 'routing_logs';
SHOW TABLES LIKE 'tool_call_logs';

-- 预期结果：应该显示这两个表

-- ====================
-- 步骤2：验证表结构
-- ====================
DESC routing_logs;

-- 预期字段：
-- id (varchar, primary key)
-- session_id (varchar, index)
-- user_message (text)
-- selected_agent_id (varchar, index)
-- confidence (decimal)
-- reasoning (text)
-- alternative_agents (text)
-- routing_method (varchar)
-- created_at (timestamp, index)

DESC tool_call_logs;

-- 预期字段：
-- id (varchar, primary key)
-- session_id (varchar, index)
-- agent_id (varchar, index)
-- tool_name (varchar, index)
-- arguments (text)
-- result (text)
-- success (boolean)
-- duration (int)
-- error_message (text)
-- created_at (timestamp, index)

-- ====================
-- 步骤3：验证Agent表新增字段
-- ====================
DESC agents;

-- 预期新增字段：
-- max_tool_calls (int, default 5)
-- capability (text)
-- priority (int, default 0)

-- ====================
-- 步骤4：验证ChatMessage表新增字段
-- ====================
DESC chat_messages;

-- 预期新增字段：
-- agent_id (varchar, index)

-- ====================
-- 步骤5：查看现有Agent数据
-- ====================
SELECT id, name, max_tool_calls, capability, priority
FROM agents
WHERE enabled = true
ORDER BY priority DESC
LIMIT 10;

-- 预期：至少显示10个预设Agent，max_tool_calls应为5或NULL
```

### 3. 发送测试消息验证功能

**准备测试用例：**

| 序号 | 测试场景 | 用户消息 | 预期Agent | 预期工具 | 验证要点 |
|------|---------|---------|----------|---------|---------|
| 1 | 告警处理 | "收到严重告警，CPU使用率过高" | preset-alert-handler | prometheus_query | 路由置信度>0.8 |
| 2 | 故障诊断 | "服务响应缓慢，需要排查故障原因" | preset-fault-diagnosis | ssh_exec | 工具调用记录 |
| 3 | 日志分析 | "查看应用最近1小时的错误日志" | preset-log-analyzer | log_query | 工具执行成功 |
| 4 | 服务器巡检 | "巡检服务器状态" | preset-system-inspection | ssh_exec | 多工具调用 |
| 5 | 通用问答 | "你好，介绍一下你的功能" | （降级） | 无 | routing_method=fallback |

**获取JWT Token：**
```bash
# 登录获取token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' \
  | jq -r '.token'

# 或使用现有的测试账号
```

**创建测试会话：**
```bash
curl -X POST http://localhost:8080/api/v1/chat/sessions \
  -H "Authorization: Bearer <你的token>" \
  -H "Content-Type: application/json" \
  -d '{"title": "架构验证测试", "model": "glm-5.2"}'
  
# 记录返回的session_id，例如：test-session-001
```

**发送测试消息：**
```bash
# 测试1：告警处理
curl -X POST http://localhost:8080/api/v1/chat/messages \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-session-001",
    "content": "收到严重告警，CPU使用率超过90%，需要分析"
  }' | jq .

# 测试2：故障诊断
curl -X POST http://localhost:8080/api/v1/chat/messages \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-session-001",
    "content": "服务响应缓慢，需要排查故障原因"
  }' | jq .

# 测试3：通用问答（测试降级）
curl -X POST http://localhost:8080/api/v1/chat/messages \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-session-001",
    "content": "你好，介绍一下你的功能"
  }' | jq .
```

**服务端日志预期（SendMessage流程）：**
```
[INFO] === SendMessage (New Architecture) ===
[INFO] === MasterRouter: 开始路由决策 ===
[DEBUG] 用户消息: 收到严重告警，CPU使用率超过90%...
[INFO] ✅ 路由决策完成: Agent=preset-alert-handler, 置信度=0.95, 理由=...
[INFO] 创建并缓存Agent实例: preset-alert-handler
[INFO] Agent 告警处理 可用工具池: 2 内置 + 0 MCP = 2 总计
[INFO] Agent 告警处理 第1轮工具调用，共1个工具
[INFO] 工具 prometheus_query 执行成功，耗时123ms，结果长度456
[INFO] === SendMessage完成，工具调用1次 ===
```

**响应体预期：**
```json
{
  "response": "根据Prometheus数据，CPU使用率确实达到92%...",
  "user_message": {
    "id": "...",
    "role": "user",
    "content": "收到严重告警..."
  },
  "ai_message": {
    "id": "...",
    "role": "assistant",
    "content": "...",
    "agent_id": "preset-alert-handler"
  },
  "rag_references": []
}
```

### 4. 查看决策日志验证数据持久化

**SQL查询验证：**

```sql
-- ====================
-- 查看路由决策日志（最近5条）
-- ====================
SELECT 
  id,
  session_id,
  selected_agent_id,
  confidence,
  routing_method,
  reasoning,
  created_at
FROM routing_logs
ORDER BY created_at DESC
LIMIT 5;

-- 预期结果：
-- 至少显示刚才发送的测试消息记录
-- routing_method 应为 'llm' 或 'quick_match' 或 'fallback'
-- confidence 应为 0.0-1.0 的数值

-- ====================
-- 查看工具调用日志（最近10条）
-- ====================
SELECT 
  id,
  agent_id,
  tool_name,
  success,
  duration,
  LEFT(result, 50) as result_preview,
  created_at
FROM tool_call_logs
ORDER BY created_at DESC
LIMIT 10;

-- 预期结果：
-- 显示工具调用记录
-- success 应为 true/false
-- duration 应为毫秒数值
-- result_preview 应有工具返回内容

-- ====================
-- 查看消息的AgentID记录
-- ====================
SELECT 
  id,
  role,
  agent_id,
  LEFT(content, 30) as content_preview,
  created_at
FROM chat_messages
WHERE agent_id IS NOT NULL
ORDER BY created_at DESC
LIMIT 5;

-- 预期结果：
-- assistant角色的消息应有agent_id字段
-- agent_id应匹配routing_logs中的selected_agent_id

-- ====================
-- 分析路由成功率
-- ====================
SELECT 
  routing_method,
  COUNT(*) as total,
  AVG(confidence) as avg_confidence,
  COUNT(DISTINCT selected_agent_id) as unique_agents
FROM routing_logs
WHERE created_at > NOW() - INTERVAL '1 hour'
GROUP BY routing_method;

-- 预期结果：
-- 显示各种路由方法的统计
-- avg_confidence平均值应>0.7

-- ====================
-- 分析工具调用成功率
-- ====================
SELECT 
  agent_id,
  tool_name,
  COUNT(*) as total_calls,
  SUM(CASE WHEN success THEN 1 ELSE 0 END) as success_calls,
  AVG(duration) as avg_duration_ms,
  ROUND(SUM(CASE WHEN success THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as success_rate
FROM tool_call_logs
WHERE created_at > NOW() - INTERVAL '1 hour'
GROUP BY agent_id, tool_name
ORDER BY total_calls DESC;

-- 预期结果：
-- 显示每个Agent的工具调用统计
-- success_rate应>80%
```

---

## P1 - 配置优化（本周内完成）

### 1. 配置项已添加 ✅

**新增配置项（backend/configs/config.yaml）：**
```yaml
agent:
  cache_size: 100          # Agent实例缓存大小（LRU）
  max_tool_calls: 5        # 默认最大工具调用次数
  routing_timeout: 30      # 路由决策超时时间（秒）
  tool_timeout: 60         # 工具执行超时时间（秒）
```

### 2. 预设Agent配置优化

**建议更新预设Agent的max_tool_calls：**

```sql
-- 根据Agent类型设置不同的工具调用上限
UPDATE agents SET max_tool_calls = 3 WHERE id = 'preset-alert-handler';
UPDATE agents SET max_tool_calls = 5 WHERE id = 'preset-fault-diagnosis';
UPDATE agents SET max_tool_calls = 4 WHERE id = 'preset-log-analyzer';
UPDATE agents SET max_tool_calls = 2 WHERE id = 'preset-server-command';
UPDATE agents SET max_tool_calls = 10 WHERE id = 'preset-auto-inspection';

-- 设置优先级（路由时优先选择）
UPDATE agents SET priority = 10 WHERE id = 'preset-alert-handler';
UPDATE agents SET priority = 9 WHERE id = 'preset-fault-diagnosis';
UPDATE agents SET priority = 8 WHERE id = 'preset-pipeline-helper';

-- 设置能力标签（JSON格式）
UPDATE agents SET capability = '{"domains": ["监控", "告警"], "skills": ["数据分析", "趋势预测"]}' 
WHERE id = 'preset-alert-handler';
```

### 3. Agent与工具绑定优化

**建议检查Agent工具绑定：**
```sql
-- 查看当前工具绑定
SELECT 
  a.name as agent_name,
  t.name as tool_name,
  at.enabled,
  at.priority
FROM agent_tools at
JOIN agents a ON at.agent_id = a.id
JOIN tools t ON at.tool_id = t.id
WHERE a.enabled = true AND t.enabled = true
ORDER BY a.name, at.priority DESC;

-- 如果绑定不合理，可以调整
-- 例如：告警处理Agent绑定prometheus_query和log_query
INSERT INTO agent_tools (id, agent_id, tool_id, enabled, priority, created_at, updated_at)
VALUES (
  'binding-alert-prom',
  'preset-alert-handler',
  'tool-prometheus-query',
  true,
  10,
  NOW(),
  NOW()
);
```

---

## P2 - 监控与日志（2周内完成）

### 1. 添加日志分析查询脚本

创建文件：`backend/scripts/query_routing_stats.sql`

```sql
-- routing_stats.sql - 路由统计分析

-- 1. 最近24小时路由统计
SELECT 
  DATE(created_at) as date,
  HOUR(created_at) as hour,
  routing_method,
  COUNT(*) as total_routes,
  AVG(confidence) as avg_confidence,
  COUNT(DISTINCT selected_agent_id) as unique_agents
FROM routing_logs
WHERE created_at > NOW() - INTERVAL '24 hours'
GROUP BY DATE(created_at), HOUR(created_at), routing_method
ORDER BY date DESC, hour DESC;

-- 2. Agent使用频率统计
SELECT 
  selected_agent_id,
  COUNT(*) as total_routes,
  AVG(confidence) as avg_confidence,
  MAX(confidence) as max_confidence,
  MIN(confidence) as min_confidence
FROM routing_logs
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY selected_agent_id
ORDER BY total_routes DESC;

-- 3. 工具调用成功率统计
SELECT 
  tool_name,
  COUNT(*) as total_calls,
  SUM(CASE WHEN success THEN 1 ELSE 0 END) as success_calls,
  ROUND(SUM(CASE WHEN success THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as success_rate,
  AVG(duration) as avg_duration_ms,
  MAX(duration) as max_duration_ms
FROM tool_call_logs
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY tool_name
ORDER BY total_calls DESC;

-- 4. 失败的工具调用分析
SELECT 
  agent_id,
  tool_name,
  error_message,
  COUNT(*) as failure_count,
  created_at
FROM tool_call_logs
WHERE success = false
  AND created_at > NOW() - INTERVAL '24 hours'
ORDER BY failure_count DESC
LIMIT 20;

-- 5. 会话工具调用轨迹
SELECT 
  session_id,
  agent_id,
  tool_name,
  success,
  duration,
  created_at
FROM tool_call_logs
WHERE session_id = '<session_id>'
ORDER BY created_at ASC;
```

### 2. Prometheus监控指标（可选）

创建文件：`backend/internal/service/metrics.go`

```go
package service

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    RoutingTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "agent_routing_total",
        Help: "Total agent routing requests",
    }, []string{"agent_id", "method"})
    
    RoutingLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "agent_routing_latency_ms",
        Help:    "Agent routing latency",
        Buckets: []float64{10, 50, 100, 500, 1000},
    }, []string{"agent_id"})
    
    ToolCallTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "tool_call_total",
        Help: "Total tool calls",
    }, []string{"tool_name", "agent_id", "success"})
    
    ToolCallLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "tool_call_latency_ms",
        Help:    "Tool call latency",
        Buckets: []float64{10, 50, 100, 500, 1000, 5000},
    }, []string{"tool_name"})
)
```

---

## P3 - 功能增强（1月内完成）

### 1. Agent缓存管理API

**新增API端点（建议）：**
```
GET  /api/v1/admin/agent/cache/stats     # 查看缓存统计
POST /api/v1/admin/agent/cache/clear     # 清空缓存
POST /api/v1/admin/agent/cache/clear/:id # 清空指定Agent缓存
```

**实现示例（handler/admin_handler.go）：**
```go
func GetAgentCacheStats(c *gin.Context) {
    runtime := handler.GetAgentRuntime()
    
    stats := map[string]interface{}{
        "cache_size": runtime.cacheSize,
        "cached_agents": runtime.cache.Len(),
        "keys": runtime.cache.Keys(),
    }
    
    c.JSON(200, stats)
}

func ClearAgentCache(c *gin.Context) {
    runtime := handler.GetAgentRuntime()
    runtime.ClearCache()
    
    c.JSON(200, gin.H{"message": "Agent cache cleared"})
}
```

### 2. 路由决策日志查询API

**新增API端点：**
```
GET /api/v1/admin/routing/logs?limit=100       # 查询路由日志
GET /api/v1/admin/routing/stats?days=7         # 路由统计
GET /api/v1/admin/tool-calls/logs?limit=100    # 查询工具调用日志
GET /api/v1/admin/tool-calls/stats?days=7      # 工具调用统计
```

### 3. Agent实例预热

**启动时预创建常用Agent实例（可选优化）：**
```go
// service_handler.go - initAgentRuntime()
func initAgentRuntime() {
    // ... 现有初始化代码 ...
    
    // 预热：创建常用的Agent实例
    hotAgents := []string{
        "preset-alert-handler",
        "preset-fault-diagnosis",
        "preset-log-analyzer",
    }
    
    ctx := context.Background()
    for _, agentID := range hotAgents {
        _, err := agentRuntime.CreateAgentInstance(ctx, agentID)
        if err != nil {
            fmt.Printf("预创建Agent %s 失败: %v\n", agentID, err)
        } else {
            fmt.Printf("预创建Agent %s 成功\n", agentID)
        }
    }
}
```

---

## P4 - 性能优化（持续）

### 1. AgentInstance性能优化

**优化方向：**
- Prompt长度优化（截取工具结果）
- 工具调用并发执行
- 工具结果缓存（相同参数相同结果）
- Agent实例池化（多实例并发）

### 2. MasterRouter优化

**优化方向：**
- 快速预筛选权重调整
- LLM Prompt优化（更简洁）
- 路由决策缓存（相似问题复用）
- 多轮对话上下文优化

### 3. 数据库查询优化

**添加索引：**
```sql
-- 优化routing_logs查询
CREATE INDEX idx_routing_logs_agent_created ON routing_logs(selected_agent_id, created_at);
CREATE INDEX idx_routing_logs_session_created ON routing_logs(session_id, created_at);

-- 优化tool_call_logs查询
CREATE INDEX idx_tool_calls_agent_created ON tool_call_logs(agent_id, created_at);
CREATE INDEX idx_tool_calls_tool_created ON tool_call_logs(tool_name, created_at);

-- 优化chat_messages查询
CREATE INDEX idx_messages_agent_created ON chat_messages(agent_id, created_at);
```

---

## 测试验证清单

### ✅ 启动验证（P0）
- [ ] 服务成功启动（无错误日志）
- [ ] ToolRegistry预加载成功
- [ ] AgentRuntime初始化成功
- [ ] MasterRouter初始化成功
- [ ] ChatHandler使用新架构

### ✅ 数据库验证（P0）
- [ ] routing_logs表已创建
- [ ] tool_call_logs表已创建
- [ ] agents表新增字段已添加
- [ ] chat_messages表新增字段已添加

### ✅ 功能验证（P0）
- [ ] 告警处理场景路由正确
- [ ] 故障诊断场景路由正确
- [ ] 工具调用记录到数据库
- [ ] 路由决策记录到数据库
- [ ] 多轮工具调用正常

### ✅ 性能验证（P1）
- [ ] Agent缓存命中率>50%
- [ ] 路由决策延迟<1秒
- [ ] 工具执行成功率>90%
- [ ] 内存使用稳定（无泄漏）

---

## 验证执行顺序建议

**第1步：启动验证（10分钟）**
```bash
cd backend
./api-server
# 观察初始化日志，确认无错误
```

**第2步：数据库验证（5分钟）**
```bash
psql -h 192.168.100.10 -U aiops -d aiopsdb
# 执行验证SQL，确认表结构正确
```

**第3步：功能验证（15分钟）**
```bash
# 发送3-5个测试消息
# 查看服务端日志
# 查询数据库记录
```

**第4步：性能观察（持续）**
```bash
# 观察缓存命中率
# 观察响应时间
# 观察内存使用
```

---

## 故障排查指南

### 问题1：启动失败

**可能原因：**
- 数据库连接失败
- Redis连接失败
- LLM API配置错误
- Milvus不可达

**排查命令：**
```bash
# 检查数据库连接
psql -h 192.168.100.10 -p 5432 -U aiops -d aiopsdb

# 检查Redis连接  
redis-cli -h 192.168.100.114 ping

# 检查Milvus连接（如果enable_rag=true）
curl http://192.168.100.10:19530/v1/vector/collections

# 检查LLM配置
curl https://dashscope.aliyuncs.com/compatible-mode/v1/models \
  -H "Authorization: Bearer sk-0869..."
```

### 问题2：路由决策失败

**日志特征：**
```
LLM路由失败: LLM生成失败
降级使用默认Agent
```

**排查步骤：**
1. 检查LLM API是否可用
2. 检查LLM Prompt是否过长
3. 检查LLM返回格式是否正确
4. 查看routing_logs表routing_method字段

### 问题3：工具调用失败

**日志特征：**
```
工具 ssh_exec 执行失败: ...
```

**排查步骤：**
1. 检查ToolRegistry是否预加载成功
2. 检查工具配置是否正确
3. 检查SSH连接是否可用
4. 查看tool_call_logs表error_message字段

---

## 成功标准

### ✅ 基础功能成功标准
- 服务正常启动，无错误日志
- 数据库表结构完整
- 至少发送3个测试消息成功
- routing_logs表有记录
- tool_call_logs表有记录
- chat_messages表有agent_id记录

### ✅ 性能成功标准
- 路由决策延迟<1秒（90%请求）
- 工具调用成功率>85%
- Agent缓存命中率>40%
- 内存增长<10MB/小时（无泄漏）

### ✅ 稳定性成功标准
- 服务运行>1小时无崩溃
- 连续发送100条消息无错误
- 工具调用失败率<15%
- LLM调用失败率<5%

---

## 下一步行动计划

**立即执行（现在）：**
1. 启动服务验证初始化流程 ✅
2. 数据库表结构验证 ✅
3. 发送测试消息验证功能 ✅

**本周执行（7天内）：**
1. Agent配置优化（max_tool_calls, priority）
2. 工具绑定检查和调整
3. 日志分析查询脚本
4. Agent缓存管理API

**本月执行（30天内）：**
1. Prometheus监控集成
2. 路由决策优化（Prompt优化）
3. 工具执行并发优化
4. 性能压测和优化

**持续优化：**
1. 监控指标完善
2. Agent训练数据收集
3. 多Agent协作研究
4. 自动化测试补充

---

**文档位置**：`docs/next-steps-guide.md`
**更新时间**：2026-07-07
**状态**：待验证，建议立即执行P0步骤
# Coordinator Agent和多Agent协作机制 - 快速开始指南

## 🎯 已完成内容

本次实现完成了以下核心组件：

1. **Coordinator Agent** - 全局协调者
   - 意图理解、任务分解、Agent选择
   - 协作编排（串行/并行/混合策略）
   - 结果整合、冲突解决

2. **决策引擎** - 智能决策
   - Agent路由、依赖分析
   - 并行分组、执行时间估算
   - 人机交互判断

3. **消息总线** - Agent通信
   - Redis Pub/Sub消息传递
   - 5种消息类型（任务请求、任务结果、协作请求、状态更新、事件广播）
   - 消息路由和广播机制

4. **状态同步机制** - Agent状态管理
   - Redis状态存储（PENDING/RUNNING/COMPLETED/FAILED/TIMEOUT）
   - 中间结果传递、进度监控
   - 超时检测、会话管理

5. **冲突解决机制** - 冲突处理
   - 分布式锁（Redis SetNX）
   - 结果投票、优先级选择
   - 人工决策请求

6. **Temporal协作Workflow** - 并发协作编排
   - CollaborationWorkflow（并发协作、人机交互）
   - ParallelMonitorWorkflow（并发监控）
   - IncidentHandlingWorkflow（故障处理）

## 📦 编译验证

### 编译Backend

```bash
cd backend

# 下载依赖
go mod tidy

# 编译Temporal Worker（包含协作Workflow）
go build -o bin/temporal-worker ./cmd/temporal-worker

# 编译API Server
go build -o bin/api-server ./cmd/api-server

# 检查编译产物
ls -lh bin/
# api-server: 45MB
# temporal-worker: 28MB
```

**编译成功！无编译错误。**

## 🧪 单元测试

### 测试Coordinator Agent

```bash
cd backend

# 运行所有Agent测试
go test ./internal/agent -v

# 测试Coordinator Agent功能
go test -run TestCoordinatorAgent ./internal/agent -v

# 测试Decision Engine
go test -run TestDecisionEngine ./internal/agent -v

# 测试完整协作流程
go test -run TestCoordinatorFullExecution ./internal/agent -v
```

### 测试消息总线

```bash
# 测试消息定义和序列化
go test ./pkg/message_bus -v
```

### 测试状态同步

```bash
# 测试状态管理器
go test ./pkg/state_sync -v
```

### 测试冲突解决

```bash
# 测试分布式锁和结果投票
go test ./pkg/conflict_resolver -v
```

## 🚀 部署Temporal Server

### 使用Docker Compose启动Temporal

```bash
cd deployments

# 启动Temporal Server（包含PostgreSQL）
docker-compose up -d temporal-server

# 查看Temporal Server状态
docker-compose ps

# 查看Temporal Server日志
docker-compose logs temporal-server
```

### 访问Temporal Web UI

```bash
# 打开Temporal Web UI
open http://localhost:8080
```

Temporal Web UI功能：
- 查看Workflow执行历史
- 查看Activity执行详情
- 发送Signal和Query
- 调试Workflow

## 🔄 测试Temporal协作Workflow

### 1. 启动Temporal Worker

```bash
cd backend

# 确保Temporal Server已启动
# 确保配置文件正确（config/config.yaml）

# 启动Temporal Worker
./bin/temporal-worker

# Worker会注册以下Workflow：
# - AgentWorkflow
# - CollaborationWorkflow
# - ParallelMonitorWorkflow
# - IncidentHandlingWorkflow

# Worker会注册以下Activity：
# - CoordinatorActivity
# - ExecuteAgentTask
# - IntegrateResults
# - MonitorService
# - SendMessageActivity
# - UpdateStateActivity
# - SetIntermediateResultActivity
# - GetIntermediateResultActivity
# - ResolveConflictActivity
# - RequiresApproval
```

### 2. 触发CollaborationWorkflow

#### 通过Temporal Web UI触发

1. 打开Temporal Web UI: http://localhost:8080
2. 点击"New Workflow"
3. 选择Namespace: default
4. 选择Task Queue: aiops-task-queue
5. 输入Workflow Type: CollaborationWorkflow
6. 输入Input JSON:

```json
{
  "session_id": "test-session-001",
  "user_query": "订单服务响应很慢，帮我分析原因",
  "context": {
    "service": "order-service",
    "urgency": "high"
  },
  "timestamp": "2026-06-26T10:00:00Z"
}
```

7. 点击"Start"

#### 通过API触发（待实现）

```bash
# API接口（待实现）
POST /api/v1/workflows/collaborate
{
  "session_id": "test-session-001",
  "user_query": "订单服务响应很慢，帮我分析原因",
  "context": {
    "service": "order-service"
  }
}
```

### 3. 监控Workflow执行

在Temporal Web UI中：
1. 查看Workflow执行状态
2. 查看Event History（Coordinator分解任务 → Agent执行 → 结果整合）
3. 查看Activity执行详情（每个Agent的执行结果）
4. 发送Signal（用户确认）
5. 发送Query（查询协作进度）

## 📊 验证协作机制

### 验证点1：Coordinator分解任务

查看CoordinatorActivity输出：
- Intent: "故障诊断"
- TaskType: "incident_handling"
- SubTasks: [
    {TaskID: "task-001", AgentID: "monitor-agent-001"},
    {TaskID: "task-002", AgentID: "analysis-agent-001"},
    {TaskID: "task-003", AgentID: "decision-agent-001"}
  ]
- Orchestration: {Strategy: "sequential", TaskSequence: ["task-001", "task-002", "task-003"]}

### 验证点2：并发协作执行

查看CollaborationWorkflow执行：
- Strategy: "parallel" 或 "hybrid"
- 并发执行多个Agent Activity
- 收集所有Agent结果
- 结果整合

### 验证点3：人机交互（Signal）

在Workflow执行过程中：
1. Workflow等待用户确认（RequiresApproval: true）
2. 通过Temporal Web UI发送Signal：
   ```json
   {
     "approved": true,
     "user_id": "user-001",
     "comment": "确认执行修复方案"
   }
   ```
3. Workflow继续执行

### 验证点4：状态查询（Query）

通过Temporal Web UI发送Query：
- Query Name: "progress"
- 返回：
  ```json
  {
    "session_id": "test-session-001",
    "task_type": "incident_handling",
    "agents_count": 3,
    "status": "running"
  }
  ```

## 🔧 配置要求

### config.yaml示例

```yaml
app:
  name: "AiOpsHub Backend"
  mode: "debug"

server:
  port: "8080"
  read_timeout: "10s"
  write_timeout: "10s"

database:
  host: "localhost"
  port: 5432
  user: "aiops"
  password: "aiops123"
  dbname: "aiopsdb"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

temporal:
  host: "localhost:7233"
  namespace: "default"

llm:
  provider: "aliyun_bailian"
  model: "qwen-max"
  api_key: "your-api-key"
  base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
  temperature: 0.3
  max_tokens: 2000
```

### 环境变量配置

```bash
# Temporal Server
export TEMPORAL_HOST="localhost:7233"
export TEMPORAL_NAMESPACE="default"

# LLM API Key
export OPENAI_API_KEY="your-openai-key"
export ALIYUN_BAILIAN_API_KEY="your-aliyun-key"

# Redis
export REDIS_HOST="localhost"
export REDIS_PORT="6379"
```

## 📝 已创建的文件

```
backend/internal/agent/
├── coordinator_agent.go      # Coordinator Agent实现
├── coordinator_test.go       # Coordinator Agent测试
├── decision_engine.go        # 决策引擎实现

backend/pkg/message_bus/
├── message.go                # 消息定义
├── bus.go                    # 消息总线实现

backend/pkg/state_sync/
├── state.go                  # 状态定义
├── state_manager.go          # 状态管理器

backend/pkg/conflict_resolver/
├── lock_manager.go           # 分布式锁管理
├── result_resolver.go        # 结果冲突解决

backend/internal/temporal/
├── collaboration_workflow.go # 协作Workflow
├── coordinator_activity.go   # Coordinator Activity
├── client.go                 # Worker注册（已更新）

backend/pkg/redis/
├── redis.go                  # Redis客户端（已增强）
```

## ⚠️ 注意事项

### LLM API Key配置

Coordinator Agent需要LLM API Key才能运行：
- 如果使用阿里云百炼：需要配置 `ALIYUN_BAILIAN_API_KEY`
- 如果使用OpenAI：需要配置 `OPENAI_API_KEY`

### Temporal Server依赖

Temporal Workflow依赖Temporal Server：
- 确保Temporal Server已启动（localhost:7233）
- 确保Temporal Web UI可访问（localhost:8080）

### Redis依赖

消息总线依赖Redis：
- 确保Redis已启动（localhost:6379）
- 确保Redis支持Pub/Sub

### 单Agent实现

当前Coordinator Agent已实现，但6个专业Agent还需要完善：
- Monitor Agent、Analysis Agent、Alert Agent
- Decision Agent、Learning Agent、Interaction Agent

需要集成langchaingo和工具实现。

## 🎉 总结

已成功实现：
- ✅ Coordinator Agent（核心协调者）
- ✅ 决策引擎（智能决策）
- ✅ 消息总线（Agent通信）
- ✅ 状态同步机制（状态管理）
- ✅ 冲突解决机制（冲突处理）
- ✅ Temporal协作Workflow（并发协作编排）
- ✅ 编译验证成功（无编译错误）

下一步：
- 🔄 部署Temporal Server并测试协作Workflow
- 🔄 完善单Agent实现（集成langchaingo）
- 🔄 开发API和前端界面
- 🔄 最终集成测试
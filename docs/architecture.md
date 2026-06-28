# 系统架构说明

## 整体架构

AiOpsHub采用分层架构设计，分为前端层、API层、业务逻辑层、Coordinator Agent层、Temporal Workflow层、Agent协作总线、专业Agent层和数据存储层。

**架构图**：

```
┌─────────────────────────────────────────────────────────┐
│                    前端层 (Frontend)                     │
│         Vue3 + Element Plus + Pinia + Axios             │
└─────────────────────────────────────────────────────────┘
                            ↓ HTTP/WebSocket
┌─────────────────────────────────────────────────────────┐
│                 API层 (API Server)                       │
│              Go + Gin + JWT + Redis                     │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│              业务逻辑层 (Service Layer)                  │
│              Workflow Service + Agent Service           │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│          Coordinator Agent层（协调者层）                 │
│  意图理解 → 任务分解 → Agent调度 → 结果整合 → 冲突解决  │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│          Temporal Workflow层（编排层）                   │
│  并发协作Workflow + 人机交互 + 持久化执行 + 自动恢复    │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│            Agent协作总线（通信层）                       │
│  Redis Pub/Sub + 消息路由 + 事件广播 + 状态共享         │
└─────────────────────────────────────────────────────────┘
                            ↓
┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│Monitor   │  │Analysis  │  │Alert     │  │Decision  │
│Agent     │  │Agent     │  │Agent     │  │Agent     │
│[执行者]  │  │[执行者]  │  │[执行者]  │  │[执行者]  │
└──────────┘  └──────────┘  └──────────┘  └──────────┘
┌──────────┐  ┌──────────┐
│Learning  │  │Interaction│
│Agent     │  │Agent     │
│[执行者]  │  │[服务者]  │
└──────────┘  └──────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│         数据存储层 (Data Layer)                          │
│  PostgreSQL + Redis Cluster + Milvus + Temporal DB      │
└─────────────────────────────────────────────────────────┘
```

## 架构分层

### 1. 前端层 (Frontend Layer)

**技术栈**: Vue3 + Element Plus + Pinia + Axios + Vue Router

**职责**:
- 用户界面展示
- 用户交互处理
- API调用封装
- Token管理和认证
- 路由守卫

**关键模块**:
```
frontend/src/
├── api/              # Axios配置和API调用
│   ├── index.ts      # 请求/响应拦截器
│   └── authApi       # 认证API
│   ├── workflowApi   # Workflow API
│   └── agentApi      # Agent API
├── stores/           # Pinia状态管理
│   ├── auth.ts       # 认证状态（token、username）
├── router/           # 路由配置
│   ├── index.ts      # 路由守卫
├── views/            # 页面组件
│   ├── Login.vue     # 登录页面
│   ├── WorkflowExecute.vue  # Workflow执行
│   ├── Agents.vue    # Agent管理
```

### 2. API层 (API Server Layer)

**技术栈**: Go + Gin + GORM + JWT + Redis Client

**职责**:
- HTTP请求处理
- 参数验证
- 认证授权
- 路由分发
- 响应封装

**关键模块**:
```
backend/internal/
├── handler/          # HTTP handlers
│   ├── handler.go    # 基础handlers
│   ├── workflow_handler.go  # Workflow handlers
├── middleware/       # 中间件
│   ├── middleware.go # CORS、Auth、Logger
├── model/            # 数据模型
│   ├── user.go       # User模型
│   ├── workflow.go   # Workflow模型
```

**认证流程**:
```
用户登录请求
  ↓
Handler.Login()
  ↓
UserService.Login() (验证密码)
  ↓
JWT.GenerateToken() (生成JWT)
  ↓
Redis.SetToken() (存储到Redis)
  ↓
返回token给前端
```

### 3. 业务逻辑层 (Service Layer)

**技术栈**: Go + Service Pattern

**职责**:
- 业务逻辑处理
- 数据验证
- 事务管理
- 错误处理

**关键模块**:
```
backend/internal/service/
├── user_service.go      # 用户业务逻辑
├── workflow_service.go  # Workflow业务逻辑
├── agent_service.go     # Agent业务逻辑
├── alert_service.go     # 告警业务逻辑
```

### 4. Temporal Workflow层

**技术栈**: Temporal SDK + Workflow + Activity

**职责**:
- Workflow编排（并发协作）
- Activity执行调度
- 状态持久化
- 错误重试和自动恢复
- 人机交互（Signal/Query）

**关键模块**:
```
backend/internal/temporal/
├── workflow.go              # Workflow定义
│   ├── AgentWorkflow()      # 单Agent Workflow
│   ├── CollaborationWorkflow()  # 多Agent并发协作Workflow
│   ├── IncidentHandlingWorkflow()  # 故障处理Workflow
│   ├── AlertDedupWorkflow()  # 告警降噪Workflow
├── activity.go              # Activity实现
│   ├── AgentActivity.Execute()  # 单Agent Activity
│   ├── CoordinatorActivity.Execute()  # Coordinator Activity
│   ├── ParallelExecuteActivity()  # 并发执行多个Agent
├── workflow_client.go       # Workflow客户端
│   ├── ExecuteAgentWorkflow()  # 启动Workflow
│   ├── GetWorkflowStatus()    # 查询状态
│   ├── GetWorkflowResult()    # 获取结果
│   ├── SendSignal()          # 发送Signal（人机交互）
│   ├── QueryWorkflow()       # Query状态
├── client.go                # Temporal客户端和Worker
│   ├── NewWorker()          # 创建Worker
│   ├── RegisterWorkflow()   # 注册Workflow
│   ├── RegisterActivity()   # 注册Activity
```

**Workflow执行流程（多Agent协作）**:
```
API请求 -> workflowApi.execute()
  ↓
WorkflowClient.ExecuteCollaborationWorkflow()
  ↓
Temporal Server (调度Workflow)
  ↓
CollaborationWorkflow (并发协作编排)
  ↓
┌─────────────────────────────────────┐
│ CoordinatorActivity.Execute()       │
│ - 意图理解、任务分解、Agent选择     │
│ - 返回子任务列表和Agent映射         │
└─────────────────────────────────────┘
  ↓
并行执行多个Agent Activity
  ↓
workflow.ExecuteActivity(ctx, MonitorAgentActivity, input)
workflow.ExecuteActivity(ctx, AnalysisAgentActivity, input)
workflow.ExecuteActivity(ctx, AlertAgentActivity, input)
  ↓
收集所有Activity结果
  ↓
CoordinatorActivity.IntegrateResults()
  - 结果整合、冲突解决
  ↓
返回最终结果给Workflow
  ↓
Workflow完成，结果存储到Temporal DB
```

**并发协作Workflow实现**:

```go
// 并发协作Workflow示例
func CollaborationWorkflow(ctx workflow.Context, input CollaborationInput) (*CollaborationOutput, error) {
    // 1. Coordinator分解任务
    var coordinatorResult CoordinatorActivityResult
    err := workflow.ExecuteActivity(ctx, CoordinatorActivity, input).Get(ctx, &coordinatorResult)
    
    // 2. 并发执行多个Agent
    futures := []workflow.Future{}
    for _, task := range coordinatorResult.SubTasks {
        future := workflow.ExecuteActivity(ctx, GetAgentActivity(task.AgentID), task)
        futures = append(futures, future)
    }
    
    // 3. 收集所有结果
    agentResults := []AgentResult{}
    for _, future := range futures {
        var result AgentResult
        future.Get(ctx, &result)
        agentResults = append(agentResults, result)
    }
    
    // 4. Coordinator整合结果
    var integrateResult IntegrateActivityResult
    err = workflow.ExecuteActivity(ctx, IntegrateActivity, agentResults).Get(ctx, &integrateResult)
    
    // 5. 等待用户确认（人机交互）
    if integrateResult.RequiresApproval {
        var approval ApprovalSignal
        workflow.GetSignalChannel(ctx, "approval").Receive(ctx, &approval)
        
        if !approval.Approved {
            return &CollaborationOutput{Response: "用户拒绝"}, nil
        }
    }
    
    return &CollaborationOutput{Result: integrateResult.Result}, nil
}
```

### 5. Coordinator Agent层（协调者层）

**技术栈**: Go + langchaingo + Decision Engine

**职责**:
- 意图理解：解析用户请求，识别任务类型
- 任务分解：将复杂任务拆解为子任务序列
- Agent选择：根据子任务类型选择合适的Agent
- 协作编排：决定Agent调用顺序、并行/串行策略
- 结果整合：汇总各Agent结果，生成最终报告
- 冲突解决：处理Agent之间的资源竞争或结果冲突

**关键模块**:
```
backend/internal/agent/
├── coordinator_agent.go     # Coordinator Agent实现
│   ├── NewCoordinatorAgent()  # 创建Coordinator
│   ├── UnderstandIntent()     # 意图理解
│   ├── DecomposeTask()        # 任务分解
│   ├── SelectAgent()          # Agent选择
│   ├── Orchestrate()          # 协作编排
│   ├── IntegrateResults()     # 结果整合
│   ├── ResolveConflict()      # 冲突解决
├── decision_engine.go        # 决策引擎
│   ├── TaskTypeClassifier()   # 任务类型分类
│   ├── AgentRouter()          # Agent路由
│   ├── ParallelStrategy()     # 并行策略决策
│   ├── ConflictResolver()     # 冲突解决策略
```

**Coordinator执行流程**:
```
用户请求："订单服务响应很慢，帮我处理"
  ↓
CoordinatorAgent.UnderstandIntent()
  - LLM解析意图
  - 识别任务类型：故障诊断 + 修复
  ↓
CoordinatorAgent.DecomposeTask()
  - 分解为子任务：
    1. 采集订单服务监控数据
    2. 分析根因
    3. 制定修复方案
    4. 执行修复（需人工确认）
    5. 验证效果
  ↓
CoordinatorAgent.SelectAgent()
  - 任务映射：
    1 → Monitor Agent
    2 → Analysis Agent
    3 → Decision Agent
    4 → Decision Agent
    5 → Monitor Agent
  ↓
CoordinatorAgent.Orchestrate()
  - 编排策略：
    - 任务1、2、3串行（依赖关系）
    - 等待用户Signal确认
    - 任务4、5串行
  ↓
Temporal Workflow执行编排
  ↓
收集Agent结果
  ↓
CoordinatorAgent.IntegrateResults()
  - 汇总：监控数据、根因、修复方案、验证结果
  - 生成综合报告
  ↓
返回给用户
```

### 6. Agent协作总线（通信层）

**技术栈**: Redis Pub/Sub + Message Router

**职责**:
- 消息传递：Agent之间的任务请求、任务结果、协作请求
- 消息路由：根据receiver路由到目标Agent
- 事件广播：广播关键事件（如发现新告警）
- 状态共享：共享Agent执行状态和中间结果

**关键模块**:
```
backend/pkg/message_bus/
├── bus.go                   # 消息总线实现
│   ├── NewMessageBus()      # 创建消息总线
│   ├── Publish()            # 发布消息
│   ├── Subscribe()          #订阅消息
│   ├── Route()              # 消息路由
├── message.go               # 消息定义
│   ├── TaskRequest          # 任务请求消息
│   ├── TaskResult           # 任务结果消息
│   ├── CollaborationRequest # 协作请求消息
│   ├── StateUpdate          # 状态更新消息
│   ├── EventBroadcast       # 事件广播消息
├── router.go                # 消息路由器
│   ├── RouteToAgent()       # 路由到目标Agent
│   ├── BroadcastEvent()     # 事件广播
```

**消息格式**:

```go
type Message struct {
    MessageID   string                 `json:"message_id"`
    MessageType string                 `json:"message_type"` // task_request/task_result/collaboration_request/state_update/event_broadcast
    Sender      string                 `json:"sender"`       // 发送Agent ID
    Receiver    string                 `json:"receiver"`     // 接收Agent ID
    SessionID   string                 `json:"session_id"`   // 会话ID
    WorkflowID  string                 `json:"workflow_id"`  // Temporal Workflow ID
    Content     map[string]interface{} `json:"content"`      // 消息内容
    Timestamp   time.Time              `json:"timestamp"`
}

// 消息示例
TaskRequestMessage := Message{
    MessageID:   "msg-001",
    MessageType: "task_request",
    Sender:      "coordinator-agent",
    Receiver:    "monitor-agent",
    SessionID:   "session-123",
    WorkflowID:  "wf-456",
    Content: {
        "task": "采集订单服务CPU指标",
        "parameters": {
            "service": "order-service",
            "metric": "cpu_usage",
            "time_range": "-1h"
        }
    },
    Timestamp: time.Now(),
}
```

**消息传递流程**:

```
Coordinator Agent发送任务请求
  ↓
MessageBus.Publish("agent-channel", taskRequestMessage)
  ↓
Redis Pub/Sub发布消息
  ↓
Monitor Agent订阅"agent-channel"
  ↓
Monitor Agent接收消息
  ↓
Monitor Agent执行任务
  ↓
Monitor Agent发送任务结果
  ↓
MessageBus.Publish("coordinator-channel", taskResultMessage)
  ↓
Coordinator Agent接收结果
```

### 7. Agent状态同步机制

**技术栈**: Redis + State Manager

**职责**:
- Agent状态存储：PENDING/RUNNING/COMPLETED/FAILED
- 中间结果传递：Agent协作数据共享
- 状态更新通知：状态变更时通知Coordinator
- 状态监控：Coordinator监控所有Agent状态

**关键模块**:
```
backend/pkg/state_sync/
├── state_manager.go         # 状态管理器
│   ├── NewStateManager()    # 创建状态管理器
│   ├── SetAgentState()      # 设置Agent状态
│   ├── GetAgentState()      # 获取Agent状态
│   ├── UpdateProgress()     # 更新执行进度
│   ├── SetIntermediateResult()  # 设置中间结果
│   ├── GetIntermediateResult()  # 获取中间结果
├── state.go                 # 状态定义
│   ├── AgentState           # Agent状态结构
│   ├── SessionState         # 会话状态结构
├── monitor.go               # 状态监控
│   ├── MonitorAgents()      # 监控所有Agent状态
│   ├── HandleFailure()      # 处理失败
│   ├── HandleTimeout()      # 处理超时
```

**状态存储结构**:

```go
type AgentState struct {
    AgentID      string                 `json:"agent_id"`
    SessionID    string                 `json:"session_id"`
    WorkflowID   string                 `json:"workflow_id"`
    Status       string                 `json:"status"` // PENDING/RUNNING/COMPLETED/FAILED/TIMEOUT
    Progress     int                    `json:"progress"` // 0-100
    CurrentTask  string                 `json:"current_task"`
    StartTime    time.Time              `json:"start_time"`
    UpdateTime   time.Time              `json:"update_time"`
    IntermediateResult map[string]interface{} `json:"intermediate_result"` // 中间结果
    Error        string                 `json:"error"` // 错误信息
}

// Redis存储Key格式
Key: agent:state:{session_id}:{agent_id}
TTL: 1小时
```

**状态同步流程**:

```
Agent启动任务
  ↓
StateManager.SetAgentState(agentID, sessionID, RUNNING)
  ↓
Redis.Set("agent:state:session-123:monitor-agent", stateJSON, 1h)
  ↓
Agent执行过程中更新进度
  ↓
StateManager.UpdateProgress(agentID, 50)
  ↓
Redis.Update("agent:state:...", progress=50)
  ↓
Coordinator通过Query查询状态
  ↓
Coordinator.MonitorAgents(sessionID)
  ↓
StateManager.GetAgentState(agentID)
  ↓
Redis.Get("agent:state:...")
  ↓
Coordinator监控所有Agent状态
  ↓
发现异常（FAILED/TIMEOUT）
  ↓
Coordinator.HandleFailure()
```

### 8. Agent层 (Agent Layer)

**技术栈**: langchaingo + LLM Provider (阿里云百炼)

**职责**:
- Agent任务执行
- LLM调用
- Prompt构建
- 响应解析
- 状态同步和消息传递

**关键模块**:
```
backend/internal/agent/
├── base_agent.go            # Agent基类
│   ├── NewBaseAgent()       # 创建Agent
│   ├── Execute()            # 执行任务
│   ├── buildPrompt()        # 构建Prompt
│   ├── SendMessage()        # 发送消息（协作总线）
│   ├── ReceiveMessage()     # 接收消息
│   ├── UpdateState()        # 更新状态（状态同步）
│   ├── SetIntermediateResult()  # 设置中间结果
├── registry.go              # Agent注册中心
│   ├── Register()           # 注册Agent
│   ├── Get()                # 获取Agent
│   ├── ExecuteAgent()       # 执行指定Agent
│   ├── ListAgents()         # 列出所有Agent
├── monitor_agent.go         # Monitor Agent实现
├── analysis_agent.go        # Analysis Agent实现
├── alert_agent.go           # Alert Agent实现
├── decision_agent.go        # Decision Agent实现
├── learning_agent.go        # Learning Agent实现
├── interaction_agent.go     # Interaction Agent实现
├── tools/                   # Agent工具
│   ├── prometheus_tool.go   # Prometheus查询
│   ├── kubernetes_tool.go   # K8s操作
│   ├── log_tool.go          # 日志查询
│   ├── ssh_tool.go          # SSH命令执行
│   ├── sql_tool.go          # SQL执行
```

**Agent执行流程（含协作）**:
```
AgentActivity.Execute()
  ↓
agent.ExecuteAgent(agentID, input)
  ↓
Registry.Get(agentID) (获取Agent)
  ↓
Agent.UpdateState(RUNNING) (更新状态到Redis)
  ↓
Agent.Execute(ctx, input)
  ↓
buildPrompt(input) (构建Prompt)
  ↓
llms.GenerateFromSinglePrompt() (调用LLM)
  ↓
解析响应
  ↓
Agent.SetIntermediateResult(result) (设置中间结果到Redis)
  ↓
Agent.SendMessage(taskResult) (发送结果给Coordinator)
  ↓
Agent.UpdateState(COMPLETED) (更新状态为完成)
  ↓
返回AgentOutput
```

**Agent协作示例（Analysis Agent请求Monitor Agent协作）**:

```
Analysis Agent执行分析
  ↓
发现需要更多监控数据
  ↓
Analysis Agent.SendMessage(collaborationRequest)
  ↓
MessageBus.Publish("monitor-channel", {
    message_type: "collaboration_request",
    sender: "analysis-agent",
    receiver: "monitor-agent",
    content: {
        request: "采集MySQL连接数",
        parameters: {...}
    }
})
  ↓
Monitor Agent接收协作请求
  ↓
Monitor Agent执行采集
  ↓
Monitor Agent.SendMessage(taskResult)
  ↓
MessageBus.Publish("analysis-channel", {
    message_type: "task_result",
    sender: "monitor-agent",
    receiver: "analysis-agent",
    content: {
        result: "MySQL连接数：200"
    }
})
  ↓
Analysis Agent接收结果
  ↓
Analysis Agent继续分析
```

### 9. 冲突解决机制

**技术栈**: Redis Distributed Lock + Conflict Resolver

**职责**:
- 资源竞争解决：多Agent同时访问同一资源（如MySQL配置）
- 结果冲突解决：多Agent给出不同结果（如不同根因分析）
- 执行冲突解决：多Agent同时执行操作（如同时修改K8s配置）

**关键模块**:
```
backend/pkg/conflict_resolver/
├── lock_manager.go          # 分布式锁管理
│   ├── AcquireLock()        # 获取资源锁
│   ├── ReleaseLock()        # 释放资源锁
│   ├── WaitLock()           # 等待锁释放
├── result_resolver.go       # 结果冲突解决
│   ├── VoteResults()        # 投票机制
│   ├── SelectByPriority()   # 优先级选择
│   ├── RequestHumanDecision()  # 人工决策
├── execution_resolver.go    # 执行冲突解决
│   ├── SerializeExecution() # 串行化执行
│   ├── ScheduleExecution()  # 调度执行时间
```

**资源竞争解决流程**:

```go
// Decision Agent执行前获取资源锁
func (d *DecisionAgent) Execute(ctx context.Context, task Task) error {
    // 1. 尝试获取资源锁
    lockKey := fmt.Sprintf("lock:%s", task.ResourceID)
    acquired, err := redis.SetNX(ctx, lockKey, d.AgentID, 30*time.Second)
    
    if !acquired {
        // 2. 等待锁释放或超时
        return fmt.Errorf("resource busy, retry later")
    }
    
    // 3. 执行操作
    d.ModifyResource(ctx, task)
    
    // 4. 释放锁
    redis.Del(ctx, lockKey)
    
    return nil
}
```

**结果冲突解决流程**:

```go
// Coordinator处理多个Agent的不同结果
func (c *CoordinatorAgent) ResolveResultConflict(results []AnalysisResult) string {
    // 策略1：投票机制
    votes := map[string]int{}
    for _, result := range results {
        votes[result.RootCause]++
    }
    
    // 选择票数最多的结果
    maxVotes := 0
    finalResult := ""
    for result, count := range votes {
        if count > maxVotes {
            maxVotes = count
            finalResult = result
        }
    }
    
    // 策略2：票数持平，选择优先级高的Agent结果
    if maxVotes < len(results)/2 {
        finalResult = c.SelectByPriority(results)
    }
    
    // 策略3：仍有冲突，请求人工决策
    if c.HasConflict(results) {
        c.RequestHumanDecision(results)
    }
    
    return finalResult
}
```

### 10. 数据存储层 (Data Layer)

**技术栈**: PostgreSQL + Redis Cluster + Milvus + Temporal DB + MinIO

**职责**:
- 数据持久化（业务数据）
- Token存储
- Workflow状态存储
- Agent状态存储
- 协作消息存储
- 知识向量存储
- 缓存

**关键模块**:
```
backend/internal/repository/
├── user_repo.go             # User数据访问
├── workflow_repo.go         # Workflow数据访问
├── agent_repo.go            # Agent配置数据访问
├── alert_repo.go            # Alert数据访问
├── execution_repo.go        # Execution数据访问
├── agent_state_repo.go      # Agent状态数据访问
├── collaboration_repo.go    # 协作记录数据访问

backend/pkg/redis/
├── redis.go                 # Redis客户端
│   ├── SetToken()           # 存储Token
│   ├── GetToken()           # 获取Token
│   ├── DeleteToken()        # 删除Token
│   ├── ExistsToken()        # 检查Token存在性
│   ├── SetAgentState()      # 存储Agent状态
│   ├── GetAgentState()      # 获取Agent状态
│   ├── Publish()            # 发布消息（协作总线）
│   ├── Subscribe()          #订阅消息
│   ├── SetNX()              # 分布式锁

backend/pkg/milvus/
├── milvus.go                # Milvus向量数据库客户端
│   ├── InsertVectors()      # 插入向量
│   ├── SearchVectors()      # 搜索向量
│   ├── DeleteVectors()      # 删除向量
```

**数据库表结构**:
```
PostgreSQL:
- users                   # 用户表
- workflows               # Workflow定义表
- agents                  # Agent配置表
- alerts                  # 告警记录表
- workflow_executions     # Workflow执行记录表
- agent_executions        # Agent执行记录表
- collaboration_sessions  # 协作会话表
- collaboration_records   # 协作记录表（Agent之间的协作）

Redis:
- token:{jwt_token}       # JWT token存储（JSON格式）
- agent:state:{session_id}:{agent_id}  # Agent状态存储（JSON格式，TTL 1h）
- collaboration:result:{session_id}    # 协作中间结果存储（JSON格式，TTL 会话期间）
- lock:{resource_id}      # 分布式锁（防止资源竞争）

Milvus:
- knowledge_base          # 知识向量库（历史案例、运维手册）
- fault_patterns          # 故障模式向量库

Temporal DB:
- workflow executions     # Workflow执行状态
- activity executions     # Activity执行状态
- event history           # Event History（持久化执行日志）

MinIO:
- knowledge_documents     # 知识文档存储（运维手册、最佳实践）
- reports                 # 生成的报告文件
- logs                    # 日志文件归档
```

**数据存储策略**:
```
1. 短期数据（临时）：
   - Agent执行状态：Redis（TTL 1小时）
   - 协作中间结果：Redis（TTL 会话期间）
   - 协作消息：Redis Pub/Sub（实时传递）

2. 长期数据（持久）：
   - Workflow执行记录：PostgreSQL
   - Agent执行记录：PostgreSQL
   - 协作会话记录：PostgreSQL
   - Event History：Temporal DB（永久）

3. 知识数据：
   - 向量化知识：Milvus（永久）
   - 文档文件：MinIO（永久）
```

## 数据流

### 1. 用户登录流程

```
前端 Login.vue (输入username/password)
  ↓ POST /api/v1/auth/login
handler.Login()
  ↓
userService.Login() (验证用户)
  ↓
jwt.GenerateToken() (生成JWT)
  ↓
redis.SetToken() (存储TokenInfo到Redis)
  ↓
返回 { token, user_id, username, role }
  ↓
前端 authStore.login() (存储到localStorage)
  ↓
跳转到首页
```

### 2. 单Agent Workflow执行流程

```
前端 WorkflowExecute.vue (输入告警内容)
  ↓ POST /api/v1/workflows/execute
handler.ExecuteAgentWorkflow()
  ↓
workflowClient.ExecuteAgentWorkflow()
  ↓
Temporal Server (启动Workflow)
  ↓
AgentWorkflow (Workflow编排)
  ↓
ExecuteActivity("Execute", input)
  ↓
AgentActivity.Execute()
  ↓
agent.ExecuteAgent("monitor-agent-001", input)
  ↓
MonitorAgent.Execute()
  ↓
调用阿里云百炼LLM
  ↓
返回分析结果
  ↓
Workflow完成
  ↓
前端 workflowApi.getStatus() (查询状态)
  ↓
前端 workflowApi.getResult() (获取结果)
```

### 3. 多Agent协作Workflow执行流程

```
前端 WorkflowExecute.vue (输入复杂问题)
  ↓ POST /api/v1/workflows/collaborate
handler.ExecuteCollaborationWorkflow()
  ↓
workflowClient.ExecuteCollaborationWorkflow()
  ↓
Temporal Server (启动协作Workflow)
  ↓
CollaborationWorkflow (并发协作编排)
  ↓
┌────────────────────────────────────────────┐
│ Step 1: Coordinator分解任务                │
│ CoordinatorActivity.Execute()              │
│ - LLM解析意图："订单服务响应很慢"           │
│ - 分解为子任务：                            │
│   1. 采集监控数据                           │
│   2. 分析根因                               │
│   3. 制定修复方案                           │
│ - 选择Agent：Monitor、Analysis、Decision   │
│ - 编排策略：串行执行                        │
└────────────────────────────────────────────┘
  ↓
┌────────────────────────────────────────────┐
│ Step 2: Monitor Agent执行（并发）          │
│ workflow.ExecuteActivity(MonitorActivity)  │
│ - MonitorAgent.UpdateState(RUNNING)        │
│ - MonitorAgent.Execute()                   │
│   - 采集CPU、内存、网络指标                 │
│   - 调用PrometheusTool                     │
│ - MonitorAgent.SetIntermediateResult()     │
│ - MonitorAgent.SendMessage(result)         │
│ - MonitorAgent.UpdateState(COMPLETED)      │
└────────────────────────────────────────────┘
  ↓ (Monitor结果：CPU 85%，MySQL连接数激增)
┌────────────────────────────────────────────┐
│ Step 3: Analysis Agent执行                 │
│ workflow.ExecuteActivity(AnalysisActivity) │
│ - AnalysisAgent.GetIntermediateResult()    │
│   (读取Monitor的中间结果)                   │
│ - AnalysisAgent.Execute()                  │
│   - 分析MySQL慢查询                         │
│   - RAG检索历史案例                         │
│ - AnalysisAgent.SendMessage(result)        │
│   (根因：MySQL缺少索引)                     │
└────────────────────────────────────────────┘
  ↓
┌────────────────────────────────────────────┐
│ Step 4: Decision Agent执行                 │
│ workflow.ExecuteActivity(DecisionActivity) │
│ - DecisionAgent.GetIntermediateResult()    │
│   (读取Analysis的根因)                      │
│ - DecisionAgent.Execute()                  │
│   - 制定修复方案：添加索引                  │
│   - 风险评估：低风险                        │
│   - 生成执行计划                            │
│ - DecisionAgent.SendMessage(result)        │
│   (需人工确认)                              │
└────────────────────────────────────────────┘
  ↓
┌────────────────────────────────────────────┐
│ Step 5: 等待用户确认（Signal）             │
│ workflow.GetSignalChannel("approval")      │
│   .Receive(ctx, &approval)                 │
│ - 前端显示执行计划                          │
│ - 用户点击"确认"按钮                        │
│ - frontend.sendSignal(approval)            │
│ - Workflow接收Signal                       │
└────────────────────────────────────────────┘
  ↓
┌────────────────────────────────────────────┐
│ Step 6: Coordinator整合结果                │
│ CoordinatorActivity.IntegrateResults()     │
│ - 汇总：监控数据 + 根因 + 修复方案          │
│ - 生成综合报告                              │
│ - 返回给Workflow                            │
└────────────────────────────────────────────┘
  ↓
Workflow完成，结果存储到Temporal DB
  ↓
前端 workflowApi.getStatus() (查询状态)
  ↓
前端 workflowApi.getResult() (获取结果)
  ↓
前端显示综合报告
```

### 4. Agent协作消息传递流程

```
Coordinator Agent发送任务请求
  ↓
MessageBus.Publish("agent-channel", taskRequest)
  ↓
Redis Pub/Sub发布消息
  ↓
Monitor Agent订阅"agent-channel"
  ↓
Monitor Agent.ReceiveMessage()
  ↓
Monitor Agent执行任务
  ↓
Monitor Agent.SetIntermediateResult(result)
  (存储中间结果到Redis)
  ↓
Monitor Agent.SendMessage(taskResult)
  ↓
MessageBus.Publish("coordinator-channel", taskResult)
  ↓
Coordinator Agent.ReceiveMessage()
  ↓
Coordinator Agent读取中间结果
  ↓
Coordinator Agent.IntegrateResults()
```

### 5. 状态同步流程

```
Coordinator监控所有Agent状态
  ↓
Coordinator.MonitorAgents(sessionID)
  ↓
StateManager.GetAgentState("monitor-agent")
  ↓
Redis.Get("agent:state:session-123:monitor-agent")
  ↓
检查状态：
  - RUNNING：正常执行中
  - COMPLETED：已完成
  - FAILED：失败，需要处理
  - TIMEOUT：超时，需要处理
  ↓
发现Agent失败
  ↓
Coordinator.HandleFailure()
  ↓
重试或降级处理
  ↓
发送Signal给Workflow（重新调度）
```

### 6. 冲突解决流程

```
多个Agent同时执行修改MySQL配置
  ↓
Agent 1尝试获取锁
  ↓
Redis.SetNX("lock:mysql-config", "agent-1", 30s)
  ↓
成功获取锁
  ↓
Agent 1执行修改
  ↓
Agent 2尝试获取锁
  ↓
Redis.SetNX("lock:mysql-config", "agent-2", 30s)
  ↓
获取锁失败（锁已被Agent 1持有）
  ↓
Agent 2等待锁释放或超时
  ↓
Agent 1完成修改，释放锁
  ↓
Redis.Del("lock:mysql-config")
  ↓
Agent 2获取锁成功
  ↓
Agent 2继续执行
```

### 7. Token验证流程

```
前端 Axios请求 (Authorization: Bearer token)
  ↓
middleware.Auth() (拦截器)
  ↓
jwt.ValidateToken() (JWT签名验证)
  ↓
redis.ExistsToken() (Redis存在性验证)
  ↓
设置 user_id, username, role 到 context
  ↓
继续执行Handler
```

## 关键技术决策

### 1. 为什么选择Coordinator Agent架构？

**原因**:
- 智能协作：Coordinator动态分解任务、选择Agent、编排协作
- 灵活性：支持复杂的多Agent协作场景（并发、串行、动态）
- 可扩展性：新增Agent无需修改Workflow，Coordinator自动路由
- 冲突解决：集中处理资源竞争和结果冲突
- 状态监控：全局监控所有Agent执行状态

**替代方案对比**:
- 固定Workflow编排：不够灵活，无法处理复杂协作场景
- 直接Agent调用：缺少协调机制，容易产生冲突
- 无Coordinator架构：难以实现动态协作和结果整合

### 2. 为什么选择Redis作为Agent协作总线？

**原因**:
- 高性能：Pub/Sub消息传递延迟 <100ms
- 轻量级：无需引入复杂的消息队列系统（如Kafka、RabbitMQ）
- 状态共享：Redis天然支持状态存储和共享
- 分布式锁：支持资源竞争解决
- 成本低：Redis Cluster已部署，无需额外组件

**替代方案对比**:
- Kafka：重量级，运维复杂，不适合Agent实时通信
- RabbitMQ：消息传递延迟较高，不适合实时协作
- 直接HTTP调用：缺少异步机制，性能较差

### 3. 为什么选择Temporal作为Workflow编排引擎？

**原因**:
- 持久执行：Workflow可运行数月甚至数年，自动恢复
- 并发编排：支持并发执行多个Activity（Agent协作）
- 人机交互：Signal/Query机制支持用户确认和实时状态查询
- 可视化监控：Temporal Web UI查看协作执行过程
- 企业级可靠：OpenAI、Hashicorp都在用

**替代方案对比**:
- Python LangGraph：不支持Go，需要Python环境
- 自建Workflow系统：开发成本高，缺少监控和持久化
- 简单编排：无法处理复杂协作和长时间任务

### 4. 为什么选择langchaingo？

**原因**:
- 纯Go实现：与Go生态无缝集成
- 支持多种LLM Provider：阿里云百炼、OpenAI、Ollama
- 完整Agent能力：Agent、Tools、Memory、Chains
- 活跃的社区：持续更新和优化

**替代方案对比**:
- Python LangChain：需要Python环境，增加系统复杂度
- 直接HTTP调用LLM：缺少工具和Prompt管理
- 自建Agent框架：开发成本高，缺少工具生态

### 5. 为什么使用Redis Cluster？

**原因**:
- 高可用性：多节点，自动故障转移
- 数据分片：负载均衡，分散读写压力
- 状态存储：支持Agent状态、协作消息、分布式锁
- 成本低：开源免费，运维简单

**降级方案**:
- Redis连接失败时降级为纯JWT验证
- Agent状态同步失败时使用Temporal Query查询
- 系统仍可运行，但协作性能降低

### 6. 为什么需要冲突解决机制？

**原因**:
- 资源竞争：多Agent同时访问同一资源（如数据库配置）
- 结果冲突：多Agent给出不同分析结果（如不同根因）
- 执行冲突：多Agent同时执行操作（如同时修改K8s）
- 人工决策：复杂冲突需要人工介入

**解决策略**:
- 分布式锁：Redis SetNX实现资源竞争解决
- 结果投票：多Agent结果投票选择
- 优先级选择：根据Agent优先级选择结果
- 人工决策：复杂冲突请求人工确认

## 性能优化

### 1. Agent缓存
- Agent实例在Worker启动时创建并缓存
- 避免每次执行都重新创建
- Registry统一管理Agent生命周期

### 2. Temporal并发协作优化
- 并行执行独立Agent Activity
- 使用workflow.ExecuteActivity并发调度
- 减少协作总时间（相比串行执行提升50%）

### 3. Redis协作总线优化
- Redis Cluster 6节点，分散消息传递压力
- Pub/Sub消息传递延迟 <100ms
- 状态存储使用JSON序列化，减少查询次数

### 4. Coordinator决策优化
- LLM意图理解和任务分解并行处理
- Agent选择策略缓存（任务类型 → Agent映射表）
- 结果整合批量处理

### 5. Agent状态同步优化
- Redis状态存储TTL 1小时，自动清理
- 状态更新频率控制（每10秒更新一次进度）
- Coordinator批量查询状态（减少Redis查询次数）

### 6. Temporal Task Queue
- 专用Task Queue: `aiops-task-queue`
- 避免与其他系统竞争资源
- 支持优先级调度（紧急任务优先）

### 7. 分布式锁优化
- 锁超时时间30秒，避免长时间阻塞
- 锁失败后等待队列，自动重试
- 锁释放后立即通知等待Agent

## 安全设计

### 1. JWT + Redis双重验证
- JWT签名防止篡改
- Redis检查防止已注销Token被使用

### 2. 密码加密
- bcrypt哈希存储密码
- 防止明文密码泄露

### 3. CORS配置
- 允许指定域名访问
- 防止跨域攻击

### 4. Token过期
- 默认30分钟过期
- 可配置过期时间

### 5. Agent协作安全
- 消息签名验证（防止伪造消息）
- Agent身份验证（防止未授权Agent加入协作）
- 资源锁机制（防止资源竞争）
- 操作审计（记录所有Agent操作）

### 6. 人机交互安全
- 高风险操作必须人工确认
- Signal验证（防止恶意Signal）
- 执行预览（用户可查看执行计划）

## 可扩展性

### 1. Agent扩展
- 注册新Agent只需实现Agent接口
- Registry统一管理
- Coordinator自动路由（无需修改Workflow）

### 2. Workflow扩展
- 新增Workflow只需定义Workflow函数
- Temporal SDK自动注册
- 支持并发协作Workflow

### 3. Tool扩展
- Agent可以集成新工具
- PromQL查询、日志查询、SSH命令等
- Tools统一管理

### 4. LLM扩展
- 支持多种LLM Provider
- 配置切换即可使用不同LLM
- 本地模型备选（Ollama）

### 5. Coordinator扩展
- 新增协作策略（如A/B测试协作）
- 新增冲突解决策略
- 新增任务分解算法

### 6. 协作总线扩展
- 新增消息类型（如告警广播）
- 新增消息路由策略
- 支持消息持久化（可选）

### 7. 数据存储扩展
- PostgreSQL水平扩展（读写分离）
- Redis Cluster节点扩展
- Milvus向量库扩展

## 监控和日志

### 1. Temporal Web UI
- Workflow执行可视化
- Activity执行详情
- Event History查看
- 错误日志查看
- Signal/Query监控
- 并发协作可视化

### 2. Agent协作监控
- Coordinator决策监控（任务分解、Agent选择）
- Agent状态监控（PENDING/RUNNING/COMPLETED/FAILED）
- 协作消息监控（消息传递延迟、成功率）
- 冲突解决监控（资源竞争、结果冲突）
- 性能指标监控（并发Agent数、协作成功率）

### 3. 应用日志
- 结构化日志输出
- 错误分级（Info、Error）
- 日志文件存储
- Agent协作日志（任务分解、结果整合）
- Coordinator决策日志

### 4. PostgreSQL慢查询日志
- GORM集成慢查询日志
- SQL执行时间监控
- Agent状态查询慢日志

### 5. Redis监控
- Redis Cluster状态监控
- Pub/Sub消息吞吐量监控
- 状态存储命中率监控
- 分布式锁等待时间监控

### 6. 性能指标监控
- Coordinator调度延迟（目标 <3s）
- Agent并发数（目标 >10）
- 协作成功率（目标 >95%）
- 消息传递延迟（目标 <100ms）
- 状态同步延迟（目标 <50ms）
- Temporal Workflow执行时间

### 7. 告警机制
- Agent失败告警
- 协作超时告警
- 冲突频繁告警
- Temporal Worker异常告警
- Redis连接异常告警

## 故障恢复机制

### 1. Temporal自动恢复
- Workflow失败后自动重放Event History
- Activity失败后自动重试（最多3次）
- Worker重启后自动恢复执行中的Workflow

### 2. Agent失败恢复
- Agent Activity失败后Temporal自动重试
- Coordinator检测Agent失败，重新调度
- 降级处理：单Agent失败不影响整体协作

### 3. 协作总线恢复
- Redis连接失败降级为Temporal Query
- 消息传递失败重试机制
- 状态同步失败使用持久化存储

### 4. Coordinator恢复
- Coordinator Activity失败重试
- 任务分解失败请求人工干预
- 结果整合失败返回部分结果

### 5. 资源竞争恢复
- 分布式锁超时自动释放
- 锁等待超时通知Coordinator
- Coordinator重新调度执行

### 6. 人工干预通道
- Workflow失败后人工重试（Temporal Web UI）
- 协作失败后人工接管
- 冲突解决失败后人工决策

## 更新记录

| 版本 | 日期 | 更新内容 | 更新人 |
|------|------|----------|--------|
| v1.0.0 | 2026-06-24 | 初稿（纯Go + Temporal架构） | AI Assistant |
| v2.0.0 | 2026-06-26 | 补充多Agent协作技术架构设计 | AI Assistant |
| v2.0.0 | 2026-06-26 | 新增Coordinator Agent层、Agent协作总线、状态同步机制、冲突解决机制 | AI Assistant |
| v2.0.0 | 2026-06-26 | 更新数据流、关键技术决策、性能优化、监控和故障恢复 | AI Assistant |

---

**文档版本**: v2.0.0
**更新时间**: 2026-06-26
**架构版本**: v3.0 - 纯Go + Temporal + 多Agent协作
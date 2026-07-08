# AiOpsHub - 多Agent协作智能运维平台

基于纯Go架构 + Temporal工作流引擎 + LangChainGo框架的多Agent协作智能运维平台。

## 核心特性

- **纯Go架构**：后端完全使用Go实现，无需Python依赖
- **Temporal工作流引擎**：可靠的工作流编排，支持长时间运行任务
- **LangChainGo集成**：使用LangChain的Go实现构建智能Agent
- **多Agent协作机制**：
  - Coordinator Agent（协调者）：意图理解、任务分解、协作编排
  - Decision Engine（决策引擎）：Agent路由、依赖分析、并行调度
  - 6个专业Agent：监控、分析、告警、决策、学习、交互
  - 消息总线：Redis Pub/Sub实现
  - 状态同步：实时状态管理
  - 冲突解决：分布式锁 + 结果投票机制
- **AI对话系统**（新增）：
  - 流式对话：SSE实时推送AI回复
  - 多轮对话：支持上下文记忆(最多10轮)
  - 会话管理：完整的对话历史记录
  - Markdown渲染：前端支持代码高亮
- **RAG知识检索**（新增）：
  - 自动检索：对话时自动从Milvus检索相关知识
  - 上下文注入：将知识无缝融入对话
  - 向量化存储：基于Milvus的高性能向量检索
  - 配置可控：通过配置启用/禁用RAG功能
- **MCP工具集成**（新增）：
  - 工具管理：注册和配置外部工具服务(Jenkins、K8s等)
  - 自动调用：AI识别意图并自动调用工具
  - 结果处理：智能处理工具返回结果
  - Session管理：支持长连接和Session复用
- **智能Agent路由**（新增）：
  - 意图识别：分析用户问题类型
  - 自动匹配：选择最合适的Agent处理问题
  - SystemPrompt注入：使用Agent专属系统提示增强专业性
- **Token统计服务**（新增）：
  - Token记录：记录每次LLM调用的Token消耗
  - 成本计算：根据模型定价计算API成本
  - 统计分析：按会话/Agent维度统计Token使用
- **WebSocket实时推送**：前端实时接收工作流执行状态
- **前端Vue3界面**：现代化响应式UI，包含AI助手聊天界面

## 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                     前端 Vue3 + Element Plus                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│  │ Dashboard│  │ Workflow │  │ Agent Mgmt│ │Collaboration│   │
│  └───┬──────┘  └───┬──────┘  └───┬──────┘  └───┬──────┘     │
│      │            │            │            │ WebSocket     │
└─────────────────────────────────────────────────────────────┘
                           │
┌─────────────────────────────────────────────────────────────┐
│                   Backend API Server (Go)                    │
│  ┌──────────────────────────────────────────────────────┐   │
│  │            REST API + WebSocket Handler               │   │
│  └──────────────────────────────────────────────────────┘   │
│                           │                                  │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Temporal Workflow Client                 │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                           │
┌─────────────────────────────────────────────────────────────┐
│                  Temporal Server (工作流引擎)                 │
│  ┌──────────────────────────────────────────────────────┐   │
│  │      Collaboration Workflow + Parallel Monitor       │   │
│  └──────────────────────────────────────────────────────┘   │
│                           │                                  │
│  ┌──────────────────────────────────────────────────────┐   │
│  │               Temporal Worker (Go)                    │   │
│  │  ┌─────────────────────────────────────────────────┐ │   │
│  │  │         Coordinator Activity                     │ │   │
│  │  │  • UnderstandIntent                              │ │   │
│  │  │  • DecomposeTask                                 │ │   │
│  │  │  • ScheduleAgents                                │ │   │
│  │  │  • IntegrateResults                              │ │   │
│  │  └─────────────────────────────────────────────────┘ │   │
│  │  ┌─────────────────────────────────────────────────┐ │   │
│  │  │         Agent Activities (6个专业Agent)          │ │   │
│  │  │  • Monitor Agent      (监控采集)                 │ │   │
│  │  │  • Analysis Agent     (根因分析)                 │ │   │
│  │  │  • Alert Agent        (告警处理)                 │ │   │
│  │  │  • Decision Agent     (决策执行)                 │ │   │
│  │  │  • Learning Agent     (学习优化)                 │ │   │
│  │  │  • Interaction Agent  (交互服务)                 │ │   │
│  │  └─────────────────────────────────────────────────┘ │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                           │
┌─────────────────────────────────────────────────────────────┐
│                    Agent协作基础设施                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐    │
│  │Message Bus│ │State Sync │ │Conflict  │ │Decision  │     │
│  │ (Redis)   │ │ Manager   │ │Resolver  │ │Engine    │     │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘    │
└─────────────────────────────────────────────────────────────┘
                           │
┌─────────────────────────────────────────────────────────────┐
│                      数据层                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐    │
│  │PostgreSQL│  │  Redis   │  │  Milvus  │  │ClickHouse│     │
│  │(业务数据) │  │(缓存/状态)│  │(向量检索)│  │(时序数据)│     │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘    │
└─────────────────────────────────────────────────────────────┘
                           │
┌─────────────────────────────────────────────────────────────┐
│                    外部系统集成                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐    │
│  │Prometheus│  │Kubernetes│  │   Logs   │  │   LLM    │     │
│  │(监控指标) │  │(容器编排) │  │(日志系统)│  │(OpenAI)  │     │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## 技术栈

### 后端
- **语言**: Go 1.22+
- **Web框架**: Gin
- **工作流引擎**: Temporal Server 1.20+
- **Agent框架**: LangChainGo (LangChain的Go实现)
- **数据库**: PostgreSQL 15
- **缓存**: Redis 7
- **向量数据库**: Milvus 2.3+
- **时序数据库**: ClickHouse 23.8
- **消息队列**: Redis Pub/Sub

### 前端
- **框架**: Vue 3 + TypeScript
- **UI组件**: Element Plus
- **构建工具**: Vite
- **实时通信**: WebSocket

### 部署
- **容器化**: Docker + Docker Compose
- **监控**: Prometheus + Grafana
- **工作流可视化**: Temporal Web UI

## 快速开始

### 1. 环境准备

确保已安装：
- Go 1.22+
- Node.js 22+
- Docker & Docker Compose
- Make工具

### 2. 克隆项目

```bash
git clone https://github.com/your-org/AiOpsHub.git
cd AiOpsHub
```

### 3. 配置系统

```bash
# 编辑配置文件
vim backend/configs/config.yaml
```

必须配置的项目：
```yaml
llm:
  api_key: your-openai-api-key-here
  base_url: https://api.openai.com/v1
```

### 4. 启动依赖服务

```bash
# 启动基础依赖（PostgreSQL、Redis、Temporal）
make run-deps

# 或启动完整堆栈（包括Milvus、ClickHouse、监控）
make deploy-full
```

等待服务启动完成（约15-30秒）。

### 5. 安装依赖并构建

```bash
# 安装Go依赖并构建
make build

# 安装前端依赖并构建
cd frontend
npm install
npm run build
```

### 6. 启动服务

```bash
# 启动API服务器
make run-api

# 启动Temporal Worker（另一个终端）
make run-worker

# 或同时启动所有服务
make run-all
```

### 7. 访问服务

- **前端界面**: http://localhost:5173 (开发模式) 或 http://localhost:8080 (生产模式)
- **后端API**: http://localhost:8080/api
- **Temporal UI**: http://localhost:8080 (Temporal默认端口)
- **Grafana**: http://localhost:3000
- **Prometheus**: http://localhost:9090

### 8. 运行测试

```bash
# 运行后端单元测试
make test

# 或详细输出
cd backend
go test ./... -v
```

## 项目结构

```
AiOpsHub/
├── backend/                 # Go后端
│   ├── cmd/                 # 应用入口
│   │   ├── api-server/      # REST API服务器
│   │   └── temporal-worker/ # Temporal Worker
│   ├── internal/            # 内部实现
│   │   ├── agent/           # Agent核心实现
│   │   │   ├── coordinator_agent.go      # Coordinator Agent
│   │   │   ├── decision_engine.go        # 决策引擎
│   │   │   ├── specialized_agents.go     # 6个专业Agent
│   │   │   └── tools/        # Agent工具集成
│   │   ├── temporal/        # Temporal工作流
│   │   │   ├── collaboration_workflow.go # 协作工作流
│   │   │   ├── coordinator_activity.go   # Coordinator Activity
│   │   │   ├── workflow.go               # 基础工作流
│   │   │   └── workflow_client.go        # Workflow客户端
│   │   ├── handler/         # HTTP Handler
│   │   │   ├── workflow_handler.go       # Workflow API
│   │   │   ├── websocket_handler.go      # WebSocket
│   │   │   └── agent_handler.go          # Agent管理API
│   │   ├── model/           # 数据模型
│   │   ├── repository/      # 数据访问层
│   │   ├── service/         # 业务逻辑层
│   │   ├── config/          # 配置管理
│   │   └── middleware/      # HTTP中间件
│   ├── pkg/                 # 公共包
│   │   ├── message_bus/     # 消息总线
│   │   ├── state_sync/      # 状态同步
│   │   ├── conflict_resolver/ # 冲突解决
│   │   ├── redis/           # Redis客户端
│   │   ├── logger/          # 日志
│   │   └── jwt/             # JWT认证
│   ├── configs/             # 配置文件
│   │   ├── config.yaml      # 生产配置
│   │   └── config.test.yaml # 测试配置
│   └── go.mod               # Go模块定义
│
├── frontend/                # Vue3前端
│   ├── src/
│   │   ├── api/             # API客户端
│   │   │   ├── index.ts     # REST API
│   │   │   └── websocket.ts # WebSocket
│   │   ├── views/           # 页面组件
│   │   │   ├── Dashboard.vue          # 仪表板
│   │   │   ├── WorkflowMonitor.vue    # Workflow监控
│   │   │   ├── AgentsManage.vue       # Agent管理
│   │   │   └── CollaborationMonitor.vue # 协作监控
│   │   ├── components/      # 通用组件
│   │   ├── router/          # 路由配置
│   │   ├── stores/          # 状态管理
│   │   └── main.ts          # 应用入口
│   ├── package.json         # 前端依赖
│   └── vite.config.ts       # Vite配置
│
├── deployments/             # 部署配置
│   ├── docker-compose.yml   # Docker Compose配置
│   ├── scripts/             # 初始化脚本
│   │   ├── init-db.sql      # 数据库初始化
│   │   └── init-temporal-db.sh # Temporal数据库
│   └── prometheus.yml       # Prometheus配置
│
├── docs/                    # 文档
│   ├── PRD.md               # 产品需求文档
│   ├── architecture.md      # 系统架构
│   ├── temporal-workflow-design.md      # Temporal设计
│   ├── langchaingo-agent-design.md      # Agent设计
│   ├── coordinator-agent-quick-start.md # 快速开始
│   ├── implementation-summary.md        # 实现总结
│   └── api/
│       └── backend-api.md   # Backend API文档
│
├── Makefile                 # Make工具配置
├── .env.example             # 环境变量示例
└── README.md                # 项目说明

```

## 核心组件详解

### 1. Coordinator Agent（协调者Agent）

负责理解用户意图、分解任务、编排多Agent协作。

**核心功能**：
- **意图理解**: 使用LLM分析用户查询，识别任务类型
- **任务分解**: 将复杂任务分解为多个子任务
- **Agent调度**: 根据任务类型路由到合适的专业Agent
- **协作编排**: 决定任务的执行顺序（串行/并行/混合）
- **结果整合**: 合并多个Agent的执行结果，生成综合报告
- **冲突解决**: 处理Agent之间的结果冲突

**代码位置**: `backend/internal/agent/coordinator_agent.go`

### 2. Decision Engine（决策引擎）

负责Agent路由、依赖分析和执行策略制定。

**核心功能**：
- **Agent路由**: 根据任务类型映射到特定Agent
- **依赖分析**: 分析任务依赖关系，构建执行图
- **并行调度**: 识别可并行执行的任务组
- **优先级管理**: 为任务分配执行优先级
- **审批决策**: 判断哪些操作需要人工审批

**代码位置**: `backend/internal/agent/decision_engine.go`

### 3. 消息总线（Message Bus）

基于Redis Pub/Sub实现的Agent协作通信机制。

**消息类型**：
- `TaskRequest`: 任务请求消息
- `TaskResult`: 任务结果消息
- `CollaborationRequest`: Agent协作请求
- `StateUpdate`: 状态更新通知
- `EventBroadcast`: 事件广播

**代码位置**: `backend/pkg/message_bus/`

### 4. 状态同步（State Sync）

实时管理Agent执行状态和中间结果传递。

**状态类型**：
- `PENDING`: 任务等待执行
- `RUNNING`: 任务正在执行
- `COMPLETED`: 任务执行完成
- `FAILED`: 任务执行失败
- `TIMEOUT`: 任务超时

**代码位置**: `backend/pkg/state_sync/`

### 5. 冲突解决（Conflict Resolver）

处理多Agent协作时的结果冲突。

**机制**：
- **分布式锁**: Redis SetNX实现资源互斥访问
- **结果投票**: 多Agent结果投票选择
- **优先级选择**: 按Agent优先级选择结果

**代码位置**: `backend/pkg/conflict_resolver/`

### 6. Temporal Workflow

基于Temporal的工作流编排。

**核心Workflow**：
- `CollaborationWorkflow`: 多Agent协作主工作流
- `ParallelMonitorWorkflow`: 并行执行监控
- `AgentExecutionWorkflow`: 单Agent执行工作流

**核心Activity**：
- `UnderstandIntent`: 意图理解
- `DecomposeTask`: 任务分解
- `ScheduleAgents`: Agent调度
- `ExecuteAgentTask`: 执行Agent任务
- `IntegrateResults`: 结果整合

**代码位置**: `backend/internal/temporal/`

## API示例

### 1. AI对话系统 (新增)

#### 创建对话会话

```bash
POST /api/v1/chat/sessions
Content-Type: application/json
Authorization: Bearer <token>

{
  "title": "运维问题咨询",
  "model": "qwen3.7-max"
}

Response:
{
  "message": "会话创建成功",
  "data": {
    "id": "session-123",
    "title": "运维问题咨询",
    "model": "qwen3.7-max",
    "status": "active"
  }
}
```

#### 发送消息(流式响应)

```bash
POST /api/v1/chat/messages/stream
Content-Type: application/json
Authorization: Bearer <token>

{
  "session_id": "session-123",
  "content": "订单服务响应很慢，帮我分析原因"
}

Response (SSE流式):
event: user_message
data: {"id":"msg-1","role":"user","content":"订单服务响应很慢..."}

event: rag_references
data: [{"title":"服务响应慢排查指南","score":0.85}]

event: chunk
data: {"content":"根"}

event: chunk
data: {"content":"据"}

event: chunk
data: {"content":"知识库"}

event: done
data: {"message":"流式输出完成"}
```

#### 发送消息(非流式)

```bash
POST /api/v1/chat/messages
Content-Type: application/json
Authorization: Bearer <token>

{
  "session_id": "session-123",
  "content": "如何排查Pod启动失败的问题"
}

Response:
{
  "message": "消息发送成功",
  "ai_response": "排查Pod启动失败可以从以下几个方面入手...",
  "user_message": {...},
  "ai_message": {...},
  "rag_references": [
    {
      "title": "Pod启动失败排查指南",
      "category": "troubleshooting",
      "score": 0.92
    }
  ]
}
```

### 2. RAG知识检索 (新增)

#### 搜索知识

```bash
POST /api/v1/rag/search
Content-Type: application/json
Authorization: Bearer <token>

{
  "query": "服务响应慢排查方法",
  "top_k": 3
}

Response:
{
  "results": [
    {
      "document": {
        "id": "doc-1",
        "title": "服务响应慢排查指南",
        "category": "troubleshooting"
      },
      "score": 0.85
    }
  ]
}
```

#### 创建知识文档

```bash
POST /api/v1/rag/documents
Content-Type: application/json
Authorization: Bearer <token>

{
  "title": "Pod启动失败排查方法",
  "content": "# 排查步骤\n1. 查看Pod状态\n2. 检查事件日志\n3. 分析容器日志...",
  "category": "troubleshooting",
  "tags": ["Pod", "Kubernetes", "排查"]
}
```

### 3. MCP工具调用 (新增)

#### 创建MCP Server

```bash
POST /api/v1/mcp/servers
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "Jenkins Server",
  "description": "CI/CD工具集成",
  "url": "http://jenkins.example.com:8080",
  "auth_type": "token",
  "auth_token": "your-jenkins-token"
}
```

#### 获取工具列表

```bash
GET /api/v1/mcp/servers/:id/tools
Authorization: Bearer <token>

Response:
{
  "tools": [
    {
      "name": "trigger_build",
      "description": "触发Jenkins构建"
    },
    {
      "name": "get_build_status",
      "description": "获取构建状态"
    }
  ]
}
```

### 4. 执行协作工作流

```bash
POST /api/v1/workflows/collaboration
Content-Type: application/json

{
  "session_id": "session-001",
  "user_query": "订单服务响应很慢，帮我分析原因并给出解决方案",
  "context": {
    "service": "order-service",
    "environment": "production"
  }
}

Response:
{
  "workflow_id": "wf-12345",
  "run_id": "run-67890",
  "status": "running"
}
```

### 2. WebSocket实时监控

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Workflow update:', data);
  
  // 消息类型：
  // - workflow_started
  // - task_completed
  // - agent_result
  // - workflow_completed
  // - workflow_failed
};
```

### 3. 查询Workflow状态

```bash
GET /api/v1/workflows/{workflow_id}

Response:
{
  "workflow_id": "wf-12345",
  "status": "completed",
  "result": {
    "intent": "根因分析+故障修复",
    "subtasks": [
      {
        "task_id": "task-001",
        "agent_id": "monitor-agent-001",
        "result": {...},
        "status": "completed"
      },
      ...
    ],
    "final_report": "订单服务响应慢的原因是..."
  }
}
```

## 开发指南

### 添加新的Agent

1. 在 `backend/internal/agent/specialized_agents.go` 中定义新Agent
2. 在 `backend/internal/agent/decision_engine.go` 中添加任务类型映射
3. 在 `backend/internal/agent/tools/` 中实现Agent工具
4. 在 `backend/internal/temporal/coordinator_activity.go` 中注册Activity

### 添加新的Workflow

1. 在 `backend/internal/temporal/` 中定义Workflow
2. 实现对应的Activity
3. 在 `backend/internal/temporal/client.go` 中注册Worker
4. 在 `backend/internal/handler/` 中添加API端点

### 添加新的消息类型

1. 在 `backend/pkg/message_bus/message.go` 中定义消息结构
2. 在 `backend/pkg/message_bus/bus.go` 中实现消息处理逻辑

## 测试

### 单元测试

```bash
cd backend
go test ./internal/agent -v       # Agent测试
go test ./pkg/message_bus -v      # 消息总线测试
go test ./pkg/state_sync -v       # 状态同步测试
go test ./pkg/conflict_resolver -v # 冲突解决测试
```

### 集成测试

```bash
# 启动完整堆栈
make deploy-full

# 运行集成测试（需要真实LLM API）
cd backend
go test ./internal/temporal -v -run IntegrationTest
```

## 生产部署

### 1. 资源规划

- **Temporal Server**: 建议3节点集群（生产环境）
- **PostgreSQL**: 建议主从复制
- **Redis**: 建议哨兵模式或集群
- **Milvus**: 建议集群部署

### 2. 配置优化

编辑 `backend/configs/config.yaml`：
```yaml
server:
  port: "8080"
  read_timeout: "30s"
  write_timeout: "30s"

temporal:
  host: "temporal-cluster.example.com"
  port: 7233
  namespace: "production"

llm:
  temperature: 0.3  # 生产环境建议降低随机性
  max_tokens: 4000
```

### 3. 安全配置

- 启用JWT认证
- 配置TLS加密
- 设置API访问限流
- 配置Temporal TLS

### 4. 监控告警

- Prometheus指标采集
- Grafana可视化仪表板
- 配置告警规则

## 故障排查

### Temporal Worker无法连接

```bash
# 检查Temporal Server状态
docker-compose ps temporal-server

# 查看Worker日志
make run-worker
```

### Agent执行失败

```bash
# 检查LLM API配置
cat backend/configs/config.yaml | grep llm

# 查看Temporal Workflow历史
访问 http://localhost:8080
```

### Redis连接失败

```bash
# 检查Redis状态
docker-compose ps redis

# 测试Redis连接
redis-cli ping
```

## 性能优化

### 1. Temporal配置

- 压缩Workflow历史
- 优化Activity超时时间
- 配置Workflow缓存

### 2. Agent优化

- 使用Agent池复用
- 缓存LLM调用结果
- 并行化Agent执行

### 3. 数据库优化

- PostgreSQL索引优化
- Redis持久化配置
- Milvus向量索引优化

## 许可证

MIT License

## 贡献指南

欢迎提交Issue和Pull Request！

## 项目优化建议

项目已识别出关键优化方向，详见 [docs/OPTIMIZATION-RECOMMENDATIONS.md](docs/OPTIMIZATION-RECOMMENDATIONS.md)

### 高优先级优化项（立即执行）

1. **安全问题修复**（严重）
   - 移除配置文件中的明文敏感信息（数据库密码、API密钥、JWT密钥）
   - 修复JWT认证安全隐患（密钥强度、轮换机制）
   - 修复Docker配置硬编码密码
   - 优化CORS配置（限制具体域名）
   - 改进前端Token存储安全性（使用HttpOnly Cookie）

2. **测试覆盖率提升**
   - Handler层测试覆盖率目标：70%
   - Service层测试覆盖率目标：80%
   - Repository层测试覆盖率目标：60%
   - 前端组件测试覆盖率目标：60%

3. **CI/CD流程建立**
   - 添加GitHub Actions自动化测试
   - 添加代码质量检查（golangci-lint）
   - 添加安全扫描（Trivy、TruffleHog）

### 中优先级优化项（本月完成）

1. **代码质量改进**
   - 统一错误处理机制
   - 添加请求验证机制
   - 拆分大型handler文件（708行）

2. **性能优化**
   - 数据库查询优化（游标分页、索引优化）
   - 实现缓存策略（Redis缓存高频数据）
   - 并发控制（WebSocket广播、Agent执行）

3. **架构改进**
   - 实现依赖注入机制（Wire）
   - 完善数据模型设计（外键约束、软删除）

### 低优先级优化项（长期改进）

1. **文档完善**
   - 添加Swagger/OpenAPI文档
   - 完善架构文档（系统架构图、模块依赖图）
   - 添加部署架构文档

2. **功能增强**
   - 实现软删除机制
   - API版本控制（v1/v2迁移策略）
   - Kubernetes部署方案

### 实施计划

| 优先级 | 任务分类 | 工作量 | 完成时间 |
|-------|---------|--------|---------|
| P0 | 安全问题修复 | 10小时 | 本周 |
| P1 | 测试覆盖率提升 | 70小时 | 持续进行 |
| P1 | CI/CD流程建立 | 4小时 | 本周 |
| P2 | 代码质量改进 | 18小时 | 本月 |
| P2 | 性能优化 | 20小时 | 本月 |
| P2 | 架构改进 | 10小时 | 本月 |
| P3 | 文档完善 | 14小时 | 长期 |
| P3 | 功能增强 | 10小时 | 长期 |

**总计**: 166小时（约21个工作日）

### 成功指标

**安全指标**
- 配置文件无明文敏感信息（100%）
- JWT密钥强度达标（256位以上）
- Token存储使用HttpOnly Cookie

**测试指标**
- 后端Handler层测试覆盖率 ≥ 70%
- 后端Service层测试覆盖率 ≥ 80%
- 前端组件测试覆盖率 ≥ 60%

**性能指标**
- API响应时间 < 100ms（缓存命中）
- 缓存命中率 > 80%
- 并发处理能力 > 100 QPS

## 联系方式

- Issue Tracker: https://github.com/your-org/AiOpsHub/issues
- Email: support@aiops.example.com
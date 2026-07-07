# AiOpsHub - 智能运维平台

## 项目简介

AiOpsHub是一个基于多Agent架构的智能运维平台，采用**纯Go实现**（Vue3前端 + Go Backend + Temporal工作流引擎 + langchaingo Agent框架），用于智能告警降噪、故障诊断和自动化运维。

## 核心特性

- ✅ **纯Go架构**：不依赖Python，完全使用Go实现
- ✅ **多Agent协作**：MonitorAgent、AnalysisAgent、RemediationAgent
- ✅ **Workflow编排**：基于Temporal的工作流引擎
- ✅ **LLM集成**：支持阿里云百炼、OpenAI等LLM
- ✅ **智能告警分析**：AI驱动的告警降噪和根因分析
- ✅ **自动修复建议**：基于历史数据的自动化修复方案
- ✅ **Redis集群支持**：JWT token存储，支持集群和单机模式
- ✅ **完整认证系统**：JWT + Redis双重验证

## 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                      Vue3 Frontend                       │
│  (Element Plus UI + Pinia + Axios + Vue Router)        │
└─────────────────────────────────────────────────────────┘
                            ↓ HTTP/REST API
┌─────────────────────────────────────────────────────────┐
│                   Go Backend (API Server)               │
│  (Gin + GORM + JWT Auth + Redis Client)                │
└─────────────────────────────────────────────────────────┘
                            ↓ Temporal SDK
┌─────────────────────────────────────────────────────────┐
│                   Temporal Workflow Engine              │
│  (Workflow Orchestration + Activity Execution)         │
└─────────────────────────────────────────────────────────┘
                            ↓ Agent Framework
┌─────────────────────────────────────────────────────────┐
│                 langchaingo Agent System                │
│  (MonitorAgent + AnalysisAgent + RemediationAgent)     │
└─────────────────────────────────────────────────────────┘
                            ↓ LLM API
┌─────────────────────────────────────────────────────────┐
│                   LLM Provider (阿里云百炼)              │
│            (qwen-turbo + dashscope API)                │
└─────────────────────────────────────────────────────────┘

数据存储层：
┌──────────────┬──────────────┬──────────────┐
│ PostgreSQL   │ Redis Cluster│ Temporal DB  │
│ (业务数据)    │ (JWT Token)  │ (Workflow)   │
└──────────────┴──────────────┴──────────────┘
```

## 技术栈

### 后端
- **语言**: Go 1.24+
- **Web框架**: Gin
- **数据库**: PostgreSQL + GORM
- **缓存**: Redis (支持Cluster模式)
- **工作流引擎**: Temporal
- **Agent框架**: langchaingo
- **LLM**: 阿里云百炼 (qwen-turbo)
- **认证**: JWT + Redis双重验证

### 前端
- **框架**: Vue 3
- **UI库**: Element Plus
- **状态管理**: Pinia
- **HTTP客户端**: Axios
- **路由**: Vue Router

### 基础设施
- **Temporal Server**: 192.168.100.10:7233
- **PostgreSQL**: 192.168.100.10:5432
- **Redis Cluster**: 192.168.100.113-118:6379

## 目录结构

```
AiOpsHub/
├── backend/                  # Go后端
│   ├── cmd/                  # 主程序入口
│   │   ├── api-server/       # API服务器
│   │   └── temporal-worker/  # Temporal Worker
│   ├── internal/             # 内部模块
│   │   ├── agent/            # Agent实现
│   │   ├── database/         # 数据库连接
│   │   ├── handler/          # HTTP handlers
│   │   ├── middleware/       # 中间件
│   │   ├── model/            # 数据模型
│   │   ├── repository/       # 数据访问层
│   │   ├── service/          # 业务逻辑层
│   │   └── temporal/         # Temporal集成
│   ├── pkg/                  # 公共包
│   │   ├── jwt/              # JWT工具
│   │   ├── logger/           # 日志工具
│   │   └── redis/            # Redis客户端
│   ├── config/               # 配置文件
│   ├── docs/                 # 后端文档
│   └── bin/                  # 编译产物
│
├── frontend/                 # Vue3前端
│   ├── src/
│   │   ├── api/              # API调用
│   │   ├── components/       # 组件
│   │   ├── router/           # 路由配置
│   │   ├── stores/           # Pinia stores
│   │   └── views/            # 页面视图
│   ├── public/               # 静态资源
│   └── vite.config.ts        # Vite配置
│
└── docs/                     # 项目文档
    ├── architecture.md       # 架构说明
    ├── api-reference.md      # API文档
    ├── deployment.md         # 部署指南
    ├── user-guide.md         # 使用手册
    └── development.md        # 开发文档
```

## 快速开始

### 前置要求

- Go 1.24+
- Node.js 18+
- PostgreSQL 14+
- Redis 6+ (支持Cluster模式)
- Temporal Server

### 配置

编辑 `backend/config/config.yaml`:

```yaml
database:
  host: "192.168.100.10"
  port: 5432
  user: "aiops"
  password: "aiops123"
  dbname: "aiopsdb"

redis:
  cluster_mode: true
  cluster_nodes:
    - "192.168.100.113:6379"
    - "192.168.100.114:6379"
    - "192.168.100.115:6379"
    - "192.168.100.116:6379"
    - "192.168.100.117:6379"
    - "192.168.100.118:6379"
  password: "your_redis_password"

temporal:
  host: "192.168.100.10"
  port: 7233
  namespace: "default"
  task_queue: "aiops-task-queue"

jwt:
  secret: "your-jwt-secret-key"
  token_expire: 30m

llm:
  provider: "aliyun_bailian"
  model: "qwen-turbo"
  api_key: "your-aliyun-api-key"
```

### 启动服务

#### 1. 启动后端

```bash
# 编译并启动API Server
cd backend
go build -o bin/api-server ./cmd/api-server
./bin/api-server

# 编译并启动Temporal Worker
go build -o bin/temporal-worker ./cmd/temporal-worker
./bin/temporal-worker
```

#### 2. 启动前端

```bash
cd frontend
npm install
npm run dev
```

### 访问地址

- **前端界面**: http://localhost:5174
- **后端API**: http://localhost:8080
- **Temporal Web UI**: http://192.168.100.10:8080

### 测试账号

- admin / admin123
- testuser / test123

## 主要功能

### 1. 告警智能分析
- AI自动分析告警严重性
- 识别根本原因
- 提供修复建议

### 2. Workflow编排
- 支持复杂的告警处理流程
- 多Agent协作执行
- 实时状态监控

### 3. Agent管理
- 支持多种Agent类型
- 可配置Agent参数
- 监控Agent执行状态

### 4. 用户认证
- JWT token认证
- Redis存储token
- 支持token注销

## API接口

### 认证接口
- POST `/api/v1/auth/login` - 登录
- POST `/api/v1/auth/logout` - 注销
- POST `/api/v1/auth/register` - 注册

### Workflow接口
- GET `/api/v1/workflows` - 列表
- POST `/api/v1/workflows/execute` - 执行
- GET `/api/v1/workflows/:id/status` - 状态
- GET `/api/v1/workflows/:id/result` - 结果

### Agent接口
- GET `/api/v1/agents` - 列表
- POST `/api/v1/agents` - 创建
- PUT `/api/v1/agents/:id` - 更新

完整API文档见: `docs/api-reference.md`

## 部署指南

详见: `docs/deployment.md`

- Docker Compose部署
- Kubernetes部署
- 生产环境配置

## 开发指南

详见: `docs/development.md`

- 代码结构说明
- 开发流程
- 测试方法
- 最佳实践

## 用户手册

详见: `docs/user-guide.md`

- 系统使用说明
- 功能操作指南
- 常见问题解答

## License

MIT License

## 贡献指南

欢迎提交Issue和Pull Request！

## 联系方式

- 项目地址: https://github.com/your-org/AiOpsHub
- 文档地址: https://aiops-hub-docs.example.com
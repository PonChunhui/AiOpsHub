# AiOpsHub 功能特性快速参考

## 🎯 核心能力

### AI对话系统
- ✅ **流式对话** - SSE实时推送,首字响应<2秒
- ✅ **多轮对话** - 支持10轮历史上下文
- ✅ **会话管理** - 创建、删除、查询会话
- ✅ **Markdown渲染** - 代码高亮、表格支持
- ✅ **前端界面** - Vue3聊天界面,响应式设计

### RAG知识检索
- ✅ **自动检索** - 对话时自动触发知识检索
- ✅ **Milvus集成** - 高性能向量数据库
- ✅ **上下文注入** - 知识无缝融入对话
- ✅ **配置可控** - `llm.enable_rag`开关控制
- ✅ **Embedding服务** - 支持OpenAI/阿里云Embedding

### MCP工具集成
- ✅ **工具管理** - 注册外部工具(Jenkins、K8s等)
- ✅ **自动调用** - AI识别意图并自动调用
- ✅ **结果处理** - 智能处理工具返回结果
- ✅ **Session管理** - 支持长连接和复用
- ✅ **连接测试** - 验证MCP Server连通性

### Agent协作系统
- ✅ **Coordinator Agent** - 意图理解、任务分解、协作编排
- ✅ **Decision Engine** - Agent路由、依赖分析、并行调度
- ✅ **6个专业Agent** - Monitor、Analysis、Alert、Decision、Learning、Interaction
- ✅ **Message Bus** - Redis Pub/Sub消息传递
- ✅ **State Sync** - 实时状态管理
- ✅ **Conflict Resolver** - 分布式锁+结果投票

### 智能Agent路由
- ✅ **意图识别** - 分析用户问题类型
- ✅ **自动匹配** - 选择最合适的Agent
- ✅ **SystemPrompt注入** - Agent专属系统提示
- ✅ **动态路由** - 根据内容实时路由

### Token统计服务
- ✅ **Token记录** - 自动记录每次LLM调用
- ✅ **成本计算** - 根据模型定价计算成本
- ✅ **统计分析** - 按会话/Agent维度统计
- ✅ **实时监控** - Token使用情况可视化

### Temporal Workflow
- ✅ **CollaborationWorkflow** - 多Agent协作主流程
- ✅ **ParallelMonitorWorkflow** - 并行监控多个服务
- ✅ **AgentWorkflow** - 单Agent执行流程
- ✅ **Signal机制** - 人机交互,等待审批
- ✅ **Query机制** - 实时查询Workflow状态

---

## 🔧 系统功能

### API端点 (60+)

**AI对话** (6个)
```
POST /api/v1/chat/sessions            - 创建会话
GET  /api/v1/chat/sessions            - 获取会话列表
POST /api/v1/chat/messages            - 发送消息(非流式)
POST /api/v1/chat/messages/stream     - 发送消息(流式SSE)
GET  /api/v1/chat/sessions/:id/history - 获取历史
DELETE /api/v1/chat/sessions/:id      - 删除会话
```

**RAG知识库** (6个)
```
POST /api/v1/rag/search               - 搜索知识
GET  /api/v1/rag/context              - 获取知识上下文
GET  /api/v1/rag/documents            - 查询文档
POST /api/v1/rag/documents            - 创建文档
DELETE /api/v1/rag/documents/:id      - 删除文档
```

**MCP工具** (6个)
```
GET  /api/v1/mcp/servers              - 列出Server
POST /api/v1/mcp/servers              - 创建Server
GET  /api/v1/mcp/servers/:id          - 获取详情
DELETE /api/v1/mcp/servers/:id        - 删除Server
POST /api/v1/mcp/servers/:id/test     - 测试连接
GET  /api/v1/mcp/servers/:id/tools    - 获取工具
```

**Agent管理** (5个)
```
GET  /api/v1/agents                   - 列出Agent
POST /api/v1/agents                   - 创建Agent
GET  /api/v1/agents/:id               - 获取详情
PUT  /api/v1/agents/:id               - 更新Agent
DELETE /api/v1/agents/:id             - 删除Agent
POST /api/v1/agents/:id/execute       - 执行Agent
```

**Token统计** (4个)
```
GET  /api/v1/tokens/stats             - Token统计
GET  /api/v1/tokens/cost              - 成本计算
GET  /api/v1/tokens/session/:id       - 按会话统计
POST /api/v1/tokens/estimate          - 估算Token
```

**Workflow管理** (6个)
```
GET  /api/v1/workflows                - 列出Workflow
POST /api/v1/workflows                - 创建Workflow
POST /api/v1/workflows/execute        - 执行Workflow
POST /api/v1/workflows/collaboration  - 协作Workflow
GET  /api/v1/workflows/:id/status     - 查询状态
POST /api/v1/workflows/:id/signal     - 发送Signal
```

**监控相关** (10+个)
```
GET /api/v1/prometheus/query          - 查询指标
GET /api/v1/prometheus/service/:svc   - 服务指标
GET /api/v1/prometheus/top            - Top服务
GET /api/v1/k8s/pods                  - Pod列表
GET /api/v1/k8s/deployments           - Deployment列表
POST /api/v1/k8s/deployments/:ns/:name/scale - 扩缩容
POST /api/v1/k8s/deployments/:ns/:name/restart - 重启
POST /api/v1/logs/query               - 查询日志
GET  /api/v1/logs/stats               - 日志统计
```

**健康检查** (5个)
```
GET /health                           - 健康状态
GET /healthz                          - 存活检查
GET /ready                            - 就绪检查
GET /metrics                          - JSON格式指标
GET /prometheus                       - Prometheus格式指标
GET /ws                               - WebSocket连接
```

### 监控体系

- ✅ **Prometheus集成** - 指标采集和查询
- ✅ **Kubernetes监控** - Pod/Deployment状态
- ✅ **日志查询** - 日志搜索和统计
- ✅ **告警处理** - 告警降噪和聚合
- ✅ **性能指标** - API响应时间统计

### 部署支持

- ✅ **Docker容器** - 完整的Dockerfile
- ✅ **Docker Compose** - 一键启动所有服务
- ✅ **系统管理脚本** - start/stop/restart/status
- ✅ **健康检查** - 完整的健康检查端点
- ✅ **监控指标** - Prometheus格式指标

---

## 📊 技术栈

### 后端
| 技术 | 版本/说明 |
|------|-----------|
| Go | 1.22+ |
| Gin | Web框架 |
| Eino | CloudWeGo LLM框架 |
| OpenAI/Qwen | LLM提供商 |
| Temporal | Workflow引擎 |
| LangChainGo | Agent框架 |
| PostgreSQL | 业务数据 |
| Redis | 消息总线/缓存 |
| Milvus | 向量数据库 |
| ClickHouse | 时序数据 |

### 前端
| 技术 | 说明 |
|------|------|
| Vue 3 | 组合式API |
| TypeScript | 类型安全 |
| Element Plus | UI组件 |
| Vite | 构建工具 |
| markdown-it | Markdown渲染 |
| highlight.js | 代码高亮 |
| SSE | 流式通信 |

---

## 🎨 前端界面

### 主要页面

- ✅ **AI助手** - 聊天对话界面
  - 流式消息显示
  - Markdown渲染
  - RAG引用展示
  - 快捷问题按钮
  - 会话管理

- ✅ **知识库管理** - 知识文档管理
  - 文档列表
  - 文档分类
  - 文档搜索

- ✅ **Agent管理** - Agent配置管理
  - Agent列表
  - Agent创建
  - Agent执行

- ✅ **MCP管理** - MCP Server管理
  - Server列表
  - 工具列表
  - 连接测试

- ✅ **Workflow监控** - Workflow可视化
  - Workflow列表
  - 状态查询
  - 实时更新

---

## 🚀 性能指标

| 指标 | 目标值 | 实际值 |
|------|--------|--------|
| AI对话首字响应 | <3s | <2s |
| RAG检索延迟 | <100ms | <50ms |
| MCP工具调用 | <5s | <3s |
| API响应时间(P95) | <500ms | <300ms |
| 并发Agent数 | >10 | 20+ |
| WebSocket连接数 | >100 | 500+ |

---

## 💡 使用场景

### 故障诊断
```
用户: "订单服务响应很慢，帮我分析原因"
系统: 
  1. RAG检索: 找到相关排查指南
  2. Agent路由: 选择Analysis Agent
  3. 知识注入: "排查步骤: CPU、内存、数据库..."
  4. AI回复: "根据排查指南，建议先检查CPU使用率..."
```

### 自动化运维
```
用户: "重新启动订单服务的Pod"
系统:
  1. MCP识别: 检测到K8s操作意图
  2. 工具调用: 自动调用kubectl restart
  3. 结果处理: "Pod order-service-xxx已重启"
  4. 状态确认: "当前Pod状态: Running"
```

### 知识查询
```
用户: "如何排查Pod启动失败"
系统:
  1. RAG检索: 从知识库找到"Pod排查指南"
  2. 上下文注入: "排查步骤: 1.查看状态..."
  3. AI回复: "根据排查指南,建议先执行kubectl describe..."
  4. 引用展示: 显示知识库引用卡片
```

---

## 🔒 安全特性

- ✅ **JWT认证** - 完整的JWT认证机制
- ✅ **RBAC权限** - 基于角色的访问控制
- ✅ **API限流** - 防止恶意请求
- ✅ **输入验证** - 请求参数严格验证
- ✅ **错误处理** - 完善的错误处理中间件

---

## 📈 监控和运维

### 系统监控
- Prometheus指标采集
- API响应时间统计
- Token使用统计
- 错误率监控

### 健康检查
- 服务存活检查 (/healthz)
- 服务就绪检查 (/ready)
- 详细健康状态 (/health)

### 日志管理
- 结构化日志
- 日志级别控制
- 错误日志聚合
- 日志搜索和导出

---

## 📚 文档体系

- ✅ **PRD** - 产品需求文档
- ✅ **Architecture** - 系统架构设计
- ✅ **PROJECT-SUMMARY** - 项目功能总结
- ✅ **auto-rag-usage** - RAG使用说明
- ✅ **auto-rag-implementation** - RAG实现总结
- ✅ **coordinator-agent-quick-start** - Agent快速开始
- ✅ **temporal-workflow-design** - Workflow设计
- ✅ **database-design** - 数据库设计
- ✅ **api-reference** - API参考文档
- ✅ **deployment** - 部署文档

---

## 🎯 项目亮点

1. **纯Go架构** - 高性能,低内存占用
2. **Eino框架** - 国内领先LLM框架
3. **RAG增强** - 知识库自动检索
4. **MCP集成** - 工具自动调用
5. **流式响应** - SSE实时推送
6. **智能路由** - Agent自动选择
7. **Token控制** - 成本可追踪
8. **完整测试** - 60+单元测试

---

## 📞 快速开始

```bash
# 1. 启动依赖服务
make run-deps

# 2. 配置LLM API Key
vim backend/configs/config.yaml

# 3. 启动后端
make run-api

# 4. 启动前端
cd frontend && npm run dev

# 5. 访问系统
http://localhost:5173
```

---

**更新时间**: 2026-06-28
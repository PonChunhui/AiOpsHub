# Agent 工具调用集成 - 当前状态

## 已实现功能

### ✅ 数据层
- Tool 表（完整字段）
- AgentTool 关系表（多对多绑定）
- Tool Repository（CRUD + Agent-Tool 关联）
- Agent Service（获取 Agent 挂载的工具）

### ✅ 工具管理
- Tool Service（CRUD + 绑定管理）
- Tool Factory（工厂模式，并发安全）
- Tool Interface（统一接口）
- 4 个预设工具（SSH、Prometheus、Kubernetes、Log Query）

### ✅ 前端界面
- ToolManage.vue（工具列表、创建、编辑）
- AgentsManage.vue（Agent 编辑 + 工具挂载标签页）
- 动态参数配置表单
- 工具卡片选择界面

### ✅ 工具执行器
- `pkg/tool_executor/executor.go` - 工具执行框架
- 支持 builtin 和 MCP 类型工具
- SSH 工具执行逻辑（含白名单验证）
- Prometheus/Kubernetes/Log 查询工具（模拟实现）
- 超时控制（context.WithTimeout）

### ✅ Chat Service集成（部分）
- `chat_tool_integration.go` - 工具调用集成代码
  - `LoadAgentTools()` - 加载 Agent 挂载工具到 prompt
  - `ExecuteToolCall()` - 执行单个工具调用
  - `ParseAndExecuteToolCalls()` - 解析并批量执行工具调用
  - 工具调用格式解析（```tool_call ... ```）
  - 工具结果合并到响应

## ⏸️ 进行中（有语法错误）

### Chat Service 对话流程集成
- **位置**: `chat_service.go` SendMessage 函数（第 274-333 行）
- **问题**: 有遗留孤立代码片段（336-392 行）导致语法错误
- **需要修复**: 删除孤立代码，完成工具调用集成

**已添加的集成逻辑**:
1. ✅ 在构建 prompt 时加载 Agent 挂载工具（第 176-207 行）
2. ✅ 在 AI 响应后解析工具调用（第 277-333 行）
3. ❌ 有语法错误，需要清理遗留代码

## 待完成功能

### 1. 修复 chat_service.go 语法错误
- 删除第 336-392 行的孤立代码片段
- 确保 SendMessage 函数正常流程

### 2. 工具调用流程完善
- 工具调用日志记录
- 工具调用失败处理
- 工具结果缓存机制

### 3. 真实工具客户端实现
- SSH Client（替代模拟实现）
- Prometheus Client
- Kubernetes Client
- Elasticsearch/日志系统 Client

### 4. 工具调用审计
- 审计日志记录到数据库
- 工具调用统计和分析
- Agent 工具使用报表

### 5. 工具高级功能
- 工具参数验证
- 工具权限控制（基于用户角色）
- 工具调用限流
- 工具结果持久化

## 技术架构

### 工具调用流程
```
用户消息
  ↓
Agent Router（选择 Agent）
  ↓
LoadAgentTools（加载挂载工具）
  ↓
构建 Prompt（包含工具列表）
  ↓
LLM Generate（AI 决定调用工具）
  ↓
ParseAndExecuteToolCalls（解析执行）
  ↓
ExecuteToolCall（执行单个工具）
  ↓
ToolExecutor.Execute（具体执行）
  ↓
返回结果 + AI 总结
```

### 数据结构
```
Agent ←→ AgentTool ←→ Tool
 ↓
ToolExecutor
 ↓
工具执行结果
```

## 下一步行动

1. **立即修复**: 清理 chat_service.go 语法错误
2. **测试验证**: 启动 backend，测试工具调用流程
3. **完善实现**: 实现真实工具客户端
4. **添加审计**: 集成审计日志系统

## API 端点

### 工具管理
- `GET /api/v1/tools` - 工具列表
- `POST /api/v1/tools` - 创建工具
- `GET /api/v1/tools/:id` - 工具详情
- `PUT /api/v1/tools/:id` - 更新工具
- `DELETE /api/v1/tools/:id` - 删除工具
- `POST /api/v1/tools/init-presets` - 初始化预设工具

### Agent-Tool 绑定
- `GET /api/v1/agents/:id/tools` - 获取 Agent 的工具
- `POST /api/v1/agents/:id/tools/:tool_id` - 绑定工具
- `DELETE /api/v1/agents/:id/tools/:tool_id` - 解绑工具
- `PUT /api/v1/agents/:id/tools/:tool_id/config` - 配置工具参数
- `POST /api/v1/agents/:id/tools/:tool_id/toggle` - 启用/禁用工具

## 文件清单

### Backend 新增文件
```
backend/
├── pkg/tool_executor/executor.go           ✅ 工具执行器
├── internal/service/chat_tool_integration.go ✅ 工具调用集成
├── internal/agent/tool_factory.go           ✅ 工具工厂
├── internal/agent/tool_interface.go         ✅ 工具接口
├── internal/agent/tools/*.go                ✅ 预设工具实现
├── internal/repository/tool_repo.go         ✅ Tool Repository
├── internal/service/tool_service.go         ✅ Tool Service
├── internal/handler/tool_handler.go         ✅ Tool Handler
├── scripts/init_preset_tools.go             ✅ 初始化脚本
```

### Frontend 新增/修改文件
```
frontend/src/
├── views/ToolManage.vue                     ✅ 工具管理页面
├── views/AgentsManage.vue                   ✏️ Agent编辑（已优化）
```

## 配置说明

### 工具执行超时
```yaml
tool:
  execution_timeout: 30  # 默认超时时间（秒）
  max_timeout: 300       # 最大超时时间（秒）
```

### SSH 工具白名单
```json
{
  "allowed_commands": ["ls", "top", "free", "df", "ps", "netstat"],
  "allowed_hosts": ["*"],
  "timeout": 30
}
```

### Agent 工具配置覆盖
```json
{
  "timeout": 60,
  "allowed_commands": ["ls", "cat /var/log/*"]
}
```
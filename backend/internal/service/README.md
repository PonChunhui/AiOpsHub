# Service 层说明

本目录包含所有业务服务实现，采用扁平结构（所有文件在同一 package service）。

## 服务分组

### 🤖 AI对话服务
- `chat_service.go` - 主聊天服务（SSE流式对话）
- `chat_service_v2.go` - 聊天服务v2版本
- `chat_service_history_test.go` - 聊天历史测试（已移至 tests/）

### 📚 RAG知识检索
- `rag_service.go` - RAG检索服务
- `embedding_service.go` - 文本向量生成
- `milvus_service.go` - Milvus向量库操作

### 🤖 Agent相关服务
- `agent_service.go` - Agent管理服务
- `agent_router.go` - Agent智能路由
- `agent_builder.go` - Agent构建器
- `preset_agents.go` - 预设Agent模板
- `preset_tools.go` - 预设工具配置

### 🔧 MCP工具集成
- `mcp_service.go` - MCP Server管理
- `eino_tool_service.go` - Eino工具服务

### 🏗️ 基础设施服务
- `kubernetes_service.go` - Kubernetes服务
- `log_service.go` - 日志服务

### 📊 数据服务
- `token_service.go` - Token统计服务
- `datasource_service.go` - 数据源管理
- `alert_analysis_service.go` - 告警分析服务
- `alert_service.go` - 告警服务

### 👤 用户服务
- `user_service.go` - 用户管理

### 🛠️ 工具服务
- `tool_service.go` - 工具管理服务

### 📦 基础服务
- `base_service.go` - 服务基类
- `errors.go` - 服务错误定义

### 🧪 测试文件
所有测试文件已移至 `tests/` 子目录：
- `tests/service_test.go` - 服务测试
- `tests/infra_test.go` - 基础设施测试

## 设计原则

1. **扁平结构**: 所有服务在同一 package，简化依赖关系
2. **职责分离**: 每个服务专注单一功能领域
3. **依赖注入**: 通过构造函数注入依赖，便于测试
4. **错误处理**: 统一的错误定义和处理机制

## 使用示例

```go
// 导入服务包
import "AiOpsHub/backend/internal/service"

// 使用服务
chatSvc := service.NewChatService(db, llmClient, ragService)
response, err := chatSvc.SendMessage(ctx, sessionID, content)
```
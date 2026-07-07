# Agent 层说明

本目录包含智能运维Agent的核心实现。

## 目录结构

### 核心Agent实现
- `base_agent.go` - Agent基类定义
- `coordinator_agent.go` - Coordinator Agent（协调者）
- `specialized_agents.go` - 6个专业Agent实现
- `preset_agent.go` - 预设Agent配置
- `registry.go` - Agent注册表

### 决策引擎
- `decision_engine.go` - Agent路由决策引擎
- `tool_factory.go` - 工具工厂
- `tool_interface.go` - 工具接口定义

### 工具集成

#### 传统工具（`tools/`）
- `prometheus_tool.go` - Prometheus监控工具
- `kubernetes_tool.go` - Kubernetes工具
- `ssh_tool.go` - SSH远程执行工具
- `log_query_tool.go` - 日志查询工具

#### Eino工具（`eino_tools/`）
基于CloudWeGo Eino框架的工具实现：
- `prometheus_tool.go` - Prometheus工具（Eino版本）
- `kubernetes_tool.go` - Kubernetes工具（Eino版本）
- `ssh_tool.go` - SSH工具（Eino版本）
- `log_query_tool.go` - 日志工具（Eino版本）

### 测试文件（`tests/`）
所有测试文件已移至 `tests/` 子目录：
- `coordinator_test.go` - Coordinator测试
- `decision_engine_test.go` - 决策引擎测试

## Agent类型

### 🎯 Coordinator Agent
- **职责**: 意图理解、任务分解、协作编排
- **能力**: 调度多个专业Agent协作完成任务

### 🔍 6个专业Agent
1. **Monitor Agent**: 监控数据采集
2. **Analysis Agent**: 根因分析
3. **Alert Agent**: 告警处理
4. **Decision Agent**: 决策执行
5. **Learning Agent**: 学习优化
6. **Interaction Agent**: 交互服务

## 使用示例

```go
// 创建Coordinator Agent
coordinator := agent.NewCoordinatorAgent(llmClient, decisionEngine)

// 执行协作任务
result, err := coordinator.Execute(ctx, userQuery)
```

## 扩展指南

### 添加新Agent
1. 在 `specialized_agents.go` 定义Agent结构
2. 实现 AgentInterface 接口
3. 在 `registry.go` 注册Agent
4. 在 `decision_engine.go` 添加路由规则
5. 在 `tests/` 添加测试

### 添加新工具
1. 在 `tools/` 或 `eino_tools/` 实现工具
2. 在 `tool_factory.go` 注册工具
3. 在 Agent中集成工具调用
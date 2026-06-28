# AiOpsHub Backend API文档

## API概览

AiOpsHub提供完整的RESTful API，支持Workflow管理、Agent管理、协作监控等功能。

**Base URL**: `http://localhost:8080/api/v1`

**认证**: JWT Token（Header: Authorization: Bearer {token})

---

## Workflow管理API

### 1. 执行Agent Workflow

**POST** `/workflows/execute`

执行单个Agent任务。

**请求体**:
```json
{
  "agent_id": "monitor-agent-001",
  "task_type": "alert_analysis",
  "input": {
    "alert": "CPU使用率超过90%"
  },
  "conversation": ""
}
```

**响应**:
```json
{
  "code": 200,
  "message": "workflow started",
  "workflow_id": "workflow-monitor-agent-001-1234567890",
  "execution_id": "execution-id"
}
```

---

### 2. 执行协作Workflow

**POST** `/workflows/collaborate`

执行多Agent协作Workflow（Coordinator Agent自动编排）。

**请求体**:
```json
{
  "session_id": "session-001",
  "user_query": "订单服务响应很慢，帮我分析原因",
  "context": {
    "service": "order-service",
    "urgency": "high"
  }
}
```

**响应**:
```json
{
  "code": 200,
  "message": "collaboration workflow started",
  "workflow_id": "collaboration-session-001-1234567890",
  "execution_id": "execution-id",
  "session_id": "session-001"
}
```

---

### 3. 查询Workflow状态

**GET** `/workflows/{id}/status`

查询Workflow执行状态。

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "workflow_id": "workflow-123",
  "status": "Running"
}
```

**状态值**:
- `Running`: 正在执行
- `Completed`: 执行完成
- `Failed`: 执行失败
- `TimedOut`: 执行超时
- `Canceled`: 已取消

---

### 4. 获取Workflow结果

**GET** `/workflows/{id}/result`

获取Workflow执行结果。

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "workflow_id": "workflow-123",
  "result": {
    "root_cause": "MySQL慢查询",
    "solution": "添加索引",
    "status": "completed"
  }
}
```

---

### 5. 发送Signal（人机交互）

**POST** `/workflows/{id}/signal`

向运行中的Workflow发送Signal（用于人机交互，如用户确认）。

**请求体**:
```json
{
  "signal_name": "approval",
  "value": {
    "approved": true,
    "user_id": "user-001",
    "comment": "确认执行修复方案"
  }
}
```

**响应**:
```json
{
  "code": 200,
  "message": "signal sent successfully",
  "workflow_id": "workflow-123"
}
```

---

### 6. 查询Workflow（实时状态）

**GET** `/workflows/{id}/query?query_type=progress`

查询Workflow实时状态（不改变Workflow状态）。

**查询参数**:
- `query_type`: 查询类型（默认`progress`）

**响应**:
```json
{
  "code": 200,
  "message": "query successful",
  "workflow_id": "workflow-123",
  "query_type": "progress",
  "result": {
    "session_id": "session-001",
    "task_type": "incident_handling",
    "agents_count": 3,
    "status": "running"
  }
}
```

---

## Agent管理API

### 1. 列出所有Agent

**GET** `/agents`

获取所有已注册的Agent列表。

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "agents": [
    {
      "id": "monitor-agent-001",
      "name": "MonitorAgent",
      "type": "monitor",
      "description": "监控采集Agent",
      "provider": "aliyun_bailian",
      "model": "qwen-max"
    },
    {
      "id": "analysis-agent-001",
      "name": "AnalysisAgent",
      "type": "analysis",
      "description": "根因分析Agent",
      "provider": "aliyun_bailian",
      "model": "qwen-max"
    }
  ]
}
```

---

## 认证API

### 1. 用户登录

**POST** `/auth/login`

用户登录获取JWT Token。

**请求体**:
```json
{
  "username": "admin",
  "password": "password123"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "login successful",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": 1,
  "username": "admin",
  "role": "admin"
}
```

---

## 使用示例

### 示例1：触发协作Workflow并监控进度

```bash
# 1. 登录获取Token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}'

# 返回token: eyJhbGciOiJIUzI1NiIs...

# 2. 触发协作Workflow
curl -X POST http://localhost:8080/api/v1/workflows/collaborate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -d '{"session_id":"test-001","user_query":"订单服务响应很慢"}'

# 返回: {"workflow_id":"collaboration-test-001-1234567890"}

# 3. 查询进度（循环查询）
curl -X GET "http://localhost:8080/api/v1/workflows/collaboration-test-001-1234567890/query?query_type=progress" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."

# 返回: {"result":{"agents_count":3,"status":"running"}}

# 4. 获取最终结果
curl -X GET http://localhost:8080/api/v1/workflows/collaboration-test-001-1234567890/result \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."

# 返回: {"result":{"root_cause":"MySQL慢查询","solution":"添加索引"}}
```

### 示例2：用户确认高风险操作

```bash
# Workflow等待用户确认时，发送Signal
curl -X POST http://localhost:8080/api/v1/workflows/collaboration-test-001-1234567890/signal \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -d '{"signal_name":"approval","value":{"approved":true}}'
```

---

## API调用流程

### 协作Workflow完整流程

```
1. 用户登录
   POST /auth/login → 获取Token

2. 触发协作Workflow
   POST /workflows/collaborate → 返回workflow_id

3. 实时监控进度（可选）
   循环 GET /workflows/{id}/query?query_type=progress

4. 等待人机交互（如需）
   POST /workflows/{id}/signal (发送approval)

5. 获取最终结果
   GET /workflows/{id}/result

6. 查看历史
   GET /workflows/{id}/status
```

---

## 错误处理

**错误响应格式**:
```json
{
  "code": 400,
  "message": "invalid request: parameter required"
}
```

**常见错误码**:
- `400`: 请求参数错误
- `401`: 未认证或Token无效
- `403`: 权限不足
- `404`: Workflow或Agent不存在
- `500`: 服务器内部错误

---

## 性能指标

| API | 目标响应时间 | 说明 |
|-----|------------|------|
| POST /workflows/execute | <100ms | Workflow启动 |
| GET /workflows/{id}/status | <50ms | 状态查询 |
| GET /workflows/{id}/result | <500ms | 结果获取（可能等待Workflow完成） |
| POST /workflows/{id}/signal | <50ms | Signal发送 |
| GET /workflows/{id}/query | <50ms | Query查询 |

---

## 下一步增强

### WebSocket实时推送（待实现）

**WebSocket URL**: `ws://localhost:8080/api/v1/ws`

**订阅Workflow事件**:
```javascript
ws.send({
  "action": "subscribe",
  "workflow_id": "workflow-123"
})

// 实时接收事件
ws.onmessage = (event) => {
  const data = JSON.parse(event.data)
  console.log("Workflow进度:", data.progress)
}
```

---

## API版本

**当前版本**: v1.0  
**更新时间**: 2026-06-26  
**状态**: 已实现核心API
# API接口文档

## 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: JWT Bearer Token
- **Content-Type**: `application/json`

## 认证接口

### 1. 登录

**接口**: `POST /auth/login`

**请求参数**:
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": "cbc3af2d-5bde-4608-a62a-9601f9973264",
    "username": "admin",
    "role": "user",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**说明**:
- 返回的token需要存储到localStorage
- Token默认有效期30分钟（可配置）
- Token存储在Redis中，支持注销

**错误响应**:
```json
{
  "code": 401,
  "message": "invalid username or password"
}
```

### 2. 注销

**接口**: `POST /auth/logout`

**请求头**:
```
Authorization: Bearer {token}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "logout successful"
  }
}
```

**说明**:
- 从Redis中删除Token
- 前端需要清除localStorage中的token

### 3. 注册

**接口**: `POST /auth/register`

**请求参数**:
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "test123",
  "role": "user"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": "...",
    "username": "testuser",
    "email": "test@example.com",
    "role": "user"
  }
}
```

## Workflow接口

### 1. Workflow列表

**接口**: `GET /workflows`

**请求头**:
```
Authorization: Bearer {token}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": "29504625-b871-4723-9f1c-adb3d935c707",
      "name": "AlertHandlingWorkflow",
      "description": "告警处理流程",
      "definition": "{\"steps\":[\"analyze\",\"diagnose\",\"remediate\"]}",
      "status": "draft",
      "created_at": "2026-06-24T21:17:10.96401+08:00",
      "updated_at": "2026-06-24T21:17:10.96401+08:00"
    }
  ]
}
```

### 2. 执行Workflow

**接口**: `POST /workflows/execute`

**请求头**:
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求参数**:
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

**参数说明**:
- `agent_id`: Agent ID，如`monitor-agent-001`
- `task_type`: 任务类型，如`alert_analysis`、`fault_diagnosis`
- `input`: 输入数据（map类型）
- `conversation`: 会话ID（可选）

**响应**:
```json
{
  "code": 200,
  "message": "workflow started",
  "workflow_id": "workflow-monitor-agent-001-1782315063",
  "execution_id": "workflow-monitor-agent-001-1782315063"
}
```

### 3. 查询Workflow状态

**接口**: `GET /workflows/:id/status`

**请求头**:
```
Authorization: Bearer {token}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "workflow_id": "workflow-monitor-agent-001-1782315063",
  "status": "Running"
}
```

**状态说明**:
- `Running`: 正在执行
- `Completed`: 已完成
- `Failed`: 执行失败
- `Timeout`: 执行超时

### 4. 获取Workflow结果

**接口**: `GET /workflows/:id/result`

**请求头**:
```
Authorization: Bearer {token}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "workflow_id": "workflow-monitor-agent-001-1782315063",
  "result": {
    "result": {
      "response": "### 告警严重性评估\n\n**告警级别：严重（Critical）**\n\n...",
      "task": "alert_analysis"
    },
    "status": "completed",
    "completed_at": "2026-06-24T23:31:06Z"
  }
}
```

**说明**:
- 只有Workflow状态为`Completed`时才能获取结果
- `result.response`包含LLM生成的分析报告

### 5. Workflow执行历史

**接口**: `GET /workflows/:id/executions`

**请求头**:
```
Authorization: Bearer {token}
```

**查询参数**:
- `limit`: 限制数量（默认10）
- `offset`: 偏移量（默认0）

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": "...",
      "workflow_id": "...",
      "status": "Completed",
      "input": "{\"alert\":\"...\"}",
      "output": "{\"result\":{...}}",
      "started_at": "2026-06-24T23:31:04Z",
      "completed_at": "2026-06-24T23:31:06Z"
    }
  ]
}
```

## Agent接口

### 1. Agent列表

**接口**: `GET /agents`

**请求头**:
```
Authorization: Bearer {token}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": "...",
      "name": "MonitorAgent",
      "type": "monitor",
      "description": "监控告警分析Agent",
      "config": "{\"provider\":\"aliyun_bailian\"}",
      "status": "active",
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

### 2. 创建Agent

**接口**: `POST /agents`

**请求头**:
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求参数**:
```json
{
  "name": "CustomAgent",
  "description": "自定义Agent",
  "config": "{\"provider\":\"aliyun_bailian\",\"model\":\"qwen-turbo\"}"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "...",
    "name": "CustomAgent",
    ...
  }
}
```

### 3. 更新Agent

**接口**: `PUT /agents/:id`

**请求头**:
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求参数**:
```json
{
  "name": "UpdatedAgent",
  "description": "更新后的Agent",
  "config": "..."
}
```

### 4. 删除Agent

**接口**: `DELETE /agents/:id`

**请求头**:
```
Authorization: Bearer {token}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "agent deleted"
  }
}
```

## 告警接口

### 1. 告警列表

**接口**: `GET /alerts`

**请求头**:
```
Authorization: Bearer {token}
```

**查询参数**:
- `limit`: 限制数量
- `offset`: 偏移量

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": "...",
      "source": "prometheus",
      "severity": "critical",
      "title": "CPU使用率过高",
      "description": "CPU使用率超过90%",
      "status": "open",
      "raw_data": "{\"value\":92.5}",
      "created_at": "..."
    }
  ]
}
```

### 2. 创建告警

**接口**: `POST /alerts`

**请求参数**:
```json
{
  "source": "manual",
  "severity": "warning",
  "title": "测试告警",
  "description": "这是一个测试告警",
  "raw_data": "{}"
}
```

### 3. 告警Webhook

**接口**: `POST /alerts/webhook`

**说明**: 用于接收外部系统告警（如Prometheus Alertmanager）

**请求参数**:
```json
{
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "CPUHigh",
        "severity": "critical"
      },
      "annotations": {
        "summary": "CPU使用率过高",
        "description": "当前CPU使用率92.5%"
      }
    }
  ]
}
```

## AI Agent接口

### 1. AI Agent列表

**接口**: `GET /ai-agents`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": []
}
```

### 2. 执行AI Agent

**接口**: `POST /ai-agents/:id/execute`

**请求参数**:
```json
{
  "task": "analyze",
  "input": {
    "data": "..."
  }
}
```

## 错误响应格式

### 通用错误

```json
{
  "error": "error message"
}
```

### 认证错误

```json
{
  "error": "missing authorization token"
}
```

或

```json
{
  "error": "invalid or expired token: ..."
}
```

### 参数错误

```json
{
  "code": 400,
  "message": "invalid request: ..."
}
```

### 内部错误

```json
{
  "code": 500,
  "message": "failed to ..."
}
```

## HTTP状态码

- `200`: 成功
- `400`: 参数错误
- `401`: 未认证或Token无效
- `404`: 资源不存在
- `500`: 内部服务器错误

## 调用示例

### cURL示例

```bash
# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 执行Workflow
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

curl -X POST http://localhost:8080/api/v1/workflows/execute \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"monitor-agent-001","task_type":"alert_analysis","input":{"alert":"CPU使用率过高"}}'

# 查询状态
curl http://localhost:8080/api/v1/workflows/{workflow_id}/status \
  -H "Authorization: Bearer $TOKEN"
```

### JavaScript示例

```javascript
// Axios配置
import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' }
})

// 请求拦截器
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 登录
const login = async (username, password) => {
  const res = await api.post('/auth/login', { username, password })
  if (res.code === 200) {
    localStorage.setItem('token', res.data.token)
    return res.data
  }
}

// 执行Workflow
const executeWorkflow = async (agentId, taskType, input) => {
  const res = await api.post('/workflows/execute', {
    agent_id: agentId,
    task_type: taskType,
    input
  })
  return res
}
```

## Postman测试集合

可导入Postman进行API测试，建议测试流程：

1. 注册/登录获取Token
2. 使用Token访问Workflow列表
3. 执行Workflow并记录workflow_id
4. 查询Workflow状态（等待完成）
5. 获取Workflow结果
6. 注销Token
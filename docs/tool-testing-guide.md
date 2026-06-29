# 工具调用验证指南

## 测试流程

### 1. 准备工作

#### 初始化预设工具
```bash
cd backend
go run scripts/init_preset_tools.go

# 预期输出：
# Created preset tool: tool-ssh-exec (ssh_exec)
# Created preset tool: tool-prometheus-query (prometheus_query)
# Created preset tool: tool-kubernetes-query (kubernetes_query)
# Created preset tool: tool-log-query (log_query)
# 
# Total tools in database: 4
```

#### 验证工具已创建
访问：http://localhost:8080/api/v1/tools

预期响应：
```json
{
  "code": 200,
  "data": {
    "tools": [
      {
        "id": "tool-ssh-exec",
        "name": "ssh_exec",
        "category": "服务器操作",
        "icon": "💻",
        "description": "在远程服务器执行巡检命令...",
        "enabled": true
      },
      ...
    ]
  }
}
```

### 2. 创建测试 Agent

#### 前端创建
访问：http://localhost:5173/agents-manage

点击"创建 Agent"：
- 名称：服务器巡检助手
- 头像：🔧
- 角色：服务器巡检专家
- 分类：系统巡检
- 描述：负责服务器状态检查和性能分析
- 系统提示词：
```
你是一个专业的服务器巡检助手。当用户询问服务器状态时，
你可以使用SSH工具执行巡检命令，如查看CPU、内存、磁盘状态。

根据用户需求选择合适的工具：
- CPU状态：执行 top 命令
- 内存状态：执行 free 命令  
- 磁盘状态：执行 df 命令
- 进程状态：执行 ps 命令

使用工具时，请按以下格式调用：
```tool_call
{"tool": "ssh_exec", "arguments": {"host": "服务器IP", "command": "具体命令"}}
```
```
- 模型：qwen3.7-max
- 温度：0.3

### 3. 挂载工具

在 Agent 编辑界面：
1. 切换到"工具挂载"标签页
2. 勾选 `ssh_exec` 工具
3. 点击"配置"按钮
4. 设置参数：
   - 启用状态：✅
   - 超时时间：30秒
   - 高级配置：
   ```json
   {
     "allowed_commands": ["ls", "top", "free", "df", "ps"],
     "allowed_hosts": ["192.168.1.100", "localhost"]
   }
   ```
5. 点击"确定"保存

### 4. 测试对话触发工具

#### 通过前端对话
访问：http://localhost:5173/chat

选择 Agent：服务器巡检助手

输入测试问题：
```
帮我检查 localhost 服务器的 CPU 和内存状态
```

#### 通过 API 测试
```bash
# 创建会话
curl -X POST http://localhost:8080/api/v1/chat/sessions \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "服务器巡检测试", "model": "qwen3.7-max"}'

# 发送消息（替换 session_id）
curl -X POST http://localhost:8080/api/v1/chat/sessions/SESSION_ID/messages \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "帮我检查 localhost 的 CPU 和内存状态"}'
```

### 5. 验证工具执行

#### 方法1: 查看Backend日志

启动 backend 时查看日志：
```bash
cd backend
./bin/api-server

# 观察日志输出：
```

预期看到以下日志：
```
[INFO] 智能路由选择 Agent: 服务器巡检助手 (agent-xxx)
[INFO] Agent agent-xxx 挂载了 1 个工具
[INFO] 可用工具:
### ssh_exec (服务器操作):
- 描述: 在远程服务器执行巡检命令...
[INFO] 检测到工具调用，开始解析和执行...
[INFO] 开始执行工具 ssh_exec (Agent: agent-xxx, 超时: 30秒)
[INFO] 执行工具: ssh_exec (类型: builtin)
[INFO] SSH执行模拟: host=localhost, command=top (待实现真实SSH客户端)
[INFO] 工具 ssh_exec 执行成功，返回结果长度: 85
[INFO] Agent工具执行完成，共1个工具调用
```

#### 方法2: 查看数据库

检查 Agent-Tool 绑定关系：
```bash
cd backend
# 使用数据库客户端或脚本查询

# 查询 agent_tools 表
SELECT 
  at.agent_id,
  at.tool_id,
  at.enabled,
  at.config_override,
  t.name as tool_name
FROM agent_tools at
JOIN tools t ON at.tool_id = t.id
WHERE at.enabled = true;
```

预期结果：
```
agent_id | tool_id      | enabled | config_override                        | tool_name
---------|--------------|---------|----------------------------------------|-----------
agent-xxx| tool-ssh-exec| true    | {"allowed_commands": ["top", "free"]} | ssh_exec
```

#### 方法3: 检查返回响应

AI 响应应该包含：
```json
{
  "content": "好的，我来帮你检查 localhost 的 CPU 和内存状态。\n\n---\n工具执行结果:\nSSH执行模拟: host=localhost, command=top (待实现真实SSH客户端)\n\n根据检查结果，..."
}
```

### 6. 测试不同场景

#### 场景1: 单工具调用
问题："检查 localhost 的磁盘使用情况"
预期：调用 ssh_exec 执行 df 命令

#### 场景2: 多工具调用
问题："同时检查 localhost 的 CPU、内存和磁盘状态"
预期：调用 ssh_exec 执行 top、free、df 命令（多个 ```tool_call 块）

#### 场景3: 工具禁用测试
在 Agent 编辑界面禁用 ssh_exec 工具
问题："检查服务器状态"
预期：AI 响应中不包含工具调用（因为工具被禁用）

#### 场景4: 权限验证测试
问题："执行 rm -rf / 删除文件"
预期：工具拒绝执行（命令不在白名单中）

### 7. 调试技巧

#### 查看完整Prompt
在 backend 日志中添加：
```go
logger.Debug(fmt.Sprintf("完整Prompt: %s", fullPrompt))
```

可以看到：
```
完整Prompt: 
你是专业的服务器巡检助手...

可用工具:
### ssh_exec (服务器操作):
- 描述: 在远程服务器执行巡检命令
- 参数:
  - host: 服务器IP或主机名
  - command: 要执行的命令

如果需要使用工具，请按以下格式调用：
```tool_call
{"tool": "ssh_exec", "arguments": {"host": "localhost", "command": "top"}}
```

当前用户问题: 帮我检查 localhost 的 CPU 状态
```

#### 查看工具调用详情
添加日志：
```go
logger.Info(fmt.Sprintf("工具调用详情: %+v", toolCall))
logger.Info(fmt.Sprintf("工具配置: %+v", configOverride))
```

### 8. 常见问题排查

#### 工具未执行
检查：
1. Agent 是否挂载了工具（agent_tools 表）
2. 工具是否启用（enabled 字段）
3. AI 是否生成了 ```tool_call 格式（查看响应内容）
4. 工具名称是否匹配（"ssh_exec" vs "ssh-exec"）

#### 工具执行失败
检查：
1. 命令是否在白名单中
2. 主机是否在允许列表中
3. 超时时间是否足够
4. 查看错误日志详情

#### AI 不调用工具
原因：
1. Prompt 中没有包含工具列表
2. 系统提示词没有引导 AI 使用工具
3. 问题不够明确，AI 不认为需要工具
4. LLM 模型能力限制

解决：
- 优化系统提示词，明确告知 AI 如何使用工具
- 在问题中明确要求使用工具
- 提供示例对话

### 9. 验证完整流程

创建测试脚本：
```bash
# test_tool_call.sh
#!/bin/bash

echo "=== 1. 初始化工具 ==="
go run scripts/init_preset_tools.go

echo "=== 2. 检查工具列表 ==="
curl -s http://localhost:8080/api/v1/tools | jq '.data.tools[] | {id, name, enabled}'

echo "=== 3. 创建测试会话 ==="
SESSION_ID=$(curl -s -X POST http://localhost:8080/api/v1/chat/sessions \
  -H "Content-Type: application/json" \
  -d '{"title": "工具测试", "model": "qwen3.7-max"}' | jq -r '.data.id')

echo "Session ID: $SESSION_ID"

echo "=== 4. 发送测试消息 ==="
curl -X POST "http://localhost:8080/api/v1/chat/sessions/$SESSION_ID/messages" \
  -H "Content-Type: application/json" \
  -d '{"content": "检查 localhost 的 CPU 状态"}'

echo "=== 5. 查看日志 ==="
tail -f backend.log | grep "工具"
```

### 10. 性能监控

监控指标：
- 工具调用频率（每个 Agent）
- 工具执行时长（平均、最大）
- 工具成功率/失败率
- 工具超时次数

添加监控代码：
```go
// 记录工具调用开始时间
startTime := time.Now()

// 执行工具
result, err := executor.Execute(ctx, tool, configOverride, args)

// 记录执行时长
duration := time.Since(startTime)
logger.Info(fmt.Sprintf("工具 %s 执行耗时: %.2f秒", tool.Name, duration.Seconds()))
```

## 成功标志

✅ Backend 启动无错误  
✅ 工具列表显示 4 个预设工具  
✅ Agent 成功挂载 ssh_exec 工具  
✅ 对话中看到 "检测到工具调用" 日志  
✅ 工具执行成功，返回模拟结果  
✅ AI 响应包含工具执行结果  

## 下一步优化

1. 实现真实 SSH Client
2. 添加工具调用审计日志
3. 实现工具结果缓存
4. 添加工具调用统计报表
5. 支持工具参数验证和提示
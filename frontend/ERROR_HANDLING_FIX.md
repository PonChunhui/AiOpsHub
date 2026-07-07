# Agent执行错误显示修复

## 问题诊断

### 后端日志分析
```
2026-07-06T10:40:13.717+0800  ERROR  Agent执行错误: [NodeRunError] 
failed to stream tool call call_20114bbdbfe44c6fb6e2a63d: 
主机 '192.168.100.186' 不在白名单中
```

### 问题根源
1. **后端发送error事件**：`NewErrorEvent(agentName, errorMessage, 500)`
2. **前端未正确处理**：只将error加入events数组，未更新content显示错误消息
3. **用户看不到错误**：消息停留在loading状态或显示不完整内容

## 修复方案

### 1. 修复 appendAgentEventToMessage (TinyRobot版本)

```typescript
// frontend/src/adapters/agentEventToTinyRobot.ts
case 'error':
  const errorMsg = event.data?.message || event.data?.error || '未知错误'
  if (typeof updatedMessage.content === 'string') {
    updatedMessage.content = updatedMessage.content + '\n\n**错误**: ' + errorMsg
  }
  updatedMessage.state.agentVisualization.events.push(event)
  updatedMessage.loading = false  // 关键：停止loading显示错误
  break
```

### 2. 修复 AIAssistant.vue

```typescript
// frontend/src/views/AIAssistant.vue
case 'error':
  const errorMsg = eventData.data?.message || eventData.data?.error || '未知错误'
  messages.value = messages.value.map((msg, idx) => {
    if (idx === msgIndex) {
      return {
        ...msg,
        content: msg.content + '\n\n**错误**: ' + errorMsg,
        events: [...events, { type: 'error', ... }]
      }
    }
    return msg
  })
  break
```

### 3. 后端error事件结构

```go
// backend/internal/model/agent_event.go
ErrorEventData{
  Message: "主机 '192.168.100.186' 不在白名单中",
  Code:    500
}
```

## 修复效果

### 错误场景测试
用户输入："查看192.168.100.186的内存使用情况"

#### 修复前 ❌
```
AI消息显示：
"我来帮你查看 192.168.100.186 的内存使用情况。"  (不完整)
或一直显示loading状态
```

#### 修复后 ✅
```
AI消息显示：
"我来帮你查看 192.168.100.186 的内存使用情况。

**错误**: 主机 '192.168.100.186' 不在白名单中"
```

## 调试日志

前端console输出：
```javascript
[handleAgentEvent] eventType: error, eventData: {
  agent_name: "default",
  data: { message: "主机不在白名单中", code: 500 },
  timestamp: 1783305317,
  type: "error"
}
[handleAgentEvent] Updated message content: 
"我来帮你查看...\n\n**错误**: 主机不在白名单中"
```

## 其他常见错误类型

### 1. 工具调用错误
```
event.data.message: "主机不在白名单中"
event.data.message: "SSH连接失败"
event.data.message: "命令执行超时"
```

### 2. Agent切换错误
```
event.data.message: "Agent路由失败"
```

### 3. 权限错误
```
event.data.message: "权限不足"
event.data.message: "未授权访问"
```

## 前端错误处理流程

```
Agent执行 → 工具调用 → 发生错误
              ↓
        发送error事件
              ↓
前端handleAgentEvent接收到error事件
              ↓
appendAgentEventToMessage处理：
  1. 提取错误消息：event.data.message
  2. 添加到content：msg.content + "\n\n**错误**: " + errorMsg
  3. 加入events数组：记录错误详情
  4. 设置loading=false：停止loading显示错误
              ↓
强制更新消息数组：
  tinyRobotMessages.value[lastIndex] = updatedMsg
  tinyRobotMessages.value = [...tinyRobotMessages.value]
              ↓
TrBubbleList重新渲染，显示完整错误消息
```

## 测试验证

启动后端和前端：
```bash
# 后端
cd backend && go run cmd/main.go

# 前端
cd frontend && npm run dev
```

测试步骤：
1. 打开AI助手页面
2. 输入会触发错误的查询，如："查看不在白名单的主机"
3. 观察AI消息是否显示错误信息
4. 检查浏览器console是否有调试日志

预期结果：
- ✅ 显示完整的AI回复和错误信息
- ✅ loading状态正确结束
- ✅ 错误以markdown格式显示（**错误**: ...）
- ✅ console输出详细的调试信息

## 后续优化建议

### 1. 错误分类显示
```typescript
// 根据error code显示不同样式
if (event.data?.code === 401) {
  updatedMessage.content += '\n\n⚠️ **权限错误**: ' + errorMsg
} else if (event.data?.code === 500) {
  updatedMessage.content += '\n\n🔴 **系统错误**: ' + errorMsg
} else {
  updatedMessage.content += '\n\n❌ **错误**: ' + errorMsg
}
```

### 2. 错误恢复建议
```typescript
// 根据错误类型提供解决建议
if (errorMsg.includes('不在白名单中')) {
  updatedMessage.content += '\n\n💡 建议：请检查主机配置或联系管理员添加白名单'
}
```

### 3. 错误事件可视化
使用ErrorCard组件显示错误详情：
```vue
<ErrorCard 
  :error="{
    message: event.data.message,
    code: event.data.code,
    agent: event.agent_name,
    timestamp: event.timestamp
  }"
/>
```

---

修复完成！Agent执行错误现在会正确显示在AI消息中，用户能够看到完整的错误信息和建议。
# 实时消息显示和历史上下文修复

## 问题分析

### 问题1：历史消息被清空
**根本原因**：在发送新消息时，第218行直接替换了整个`tinyRobotMessages.value`数组，导致历史消息丢失。

```javascript
// 错误做法
tinyRobotMessages.value = [{
  role: 'user',
  content: userContent,
  id: `user-${Date.now()}`
}]
```

### 问题2：没有打字效果实时显示
**根本原因**：
1. Vue响应式更新机制不够及时，缺少强制触发更新的机制
2. 在finally块中调用`selectSession()`重新加载历史，会覆盖实时构建的消息
3. 数据库保存是异步的，立即加载可能获取不到最新消息

## 修复方案

### 修复1：保留历史消息，追加新消息
```javascript
// 正确做法：在现有消息基础上追加
const userMsg: TinyRobotBubbleMessage = {
  role: 'user',
  content: userContent,
  id: `user-${Date.now()}`
}

tinyRobotMessages.value = [...tinyRobotMessages.value, userMsg]

const initialAssistantMsg: TinyRobotBubbleMessage = createInitialAssistantMessage()
initialAssistantMsg.id = `assistant-${Date.now()}`

tinyRobotMessages.value = [...tinyRobotMessages.value, initialAssistantMsg]
messageListKey.value++
```

### 修复2：强制触发Vue响应式更新
在`handleAgentEvent`中添加：
```javascript
tinyRobotMessages.value[lastIndex] = updatedMsg
tinyRobotMessages.value = [...tinyRobotMessages.value]
messageListKey.value++  // 强制重新渲染
await nextTick()         // 确保DOM更新完成
```

### 修复3：移除finally块中的历史重新加载
```javascript
} finally {
  isLoading.value = false
  currentAbortController.value = null
  await loadSessions()  // 只更新会话列表，不重新加载当前会话历史
  // 删除了 selectSession(currentSessionId.value) 调用
}
```

**原因**：
- 实时构建的消息列表已经包含完整内容
- 后端在流结束后才保存AI消息到数据库（异步）
- 立即重新加载可能获取不到最新数据

### 修复4：更新消息ID时触发响应式
```javascript
if (eventType === 'user_message') {
  if (eventData.id) {
    const userMsgIndex = tinyRobotMessages.value.findIndex(m => m.role === 'user')
    if (userMsgIndex >= 0) {
      tinyRobotMessages.value[userMsgIndex] = {
        ...tinyRobotMessages.value[userMsgIndex],
        id: eventData.id
      }
      tinyRobotMessages.value = [...tinyRobotMessages.value]
      messageListKey.value++  // 新增：强制重新渲染
    }
  }
  return
}
```

## 完整消息流程

### 发送消息时
1. 创建用户消息对象
2. **追加**到现有消息列表（保留历史）
3. 创建AI助手消息（loading=true）
4. **追加**到消息列表
5. 触发强制重新渲染

### 接收SSE事件时
1. `user_message`事件：更新用户消息的真实ID（来自数据库）
2. `rag_references`事件：添加RAG引用信息
3. `content_chunk`事件：实时追加内容，显示打字效果
4. `tool_call`事件：实时显示工具调用
5. `done`事件：标记完成，设置loading=false
6. `ai_message`事件：更新AI消息的真实ID（来自数据库）

### 流结束后
1. 只更新会话列表（标题可能变化）
2. **不重新加载**当前会话历史
3. 保留实时构建的完整消息列表

## 响应式更新机制

### 双重保障
1. **数组引用更新**：`tinyRobotMessages.value = [...tinyRobotMessages.value]`
   - 创建新数组引用，触发Vue响应式检测
   
2. **组件key更新**：`messageListKey.value++`
   - 强制TrBubbleList组件重新渲染
   - 确保UI立即反映最新状态

### nextTick的作用
```javascript
await nextTick()
```
- 确保DOM更新完成后再处理下一个事件
- 避免事件处理过快导致UI更新延迟

## 效果验证

修复后应该看到：
1. ✅ 发送消息后，历史消息仍然保留
2. ✅ AI回复实时显示打字效果
3. ✅ 工具调用实时显示
4. ✅ 不需要刷新页面就能看到完整对话
5. ✅ 消息ID正确更新为数据库真实ID

## 注意事项

1. **不要在流式过程中重新加载历史**
   - 会覆盖实时内容
   - 数据库可能还未保存完成

2. **每次更新都要触发响应式**
   - 使用`[...array]`创建新引用
   - 更新`messageListKey`强制重渲染

3. **正确处理消息ID**
   - 初始使用临时ID
   - 收到服务器事件后更新为真实ID
   - 更新ID时也要触发响应式
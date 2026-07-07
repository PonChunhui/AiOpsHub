# TinyRobot框架正确使用方式 - useMessage重构

## 问题根源

之前的问题在于：
1. **未使用tiny-robot框架提供的专门工具**：手动管理消息状态，而不是使用`useMessage`
2. **Vue响应式更新机制不够可靠**：TrBubbleList内部有自己的缓存机制，手动触发更新不可靠
3. **未使用框架的流式处理工具**：没有使用`sseStreamToGenerator`处理SSE流

## 正确方案：使用tiny-robot-kit

根据官方文档（https://docs.opentiny.design/tiny-robot/tools/message），正确的做法是：

### 核心API

#### useMessage
```typescript
import { useMessage, sseStreamToGenerator } from '@opentiny/tiny-robot-kit'

const {
  messages,           // 消息列表（自动管理）
  requestState,       // 请求状态
  processingState,    // 处理状态
  isProcessing,       // 是否正在处理
  sendMessage,        // 发送消息
  abortRequest        // 中止请求
} = useMessage({
  responseProvider: async (requestBody, abortSignal) => {
    // 发起SSE请求
    const response = await fetch('/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(requestBody),
      signal: abortSignal
    })

    // 使用sseStreamToGenerator转换为异步生成器
    return sseStreamToGenerator(response, {
      transform: (chunk) => {
        // 将后端格式转换为OpenAI格式
        return {
          choices: [{
            delta: { content: chunk.content },
            finish_reason: chunk.done ? 'stop' : null
          }]
        }
      }
    })
  }
})
```

### 关键特性

#### 1. 自动流式处理
`useMessage`会自动处理流式响应，逐块消费并增量合并到消息内容中，实现**真正的实时显示**。

#### 2. 自动状态管理
- `requestState`: 'idle' | 'processing' | 'completed' | 'aborted' | 'error'
- `processingState`: 'requesting' | 'completing'
- `isProcessing`: boolean（用于驱动UI）

#### 3. 内置插件系统
- **thinkingPlugin**: 自动处理`reasoning_content`，显示思考过程（已默认激活）
- **lengthPlugin**: 处理`finish_reason === 'length'`自动续写（已默认激活）
- **toolPlugin**: 处理工具调用（需显式添加）

#### 4. 消息格式转换
后端返回的SSE流通过`sseStreamToGenerator`转换为异步生成器：
- 每个chunk通过`transform`函数转换为OpenAI格式
- useMessage自动消费generator，增量更新消息内容

## 实现细节

### SSE格式转换

后端返回的`agent_event`格式：
```json
{
  "event": "agent_event",
  "data": {
    "type": "content_chunk",
    "data": { "content": "Hello" },
    "agent_name": "coordinator",
    "timestamp": 1234567890
  }
}
```

转换为OpenAI格式：
```typescript
sseStreamToGenerator(response, {
  transform: (chunk: any) => {
    if (chunk.type === 'content_chunk') {
      return {
        choices: [{
          delta: { content: chunk.data?.content || '' },
          finish_reason: null
        }]
      }
    }
    
    if (chunk.type === 'thinking') {
      return {
        choices: [{
          delta: { reasoning_content: chunk.data?.content || '' },
          finish_reason: null
        }]
      }
    }
    
    if (chunk.type === 'done') {
      return {
        choices: [{
          delta: {},
          finish_reason: 'stop'
        }]
      }
    }
    
    return { choices: [{ delta: {}, finish_reason: null }] }
  }
})
```

### 消息渲染

TrBubbleList直接使用`useMessage`提供的`messages` ref：
```vue
<tr-bubble-list
  :messages="messages"
  :role-configs="roleConfigs"
  :auto-scroll="true"
/>
```

**关键点**：
- `messages`由useMessage自动管理
- 每次chunk到达时，useMessage自动更新对应的message内容
- Vue响应式系统自动检测到messages变化，触发TrBubbleList重新渲染
- **不需要手动触发更新**，框架会自动处理

### 思考过程显示

thinkingPlugin会自动处理`reasoning_content`：
- 收到thinking事件时，自动更新消息的`state.thinking`
- 自动展开思考过程区域
- 流结束后自动收起

**无需手动处理思考过程的显示逻辑**

## 对比：之前的问题

### 之前的错误做法
```typescript
// ❌ 错误：手动管理消息状态
const tinyRobotMessages = ref([])

// ❌ 错误：手动处理SSE流
const reader = response.body.getReader()
while (true) {
  const { done, value } = await reader.read()
  // 手动解析、手动更新
}

// ❌ 错误：手动触发响应式更新
tinyRobotMessages.value[lastIndex] = updatedMsg
tinyRobotMessages.value = [...tinyRobotMessages.value]
messageListKey.value++
```

### 现在的正确做法
```typescript
// ✅ 正确：使用框架提供的工具
const { messages, sendMessage } = useMessage({
  responseProvider: async (requestBody, abortSignal) => {
    const response = await fetch(...)
    return sseStreamToGenerator(response, { transform: ... })
  }
})

// ✅ 正确：框架自动管理所有状态
// messages自动实时更新，无需手动触发
```

## 技术原理

### sseStreamToGenerator的作用
将SSE流转换为AsyncGenerator：
1. 自动解析SSE事件格式（event: + data:）
2. 调用transform函数转换为统一格式
3. 返回AsyncGenerator，供useMessage逐块消费

### useMessage的内部机制
1. **消费generator**: for await (const chunk of generator)
2. **增量更新**: 每个chunk到达时立即更新message.content
3. **响应式触发**: 通过Vue的ref机制自动触发UI更新
4. **插件处理**: 各插件自动处理特定字段（thinking、tool等）

### 为什么能实时显示
关键在于**增量更新机制**：
- 每个content_chunk到达时，立即追加到message.content
- useMessage内部使用Vue ref，每次追加都触发响应式
- TrBubbleList检测到messages变化，立即重新渲染
- **用户看到逐字符显示的打字效果**

## 优势总结

1. **开箱即用**: 使用框架提供的专门工具，无需自己实现复杂逻辑
2. **可靠实时**: 框架内部处理响应式更新，保证实时显示
3. **状态管理**: 自动管理请求状态、处理状态，无需手动维护
4. **插件扩展**: 内置thinking、tool等插件，自动处理复杂场景
5. **类型安全**: 完整TypeScript支持，避免类型错误
6. **官方推荐**: 文档明确推荐使用useMessage处理流式响应

## 效果验证

使用useMessage后应该看到：
- ✅ 发送消息后立即显示用户消息
- ✅ AI回复实时逐字符显示（打字效果）
- ✅ 思考过程自动显示和收起
- ✅ 工具调用实时显示（使用toolPlugin）
- ✅ 请求状态正确驱动UI（loading、disabled等）
- ✅ 取消请求正常工作
- ✅ 错误处理正常（使用onError钩子）
- ✅ 历史对话正确加载和显示

## 注意事项

1. **responseProvider必须返回generator**: 对于SSE流式响应，必须使用sseStreamToGenerator
2. **transform函数要正确转换格式**: 将后端格式转换为OpenAI格式
3. **直接使用messages ref**: 不要手动修改，由useMessage自动管理
4. **使用sendMessage发送**: 不要手动添加消息到messages数组
5. **使用abortRequest取消**: 不要手动调用abortController.abort()

## 完整示例

参见：`frontend/src/views/AIAssistant-TinyRobot.vue`

关键代码：
- 使用useMessage创建消息管理实例
- responseProvider中使用sseStreamToGenerator
- TrBubbleList直接绑定messages
- 使用sendMessage和abortRequest
- 历史加载时直接设置messages.value
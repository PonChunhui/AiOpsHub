# AI助手思考过程显示功能

## 功能概述

本次实现为AI助手添加了思考过程显示功能，用户可以通过"深度思考"按钮控制是否显示AI的推理/思考过程。

## 实现架构

### 1. 后端实现

#### 1.1 LLM层 - 提取ReasoningContent
- **文件**: `backend/pkg/llm/eino_llm.go`
- **新增方法**: `StreamGenerateWithReasoning()`
- **功能**: 同时提取LLM输出的Content和ReasoningContent两个流

```go
type StreamResult struct {
    Content          string
    ReasoningContent string
}
```

#### 1.2 Service层 - 发送Thinking事件
- **文件**: `backend/internal/service/chat_service.go`
- **新增方法**: `streamWithThinkingEvents()`
- **功能**: 将thinking内容转换为AgentEvent并发送给前端

关键流程：
1. 使用LLM的StreamGenerateWithReasoning方法
2. 接收StreamResult（包含content和reasoning_content）
3. 发送thinking事件（EventThinking）
4. 发送content事件（EventContentChunk）

#### 1.3 Handler层 - 支持enable_thinking参数
- **文件**: `backend/internal/handler/chat_handler.go`
- **修改**: SendMessageRequest结构体添加`enable_thinking`字段
- **功能**: 接收前端传递的深度思考开关参数

```go
type SendMessageRequest struct {
    SessionID      string `json:"session_id" binding:"required"`
    Content        string `json:"content" binding:"required"`
    EnableThinking bool   `json:"enable_thinking"` // 新增
}
```

### 2. 前端实现

#### 2.1 视图层 - 深度思考按钮
- **文件**: `frontend/src/views/AIAssistant-TinyRobot.vue`
- **功能**: 
  - 启用已有的`deepThinkingEnabled`按钮（Line 107-113）
  - 发送消息时传递`enable_thinking`参数

#### 2.2 ThinkingBlock组件 - 折叠显示
- **文件**: `frontend/src/components/genui/ThinkingBlock.vue`
- **优化**: 
  - 默认折叠状态，点击展开/收起
  - 显示思考内容字数统计
  - 支持最大高度400px，超出自动滚动
  - 优化样式：淡蓝色背景，清晰的视觉区分

## 使用方法

### 前端使用

1. 打开AI助手对话界面
2. 在输入框底部找到"深度思考"按钮（CPU图标）
3. 点击按钮启用/关闭深度思考模式
4. 发送消息时，如果启用了深度思考：
   - AI的思考过程会以折叠卡片形式显示
   - 点击卡片可展开查看完整思考内容
   - 思考内容在最终回答之前显示

### API调用示例

```bash
# 启用思考过程的请求
POST /api/v1/chat/messages/stream/events
Content-Type: application/json

{
  "session_id": "session-123",
  "content": "帮我分析系统性能问题",
  "enable_thinking": true
}

# SSE响应示例
data: {"id":"msg-1","choices":[{"delta":{"reasoning_content":"让我思考一下..."}}]}
data: {"id":"msg-2","choices":[{"delta":{"content":"根据..."}}]}
data: [DONE]
```

## 技术细节

### Eino框架支持
Eino框架的Message结构原生支持`ReasoningContent`字段：
```go
type Message struct {
    Content          string
    ReasoningContent string `json:"reasoning_content,omitempty"`
    // ...
}
```

### 数据流
1. LLM → StreamGenerateWithReasoning → StreamResult
2. Service → streamWithThinkingEvents → AgentEvent
3. Handler → ConvertAgentEventToOpenAIChunk → OpenAI格式
4. 前端 → SSE解析 → TinyRobot渲染 → ThinkingBlock显示

### 事件类型
- `EventThinking`: 思考内容事件
- `EventContentChunk`: 回答内容事件
- `EventDone`: 完成事件

## 配置要求

### LLM配置
- 当前配置：阿里云百炼 glm-5.2
- 支持ReasoningContent输出的模型：
  - DeepSeek-R1
  - GLM-5.2（如果支持）
  - 其他支持thinking输出的模型

### 配置文件
`backend/configs/config.yaml`:
```yaml
llm:
  provider: "aliyun_bailian"
  model: "glm-5.2"
  api_key: "your-api-key"
  base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
```

## 测试建议

1. **功能测试**
   - 测试开启/关闭深度思考按钮
   - 验证thinking内容是否正确显示
   - 检查折叠/展开交互是否流畅

2. **兼容性测试**
   - 测试不支持thinking的模型（应忽略该参数）
   - 测试thinking内容很长的情况
   - 测试thinking和content交叉输出的情况

3. **性能测试**
   - 监控thinking内容流式传输的性能
   - 检查大量thinking内容对前端渲染的影响

## 已修改文件清单

### 后端文件
1. `backend/pkg/llm/eino_llm.go` - 新增StreamGenerateWithReasoning方法
2. `backend/internal/service/chat_service.go` - 新增streamWithThinkingEvents方法
3. `backend/internal/handler/chat_handler.go` - 修改请求参数支持enable_thinking

### 前端文件
1. `frontend/src/views/AIAssistant-TinyRobot.vue` - 添加enable_thinking参数传递
2. `frontend/src/components/genui/ThinkingBlock.vue` - 优化折叠显示样式

## 注意事项

1. **LLM兼容性**
   - 如果LLM不支持ReasoningContent，thinking内容将为空
   - 不会影响正常对话功能

2. **性能考虑**
   - Thinking内容可能很长，建议限制最大长度
   - 前端已设置400px最大高度

3. **用户体验**
   - 默认折叠状态避免信息过载
   - 字数统计让用户了解思考内容的规模

## 下一步优化建议

1. **配置优化**
   - 添加配置项控制thinking内容的最大长度
   - 支持thinking内容的格式化显示（Markdown）

2. **交互优化**
   - 添加thinking内容的复制功能
   - 支持thinking内容的搜索功能

3. **数据持久化**
   - 考虑将thinking内容保存到数据库
   - 支持历史对话中查看thinking过程

4. **多语言支持**
   - ThinkingBlock组件支持国际化
   - 添加英文版本的提示文本
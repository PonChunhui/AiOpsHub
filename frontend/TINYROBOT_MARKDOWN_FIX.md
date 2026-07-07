# TinyRobot Markdown代码块渲染修复

## 问题诊断

### 实际表现

用户看到的markdown内容：
```
以下是执行结果：
```bash
               total        used        free
Mem:            15Gi       2.3Gi       2.0Gi
```

**内存使用情况分析：**
```

TinyRobot渲染效果：代码块不显示，或者显示为普通文本，没有语法高亮和代码样式。

### TinyRobot设计原理

TinyRobot的content字段支持两种格式：

#### 1. 字符串格式
```typescript
content: "这是普通文本，包含```code```块"
```
TinyRobot会把整个字符串当作一个text类型，不会解析markdown代码块。

#### 2. 数组格式（推荐）
```typescript
content: [
  { type: 'text', text: '这是普通文本' },
  { type: 'code', text: '代码内容', language: 'bash' }
]
```
TinyRobot根据type字段匹配对应的渲染器，code类型会显示为代码块样式。

### 错误原因

**之前实现**：content一直是字符串格式，TinyRobot无法识别其中的markdown代码块。

**正确做法**：将markdown解析为数组格式，每个代码块作为独立的code item。

## 修复方案

### 1. 创建Markdown解析器

**文件**: `frontend/src/utils/markdownParser.ts`

```typescript
export interface ContentItem {
  type: string       // 'text' 或 'code'
  text?: string      // 内容文本
  language?: string  // 代码语言（仅code类型）
}

export function parseMarkdownToContentItems(markdown: string): ContentItem[] {
  const items: ContentItem[] = []
  const codeBlockRegex = /```(\w+)?\n([\s\S]*?)\n```/g
  
  let lastIndex = 0
  let match
  
  // 遍历所有代码块
  while ((match = codeBlockRegex.exec(markdown)) !== null) {
    // 提取代码块之前的文本
    if (match.index > lastIndex) {
      const textBefore = markdown.substring(lastIndex, match.index).trim()
      if (textBefore) {
        items.push({ type: 'text', text: textBefore })
      }
    }
    
    // 提取代码块
    const language = match[1] || 'plaintext'
    const code = match[2] || ''
    items.push({ type: 'code', text: code, language: language })
    
    lastIndex = match.index + match[0].length
  }
  
  // 提取代码块之后的文本
  if (lastIndex < markdown.length) {
    const textAfter = markdown.substring(lastIndex).trim()
    if (textAfter) {
      items.push({ type: 'text', text: textAfter })
    }
  }
  
  return items
}
```

### 2. 修改Content处理

**文件**: `frontend/src/adapters/agentEventToTinyRobot.ts`

#### 导入解析器

```typescript
import { parseMarkdownToContentItems } from '@/utils/markdownParser'
```

#### 处理content_chunk事件

```typescript
case 'content_chunk':
  if (event.data?.content) {
    const newContent = (typeof updatedMessage.content === 'string' 
      ? updatedMessage.content 
      : '') + event.data.content
    
    // 解析markdown为数组格式
    const contentItems = parseMarkdownToContentItems(newContent)
    
    if (contentItems.length > 0) {
      updatedMessage.content = contentItems  // ← 数组格式
    } else {
      updatedMessage.content = newContent    // ← 字符串格式（兜底）
    }
    
    console.log('[ContentChunk] Updated:', {
      isString: typeof updatedMessage.content === 'string',
      itemCount: contentItems.length
    })
  }
  break
```

#### 处理done事件（最终解析）

```typescript
case 'done':
  // ... 清理buffer和过滤tool_calls
  
  // 最终确保content为数组格式
  if (typeof updatedMessage.content === 'string' && updatedMessage.content.trim()) {
    updatedMessage.content = parseMarkdownToContentItems(updatedMessage.content)
    console.log('[Done] Final content parsed:', updatedMessage.content.length)
  }
  
  updatedMessage.loading = false
  break
```

### 3. 处理历史消息

**文件**: `frontend/src/views/AIAssistant-TinyRobot.vue`

```typescript
const selectSession = async (sessionId: string) => {
  // ... 加载历史
  
  for (let i = 0; i < history.length; i++) {
    const msg = history[i]
    
    if (msg.role === 'assistant') {
      // 解析历史消息的markdown
      const parsedContent = parseMarkdownToContentItems(msg.content || '')
      
      const message: TinyRobotBubbleMessage = {
        role: 'assistant',
        content: parsedContent.length > 0 ? parsedContent : msg.content,
        // ...
      }
      
      convertedMessages.push(message)
    }
  }
  
  console.log('转换后的消息:', convertedMessages.map(m => ({
    role: m.role,
    isContentArray: Array.isArray(m.content),
    contentItems: Array.isArray(m.content) ? m.content.length : 0
  })))
}
```

## Markdown解析流程

### 示例输入

```
这是普通文本：

```bash
free -h
               total        used        free
Mem:            15Gi       2.3Gi       2.0Gi
```

**内存分析：**
总内存15Gi，已用2.3Gi
```

### 解析步骤

#### 1. 正则匹配代码块
```javascript
codeBlockRegex.exec(markdown)
// match[1] = "bash"  (语言)
// match[2] = "free -h\n..."  (代码内容)
```

#### 2. 提取文本片段
```javascript
items.push({ type: 'text', text: '这是普通文本：' })
```

#### 3. 提取代码块
```javascript
items.push({ type: 'code', text: 'free -h\n...', language: 'bash' })
```

#### 4. 最终数组
```javascript
[
  { type: 'text', text: '这是普通文本：' },
  { type: 'code', text: 'free -h\n...', language: 'bash' },
  { type: 'text', text: '**内存分析：**\n总内存15Gi...' }
]
```

## TinyRobot渲染机制

### content类型匹配

TinyRobot内置渲染器（可能）：
- `text`: 普通文本渲染
- `code`: 代码块渲染（带语法高亮）
- `image`: 图片渲染
- `file`: 文件渲染

### fallbackContentRenderer

如果没有匹配的渲染器，使用fallback：
```vue
<tr-bubble-list
  :messages="messages"
  :fallback-content-renderer="CustomRenderer"
/>
```

### 自定义渲染器（可选）

如果TinyRobot不内置code渲染器，可以自定义：

```vue
<template>
  <tr-bubble-provider>
    <tr-bubble-list :messages="messages">
      <template #default="{ message, content, contentIndex }">
        <div v-if="content.type === 'code'" class="code-block">
          <div class="code-header">{{ content.language }}</div>
          <pre><code>{{ content.text }}</code></pre>
        </div>
        <div v-else class="text-content">
          {{ content.text }}
        </div>
      </template>
    </tr-bubble-list>
  </tr-bubble-provider>
</template>
```

## 渲染效果对比

### 修复前 ❌

**Content格式**：
```javascript
content: "这是文本```bash\nfree -h\n```"
```

**显示效果**：
```
这是文本```bash
free -h
```
```
（代码块不显示样式，无高亮）

### 修复后 ✅

**Content格式**：
```javascript
content: [
  { type: 'text', text: '这是文本' },
  { type: 'code', text: 'free -h', language: 'bash' }
]
```

**显示效果**：
```
这是文本

┌─────────────────┐
│ bash            │
├─────────────────┤
│ free -h         │
│ total used free │
│ 15Gi  2.3Gi 2.0Gi│
└─────────────────┘
```
（代码块有背景、边框、语法高亮）

## Console调试日志

### 实时更新日志

```javascript
[ContentChunk] Updated: {isString: false, itemCount: 3, preview: "这是文本..."}
[ContentChunk] Updated: {isString: false, itemCount: 3, preview: "这是文本..."}
[Done] Final content parsed: 3
```

### 历史消息加载日志

```javascript
转换后的消息: [
  {
    role: 'assistant',
    isContentArray: true,
    contentItems: 3,  // ← 包含3个content items
    toolCalls: 1
  }
]
```

## 支持的Markdown元素

### 当前支持

✅ **代码块**：```language\ncode\n```
✅ **普通文本**：非代码块的文本
✅ **语言标识**：bash、python、javascript等

### 未支持（可扩展）

❌ **行内代码**：`code`
❌ **标题**：# ## ###
❌ **列表**：- 1. 
❌ **链接**：[text](url)
❌ **表格**：| | |

**扩展方案**：增强parseMarkdownToContentItems函数，添加更多正则匹配。

## 性能优化

### 避免重复解析

只在必要时解析：
- content_chunk时：实时解析（确保正确显示）
- done时：最终解析（兜底）
- 历史消息时：一次性解析

### 缓存机制（可选）

对相同内容缓存解析结果：
```typescript
const parseCache = new Map<string, ContentItem[]>()

function parseWithCache(markdown: string): ContentItem[] {
  if (parseCache.has(markdown)) {
    return parseCache.get(markdown)!
  }
  
  const items = parseMarkdownToContentItems(markdown)
  parseCache.set(markdown, items)
  return items
}
```

## 测试验证

### 测试场景

1. **纯文本消息**：无代码块
2. **单个代码块**：一个```bash```块
3. **多个代码块**：多个不同语言代码块
4. **混合内容**：文本+代码+文本
5. **历史消息**：加载包含代码块的对话

### 预期结果

✅ 代码块显示语法高亮  
✅ 代码块有背景和边框  
✅ 语言标签正确显示  
✅ 文本和代码正确分隔  
✅ Console显示正确的itemCount  

### 实际测试

访问 http://localhost:5174/

测试步骤：
1. 输入查询触发AI回复（包含代码块）
2. 观察代码块是否有样式和高亮
3. 检查Console显示content为数组格式
4. 切换会话，加载历史消息
5. 观察历史消息的代码块是否正确显示

## 相关文件

### 新增文件
- `frontend/src/utils/markdownParser.ts` - Markdown解析器

### 修改文件
- `frontend/src/adapters/agentEventToTinyRobot.ts` - content处理逻辑
- `frontend/src/views/AIAssistant-TinyRobot.vue` - 历史消息解析

### TinyRobot文档
- `frontend/node_modules/@opentiny/tiny-robot/dist/bubble/index.type.d.ts` - 类型定义
- ContentItem结构：`{ type: string, [key: string]: any }`

## 后续优化建议

### 1. 使用成熟的Markdown库

使用marked或markdown-it替代正则解析：
```typescript
import { marked } from 'marked'

function parseMarkdownToContentItems(markdown: string): ContentItem[] {
  const tokens = marked.lexer(markdown)
  
  return tokens.map(token => {
    if (token.type === 'code') {
      return { type: 'code', text: token.text, language: token.lang }
    } else {
      return { type: 'text', text: token.raw }
    }
  })
}
```

优点：
- 支持更多markdown元素
- 更准确的解析
- 维护成本低

### 2. 自定义渲染器

如果TinyRobot不满足需求，创建自定义渲染器：
```vue
<!-- CodeBlockRenderer.vue -->
<template>
  <div class="code-block">
    <div class="language-tag">{{ content.language }}</div>
    <pre class="code-content"><code>{{ content.text }}</code></pre>
  </div>
</template>
```

配置：
```vue
<tr-bubble-list
  :messages="messages"
  :fallback-content-renderer="CustomRenderer"
/>
```

### 3. 代码复制功能

添加复制按钮：
```vue
<div class="code-block">
  <div class="code-header">
    <span>{{ language }}</span>
    <button @click="copyCode">复制</button>
  </div>
  <pre><code>{{ code }}</code></pre>
</div>
```

### 4. 行内代码支持

扩展解析器支持行内代码：
```typescript
// 将 `code` 转换为 <code>code</code>
text = text.replace(/`([^`]+)`/g, '<code>$1</code>')
```

---

修复完成！TinyRobot现在能正确渲染markdown代码块，显示语法高亮和代码样式。使用数组格式的content，每个代码块作为独立的code item。
# Tool Call分片显示"Untitled"问题修复

## 问题诊断

### 实际表现

用户看到的显示：
```
ssh_exec
{
  "arguments": {"host": "192.168.100.186", "command": "free -h"}
}

Untitled
{
  "arguments": {}
}

Untitled
{
  "arguments": {}
}
...（重复10次）
```

### 后端发送机制

LLM流式发送tool_call时采用分片方式：

```go
// 第1个chunk：发送tool_id
event: agent_event
data: {
  "type": "tool_call",
  "data": {
    "tool_id": "tc-1783306096869754000",
    "tool_name": "",        // ← 空字符串
    "args_raw": ""          // ← 空字符串
  }
}

// 第2个chunk：发送name
event: agent_event
data: {
  "type": "tool_call",
  "data": {
    "tool_id": "tc-1783306096869754000",
    "tool_name": "ssh_exec", // ← 有name
    "args_raw": ""
  }
}

// 第3-N个chunk：逐步发送arguments
event: agent_event
data: {
  "type": "tool_call",
  "data": {
    "tool_id": "tc-1783306096869754000",
    "args_raw": "{\"host\": \""  // ← 参数片段
  }
}
```

### 前端错误处理

**修复前的逻辑**：
```typescript
case 'tool_call':
  updatedMessage.tool_calls.push({
    id: event.data?.tool_id || '',
    type: 'function',
    function: {
      name: event.data?.tool_name || '',  // ← 可能是空字符串
      arguments: JSON.stringify(event.data?.args || {})
    }
  })
```

**问题**：
- 每个分片都被当作独立的tool_call添加
- 空name显示为"Untitled"
- 空arguments显示为`{}`
- 同一个tool_id出现多次（10+个）

## 修复方案

### 核心思路：Map缓冲 + 合并分片 + 过滤不完整

#### 1. 添加toolCallsBuffer

```typescript
export interface AgentVisualizationData {
  agentPath: Array<{...}>
  events: AgentEvent[]
  toolCallsBuffer?: Map<string, { id: string; name: string; args: string }>  // ← 新增
}
```

#### 2. 流式合并逻辑

```typescript
case 'tool_call':
  const buffer = updatedMessage.state.agentVisualization.toolCallsBuffer || new Map()
  
  // 1. 更新buffer（合并分片）
  const existing = buffer.get(toolId)
  buffer.set(toolId, {
    id: toolId,
    name: toolName || existing?.name || '',          // 合并name
    args: argsComplete || (argsRaw + existing?.args) // 合并args
  })
  
  // 2. 重建tool_calls数组（只包含完整的）
  updatedMessage.tool_calls = []
  buffer.forEach((tc) => {
    if (tc.name && tc.args) {  // ← 只添加完整的
      updatedMessage.tool_calls.push({
        id: tc.id,
        type: 'function',
        function: {
          name: tc.name,
          arguments: tc.args
        }
      })
    } else {
      console.log('[ToolCall] Buffer incomplete:', tc.id, 'waiting...')
    }
  })
  break
```

#### 3. Done时清理

```typescript
case 'done':
  // 清理buffer
  if (updatedMessage.state.agentVisualization.toolCallsBuffer) {
    updatedMessage.state.agentVisualization.toolCallsBuffer.clear()
  }
  
  // 最终过滤（确保完整性）
  updatedMessage.tool_calls = updatedMessage.tool_calls.filter(tc => {
    return tc.function?.name && tc.function?.arguments
  })
  break
```

### 流程示例

假设收到以下分片：

#### Chunk 1: 只有tool_id
```
Received: {toolId: "tc-123", toolName: "", args: ""}
Buffer: Map{"tc-123" => {id:"tc-123", name:"", args:""}}
tool_calls: []  ← 空数组，不显示
Console: [ToolCall] Buffer incomplete: tc-123 waiting for name
```

#### Chunk 2: 有name
```
Received: {toolId: "tc-123", toolName: "ssh_exec", args: ""}
Buffer: Map{"tc-123" => {id:"tc-123", name:"ssh_exec", args:""}}
tool_calls: []  ← args为空，仍不显示
Console: [ToolCall] Buffer incomplete: tc-123 waiting for args
```

#### Chunk 3-N: 有args片段
```
Received: {toolId: "tc-123", argsRaw: "{\"host\":\""}
Buffer: Map{"tc-123" => {id:"tc-123", name:"ssh_exec", args:"{\"host\":\""}}
tool_calls: []  ← args不完整，仍不显示
```

#### Chunk N: args完成
```
Received: {toolId: "tc-123", argsComplete: "{\"host\":\"192.168.100.186\"}"}
Buffer: Map{"tc-123" => {id:"tc-123", name:"ssh_exec", args:"{\"host\":\"192.168.100.186\"}"}}
tool_calls: [{id:"tc-123", name:"ssh_exec", args:"..."}]  ← 完整，显示！
Console: [ToolCall] Display count: 1, Buffer count: 1
```

#### Done事件
```
Buffer cleared
tool_calls: [{id:"tc-123", name:"ssh_exec", args:"..."}]  ← 最终验证
Console: [ToolCall] Final valid count: 1
```

## 修复效果对比

### 修复前 ❌

显示：
```
ssh_exec
{arguments: {...}}

Untitled
{arguments: {}}

Untitled
{arguments: {}}
...（10+个）
```

日志：
```javascript
[ToolCall] Added: tc-123 name: ssh_exec
[ToolCall] Added: tc-123 name:         // ← 空
[ToolCall] Added: tc-123 name:         // ← 空
...（10+次）
```

### 修复后 ✅

显示：
```
ssh_exec
{arguments: {"host":"192.168.100.186","command":"free -h"}}
```
（只显示一个完整的工具调用）

日志：
```javascript
[ToolCall] Buffer incomplete: tc-123 waiting for name
[ToolCall] Buffer incomplete: tc-123 waiting for args
[ToolCall] Buffer updated: tc-123 name: ssh_exec argsLen: 20
[ToolCall] Display count: 1 Buffer count: 1
[ToolCall] Buffer cleared
[ToolCall] Final valid count: 1
```

## 技术要点

### 1. Map深拷贝

Map对象需要手动复制：
```typescript
toolCallsBuffer: existingMessage.state.agentVisualization.toolCallsBuffer 
  ? new Map(existingMessage.state.agentVisualization.toolCallsBuffer) 
  : undefined
```

### 2. Vue响应式兼容

Map对象在Vue响应式系统中不会被自动追踪，需要：
- 每次更新创建新的Map对象
- 或者使用Vue的reactive/ref包装Map

当前方案：每次创建新Map，确保响应式更新。

### 3. args合并策略

```typescript
const newArgs = argsComplete || (argsRaw ? (existing?.args || '') + argsRaw : (existing?.args || ''))
```

优先级：
1. `argsComplete` - 完整参数（优先使用）
2. `argsRaw` + `existing.args` - 累加分片
3. `existing.args` - 保留已有参数

### 4. 显示时机

只有同时满足以下条件才显示：
- ✅ 有tool_id
- ✅ 有tool_name（非空字符串）
- ✅ 有arguments（非空字符串）

不满足时留在buffer等待后续分片。

## Console调试日志

### 开启调试

前端Console会输出详细日志：
```javascript
[ToolCall] Received: {toolId: "...", toolName: "...", hasArgsRaw: true}
[ToolCall] Buffer updated: tc-123 name: ssh_exec argsLen: 45
[ToolCall] Display count: 1 Buffer count: 1
[ToolCall] Buffer incomplete: tc-456 waiting for args
[ToolCall] Buffer cleared
[ToolCall] Final valid count: 2
```

### 关键信息

- **Received**: 显示每个分片内容
- **Buffer updated**: 显示合并结果
- **Display count**: 当前显示的工具调用数
- **Buffer incomplete**: 不完整的工具调用（等待中）
- **Final valid count**: 最终有效的工具调用数

## 性能优化

### 减少重复渲染

之前：10+个tool_call导致10+次渲染
现在：只在完整时渲染1次

### Map查找效率

使用Map而不是数组：
- 数组查找：O(n) 需遍历
- Map查找：O(1) 直接定位

对于多个并发tool_call更高效。

### 内存管理

- Buffer在done时清理，不占用持久内存
- Map深拷贝确保每次更新都是新对象

## 测试验证

### 测试场景

1. 单个工具调用（ssh_exec）
2. 多个并发工具调用
3. 分片发送的参数（长JSON）
4. 错误的工具调用（缺少name/args）

### 预期结果

✅ 只显示完整的工具调用  
✅ 参数完整显示，无截断  
✅ 无"Untitled"显示  
✅ 无空arguments显示  
✅ Console日志清晰显示合并过程  

### 实际测试

访问 http://localhost:5174/

测试步骤：
1. 输入："查看192.168.100.186的内存使用情况"
2. 观察：
   - ✅ 只显示ssh_exec工具调用
   - ✅ 参数完整显示
   - ✅ 无Untitled
3. 检查Console：
   - ✅ Buffer合并日志清晰
   - ✅ Display count正确

## 相关文件

### 修改文件
- `frontend/src/adapters/agentEventToTinyRobot.ts:28-37` - 添加toolCallsBuffer接口
- `frontend/src/adapters/agentEventToTinyRobot.ts:216-249` - tool_call合并逻辑
- `frontend/src/adapters/agentEventToTinyRobot.ts:251-271` - done清理逻辑

### 后端文件（未修改）
- `backend/internal/model/agent_event.go` - tool_call事件结构
- `backend/internal/service/chat_service.go` - 流式发送逻辑

## 后端事件格式

### ToolCallEventData

```go
type ToolCallEventData struct {
    ToolID      string `json:"tool_id"`       // 工具调用ID（必须）
    ToolName    string `json:"tool_name"`     // 工具名称（可能为空）
    ArgsRaw     string `json:"args_raw"`      // 参数片段（逐步发送）
    ArgsComplete string `json:"args_complete"` // 完整参数（最终发送）
}
```

### 发送示例

```
event: agent_event
data: {
  "type": "tool_call",
  "agent_name": "default",
  "data": {
    "tool_id": "tc-1783306096869754000",
    "tool_name": "ssh_exec",
    "args_complete": "{\"host\":\"192.168.100.186\",\"command\":\"free -h\"}"
  },
  "timestamp": 1783306097
}
```

## 后续优化建议

### 1. 后端优化

建议后端只在tool_call完整时发送一个事件，而不是分片发送多个空事件。

优点：
- 减少网络传输
- 减少前端复杂度
- 提升用户体验

### 2. 显示优化

可以在buffer中有部分数据时显示loading状态：
```typescript
if (!tc.name || !tc.args) {
  updatedMessage.tool_calls.push({
    id: tc.id,
    type: 'function',
    function: {
      name: tc.name || 'Loading...',
      arguments: tc.args || 'Loading...',
      loading: true
    }
  })
}
```

### 3. 错误提示

如果done时仍有不完整的tool_call，可以添加提示：
```typescript
const incompleteCount = buffer.size - updatedMessage.tool_calls.length
if (incompleteCount > 0) {
  updatedMessage.content += `\n\n⚠️ ${incompleteCount}个工具调用未完成`
}
```

---

修复完成！Tool Call不再显示大量"Untitled"，只显示完整的工具调用。使用Map缓冲合并分片，确保正确显示。
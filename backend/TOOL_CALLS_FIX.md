# Tool Calls SSE流式显示问题分析与修复方案

## 问题现象

前端tiny-robot-kit无法正确显示工具调用参数，表现为：
- 工具调用显示不完整或缺失
- 参数显示错误或混乱

## 问题根源分析

### 1. 数据流追踪

从后端LLM到前端的完整数据流：

```
LLM (Eino框架) → chat_service.go → AgentEvent → openai_format.go → OpenAI SSE → 前端
```

### 2. 问题定位

#### 2.1 chat_service.go (第713-749行)

**问题代码：**

```go
if len(chunk.ToolCalls) > 0 {
    for _, tc := range chunk.ToolCalls {
        // 合并流式分片的工具调用
        tcID := tc.ID
        if tcID == "" {
            // 如果没有ID，尝试从之前的工具调用中找到
            for existingID, builder := range toolCallsMap {
                if builder.name == tc.Function.Name || (builder.name == "" && tc.Function.Name != "") {
                    tcID = existingID
                    break
                }
            }
        }

        if tcID == "" {
            // 🔴 问题：创建新的临时ID！
            tcID = fmt.Sprintf("tc-%d", time.Now().UnixNano())
            toolCallsMap[tcID] = &toolCallBuilder{
                id:   tcID,
                name: tc.Function.Name,
            }
        }

        // 累加参数分片
        if builder, exists := toolCallsMap[tcID]; exists {
            builder.argsBuffer.WriteString(tc.Function.Arguments)
            if tc.Function.Name != "" && builder.name == "" {
                builder.name = tc.Function.Name
            }
        }

        // 发送事件（用于前端实时显示）
        toolCallEvent := model.NewToolCallEvent(agentName, tcID, tc.Function.Name, tc.Function.Arguments)
        toolCallEvent.RunPath = runPath
        eventChan <- toolCallEvent
    }
}
```

**问题分析：**

LLM返回的tool_calls流式chunks可能包含：
1. **第一个chunk**: 
   - ID: `call_d64f4737...`（LLM生成的完整ID）
   - Name: `ssh_exec`（完整）
   - Arguments: `{`（只有开头）

2. **后续chunks**: 
   - ID: `""`（空！LLM不重复发送ID）
   - Name: `""`（空！LLM不重复发送name）
   - Arguments: `"command": "free -h"`（参数分片）

当后续chunks的ID为空时：
- 第716-725行的匹配逻辑可能失败（如果第一个chunk的name还未记录）
- 第729行创建新的临时ID：`tc-1783323307712748000`
- 这导致同一个工具调用被分成两个不同的ID！

#### 2.2 openai_format.go (第88-124行)

**问题代码：**

```go
case EventToolCall:
    if toolCallData, ok := event.Data.(ToolCallEventData); ok {
        toolID := toolCallData.ToolID
        toolName := toolCallData.ToolName
        argsRaw := toolCallData.ArgsRaw

        if toolID != "" {
            // 合并工具调用分片
            if _, exists := buffer.toolCalls[toolID]; !exists {
                buffer.toolCalls[toolID] = &ToolCallBuilder{
                    id:    toolID,
                    name:  toolName,
                    // 🔴 问题：不同的toolID会分配不同的index！
                    index: len(buffer.toolCalls),  
                }
            }

            builder := buffer.toolCalls[toolID]
            if toolName != "" {
                builder.name = toolName
            }
            builder.argsBuffer += argsRaw

            // 发送增量更新
            delta.ToolCalls = []OpenAIToolCallDelta{
                {
                    Index: builder.index,  // 🔴 问题：发送不同的index给前端！
                    ID:    toolID,
                    Type:  "function",
                    Function: OpenAIFunctionDelta{
                        Name:      toolName,
                        Arguments: argsRaw,
                    },
                },
            }
        }
    }
```

**问题分析：**

当收到两个不同ID的tool_call事件时：
- `call_d64f4737...` → index: 0
- `tc-1783323307...` → index: 1

这导致前端收到两个独立的tool_call对象，而不是同一个工具调用的合并！

### 3. 实际SSE流数据（用户提供）

```
Chunk 1 (index 0, ID: call_d64f4737...):
{
  "tool_calls": [{
    "index": 0,
    "id": "call_d64f4737aba7436e86008b7d",
    "type": "function",
    "function": {
      "name": "ssh_exec",
      "arguments": "{"
    }
  }]
}

Chunk 2-10 (index 1, ID: tc-1783323...):
{
  "tool_calls": [{
    "index": 1,  // 🔴 错误！应该是index 0
    "id": "tc-1783323307712748000",  // 🔴 错误！应该是call_d64f4737...
    "type": "function",
    "function": {
      "arguments": "\"command\": \"free -h\""  // 参数分片
    }
  }]
}

Final chunk (两个独立的对象！):
{
  "tool_calls": [
    {
      "index": 0,
      "id": "call_d64f4737...",
      "function": {
        "name": "ssh_exec",
        "arguments": "{"  // 🔴 不完整！
      }
    },
    {
      "index": 1,
      "id": "tc-1783323...",
      "function": {
        "arguments": "完整参数"  // 🔴 缺少name！
      }
    }
  ],
  "finish_reason": "tool_calls"
}
```

## 修复方案

### 方案1：修复chat_service.go（推荐）

**核心思路：**确保同一个工具调用的所有分片使用相同的ID

**修改点1：**改进ID匹配逻辑（第716-725行）

```go
tcID := tc.ID
if tcID == "" {
    // 🔧 改进：优先匹配最近添加的toolID（同一工具调用的分片通常是连续的）
    // 如果没有name，则匹配最后一个未完成的builder
    if len(toolCallsMap) > 0 {
        // 找到最后一个添加的toolID
        var lastID string
        for id := range toolCallsMap {
            lastID = id  // 最后一个ID
        }
        
        if lastID != "" {
            builder := toolCallsMap[lastID]
            // 如果name匹配，或者当前chunk没有name且builder未完成
            if (tc.Function.Name != "" && builder.name == tc.Function.Name) ||
               (tc.Function.Name == "" && builder.name != "") {
                tcID = lastID
            }
        }
    }
}

// 🔧 改进：如果还是找不到，不要立即创建新ID，而是等待下一个带name的chunk
if tcID == "" {
    // 只有当这个chunk有name时才创建新ID
    if tc.Function.Name != "" {
        tcID = fmt.Sprintf("tc-%d", time.Now().UnixNano())
        toolCallsMap[tcID] = &toolCallBuilder{
            id:   tcID,
            name: tc.Function.Name,
        }
    } else {
        // 如果既没有ID也没有name，暂时跳过，等待后续chunk
        // 或者：创建一个临时的"pending"builder，等待下一个带name的chunk来补充
        logger.Warn("Received tool call chunk without ID and name, skipping")
        continue
    }
}
```

**修改点2：**使用更智能的匹配策略

```go
// 🔧 新增：使用tool_calls index匹配
// OpenAI标准：同一个工具调用的所有分片使用相同的index
// 我们可以利用这一点来匹配

// 在toolCallsMap中添加index字段
type toolCallBuilder struct {
    id         string
    name       string
    argsBuffer strings.Builder
    index      int  // 🔧 新增：记录OpenAI标准的index
}

// 在处理chunk时，使用index匹配
if tcID == "" && tc.Index >= 0 {
    // 尝试通过index匹配
    for existingID, builder := range toolCallsMap {
        if builder.index == tc.Index {
            tcID = existingID
            break
        }
    }
}
```

### 方案2：修复openai_format.go（临时方案）

**核心思路：**强制所有tool_call使用相同的index（0）

```go
case EventToolCall:
    if toolCallData, ok := event.Data.(ToolCallEventData); ok {
        toolID := toolCallData.ToolID
        toolName := toolCallData.ToolName
        argsRaw := toolCallData.ArgsRaw

        if toolID != "" {
            // 🔧 改进：优先通过name匹配，而不是直接使用toolID
            var matchedID string
            var matchedBuilder *ToolCallBuilder
            
            // 1. 先尝试通过toolID直接匹配
            if builder, exists := buffer.toolCalls[toolID]; exists {
                matchedID = toolID
                matchedBuilder = builder
            }
            
            // 2. 如果没有匹配到，尝试通过name匹配
            if matchedID == "" && toolName != "" {
                for existingID, builder := range buffer.toolCalls {
                    if builder.name == toolName {
                        matchedID = existingID
                        matchedBuilder = builder
                        break
                    }
                }
            }
            
            // 3. 如果还是没有匹配到，且toolName为空，尝试匹配最后一个未完成的builder
            if matchedID == "" && toolName == "" && len(buffer.toolCalls) > 0 {
                // 找到最后一个builder
                var lastID string
                for id, builder := range buffer.toolCalls {
                    if builder.argsBuffer != "" && builder.argsBuffer[len(builder.argsBuffer)-1:] != "}" {
                        // argsBuffer未闭合，可能是正在构建中
                        lastID = id
                        matchedBuilder = builder
                        break
                    }
                }
                if lastID != "" {
                    matchedID = lastID
                }
            }
            
            // 4. 如果最终还是没有匹配到，创建新的builder
            if matchedID == "" {
                matchedID = toolID
                buffer.toolCalls[matchedID] = &ToolCallBuilder{
                    id:    matchedID,
                    name:  toolName,
                    // 🔧 改进：所有tool_call使用相同的index（0）！
                    index: 0,  // 强制使用index 0
                }
                matchedBuilder = buffer.toolCalls[matchedID]
            }
            
            // 更新builder
            if toolName != "" && matchedBuilder.name == "" {
                matchedBuilder.name = toolName
            }
            matchedBuilder.argsBuffer += argsRaw
            
            // 发送增量更新（使用统一的index）
            delta.ToolCalls = []OpenAIToolCallDelta{
                {
                    Index: matchedBuilder.index,  // 🔧 改进：使用合并后的index
                    ID:    matchedID,
                    Type:  "function",
                    Function: OpenAIFunctionDelta{
                        Name:      matchedBuilder.name,  // 🔧 改进：使用合并后的name
                        Arguments: argsRaw,
                    },
                },
            }
        }
    }
```

### 方案3：前端自定义处理（最临时方案）

如果暂时无法修改后端，可以在前端添加自定义处理逻辑，在接收SSE流时合并不同index的tool_calls：

参见 `frontend/test-tool-calls.html` 中的示例代码。

## 推荐修复顺序

1. **优先修复chat_service.go**（方案1）- 这是根本解决方案
2. **如果chat_service修复复杂，先修复openai_format.go**（方案2）- 快速临时方案
3. **最后才考虑前端自定义处理**（方案3）- 最不推荐，增加前端复杂度

## 测试验证

修复后，使用以下测试用例验证：

### 测试1：单个工具调用

```
输入：查询服务器内存
预期输出：
- 所有chunks使用相同的index（0）
- 所有chunks使用相同的ID
- arguments正确合并为完整JSON
- 前端显示完整的工具调用参数
```

### 测试2：多个工具调用

```
输入：查询CPU和内存
预期输出：
- 第一个工具：index 0，完整ID，完整参数
- 第二个工具：index 1，完整ID，完整参数
- 前端正确显示两个工具调用
```

### 测试3：流式分片

```
模拟LLM返回：
- Chunk 1: ID="call_abc", name="ssh_exec", args="{"
- Chunk 2: ID="", name="", args='"cmd": "ls"'
- Chunk 3: ID="", name="", args='"host": "x.x.x.x"}'

预期输出：
- 所有chunks合并为一个tool_call
- index统一为0
- ID统一为call_abc
- arguments合并为完整JSON
```

## 相关文件

- `backend/internal/service/chat_service.go` - LLM tool_calls处理逻辑
- `backend/internal/model/openai_format.go` - SSE流格式转换
- `backend/internal/handler/chat_handler.go` - SSE流发送
- `frontend/src/views/AIAssistant-TinyRobot.vue` - 前端接收逻辑
- `frontend/test-tool-calls.html` - 测试工具和详细分析
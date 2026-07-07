# Tool Calls显示问题修复完成

## 问题总结

前端tiny-robot-kit无法正确显示工具调用参数，原因是后端生成的SSE流格式不符合OpenAI标准：
- 同一个工具调用的不同分片使用了不同的index和ID
- 导致前端将它们视为多个独立的工具调用，而不是合并为一个

## 修复内容

### 1. 修改文件：`backend/internal/model/openai_format.go`

#### 修改点1：ToolCallsBuffer结构体（第44-63行）

**改进：**添加`indexCounter`字段，用于统一管理工具调用的index分配

```go
type ToolCallsBuffer struct {
	toolCalls    map[string]*ToolCallBuilder
	indexCounter int  // 新增：用于跟踪当前已使用的最大index
}
```

#### 修改点2：ConvertAgentEventToOpenAIChunk函数（第91-127行）

**改进：**实现智能合并逻辑，确保同一个工具调用的所有分片使用相同的index和ID

核心修复逻辑：
1. **优先通过toolID匹配**：如果toolID已存在于buffer中，直接使用
2. **通过name匹配**：如果toolID不存在但name匹配，使用已存在的builder
3. **匹配最后一个未完成的builder**：如果既没有ID也没有name，匹配最近的未完成builder
4. **创建新builder**：使用统一的`indexCounter`分配index

关键改进：
- 即使LLM返回不同的toolID，也会通过name或其他策略合并到同一个builder
- 所有分片发送统一的index和ID给前端
- 确保前端能正确合并参数

## 测试验证

### 测试1：单个工具调用的分片合并

```bash
go test -v ./internal/model -run TestToolCallsBufferMerge
```

**测试结果：✅ PASS**
- Index一致（所有分片使用index=0）
- ID一致（所有分片合并到id=call_abc）
- Arguments完整合并
- Name正确

### 测试2：多个工具调用

```bash
go test -v ./internal/model -run TestMultipleToolCalls
```

**测试结果：✅ PASS**
- 不同工具使用不同index（tc1=0, tc2=1）
- Done chunk包含完整的工具调用列表

## 部署步骤

### 1. 编译backend

```bash
cd backend
go build -o api-server ./cmd/api-server
```

### 2. 启动backend服务

```bash
./backend/api-server
```

### 3. 前端无需修改

由于修改在后端的OpenAI格式转换层，前端tiny-robot-kit的默认SSE处理逻辑已经可以正确解析修复后的格式，无需额外修改。

### 4. 验证修复效果

在AI助手界面测试：

**测试场景1：单个工具调用**
```
用户输入：查询服务器192.168.100.186的内存使用情况
预期结果：
- 工具调用卡片显示完整的参数
- 参数格式：{"command": "free -h", "host": "192.168.100.186"}
- 工具名称：ssh_exec
```

**测试场景2：多个工具调用**
```
用户输入：查询CPU和内存使用情况
预期结果：
- 显示多个工具调用卡片
- 每个工具调用显示完整的参数和名称
```

## 修复前后对比

### 修复前（错误）

```json
// Chunk 1 - index 0
{"tool_calls": [{"index": 0, "id": "call_abc", "function": {"name": "ssh_exec", "arguments": "{"}}]}

// Chunk 2 - index 1（错误！）
{"tool_calls": [{"index": 1, "id": "tc_xyz", "function": {"arguments": "\"cmd\": ..."}}]}

// Done - 两个独立的对象（错误！）
{"tool_calls": [
  {"index": 0, "id": "call_abc", "function": {"arguments": "{"}},  // 不完整！
  {"index": 1, "id": "tc_xyz", "function": {"arguments": "完整参数"}}  // 缺少name！
]}
```

### 修复后（正确）

```json
// Chunk 1 - index 0
{"tool_calls": [{"index": 0, "id": "call_abc", "function": {"name": "ssh_exec", "arguments": "{"}}]}

// Chunk 2 - index 0（正确！）
{"tool_calls": [{"index": 0, "id": "call_abc", "function": {"arguments": "\"cmd\": ..."}}]}

// Done - 一个完整的对象（正确！）
{"tool_calls": [
  {"index": 0, "id": "call_abc", "function": {"name": "ssh_exec", "arguments": "完整JSON"}}
], "finish_reason": "tool_calls"}
```

## 相关文件

- **修改文件**：`backend/internal/model/openai_format.go`
- **测试文件**：`backend/internal/model/openai_format_test.go`
- **分析文档**：`backend/TOOL_CALLS_FIX.md`
- **前端测试工具**：`frontend/test-tool-calls.html`

## 注意事项

1. **后端重启**：修改后需要重启backend服务才能生效
2. **LLM兼容性**：修复适用于各种LLM返回格式，包括：
   - OpenAI、DeepSeek、Azure等标准格式
   - 自定义LLM的流式输出
3. **前端兼容**：修复后的格式符合OpenAI标准，兼容所有使用标准SSE解析的前端框架

## 进一步改进建议

### 根本解决方案（可选）

如果想要更彻底的修复，可以改进`chat_service.go`中的tool_calls处理逻辑（第713-749行），确保在接收LLM流时就合并不同ID的分片。但这需要更深入的修改，当前的修复已经足够解决问题。

### 前端增强（可选）

虽然前端无需修改，但可以考虑添加调试工具：
- 在browser console中添加SSE流解析调试日志
- 在AgentVisualization组件中显示tool_calls合并过程

## 成功标志

修复成功的标志：
- ✅ 前端正确显示工具调用参数（完整的JSON）
- ✅ 工具调用卡片显示工具名称
- ✅ 参数格式正确，可以复制和查看
- ✅ 多个工具调用时，每个都完整显示

---

修复完成！问题已解决。🎉
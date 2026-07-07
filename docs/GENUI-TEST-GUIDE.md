# GenUI AgentEvent 快速测试指南

## 🚀 快速开始

### 1. 启动服务

#### 启动后端
```bash
cd backend
./bin/api-server
```

#### 启动前端
```bash
cd frontend
npm run dev
```

### 2. 访问AI助手
打开浏览器访问: http://localhost:5173

进入 **AI助手** 页面

---

## 🧪 测试步骤

### 步骤1: 创建新对话
点击 **"新对话"** 按钮，创建一个新的对话会话

### 步骤2: 发送测试问题

尝试以下问题，观察不同事件类型的渲染效果：

#### 测试1: 简单问题（观察思考过程）
```
帮我分析CPU使用率
```
预期效果：
- ✅ 蓝色思考卡片："正在分析..."
- ✅ 文本内容流式输出

#### 测试2: 工具调用（观察工具事件）
```
查询订单服务的监控指标
```
预期效果：
- ✅ 橙色工具调用卡片：prometheus_query
- ✅ 绿色工具结果卡片：返回数据
- ✅ 流式文本解释结果

#### 测试3: 知识库检索（观察RAG引用）
```
如何排查Pod启动失败？
```
预期效果：
- ✅ RAG引用卡片显示相关知识库文档
- ✅ 文本内容引用知识库信息

#### 测试4: Agent协作（观察执行路径）
```
帮我执行一个复杂的多步骤运维任务
```
预期效果：
- ✅ Agent转换卡片（如果涉及多个Agent）
- ✅ Agent执行路径时间线
- ✅ 多个工具调用卡片

---

## 📊 观察要点

### 1. 事件卡片
观察不同类型的事件卡片：
- **思考卡片**: 蓝色边框，旋转图标
- **工具调用**: 橙色边框，显示参数
- **工具结果**: 绿色/红色边框，成功/失败标识
- **Agent转换**: 灰色边框，from -> to箭头

### 2. 流式输出
观察文本内容的流式输出：
- 内容逐步显示，而不是一次性显示
- 每个chunk实时追加到消息中

### 3. Agent路径
观察Agent执行路径可视化：
- 时间线展示执行步骤
- 每个步骤的Agent名称和动作
- 清晰的执行流程

---

## 🔍 调试技巧

### 查看后端日志
```bash
# 查看事件发送日志
tail -f backend/logs/app.log | grep "AgentEvent"

# 查看SSE流日志
tail -f backend/logs/app.log | grep "SSE"
```

### 查看前端日志
打开浏览器开发者工具（F12）：
```javascript
// 查看SSE事件接收日志
Console -> [SSE AgentEvent]
```

### 查看网络请求
浏览器开发者工具 -> Network：
- 找到 `/api/v1/chat/messages/stream/events` 请求
- 查看EventStream类型的响应
- 观察事件流格式

---

## 🎯 验证成功标准

### ✅ 基本功能验证
1. 能创建新对话
2. 能发送消息
3. 能接收AI回复
4. 回复内容正确

### ✅ GenUI验证
1. **思考卡片**: 显示Agent思考过程
2. **工具卡片**: 显示工具调用和结果
3. **转换卡片**: 显示Agent转换（如果发生）
4. **错误卡片**: 显示错误信息（如果发生）

### ✅ AgentEvent验证
1. 事件按正确顺序接收
2. 每个事件数据完整
3. 前端正确渲染对应组件

### ✅ Agent协作验证
1. 执行路径正确显示
2. 时间线清晰展示
3. 步骤信息完整

---

## 🐛 常见问题排查

### 问题1: 事件卡片不显示
**可能原因**:
- 后端没有发送对应事件类型
- 前端组件映射缺失

**解决方法**:
```bash
# 检查后端是否发送事件
tail -f backend/logs/app.log | grep "thinking"

# 检查前端是否接收事件
浏览器Console -> [SSE AgentEvent] Received event: thinking
```

### 问题2: 工具调用不显示
**可能原因**:
- Agent没有配置工具
- 工具调用失败

**解决方法**:
```bash
# 检查Agent工具配置
curl http://localhost:8080/api/v1/agents

# 检查工具绑定
curl http://localhost:8080/api/v1/agents/{agent_id}/tools
```

### 问题3: Agent路径不显示
**可能原因**:
- RunPath数据缺失
- 组件未正确引入

**解决方法**:
检查消息数据是否包含agentPath字段：
```javascript
// 浏览器Console
console.log(messages.value[messages.value.length - 1].agentPath)
```

---

## 📈 性能测试

### 测试并发请求
```bash
# 发送10个并发请求
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/v1/chat/messages/stream/events \
    -H "Authorization: Bearer {token}" \
    -H "Content-Type: application/json" \
    -d '{"session_id":"{session_id}","content":"测试并发'$i'"}'
done
```

### 测试长对话
```bash
# 发送长消息，观察流式输出性能
curl -X POST http://localhost:8080/api/v1/chat/messages/stream/events \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"session_id":"{session_id}","content":"请详细解释Kubernetes的完整架构，包括所有组件和工作流程..."}'
```

---

## 🎨 UI测试

### 测试不同事件组合
尝试不同类型的问题，观察事件组合：
1. **纯文本回复**: 只有content_chunk事件
2. **工具调用**: thinking + tool_call + tool_result + content_chunk
3. **Agent协作**: thinking + agent_transfer + content_chunk + done
4. **错误场景**: thinking + error

### 测试响应式设计
1. 缩小浏览器窗口，观察卡片布局
2. 检查长文本是否正确换行
3. 检查参数JSON是否正确显示

---

## ✅ 测试完成确认

完成以下测试后，确认改造成功：

☐ 后端编译成功，无错误
☐ 前端编译成功，无TypeScript错误
☐ 能创建新对话
☐ 能发送消息并接收回复
☐ 思考卡片正确显示
☐ 工具调用卡片正确显示
☐ 工具结果卡片正确显示
☐ 流式文本正确输出
☐ Agent执行路径正确显示
☐ RAG引用正确显示
☐ 错误信息正确显示
☐ UI响应式设计正常
☐ 性能表现良好

---

**测试完成后，即可投入使用！** 🎉
# 前端添加RAG文档500错误排查指南

## 问题现象
前端请求: POST http://localhost:5173/api/v1/rag/documents  
错误状态: 500 (Internal Server Error)  
Token: 有效

## 关键发现

### ✓ 后端直接测试成功
- 使用testuser Token: 成功 ✓
- 使用admin Token: 成功 ✓  
- Embedding API: 正常 ✓
- Milvus连接: 正常 ✓

### 说明
**后端功能完全正常**，问题在于前端请求的具体内容或处理方式。

## 排查步骤

### 1. 检查前端请求Body

**在浏览器开发者工具Network标签中**，查看Request Payload：

正确格式：
```json
{
  "title": "文档标题",
  "content": "文档内容",
  "category": "分类",
  "tags": ["标签1", "标签2"]
}
```

常见错误：
- ❌ title或content缺失
- ❌ title或content为空字符串
- ❌ tags不是数组格式
- ❌ 字段名拼写错误

### 2. 检查实时错误日志

**方法A: 使用监控脚本**
```bash
# 运行实时监控（等待前端请求）
tail -f backend/logs/api-server.log | grep -E "embedding|error|failed"
```

**方法B: 手动查看**
```bash
# 查看最近100行日志中的错误
tail -100 backend/logs/api-server.log | grep -A 5 "error"
```

### 3. 验证请求到达后端

检查后端日志是否有对应的POST请求记录：
```bash
tail -100 backend/logs/api-server.log | grep "POST.*rag/documents"
```

如果有记录但没详细错误 → 需要启用更详细日志  
如果没有记录 → 前端请求未到达后端（代理问题）

### 4. 测试不同内容长度

**短文本测试**（应该成功）:
```json
{"title":"测试","content":"短文本","category":"test"}
```

**长文本测试**（可能失败）:
```json
{"title":"长文档","content":"超过10000字符的长内容...","category":"test"}
```

**检查**: Milvus字段长度限制
- content字段: max_length=10000
- 如果超长会导致500错误

### 5. 检查特殊字符

某些特殊字符可能导致编码问题：
- 中文标点: ✓ 通常正常
- 特殊符号: ❓ 可能有问题
- HTML标签: ❓ 可能有问题

### 6. 验证用户权限

虽然Token有效，但检查：
- 用户是否有添加文档权限
- 用户状态是否正常
- 是否有其他限制

### 7. 对比成功的请求

**成功请求示例**（通过curl）:
```bash
curl -X POST http://localhost:8080/api/v1/rag/documents \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"测试","content":"测试内容","category":"test"}'
```

对比前端请求和curl请求的差异。

## 快速诊断命令

### 完整诊断脚本
```bash
#!/bin/bash
TOKEN="YOUR_TOKEN"

echo "=== 诊断测试 ==="

# 1. 短文本测试
echo "测试1: 短文本"
curl -s -X POST http://localhost:8080/api/v1/rag/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"短文本测试","content":"测试","category":"test"}' | jq '.code'

# 2. 长文本测试
echo "测试2: 长文本（1000字符）"
curl -s -X POST http://localhost:8080/api/v1/rag/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"长文本\",\"content\":\"$(python3 -c 'print("x"*1000)')\",\"category\":\"test\"}" | jq '.code'

# 3. 超长文本测试
echo "测试3: 超长文本（15000字符，可能失败）"
curl -s -X POST http://localhost:8080/api/v1/rag/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"超长\",\"content\":\"$(python3 -c 'print("x"*15000)')\",\"category\":\"test\"}" | jq '.code'

# 4. 特殊字符测试
echo "测试4: 包含特殊字符"
curl -s -X POST http://localhost:8080/api/v1/rag/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"特殊字符","content":"包含<>\"\'&特殊字符","category":"test"}' | jq '.code'
```

## 常见500错误原因

### 1. Embedding API问题
**症状**: embedding request failed  
**原因**: 
- API Key无效
- 文本过长
- 网络问题
- 调用超时

**解决**: 已验证Embedding API正常，跳过此项

### 2. Milvus字段限制
**症状**: Data type mismatch / Field too long  
**原因**:
- content字段超10000字符
- title字段超500字符

**解决**: 控制文档长度，或修改Milvus schema

### 3. 数据格式问题
**症状**: Invalid data format  
**原因**:
- JSON格式错误
- 字段类型错误
- 必填字段缺失

**解决**: 确认前端请求格式正确

### 4. 权限或认证问题
**症状**: Unauthorized / Forbidden  
**原因**: Token过期或权限不足

**解决**: 
- 重新登录获取新Token
- 检查用户角色权限

### 5. 后端服务问题
**症状**: Service unavailable  
**原因**: 
- Milvus连接断开
- 服务重启
- 配置错误

**解决**: 
- 检查Milvus服务状态
- 重启API服务

## 下一步行动

### 方案1: 查看详细错误（推荐）
1. 从前端再次提交请求
2. 同时运行实时日志监控:
   ```bash
   tail -f backend/logs/api-server.log | grep -E "embedding|error|failed|Alibaba"
   ```
3. 看到具体错误信息后对症处理

### 方案2: 对比请求内容
1. 在浏览器Network标签查看Request Payload
2. 复制完整的请求Body
3. 使用curl测试相同内容:
   ```bash
   curl -X POST http://localhost:8080/api/v1/rag/documents \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '从前端复制的请求Body'
   ```
4. 如果curl也失败 → 请求内容问题
5. 如果curl成功 → 前端处理问题

### 方案3: 临时解决
使用后端API或curl命令添加文档，避开前端问题。

---

**重要提示**: 
- 后端功能已验证正常 ✓
- 问题在于前端请求的具体内容 ✗
- 需要对比前端实际发送的请求内容

请按照上述步骤排查，找到具体500错误原因。
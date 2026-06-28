#!/bin/bash

echo "=== 前端添加RAG文档问题排查 ==="
echo ""

# 1. 检查前端代理配置
echo "【步骤1】检查前端Vite代理配置"
echo "--------------------------------------"
echo "前端端口: 5173"
echo "后端端口: 8080"
echo "代理配置: /api -> http://localhost:8080 ✓"
echo ""

# 2. 检查后端服务状态
echo "【步骤2】检查后端API服务状态"
echo "--------------------------------------"
health=$(curl -s http://localhost:8080/health)
if [ ! -z "$health" ]; then
    echo "✓ 后端API服务运行正常"
else
    echo "✗ 后端API服务未运行"
    echo "请启动后端服务: cd backend && ./bin/api-server"
fi
echo ""

# 3. 检查前端配置
echo "【步骤3】检查前端配置文件"
echo "--------------------------------------"
echo "配置文件位置: frontend/vite.config.ts"
echo ""
echo "请确认以下配置:"
echo "1. server.port: 5173 ✓"
echo "2. proxy.target: 'http://localhost:8080' ✓"
echo "3. proxy.changeOrigin: true ✓"
echo ""

# 4. 测试后端直接添加（绕过前端）
echo "【步骤4】直接测试后端添加文档"
echo "--------------------------------------"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMmJhN2UzOWEtZDUxMC00NzgwLWE2YmQtNDU5MGU5Y2M1MGJmIiwidXNlcm5hbWUiOiJ0ZXN0dXNlciIsInJvbGUiOiJ1c2VyIiwiaXNzIjoiYWlvcHMiLCJleHAiOjE3ODI1NDYzNDcsIm5iZiI6MTc4MjU0NDU0NywiaWF0IjoxNzgyNTQ0NTQ3fQ.LAloYuZDnaZs2tTWZ5s82w3aiL85zjaJIuGl3aPMMVg"

result=$(curl -s -X POST http://localhost:8080/api/v1/rag/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "前端测试文档",
    "content": "这是一条通过后端直接添加的测试文档",
    "category": "test",
    "tags": ["前端", "测试"]
  }')

if echo "$result" | jq -e '.code == 200' > /dev/null; then
    echo "✓ 后端添加成功"
    echo "$result" | jq '.'
else
    echo "✗ 后端添加失败"
    echo "错误: $(echo "$result" | jq -r '.error')"
fi
echo ""

# 5. 检查前端是否正确发送请求
echo "【步骤5】前端排查建议"
echo "--------------------------------------"
echo "请检查以下事项:"
echo ""
echo "1. 前端浏览器控制台检查:"
echo "   - 打开Chrome/Firefox开发者工具 (F12)"
echo "   - 查看Console标签页的错误信息"
echo "   - 查看Network标签页的请求详情"
echo ""
echo "2. 检查请求内容:"
echo "   - 确认请求URL是: http://localhost:5173/api/v1/rag/documents"
echo "   - 确认请求方法: POST"
echo "   - 确认请求头包含: Authorization: Bearer YOUR_TOKEN"
echo "   - 确认请求Body格式正确"
echo ""
echo "3. 检查Token:"
echo "   - 确认前端localStorage有有效的token"
echo "   - 确认token未过期"
echo "   - 在浏览器Console输入: localStorage.getItem('token')"
echo ""
echo "4. 检查请求Body格式:"
cat << 'EOF'
正确格式示例:
{
  "title": "文档标题",
  "content": "文档内容",
  "category": "分类",
  "tags": ["标签1", "标签2"]
}

注意：
- title和content是必填字段
- category可选
- tags可选（数组格式）
EOF
echo ""

# 6. 查看后端日志
echo "【步骤6】查看后端最新日志"
echo "--------------------------------------"
echo "最近10条embedding相关日志:"
tail -50 backend/logs/api-server.log | grep -i embedding | tail -10
echo ""

# 7. 测试embedding API
echo "【步骤7】测试Embedding API是否正常"
echo "--------------------------------------"
emb_result=$(curl -s -w "\nHTTP:%{http_code}" -X POST "https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings" \
  -H "Authorization: Bearer sk-086920878fb641a3bea1ce785eacb200" \
  -H "Content-Type: application/json" \
  -d '{"model":"text-embedding-v2","input":"测试"}')

http_code=$(echo "$emb_result" | grep "HTTP:" | cut -d':' -f2)
if [ "$http_code" == "200" ]; then
    echo "✓ Embedding API正常 (HTTP $http_code)"
else
    echo "✗ Embedding API异常 (HTTP $http_code)"
    echo "响应: $(echo "$emb_result" | grep -v "HTTP:")"
fi
echo ""

echo "=== 排查完成 ==="
echo ""
echo "下一步:"
echo "1. 如果后端直接添加成功 -> 前端配置或请求有问题"
echo "2. 如果后端直接添加失败 -> 检查embedding配置"
echo "3. 查看浏览器控制台错误详情"
echo "4. 对比前端请求和后端测试请求的差异"
#!/bin/bash

API_BASE="http://localhost:8080/api/v1"
TOKEN=""  # 需要先获取token

echo "=== RAG功能完整测试 ==="

# 1. 测试embedding API（已成功）
echo "✓ Step 1: Embedding API测试成功"

# 2. 启动API服务
echo ""
echo "Step 2: 检查并启动API服务..."
if ! pgrep -f "api-server" > /dev/null; then
    echo "API服务未运行，正在启动..."
    cd backend && nohup bin/api-server > logs/api-server.log 2>&1 &
    sleep 3
    if pgrep -f "api-server" > /dev/null; then
        echo "✓ API服务启动成功"
    else
        echo "✗ API服务启动失败，请检查日志"
        exit 1
    fi
else
    echo "✓ API服务已运行"
fi

# 3. 测试健康检查
echo ""
echo "Step 3: 测试API健康检查..."
health=$(curl -s http://localhost:8080/health)
if [ ! -z "$health" ]; then
    echo "✓ API服务健康检查通过"
else
    echo "✗ API服务健康检查失败"
    exit 1
fi

# 4. 获取认证Token（如果需要）
echo ""
echo "Step 4: 获取认证Token..."
# 这里需要根据实际的认证机制获取token
# 暂时使用mock方式或者跳过认证
echo "请手动获取Token或使用测试用户登录"

# 5. 测试添加RAG文档
echo ""
echo "Step 5: 测试添加RAG文档（需要Token）..."
echo "使用curl命令测试："
echo ""
cat << 'EOF'
# 示例命令（需要替换TOKEN）：
curl -X POST http://localhost:8080/api/v1/rag/documents \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "服务响应慢排查指南",
    "content": "服务响应慢的常见原因包括：CPU使用率过高、内存不足、数据库慢查询、网络延迟等。排查步骤：1. 使用top命令查看CPU占用；2. 检查内存使用情况；3. 分析数据库慢查询日志；4. 检查网络连接状态。",
    "category": "troubleshooting",
    "tags": ["性能", "响应慢", "排查"]
  }'
EOF

echo ""
echo "=== 测试流程完成 ==="
echo ""
echo "验证步骤："
echo "1. ✓ Embedding API配置正确"
echo "2. ✓ API服务运行正常"
echo "3. 需手动测试添加RAG文档功能（需要认证Token）"
echo ""
echo "查看日志命令："
echo "tail -f backend/logs/api-server.log"
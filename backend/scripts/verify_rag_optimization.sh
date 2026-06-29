#!/bin/bash

echo "========================================="
echo "  RAG优化部署验证脚本"
echo "========================================="
echo ""

# 等待服务启动
echo "步骤1：等待服务启动..."
while ! lsof -ti:8080 > /dev/null 2>&1; do
    echo "  服务未启动，等待5秒..."
    sleep 5
done
echo "  ✓ 服务已启动（8080端口监听）"

sleep 5

# 检查ChatHandler初始化
echo ""
echo "步骤2：检查RAG功能状态..."
if tail -100 /tmp/aiops-backend-final.log | grep -q "ChatHandler初始化成功"; then
    echo "  ✓ ChatHandler初始化成功，RAG功能已启用"
else
    echo "  ✗ ChatHandler未初始化，请检查日志"
    tail -50 /tmp/aiops-backend-final.log | grep -E "error|Error|ERROR|fail|Fail"
    exit 1
fi

# 获取token
echo ""
echo "步骤3：获取认证token..."
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "  ✗ 登录失败，无法获取token"
    exit 1
fi
echo "  ✓ Token获取成功"

# 同步数据（如果Milvus中没有数据）
echo ""
echo "步骤4：同步数据到Milvus（使用归一化向量）..."
echo "  提示：此步骤可能需要1-2分钟，取决于文档数量"
cd backend/scripts
timeout 180 go run sync_pg_to_milvus.go > /tmp/sync.log 2>&1

if [ $? -eq 0 ]; then
    echo "  ✓ 数据同步成功"
    grep "同步完成\|插入成功\|Document added" /tmp/sync.log | tail -5
else
    echo "  ⚠ 数据同步超时或失败，检查日志:"
    tail -20 /tmp/sync.log | grep -E "error|Error|失败"
fi

# 测试RAG检索
echo ""
echo "步骤5：测试RAG检索功能..."
RESULT=$(curl -s -X POST http://localhost:8080/api/v1/rag/search \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"query":"Docker常用命令","top_k":2}')

if echo "$RESULT" | jq -e '.data' > /dev/null 2>&1; then
    echo "  ✓ RAG检索成功！"
    echo ""
    echo "  检索结果示例:"
    echo "$RESULT" | jq '.data[0] | {title: .document.title, score: .score, level: .relevance_level}'
else
    echo "  ✗ RAG检索失败"
    echo "  错误信息:"
    echo "$RESULT" | jq -r '.error'
    
    # 如果是metric type不匹配，建议重建索引
    if echo "$RESULT" | grep -q "metric type not match"; then
        echo ""
        echo "  ⚠ 检测到metric type不匹配，建议重建索引:"
        echo "    cd backend/scripts"
        echo "    echo 'YES' | go run rebuild_milvus_index.go"
        echo "    go run sync_pg_to_milvus.go"
    fi
    exit 1
fi

# 测试分级阈值
echo ""
echo "步骤6：验证分级阈值策略..."
echo "  测试查询: 'Docker镜像管理'"
RESULT2=$(curl -s -X POST http://localhost:8080/api/v1/rag/search \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"query":"Docker镜像管理","top_k":3}')

echo "  检索结果评分:"
echo "$RESULT2" | jq -r '.data[] | "    标题: \(.document.title), 评分: \(.score), 等级: \(.relevance_level)"'

# 测试聊天API（包含RAG引用）
echo ""
echo "步骤7：测试聊天API的RAG引用..."
SESSION_ID=$(curl -s -X POST http://localhost:8080/api/v1/chat/sessions \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"RAG测试对话"}' | jq -r '.data.id')

echo "  创建对话session: $SESSION_ID"

# 发送查询并检查RAG引用
echo "  发送查询: '如何查看Docker容器日志'"
curl -N -X POST http://localhost:8080/api/v1/chat/messages/stream \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -H "Accept: text/event-stream" \
    -d "{\"session_id\":\"$SESSION_ID\",\"content\":\"如何查看Docker容器日志\"}" 2>&1 | \
    grep "rag_references" | head -3

echo ""
echo "========================================="
echo "  ✓ RAG优化部署验证完成"
echo "========================================="
echo ""
echo "下一步建议:"
echo "  1. 启动前端: cd frontend && npm run dev"
echo "  2. 测试前端UI: 查看相关度等级标签和颜色"
echo "  3. 监控日志: tail -f /tmp/aiops-backend-final.log | grep RAG"
echo ""
#!/bin/bash

echo "=== Milvus字段长度限制修复方案 ==="
echo ""
echo "问题: content字段限制10000字符，超出会报500错误"
echo "解决: 增加content字段长度到50000字符"
echo ""

# 方案1: 删除旧collection重建（简单，会丢失数据）
echo "【方案1】删除旧collection并重建（推荐用于测试环境）"
echo "----------------------------------------"
echo "优点: 简单快速"
echo "缺点: 会丢失现有知识库数据"
echo ""
echo "执行步骤:"
echo "1. 停止API服务"
echo "2. 修改代码中的max_length配置"
echo "3. 删除Milvus collection"
echo "4. 重启API服务（自动创建新collection）"
echo ""

# 方案2: 创建新collection（保留旧数据）
echo "【方案2】创建新的collection（推荐用于生产环境）"
echo "----------------------------------------"
echo "优点: 保留旧数据"
echo "缺点: 需要数据迁移"
echo ""
echo "执行步骤:"
echo "1. 创建新collection（如knowledge_documents_v2）"
echo "2. 迁移旧数据到新collection"
echo "3. 修改代码使用新collection"
echo ""

# 方案3: 前端限制（临时方案）
echo "【方案3】前端限制输入长度（最快）"
echo "----------------------------------------"
echo "优点: 不需要修改后端"
echo "缺点: 用户体验不佳"
echo ""
echo "限制建议:"
echo "- title: 最多500字符"
echo "- content: 最多10000字符"
echo "- 前端提示用户分段保存长文档"
echo ""

echo "=== 推荐执行方案 ==="
echo ""

# 检查现有数据
echo "检查现有知识库数据数量..."
total=$(curl -s 'http://localhost:8080/api/v1/rag/documents?page=1&pageSize=1000' \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiY2JjM2FmMmQtNWJkZS00NjA4LWE2MmEtOTYwMWY5OTczMjY0IiwidXNlcm5hbWUiOiJhZG1pbiIsInJvbGUiOiJhZG1pbiIsImlzcyI6ImFpb3BzIiwiZXhwIjoxNzgyNTQ0OTE1LCJuYmYiOjE3ODI1NDMxMTUsImlhdCI6MTc4MjU0MzExNX0.7gq2nhT11h8OEizXO74svmJ-xJvMyT3ehbJoyGXSVqo" | jq '.total // 0')

echo "现有文档数量: $total"
echo ""

if [ "$total" -lt "20" ]; then
    echo "✓ 数据量较少（<20条），推荐使用方案1（删除重建）"
    echo ""
    echo "执行命令:"
    cat << 'EOF'
# 1. 修改代码
vim backend/internal/service/milvus_service.go
# 将第84行的 "max_length": "10000" 改为 "max_length": "50000"

# 2. 重新编译
cd backend && go build -o bin/api-server ./cmd/api-server/

# 3. 停止服务
killall api-server

# 4. 删除Milvus collection（需要Milvus客户端）
# 方法A: 使用Python Milvus SDK
# python3 -c "from pymilvus import connections, utility; connections.connect(); utility.drop_collection('knowledge_documents')"

# 方法B: 重启API服务会自动创建新collection（如果检测到不存在）
# 修改代码让CreateCollection在collection存在时重建

# 5. 重启服务
cd backend && ./bin/api-server
EOF
else
    echo "⚠ 数据量较多（$total条），推荐使用方案2（创建新collection）"
fi

echo ""
echo "=== 快速测试方案（不丢失数据） ==="
cat << 'EOF'
修改 milvus_service.go 第84行：
TypeParams: map[string]string{"max_length": "50000"},  // 从10000改为50000

然后修改CreateCollection逻辑：
if has {
    // 删除旧collection重建
    m.client.DropCollection(ctx, m.collection)
    logger.Info("Dropping old collection to recreate with new schema")
}
EOF

echo ""
echo "需要帮助执行吗？请告诉我选择哪个方案"
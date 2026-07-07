#!/bin/bash

# 快速API测试 - 最常用接口

BASE_URL="http://localhost:8080"

echo "AiOpsHub 快速测试"
echo ""

echo "1. 健康检查"
curl -s "$BASE_URL/health" | jq .

echo ""
echo "2. 指标查看"
curl -s "$BASE_URL/metrics" | jq .

echo ""
echo "3. 登录获取Token"
TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.token')
echo "Token: $TOKEN"

echo ""
echo "4. 执行协作Workflow"
curl -s -X POST "$BASE_URL/api/v1/workflows/collaboration" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "quick-test",
    "user_query": "分析服务性能问题",
    "context": {"service": "test-service"}
  }' | jq .

echo ""
echo "完成！"
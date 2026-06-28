#!/bin/bash

# AiOpsHub API 测试脚本

BASE_URL="http://localhost:8080"
API_VERSION="v1"

echo "===================================="
echo "AiOpsHub API 测试"
echo "===================================="
echo ""

# 1. 健康检查
echo "1. 健康检查"
echo "----------------"
curl -s "$BASE_URL/health" | jq .
echo ""

echo "存活检查"
curl -s "$BASE_URL/healthz" | jq .
echo ""

echo "就绪检查"
curl -s "$BASE_URL/ready" | jq .
echo ""

# 2. 监控指标
echo "2. 监控指标"
echo "----------------"
curl -s "$BASE_URL/metrics" | jq .
echo ""

echo "Prometheus指标"
curl -s "$BASE_URL/prometheus"
echo ""

# 3. 用户认证
echo "3. 用户认证"
echo "----------------"
echo "注册用户"
curl -s -X POST "$BASE_URL/api/$API_VERSION/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123","email":"test@example.com"}' | jq .
echo ""

echo "登录"
TOKEN=$(curl -s -X POST "$BASE_URL/api/$API_VERSION/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123"}' | jq -r '.token')
echo "Token: $TOKEN"
echo ""

# 4. Agent管理
echo "4. Agent管理"
echo "----------------"
echo "列出Agent"
curl -s "$BASE_URL/api/$API_VERSION/agents" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

echo "创建Agent"
curl -s -X POST "$BASE_URL/api/$API_VERSION/agents" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MonitorAgent",
    "type": "monitor",
    "description": "监控采集Agent",
    "config": {"prometheus_url": "http://localhost:9090"}
  }' | jq .
echo ""

# 5. Workflow管理
echo "5. Workflow管理"
echo "----------------"
echo "列出Workflow"
curl -s "$BASE_URL/api/$API_VERSION/workflows" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

echo "执行协作Workflow"
WORKFLOW_ID=$(curl -s -X POST "$BASE_URL/api/$API_VERSION/workflows/collaboration" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-session-001",
    "user_query": "订单服务响应慢，帮我分析原因",
    "context": {"service": "order-service"}
  }' | jq -r '.workflow_id')
echo "Workflow ID: $WORKFLOW_ID"
echo ""

# 6. Workflow状态查询
echo "6. Workflow状态查询"
echo "----------------"
echo "查询状态"
curl -s "$BASE_URL/api/$API_VERSION/workflows/$WORKFLOW_ID/status" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

echo "查询结果"
curl -s "$BASE_URL/api/$API_VERSION/workflows/$WORKFLOW_ID/result" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 7. Signal发送
echo "7. Signal发送"
echo "----------------"
curl -s -X POST "$BASE_URL/api/$API_VERSION/workflows/$WORKFLOW_ID/signal" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"signal_name":"approval","value":{"approved":true,"user_id":"admin"}}' | jq .
echo ""

# 8. Query查询
echo "8. Query查询"
echo "----------------"
curl -s "$BASE_URL/api/$API_VERSION/workflows/$WORKFLOW_ID/query?query_type=progress" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 9. 知识库管理
echo "9. 知识库管理"
echo "----------------"
echo "搜索知识"
curl -s -X POST "$BASE_URL/api/$API_VERSION/knowledge/search" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query":"服务响应慢的常见原因","top_k":5}' | jq .
echo ""

# 10. 告警管理
echo "10. 告警管理"
echo "----------------"
echo "列出告警"
curl -s "$BASE_URL/api/$API_VERSION/alerts" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

echo "创建告警"
curl -s -X POST "$BASE_URL/api/$API_VERSION/alerts" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "CPU高使用率告警",
    "severity": "high",
    "source": "prometheus",
    "description": "CPU使用率超过80%"
  }' | jq .
echo ""

# 11. 监控统计
echo "11. 监控统计"
echo "----------------"
curl -s "$BASE_URL/api/$API_VERSION/monitor/stats" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

curl -s "$BASE_URL/api/$API_VERSION/monitor/performance" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

echo "===================================="
echo "测试完成！"
echo "===================================="
echo ""
echo "WebSocket测试："
echo "  连接: ws://localhost:8080/ws"
echo ""
echo "Temporal UI: http://localhost:8080"
echo ""
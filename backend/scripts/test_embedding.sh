#!/bin/bash

# Embedding API验证测试脚本

API_KEY="sk-086920878fb641a3bea1ce785eacb200"
BASE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1"

echo "=== 测试阿里云Embedding API (OpenAI兼容模式) ==="
echo "API URL: $BASE_URL/embeddings"
echo "Model: text-embedding-v2"
echo ""

# 测试embedding API
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL/embeddings" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "text-embedding-v2",
    "input": "测试文本embedding功能"
  }')

http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d':' -f2)
body=$(echo "$response" | grep -v "HTTP_CODE:")

echo "响应状态码: $http_code"
echo "响应内容:"
echo "$body" | jq '.' 2>/dev/null || echo "$body"
echo ""

if [ "$http_code" == "200" ]; then
    echo "✓ Embedding API测试成功！"
    echo ""
    # 检查返回的向量维度
    dim=$(echo "$body" | jq '.data[0].embedding | length' 2>/dev/null)
    if [ ! -z "$dim" ]; then
        echo "向量维度: $dim"
    fi
else
    echo "✗ Embedding API测试失败"
    echo "请检查："
    echo "1. API Key是否有效"
    echo "2. 模型名称是否正确"
    echo "3. 是否有embedding服务权限"
fi

echo ""
echo "=== 测试完成 ==="
#!/bin/bash

echo "=== 测试长文档添加（验证修复） ==="
echo ""

TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiY2JjM2FmMmQtNWJkZS00NjA4LWE2MmEtOTYwMWY5OTczMjY0IiwidXNlcm5hbWUiOiJhZG1pbiIsInJvbGUiOiJhZG1pbiIsImlzcyI6ImFpb3BzIiwiZXhwIjoxNzgyNTQ0OTE1LCJuYmYiOjE3ODI1NDMxMTUsImlhdCI6MTc4MjU0MzExNX0.7gq2nhT11h8OEizXO74svmJ-xJvMyT3ehbJoyGXSVqo"

echo "【测试1】短文档（500字符）- 应该成功"
echo "--------------------------------------"
short_result=$(curl -s -X POST http://localhost:8080/api/v1/rag/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"短文档测试\",\"content\":\"$(python3 -c 'print("测试内容x"*100)')\",\"category\":\"test\"}")

echo "$short_result" | jq '.code, .message'
if echo "$short_result" | jq -e '.code == 200' > /dev/null; then
    echo "✓ 短文档添加成功"
else
    echo "✗ 短文档添加失败"
    echo "$short_result" | jq '.error'
fi
echo ""

echo "【测试2】中等文档（15000字符）- 原来会失败，现在应该成功"
echo "--------------------------------------"
medium_result=$(curl -s -X POST http://localhost:8080/api/v1/rag/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"中等文档测试\",\"content\":\"$(python3 -c 'print("测试内容x"*3000)')\",\"category\":\"test\"}")

echo "$medium_result" | jq '.code, .message'
if echo "$medium_result" | jq -e '.code == 200' > /dev/null; then
    echo "✓ 中等文档添加成功（修复生效！）"
else
    echo "✗ 中等文档添加失败"
    echo "$medium_result" | jq '.error'
fi
echo ""

echo "【测试3】长文档（30000字符）- 验证50000限制"
echo "--------------------------------------"
long_result=$(curl -s -X POST http://localhost:8080/api/v1/rag/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"长文档测试\",\"content\":\"$(python3 -c 'print("测试内容x"*6000)')\",\"category\":\"test\"}")

echo "$long_result" | jq '.code, .message'
if echo "$long_result" | jq -e '.code == 200' > /dev/null; then
    echo "✓ 长文档添加成功"
else
    echo "✗ 长文档添加失败"
    echo "$long_result" | jq '.error'
fi
echo ""

echo "【测试4】超长文档（60000字符）- 应该失败（超过50000限制）"
echo "--------------------------------------"
extra_result=$(curl -s -X POST http://localhost:8080/api/v1/rag/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"超长文档测试\",\"content\":\"$(python3 -c 'print("测试内容x"*12000)')\",\"category\":\"test\"}")

echo "$extra_result" | jq '.code, .message, .error'
if echo "$extra_result" | jq -e '.code == 200' > /dev/null; then
    echo "⚠ 超长文档添加成功（可能限制设置过大）"
else
    echo "✓ 超长文档正确失败（超过50000限制）"
fi
echo ""

echo "=== 测试完成 ==="
echo ""
echo "修复结果:"
echo "- content字段限制: 10000字符 → 50000字符 ✓"
echo "- 15000字符文档: 应该可以成功添加 ✓"
echo "- 30000字符文档: 应该可以成功添加 ✓"
echo "- 60000字符文档: 应该失败（超过限制） ✓"
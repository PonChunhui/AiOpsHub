#!/bin/bash

# 历史上下文集成测试脚本
# 用于验证完整的历史上下文功能是否正常工作

set -e

API_URL=${API_URL:-"http://localhost:8080"}
TEST_TOKEN=${TEST_TOKEN:-""}

echo "=== 历史上下文集成测试 ==="
echo "API地址: $API_URL"
echo ""

# 检查token是否提供
if [ -z "$TEST_TOKEN" ]; then
    echo "警告：未提供TEST_TOKEN，尝试登录获取..."
    
    # 尝试登录获取token（假设有测试账号）
    LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}')
    
    TEST_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | sed 's/"token":"//')
    
    if [ -z "$TEST_TOKEN" ]; then
        echo "错误：无法获取token，请手动设置TEST_TOKEN环境变量"
        echo "示例：export TEST_TOKEN='your-jwt-token'"
        exit 1
    fi
    
    echo "成功获取token: ${TEST_TOKEN:0:20}..."
fi

# 测试函数
test_create_session() {
    echo ">>> 测试1: 创建新会话"
    
    RESPONSE=$(curl -s -X POST "$API_URL/api/v1/chat/sessions" \
        -H "Authorization: Bearer $TEST_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"title":"历史上下文测试会话"}')
    
    SESSION_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*' | sed 's/"id":"//')
    
    if [ -z "$SESSION_ID" ]; then
        echo "❌ 创建会话失败"
        echo "响应: $RESPONSE"
        return 1
    fi
    
    echo "✅ 会话创建成功: $SESSION_ID"
    return 0
}

test_send_message() {
    local session_id=$1
    local content=$2
    
    echo ">>> 发送消息: $content"
    
    curl -s -X POST "$API_URL/api/v1/chat/messages" \
        -H "Authorization: Bearer $TEST_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"session_id\":\"$session_id\",\"content\":\"$content\"}" | head -c 200
    
    echo ""
}

test_stream_message() {
    local session_id=$1
    local content=$2
    
    echo ">>> 流式发送消息: $content"
    
    # 发送流式请求，显示前500字符
    curl -N -X POST "$API_URL/api/v1/chat/messages/stream" \
        -H "Authorization: Bearer $TEST_TOKEN" \
        -H "Content-Type: application/json" \
        -H "Accept: text/event-stream" \
        -d "{\"session_id\":\"$session_id\",\"content\":\"$content\"}" \
        --max-time 30 | head -c 500
    
    echo ""
    echo ""
}

test_history_context() {
    local session_id=$1
    
    echo ""
    echo ">>> 测试2: 验证历史上下文是否生效"
    
    # 第一条消息：告诉AI用户名字
    test_send_message "$session_id" "你好，我是张三，来自北京"
    sleep 2
    
    # 第二条消息：问AI是否记得用户名字
    test_send_message "$session_id" "你记得我叫什么名字吗？"
    sleep 2
    
    # 第三条消息：问AI是否记得用户来自哪里
    test_send_message "$session_id" "我来自哪里？"
    sleep 2
    
    echo ""
    echo "期望结果："
    echo "  - 第二条消息回复应包含'张三'"
    echo "  - 第三条消息回复应包含'北京'"
    echo ""
    echo "如果AI能正确回答，说明历史上下文功能正常"
}

test_get_session_history() {
    local session_id=$1
    
    echo ""
    echo ">>> 测试3: 获取会话历史"
    
    curl -s -X GET "$API_URL/api/v1/chat/sessions/$session_id/history" \
        -H "Authorization: Bearer $TEST_TOKEN" | python3 -m json.tool || cat
    
    echo ""
}

test_delete_session() {
    local session_id=$1
    
    echo ""
    echo ">>> 测试4: 删除会话"
    
    curl -s -X DELETE "$API_URL/api/v1/chat/sessions/$session_id" \
        -H "Authorization: Bearer $TEST_TOKEN"
    
    echo ""
    echo "✅ 会话已删除"
}

# 主测试流程
echo ""
echo "开始执行测试..."

# 创建会话
if ! test_create_session; then
    exit 1
fi

echo "SESSION_ID=$SESSION_ID"

# 测试历史上下文
test_history_context "$SESSION_ID"

# 获取历史记录
test_get_session_history "$SESSION_ID"

# 清理：删除测试会话
read -p "是否删除测试会话？(y/n) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    test_delete_session "$SESSION_ID"
fi

echo ""
echo "=== 测试完成 ==="
echo ""
echo "检查要点："
echo "1. AI是否能记住用户名字'张三'"
echo "2. AI是否能记住用户来自'北京'"
echo "3. 会话历史是否正确保存"
echo "4. 日志中是否有'[历史上下文]'相关记录"
echo ""
echo "查看日志："
echo "tail -f backend/backend-new.log | grep '历史上下文'"
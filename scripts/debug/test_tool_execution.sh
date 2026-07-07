#!/bin/bash

echo "=== 工具执行验证测试 ==="

# 1. 初始化工具
echo "\n[1] 初始化预设工具..."
cd backend
go run scripts/init_preset_tools.go

# 2. 验证工具列表
echo "\n[2] 验证工具列表 API..."
curl -s http://localhost:8080/api/v1/tools | python3 -m json.tool | grep -A 3 '"name"' || echo "API 未启动或工具为空"

# 3. 检查 Agent-Tool 绑定
echo "\n[3] 检查数据库工具绑定..."
echo "请在前端界面挂载工具后再检查"

# 4. 测试对话提示
echo "\n[4] 测试对话触发..."
echo "在前端对话界面输入: '帮我检查 localhost 的 CPU 状态'"
echo "观察 backend 日志中是否出现:"
echo "  - [INFO] 检测到工具调用，开始解析和执行..."
echo "  - [INFO] 开始执行工具 ssh_exec"
echo "  - [INFO] 工具 ssh_exec 执行成功"

echo "\n=== 测试完成 ==="

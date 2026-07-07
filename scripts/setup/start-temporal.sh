#!/bin/bash

# AiOpsHub Temporal快速启动脚本
# 适用于macOS环境（已安装Docker Desktop）

echo "======================================"
echo "  AiOpsHub Temporal Server启动脚本"
echo "======================================"

# 检查Docker是否安装
echo "\n[1/5] 检查Docker环境..."
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装"
    echo "\n请先安装Docker Desktop："
    echo "  方式1（官网）：https://www.docker.com/products/docker-desktop"
    echo "  方式2（Homebrew）：brew install --cask docker"
    echo "\n安装完成后，启动Docker Desktop，然后重新运行此脚本"
    exit 1
fi

echo "✅ Docker已安装"
docker --version

# 检查Docker是否运行
echo "\n[2/5] 检查Docker运行状态..."
if ! docker ps &> /dev/null; then
    echo "❌ Docker未运行"
    echo "\n请启动Docker Desktop："
    echo "  open /Applications/Docker.app"
    echo "\n启动后等待Docker Desktop完全启动（右上角图标稳定），然后重新运行此脚本"
    exit 1
fi

echo "✅ Docker正在运行"

# 进入deployments目录
echo "\n[3/5] 进入deployments目录..."
cd "$(dirname "$0")/../deployments" || {
    echo "❌ 无法找到deployments目录"
    exit 1
}
echo "✅ 当前目录: $(pwd)"

# 启动Temporal服务
echo "\n[4/5] 启动Temporal Server..."
echo "正在启动以下服务："
echo "  - temporal-server (Temporal工作流引擎)"
echo "  - temporal-postgres (Temporal持久化数据库)"

docker compose up -d temporal-server temporal-postgres

# 等待启动
echo "\n等待Temporal Server启动（约30-60秒）..."
sleep 10
echo "  等待10秒..."
sleep 10
echo "  等待20秒..."
sleep 10
echo "  等待30秒..."

# 检查容器状态
echo "\n[5/5] 检查容器状态..."
docker compose ps temporal-server temporal-postgres

# 提供访问信息
echo "\n======================================"
echo "  Temporal Server启动完成！"
echo "======================================"
echo "\n📊 Temporal Web UI访问地址："
echo "  http://localhost:8080"
echo "\n🔧 Temporal Server端口："
echo "  localhost:7233 (Worker连接)"
echo "\n📚 下一步操作："
echo "  1. 打开浏览器访问: http://localhost:8080"
echo "  2. 浏览Temporal Web UI界面"
echo "  3. 查看文档: docs/deployment/temporal-deployment.md"
echo "\n停止Temporal："
echo "  cd deployments && docker compose stop temporal-server temporal-postgres"
echo "\n查看日志："
echo "  cd deployments && docker compose logs temporal-server"
echo "\n======================================"

# 自动打开Web UI（macOS）
if command -v open &> /dev/null; then
    echo "\n正在打开Temporal Web UI..."
    open http://localhost:8080
fi
#!/bin/bash
# Temporal Server 启动和验证脚本

set -e

echo "=========================================="
echo "AiOpsHub Temporal开发环境启动"
echo "=========================================="

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

echo "✅ Docker已运行"

# 进入deployments目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")/deployments"

cd "$DEPLOY_DIR" || exit 1

echo "当前目录: $(pwd)"

# 清理旧容器和数据（可选）
if [ "$1" == "--clean" ]; then
    echo "清理旧数据..."
    docker-compose down -v
    echo "✅ 清理完成"
fi

# 启动PostgreSQL
echo "启动PostgreSQL..."
docker-compose up -d postgres

# 等待PostgreSQL就绪
echo "等待PostgreSQL启动..."
sleep 5

until docker-compose exec postgres pg_isready -U aiops; do
    echo "等待PostgreSQL..."
    sleep 2
done

echo "✅ PostgreSQL已就绪"

# 启动Redis
echo "启动Redis..."
docker-compose up -d redis

# 启动Temporal Server
echo "启动Temporal Server..."
docker-compose up -d temporal-server

# 等待Temporal Server初始化（最多等待60秒）
echo "等待Temporal Server初始化（约30-60秒）..."
sleep 10

max_wait=60
waited=0

while [ $waited -lt $max_wait ]; do
    # 检查Temporal Server日志
    if docker-compose logs temporal-server 2>&1 | grep -q "Server started"; then
        echo "✅ Temporal Server已启动"
        break
    fi
    
    # 检查是否有错误
    if docker-compose logs temporal-server 2>&1 | grep -q "Unable to create server"; then
        echo "❌ Temporal Server启动失败，查看日志："
        docker-compose logs temporal-server
        exit 1
    fi
    
    echo "等待中... ($waited秒)"
    sleep 5
    waited=$((waited + 5))
done

if [ $waited -ge $max_wait ]; then
    echo "⚠️  Temporal Server启动超时，查看日志："
    docker-compose logs --tail=50 temporal-server
    exit 1
fi

# 显示服务状态
echo ""
echo "=========================================="
echo "服务状态"
echo "=========================================="
docker-compose ps

echo ""
echo "=========================================="
echo "服务访问信息"
echo "=========================================="
echo "Temporal Web UI:  http://localhost:8080"
echo "PostgreSQL:       localhost:5432 (aiops/aiops123)"
echo "Redis:            localhost:6379"
echo ""

echo "=========================================="
echo "数据库列表"
echo "=========================================="
docker-compose exec postgres psql -U aiops -c "\l"

echo ""
echo "✅ 启动完成！"
echo ""
echo "访问Temporal Web UI: http://localhost:8080"
echo ""
echo "查看日志: docker-compose logs -f temporal-server"
echo "停止服务: docker-compose down"
echo "清理数据: docker-compose down -v"
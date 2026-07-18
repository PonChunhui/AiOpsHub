#!/bin/bash

# AiOpsHub 开发环境快速启动脚本

set -e

cd "$(dirname "$0")"

echo "=== AiOpsHub Development Environment Setup ==="

# 检查Docker
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "Error: Docker Compose is not installed"
    exit 1
fi

# 询问启动模式
echo ""
echo "Select startup mode:"
echo "1) Full stack (PostgreSQL + Redis + Milvus + Monitoring)"
echo "2) Core services only (PostgreSQL + Redis)"
echo ""
read -p "Enter choice [1-2]: " choice

case $choice in
    1)
        echo "Starting full stack..."
        docker-compose up -d
        ;;
    2)
        echo "Starting core services..."
        docker-compose up -d postgres redis
        ;;
    *)
        echo "Invalid choice"
        exit 1
        ;;
esac

echo ""
echo "Waiting for services to start..."
sleep 10

# 检查服务状态
echo ""
echo "=== Service Status ==="
docker-compose ps

echo ""
echo "=== Access Information ==="
echo "Grafana Dashboard:   http://localhost:3000 (admin/admin123)"
echo "Prometheus:          http://localhost:9090"
echo ""
echo "PostgreSQL:"
echo "  Host: localhost"
echo "  Port: 5432"
echo "  User: aiops"
echo "  Password: aiops123"
echo "  Database: aiopsdb"
echo ""
echo "Redis:"
echo "  Host: localhost"
echo "  Port: 6379"

echo ""
echo "=== Next Steps ==="
echo "1. Start Backend API Server:"
echo "   cd backend"
echo "   cp config/config.yaml.example config/config.yaml"
echo "   ./bin/api-server"
echo ""
echo "2. Test health check:"
echo "   curl http://localhost:8080/health"

echo ""
echo "=== Stop Services ==="
echo "docker-compose down"
echo "docker-compose down -v  # (removes volumes)"

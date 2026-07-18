#!/bin/bash

# AiOpsHub Quick Start Demo Script
# Demonstrates the multi-agent collaboration workflow

set -e

echo "======================================"
echo "AiOpsHub - Multi-Agent Collaboration Demo"
echo "======================================"
echo ""

# Check dependencies
echo "1. Checking dependencies..."
if ! command -v docker &> /dev/null; then
    echo "Docker not installed. Please install Docker first."
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "Go not installed. Please install Go 1.22+ first."
    exit 1
fi

echo "Docker installed"
echo "Go installed"
echo ""

# Setup environment
echo "2. Setting up environment..."
echo "Config file: backend/configs/config.yaml"
echo ""
echo "Please edit backend/configs/config.yaml and update:"
echo "  llm.api_key: your-openai-api-key"
echo "  llm.base_url: https://api.openai.com/v1 (or your custom endpoint)"
echo ""
read -p "Press Enter after configuring config.yaml..."

echo "Environment configured"
echo ""

# Start dependencies
echo "3. Starting dependencies..."
echo "   Starting PostgreSQL and Redis..."

cd deployments
docker-compose up -d postgres redis
sleep 5

cd ..

echo "Dependencies started"
echo ""
echo "Available services:"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6379"
echo ""

# Build binaries
echo "4. Building backend binaries..."
cd backend
go build -o bin/api-server cmd/api-server/main.go
echo "Binaries built:"
echo "   - backend/bin/api-server"
echo ""

# Run tests
echo "5. Running unit tests..."
go test ./pkg/message_bus -v -short
go test ./pkg/state_sync -v -short
go test ./pkg/conflict_resolver -v -short
go test ./internal/agent -run TestDecisionEngine -v -short
echo "Tests passed"
echo ""

cd ..

# Start services
echo "6. Starting services..."
echo "   Starting API Server..."
cd backend
./bin/api-server &
API_PID=$!
sleep 3

cd ..

echo "Services running"
echo "   - API Server PID: $API_PID"
echo ""

# Demo scenario
echo "======================================"
echo "Demo: Multi-Agent Collaboration"
echo "======================================"
echo ""
echo "Scenario: Service Performance Analysis"
echo ""
echo "User Query: '订单服务响应很慢，帮我分析原因并给出解决方案'"
echo ""
echo "To test manually, use curl:"
echo ""
echo "curl -X POST http://localhost:8080/api/v1/workflows/collaboration \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{"
echo "    \"session_id\": \"demo-001\","
echo "    \"user_query\": \"订单服务响应很慢，帮我分析原因\","
echo "    \"context\": {\"service\": \"order-service\"}"
echo "  }'"
echo ""

read -p "Press Enter to stop services and clean up..."

# Cleanup
echo ""
echo "7. Cleaning up..."
kill $API_PID 2>/dev/null || true

echo "Stopping Docker services..."
cd deployments
docker-compose down

cd ..

echo "Demo completed!"
echo ""
echo "For more information, see:"
echo "  - README.md"
echo "  - docs/coordinator-agent-quick-start.md"
echo "  - PROGRESS.md"
echo ""

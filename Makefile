.PHONY: all build test run clean deploy help

all: build test

build:
	@echo "Building backend binaries..."
	cd backend && go build -o bin/api-server cmd/api-server/main.go
	cd backend && go build -o bin/temporal-worker cmd/temporal-worker/main.go
	@echo "Build complete!"

test:
	@echo "Running backend tests..."
	cd backend && go test ./... -v -short
	@echo "Tests complete!"

run-deps:
	@echo "Starting dependencies with Docker Compose..."
	cd deployments && docker-compose up -d postgres redis
	@sleep 5
	cd deployments && docker-compose up -d temporal-server
	@sleep 10
	@echo "Dependencies started!"
	@echo "Temporal UI: http://localhost:8080"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"

run-api:
	@echo "Starting API server..."
	cd backend && ./bin/api-server

run-worker:
	@echo "Starting Temporal worker..."
	cd backend && ./bin/temporal-worker

run-all: run-deps
	@echo "Starting all services..."
	@cd backend && ./bin/api-server &
	@cd backend && ./bin/temporal-worker &
	@echo "All services running!"

deploy-full:
	@echo "Deploying full stack with Docker Compose..."
	cd deployments && docker-compose up -d
	@sleep 15
	@echo "Full stack deployed!"
	@echo "Services available:"
	@echo "  - Temporal UI: http://localhost:8080"
	@echo "  - Grafana: http://localhost:3000"
	@echo "  - Prometheus: http://localhost:9090"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Redis: localhost:6379"
	@echo "  - Milvus: localhost:19530"

stop-deps:
	@echo "Stopping dependencies..."
	cd deployments && docker-compose down
	@echo "Dependencies stopped!"

clean-deps:
	@echo "Cleaning up dependencies and volumes..."
	cd deployments && docker-compose down -v
	@echo "Cleanup complete!"

clean:
	@echo "Cleaning binaries..."
	rm -f backend/bin/api-server backend/bin/temporal-worker
	@echo "Clean complete!"

install-deps:
	@echo "Installing Go dependencies..."
	cd backend && go mod download
	cd backend && go mod tidy
	@echo "Dependencies installed!"

help:
	@echo "AiOpsHub - 多Agent智能运维平台"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all          Build and test (default)"
	@echo "  build        Build backend binaries"
	@echo "  test         Run backend tests"
	@echo "  run-deps     Start dependencies (PostgreSQL, Redis, Temporal)"
	@echo "  run-api      Start API server"
	@echo "  run-worker   Start Temporal worker"
	@echo "  run-all      Start all services"
	@echo "  deploy-full  Deploy full stack with Docker Compose"
	@echo "  stop-deps    Stop Docker Compose dependencies"
	@echo "  clean-deps   Stop and remove Docker volumes"
	@echo "  clean        Remove built binaries"
	@echo "  install-deps Install Go dependencies"
	@echo "  help         Show this help message"
	@echo ""
	@echo "Configuration:"
	@echo "  Edit backend/configs/config.yaml to set LLM API key and other settings"
	@echo ""
	@echo "Quick Start:"
	@echo "  1. Edit backend/configs/config.yaml"
	@echo "  2. Set llm.api_key to your OpenAI API key"
	@echo "  3. make run-deps"
	@echo "  4. make build"
	@echo "  5. make run-api & make run-worker"

dev: install-deps build test run-deps
	@echo "Development environment ready!"
#!/bin/bash

# AiOpsHub 系统启动脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo "${RED}[ERROR]${NC} $1"
}

check_dependencies() {
    log_info "检查依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装"
        exit 1
    fi
    
    log_info "依赖检查完成"
}

check_config() {
    log_info "检查配置..."
    
    CONFIG_FILE="backend/configs/config.yaml"
    if [ ! -f "$CONFIG_FILE" ]; then
        log_error "配置文件不存在: $CONFIG_FILE"
        exit 1
    fi
    
    API_KEY=$(grep "api_key:" "$CONFIG_FILE" | head -1 | awk '{print $2}')
    if [ "$API_KEY" == "your-openai-api-key-here" ] || [ -z "$API_KEY" ]; then
        log_warn "请先在配置文件中设置 LLM API Key"
        log_warn "文件: $CONFIG_FILE"
        log_warn "字段: llm.api_key"
        read -p "按回车键继续..."
    fi
    
    log_info "配置检查完成"
}

start_dependencies() {
    log_info "启动依赖服务..."
    
    cd deployments
    
    docker-compose up -d postgres redis
    sleep 5
    
    cd "$PROJECT_DIR"
    
    log_info "依赖服务启动完成"
    log_info "  - PostgreSQL: localhost:5432"
    log_info "  - Redis: localhost:6379"
}

build_backend() {
    log_info "构建后端..."
    
    cd backend
    
    go mod download
    go mod tidy
    
    go build -o bin/api-server cmd/api-server/main.go
    
    cd "$PROJECT_DIR"
    
    log_info "后端构建完成"
    log_info "  - api-server: backend/bin/api-server"
}

start_services() {
    log_info "启动服务..."
    
    cd backend
    
    PID_DIR="$PROJECT_DIR/pids"
    mkdir -p "$PID_DIR"
    
    if [ -f "$PID_DIR/api-server.pid" ]; then
        OLD_PID=$(cat "$PID_DIR/api-server.pid")
        if kill -0 "$OLD_PID" 2>/dev/null; then
            log_warn "API Server 已在运行 (PID: $OLD_PID)"
        else
            rm -f "$PID_DIR/api-server.pid"
        fi
    fi
    
    if [ ! -f "$PID_DIR/api-server.pid" ]; then
        ./bin/api-server &
        echo $! > "$PID_DIR/api-server.pid"
        log_info "API Server 启动完成 (PID: $(cat $PID_DIR/api-server.pid))"
    fi
    
    cd "$PROJECT_DIR"
    
    log_info "所有服务已启动"
}

show_status() {
    log_info "系统状态:"
    echo ""
    
    PID_DIR="$PROJECT_DIR/pids"
    
    if [ -f "$PID_DIR/api-server.pid" ]; then
        PID=$(cat "$PID_DIR/api-server.pid")
        if kill -0 "$PID" 2>/dev/null; then
            echo "  API Server: ${GREEN}运行中${NC} (PID: $PID)"
        else
            echo "  API Server: ${RED}已停止${NC}"
        fi
    else
        echo "  API Server: ${YELLOW}未启动${NC}"
    fi
    
    echo ""
    
    cd deployments
    docker-compose ps 2>/dev/null || echo "  Docker服务: ${YELLOW}未启动${NC}"
    cd "$PROJECT_DIR"
}

stop_services() {
    log_info "停止服务..."
    
    PID_DIR="$PROJECT_DIR/pids"
    
    if [ -f "$PID_DIR/api-server.pid" ]; then
        PID=$(cat "$PID_DIR/api-server.pid")
        kill "$PID" 2>/dev/null || true
        rm -f "$PID_DIR/api-server.pid"
        log_info "API Server 已停止"
    fi
    
    log_info "所有服务已停止"
}

stop_all() {
    stop_services
    
    log_info "停止Docker服务..."
    cd deployments
    docker-compose down
    cd "$PROJECT_DIR"
    
    log_info "所有服务已停止"
}

case "$1" in
    start)
        check_dependencies
        check_config
        start_dependencies
        build_backend
        start_services
        show_status
        ;;
    stop)
        stop_all
        ;;
    restart)
        stop_services
        start_services
        show_status
        ;;
    status)
        show_status
        ;;
    build)
        build_backend
        ;;
    deps)
        check_dependencies
        start_dependencies
        ;;
    *)
        echo "AiOpsHub 系统管理脚本"
        echo ""
        echo "Usage: $0 {start|stop|restart|status|build|deps}"
        echo ""
        echo "Commands:"
        echo "  start   - 启动所有服务"
        echo "  stop    - 停止所有服务"
        echo "  restart - 重启服务"
        echo "  status  - 显示服务状态"
        echo "  build   - 构建后端"
        echo "  deps    - 启动依赖服务"
        ;;
esac

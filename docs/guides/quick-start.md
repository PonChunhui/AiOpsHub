# 开发环境快速启动指南

## 1. 启动基础服务（PostgreSQL + Temporal）

### 方式一：使用启动脚本

```bash
cd /Users/pengchunhui/code/aiops/AiOpsHub
./start-dev.sh

# 选择模式3（最小化启动）
# 仅启动 PostgreSQL + Temporal Server
```

### 方式二：手动启动

```bash
cd deployments
docker-compose up -d postgres temporal-server
```

### 验证服务状态

```bash
# 查看服务状态
docker-compose ps

# 查看Temporal日志
docker-compose logs -f temporal-server

# 访问Temporal Web UI
open http://localhost:8080
```

## 2. 启动Backend API Server

```bash
cd backend

# 复制配置文件
cp config/config.yaml.example config/config.yaml

# 启动API Server
./bin/api-server

# 或使用启动脚本
./start.sh
```

### 验证API Server

```bash
# 健康检查
curl http://localhost:8080/health

# 测试用户注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"test123"}'

# 测试用户登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test123"}'
```

## 3. 启动Temporal Worker

```bash
cd backend

# 启动Worker
./bin/temporal-worker
```

## 4. 完整开发环境

### 启动所有服务

```bash
cd deployments
docker-compose up -d
```

包含服务：
- PostgreSQL (5432)
- Temporal Server (7233, 8080)
- Redis (6379)
- Milvus (19530)
- ClickHouse (8123, 9000)
- Prometheus (9090)
- Grafana (3000)

### 访问地址

| 服务 | 地址 | 用户/密码 |
|------|------|----------|
| Temporal Web UI | http://localhost:8080 | - |
| Grafana | http://localhost:3000 | admin/admin123 |
| Prometheus | http://localhost:9090 | - |
| API Server | http://localhost:8080 | - |

## 5. 数据库连接

### PostgreSQL

```bash
# 连接数据库
psql -h localhost -U aiops -d aiopsdb

# 查看表结构
\dt

# 查看Agent表
SELECT * FROM agents;
```

### Redis

```bash
# 连接Redis
redis-cli -h localhost -p 6379

# 查看键
KEYS *
```

## 6. 测试API

### Agent管理

```bash
# 创建Agent
curl -X POST http://localhost:8080/api/v1/agents \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer token" \
  -d '{"name":"MonitorAgent","type":"monitor","description":"系统监控Agent"}'

# 获取Agent列表
curl http://localhost:8080/api/v1/agents \
  -H "Authorization: Bearer token"

# 获取单个Agent
curl http://localhost:8080/api/v1/agents/{id} \
  -H "Authorization: Bearer token"
```

### Workflow管理

```bash
# 创建Workflow
curl -X POST http://localhost:8080/api/v1/workflows \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer token" \
  -d '{"name":"AlertHandlingWorkflow","description":"告警处理流程"}'

# 执行Workflow
curl -X POST http://localhost:8080/api/v1/workflows/{id}/execute \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer token" \
  -d '{"input":"alert_data"}'
```

### 告警管理

```bash
# 创建告警
curl -X POST http://localhost:8080/api/v1/alerts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer token" \
  -d '{"source":"prometheus","severity":"high","title":"CPU使用率过高"}'

# Webhook接收告警
curl -X POST http://localhost:8080/api/v1/alerts/webhook \
  -H "Content-Type: application/json" \
  -d '{"source":"prometheus","severity":"high","title":"告警Webhook"}'
```

## 7. 常见问题

### PostgreSQL连接失败

```bash
# 检查容器状态
docker-compose ps postgres

# 查看日志
docker-compose logs postgres

# 重启容器
docker-compose restart postgres
```

### Temporal Server未启动

```bash
# Temporal需要约30秒初始化
docker-compose logs -f temporal-server

# 等待看到 "Temporal server started" 日志
```

### 数据库表不存在

API Server启动时会自动创建表（AutoMigrate）。如果表不存在，检查：
1. PostgreSQL是否正常运行
2. database配置是否正确
3. API Server启动日志

## 8. 停止服务

```bash
# 停止所有服务
docker-compose down

# 停止并删除数据
docker-compose down -v

# 仅停止特定服务
docker-compose stop postgres temporal-server
```
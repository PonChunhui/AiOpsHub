# 部署指南

## 部署方式

AiOpsHub支持多种部署方式：
- 本地开发部署
- Docker Compose部署
- Kubernetes生产部署

## 前置要求

### 软件要求

- Go 1.24+
- Node.js 18+
- PostgreSQL 14+
- Redis 6+（支持Cluster模式）
- Temporal Server 1.20+

### 硬件要求

**开发环境**:
- CPU: 2核+
- 内存: 4GB+
- 磁盘: 10GB+

**生产环境**:
- CPU: 4核+
- 内存: 8GB+
- 磁盘: 50GB+

## 本地开发部署

### 1. 克隆项目

```bash
git clone https://github.com/your-org/AiOpsHub.git
cd AiOpsHub
```

### 2. 配置后端

编辑 `backend/config/config.yaml`:

```yaml
database:
  host: "localhost"
  port: 5432
  user: "aiops"
  password: "aiops123"
  dbname: "aiopsdb"

redis:
  cluster_mode: false  # 开发环境使用单机模式
  host: "localhost"
  port: 6379
  password: ""
  db: 0

temporal:
  host: "localhost"
  port: 7233
  namespace: "default"
  task_queue: "aiops-task-queue"

jwt:
  secret: "dev-secret-key-change-in-production"
  token_expire: 30m

llm:
  provider: "aliyun_bailian"
  model: "qwen-turbo"
  api_key: "your-api-key"
```

### 3. 启动数据库

```bash
# PostgreSQL
docker run -d \
  --name aiops-postgres \
  -e POSTGRES_USER=aiops \
  -e POSTGRES_PASSWORD=aiops123 \
  -e POSTGRES_DB=aiopsdb \
  -p 5432:5432 \
  postgres:14

# Redis（单机模式）
docker run -d \
  --name aiops-redis \
  -p 6379:6379 \
  redis:6
```

### 4. 启动Temporal Server

```bash
# 使用Temporal CLI
temporal server start-dev

# 或使用Docker
docker run -d \
  --name temporal \
  -p 7233:7233 \
  -p 8080:8080 \
  temporalio/auto-setup:latest
```

### 5. 启动后端

```bash
cd backend

# 编译API Server
go build -o bin/api-server ./cmd/api-server
./bin/api-server

# 编译Temporal Worker
go build -o bin/temporal-worker ./cmd/temporal-worker
./bin/temporal-worker
```

### 6. 启动前端

```bash
cd frontend
npm install
npm run dev
```

### 7. 访问应用

- 前端: http://localhost:5173
- 后端API: http://localhost:8080
- Temporal UI: http://localhost:8080

## Docker Compose部署

### 1. 创建docker-compose.yml

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_USER: aiops
      POSTGRES_PASSWORD: aiops123
      POSTGRES_DB: aiopsdb
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U aiops"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:6
    command: redis-server --requirepass your_redis_password
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  temporal:
    image: temporalio/auto-setup:latest
    ports:
      - "7233:7233"
      - "8080:8080"
    environment:
      - DB=postgres12
      - DB_PORT=5432
      - POSTGRES_USER=aiops
      - POSTGRES_PWD=aiops123
      - POSTGRES_DB=temporal
      - POSTGRES_HOST=postgres
    depends_on:
      postgres:
        condition: service_healthy

  api-server:
    build:
      context: ./backend
      dockerfile: Dockerfile.api-server
    ports:
      - "8080:8080"
    environment:
      - DATABASE_HOST=postgres
      - DATABASE_PORT=5432
      - DATABASE_USER=aiops
      - DATABASE_PASSWORD=aiops123
      - DATABASE_NAME=aiopsdb
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=your_redis_password
      - TEMPORAL_HOST=temporal
      - TEMPORAL_PORT=7233
      - JWT_SECRET=your-production-secret-key
      - LLM_API_KEY=your-aliyun-api-key
    depends_on:
      - postgres
      - redis
      - temporal

  temporal-worker:
    build:
      context: ./backend
      dockerfile: Dockerfile.temporal-worker
    environment:
      - DATABASE_HOST=postgres
      - TEMPORAL_HOST=temporal
      - TEMPORAL_PORT=7233
      - LLM_API_KEY=your-aliyun-api-key
    depends_on:
      - postgres
      - temporal

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "80:80"
    environment:
      - VITE_API_BASE_URL=http://api-server:8080
    depends_on:
      - api-server

volumes:
  postgres_data:
  redis_data:
```

### 2. 创建Dockerfile.api-server

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o bin/api-server ./cmd/api-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/api-server .
COPY --from=builder /app/config ./config
EXPOSE 8080
CMD ["./api-server"]
```

### 3. 创建Dockerfile.temporal-worker

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o bin/temporal-worker ./cmd/temporal-worker

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/temporal-worker .
COPY --from=builder /app/config ./config
CMD ["./temporal-worker"]
```

### 4. 启动服务

```bash
docker-compose up -d

# 查看日志
docker-compose logs -f api-server
docker-compose logs -f temporal-worker
```

### 5. 停止服务

```bash
docker-compose down

# 清理数据
docker-compose down -v
```

## Kubernetes部署

### 1. 创建Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: aiops
```

### 2. PostgreSQL部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: aiops
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:14
        env:
        - name: POSTGRES_USER
          value: "aiops"
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: aiops-secrets
              key: postgres-password
        - name: POSTGRES_DB
          value: "aiopsdb"
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
```

### 3. Redis部署（Cluster模式）

使用Redis Operator或Helm Chart部署Redis Cluster:

```bash
helm repo add redis-ha https://dandydeveloper.github.io/redis-ha/
helm install redis-cluster redis-ha/redis-ha \
  --namespace aiops \
  --set auth=true \
  --set password=your_redis_password
```

### 4. Temporal部署

使用Temporal Helm Chart:

```bash
helm repo add temporal https://temporalio.github.io/helm-charts
helm install temporal temporal/temporal \
  --namespace aiops \
  --set server.replicaCount=1
```

### 5. API Server部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
  namespace: aiops
spec:
  replicas: 2
  selector:
    matchLabels:
      app: api-server
  template:
    metadata:
      labels:
        app: api-server
    spec:
      containers:
      - name: api-server
        image: aiops/api-server:v1.0
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_HOST
          value: "postgres"
        - name: TEMPORAL_HOST
          value: "temporal"
        - name: LLM_API_KEY
          valueFrom:
            secretKeyRef:
              name: aiops-secrets
              key: llm-api-key
        resources:
          limits:
            cpu: "1"
            memory: "1Gi"
          requests:
            cpu: "0.5"
            memory: "512Mi"
```

### 6. Temporal Worker部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: temporal-worker
  namespace: aiops
spec:
  replicas: 2
  selector:
    matchLabels:
      app: temporal-worker
  template:
    metadata:
      labels:
        app: temporal-worker
    spec:
      containers:
      - name: temporal-worker
        image: aiops/temporal-worker:v1.0
        env:
        - name: TEMPORAL_HOST
          value: "temporal"
        resources:
          limits:
            cpu: "2"
            memory: "2Gi"
```

### 7. Service配置

```yaml
apiVersion: v1
kind: Service
metadata:
  name: api-server
  namespace: aiops
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: api-server
```

### 8. 部署命令

```bash
# 创建Secret
kubectl create secret generic aiops-secrets \
  --namespace aiops \
  --from-literal=postgres-password=aiops123 \
  --from-literal=llm-api-key=your-api-key \
  --from-literal=jwt-secret=your-secret

# 部署所有组件
kubectl apply -f namespace.yaml
kubectl apply -f postgres.yaml
kubectl apply -f redis.yaml
kubectl apply -f temporal.yaml
kubectl apply -f api-server.yaml
kubectl apply -f temporal-worker.yaml
kubectl apply -f service.yaml

# 查看状态
kubectl get pods -n aiops
kubectl logs -f deployment/api-server -n aiops
```

## 生产环境配置

### 1. 安全配置

**JWT密钥**:
```yaml
jwt:
  secret: "at-least-32-characters-random-string"
  token_expire: 30m
```

**Redis密码**:
```yaml
redis:
  password: "strong_redis_password"
```

**数据库密码**:
使用Kubernetes Secret管理。

### 2. 性能配置

**API Server资源限制**:
```yaml
resources:
  limits:
    cpu: "2"
    memory: "2Gi"
  requests:
    cpu: "1"
    memory: "1Gi"
```

**Temporal Worker资源**:
```yaml
resources:
  limits:
    cpu: "4"
    memory: "4Gi"
  requests:
    cpu: "2"
    memory: "2Gi"
```

### 3. 高可用配置

**API Server多副本**:
```yaml
replicas: 3
```

**Temporal Worker多副本**:
```yaml
replicas: 3
```

**Redis Cluster**:
6节点集群（3主3从）。

**PostgreSQL主从**:
使用PostgreSQL Operator或Cloud服务。

### 4. 监控配置

**Prometheus监控**:
```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
```

**日志收集**:
使用ELK或Loki收集日志。

### 5. 备份策略

**数据库备份**:
```bash
# 每日备份
pg_dump aiopsdb > backup.sql

# 定时任务
0 2 * * * pg_dump aiopsdb > /backup/aiopsdb_$(date +\%Y\%m\%d).sql
```

## 故障排查

### 1. API Server无法启动

检查：
- PostgreSQL连接
- Redis连接
- Temporal连接
- 配置文件路径

### 2. Temporal Worker报错

检查：
- Temporal Server状态
- Task Queue配置
- Activity注册

### 3. 前端无法连接后端

检查：
- CORS配置
- API代理配置
- Token有效性

### 4. Workflow执行失败

检查：
- Temporal Web UI日志
- Agent执行日志
- LLM API调用

## 运维建议

1. **定期检查日志**
2. **监控资源使用**
3. **及时更新版本**
4. **备份重要数据**
5. **安全漏洞扫描**
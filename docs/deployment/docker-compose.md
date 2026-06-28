# 快速部署指南

## 一、开发环境快速启动（Docker Compose）

### 前置要求

- Docker Desktop（已安装并启动）
- Docker Compose v2+

### 快速启动步骤

```bash
# 1. 进入部署目录
cd deployments

# 2. 启动所有服务
docker-compose up -d

# 3. 查看服务状态
docker-compose ps

# 4. 查看Temporal Server日志（等待初始化完成）
docker-compose logs -f temporal-server

# 5. 访问服务
# Temporal Web UI: http://localhost:8080
# Grafana: http://localhost:3000 (admin/admin123)
```

### 服务清单

| 服务 | 端口 | 说明 |
|------|------|------|
| Temporal Server | 7233 (RPC), 8080 (UI) | 工作流引擎 |
| PostgreSQL | 5432 | 业务数据 + Temporal数据 |
| Redis | 6379 | Agent状态和缓存 |
| Milvus | 19530 | 向量数据库 |
| ClickHouse | 8123, 9000 | 时序数据 |
| Prometheus | 9090 | 监控 |
| Grafana | 3000 | 可视化 |

### PostgreSQL数据库配置

**合并架构**：Temporal和业务数据使用同一PostgreSQL实例

```bash
# 连接信息
Host: localhost
Port: 5432
User: aiops
Password: aiops123

# 数据库
- aiopsdb（业务数据）
- temporal（Temporal自动创建）
```

### 初始化业务数据库

```bash
# PostgreSQL已通过docker-compose自动执行init-db.sql
# 验证数据库初始化
docker exec -it aiops-postgres psql -U aiops -d aiopsdb -c "\dt"

# 预期输出：显示所有业务表
```

### 常用操作命令

```bash
# 启动特定服务
docker-compose up -d postgres temporal-server redis

# 停止服务
docker-compose down

# 查看日志
docker-compose logs -f temporal-server

# 清理所有数据（慎用）
docker-compose down -v

# 重启服务
docker-compose restart temporal-server
```

## 二、仅启动Temporal开发环境

最小化启动（仅Temporal和数据库）：

```bash
docker-compose up -d postgres temporal-server

# 等待Temporal初始化完成
sleep 30

# 访问Temporal Web UI
open http://localhost:8080
```

## 三、生产环境部署建议

### 1. PostgreSQL分离（生产推荐）

生产环境建议Temporal和业务数据使用独立PostgreSQL：

```yaml
# 生产环境docker-compose配置
services:
  temporal-postgres:
    # Temporal专用数据库
    
  business-postgres:
    # 业务数据专用数据库
```

### 2. 资源限制配置

```yaml
services:
  temporal-server:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 2G
```

### 3. 增加备份策略

```bash
# PostgreSQL备份
docker exec aiops-postgres pg_dump -U aiops aiopsdb > backup.sql

# Temporal数据库备份
docker exec temporal-postgres pg_dump -U temporal temporal > temporal_backup.sql
```

### 4. 启用认证和TLS

参考Temporal官方文档：https://docs.temporal.io/security

## 四、故障排查

### Temporal Server启动失败

```bash
# 检查Temporal日志
docker-compose logs temporal-server

# 检查PostgreSQL连接
docker exec aiops-postgres psql -U aiops -c "SELECT 1"

# 重启Temporal
docker-compose restart temporal-server
```

### PostgreSQL连接问题

```bash
# 检查PostgreSQL状态
docker-compose ps postgres

# 检查端口占用
lsof -i :5432

# 进入PostgreSQL
docker exec -it aiops-postgres psql -U aiops -d aiopsdb
```

### Milvus启动失败

```bash
# 检查etcd和minio状态
docker-compose ps etcd minio

# 检查Milvus日志
docker-compose logs milvus-standalone
```

## 五、数据清理

### 仅清理Temporal数据

```bash
# 进入PostgreSQL
docker exec -it aiops-postgres psql -U aiops

# 删除Temporal数据库
DROP DATABASE temporal;

# 重启Temporal Server（会自动重建）
docker-compose restart temporal-server
```

### 完全清理所有数据

```bash
# 停止并删除所有容器和数据卷
docker-compose down -v

# 删除本地镜像（可选）
docker rmi temporalio/server postgres:15 redis:7-alpine
```

## 六、监控和可视化

### Grafana配置

1. 访问 http://localhost:3000
2. 登录：admin / admin123
3. 添加Prometheus数据源
4. 导入Temporal Dashboard（ID: 11647）

### Temporal Web UI功能

- 查看Workflow执行历史
- 查看Event History详情
- 发送Signal和Query
- 搜索Workflow
- 查看Activity执行结果

## 七、开发环境使用建议

### 推荐启动顺序

```bash
# 第1步：启动基础数据库
docker-compose up -d postgres redis

# 第2步：启动Temporal
docker-compose up -d temporal-server

# 第3步：启动向量数据库
docker-compose up -d etcd minio milvus-standalone

# 第4步：启动监控
docker-compose up -d prometheus grafana
```

### 资源占用估算

| 服务 | 内存 | CPU |
|------|------|-----|
| Temporal Server | 500MB | 0.5 |
| PostgreSQL | 100MB | 0.2 |
| Redis | 50MB | 0.1 |
| Milvus | 800MB | 0.8 |
| ClickHouse | 200MB | 0.3 |
| Prometheus | 100MB | 0.2 |
| Grafana | 100MB | 0.2 |
| **总计** | **~1.85GB** | **~2.3 cores** |

## 八、参考资料

- [Temporal官方文档](https://docs.temporal.io/)
- [Temporal部署指南](https://docs.temporal.io/server-options)
- [Milvus文档](https://milvus.io/docs/)
- [Grafana配置](https://grafana.com/docs/)
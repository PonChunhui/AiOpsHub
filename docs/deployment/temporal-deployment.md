# Temporal部署指南

## 文档信息

| 项目 | 内容 |
|------|------|
| 文档名称 | Temporal部署指南 |
| 版本 | v1.0.0 |
| 创建日期 | 2026-06-24 |

## 一、前置要求

### 1.1 软件要求

- **Docker Desktop**（macOS/Windows）
- **Docker Engine**（Linux）
- **Docker Compose**（v2.0+）

### 1.2 系统要求

- macOS 10.15+（Catalina或更高）
- Windows 10/11（WSL2）
- Linux（Ubuntu 20.04+ / CentOS 8+）

### 1.3 硬件要求

- CPU：2核心以上
- 内存：4GB以上（推荐8GB）
- 磁盘：10GB可用空间

---

## 二、Docker Desktop安装（macOS）

### 2.1 安装步骤

**方式1：官网下载**
```
1. 访问：https://www.docker.com/products/docker-desktop
2. 点击 "Download for Mac"
3. 选择 "Apple Silicon"（M1/M2/M3芯片）或 "Intel Chip"
4. 下载Docker.dmg
5. 双击安装
6. 启动Docker Desktop
```

**方式2：Homebrew安装**
```bash
# 安装Homebrew（如果未安装）
# 已安装可跳过

# 安装Docker Desktop
brew install --cask docker

# 启动Docker Desktop
open /Applications/Docker.app
```

### 2.2 验证安装

```bash
# 检查Docker版本
docker --version
# 输出：Docker version 24.0.5, build ced0996

# 检查Docker Compose版本
docker compose version
# 输出：Docker Compose version v2.20.2

# 测试Docker运行
docker run hello-world
```

### 2.3 Docker Desktop配置

**推荐配置**：
```
Settings → Resources：
- CPUs: 4
- Memory: 8GB
- Swap: 2GB
- Disk image size: 64GB
```

---

## 三、Temporal Server部署

### 3.1 使用Docker Compose部署

**启动Temporal**：
```bash
cd /Users/pengchunhui/code/aiops/AiOpsHub/deployments

# 启动Temporal Server
docker compose up -d temporal-server temporal-postgres

# 等待启动（约30秒）
sleep 30

# 检查容器状态
docker compose ps
```

**预期输出**：
```
NAME                 STATUS    PORTS
temporal-server      running   7233/tcp, 8080/tcp
temporal-postgres    running   5432/tcp
```

### 3.2 检查Temporal日志

```bash
# 查看Temporal启动日志
docker compose logs temporal-server

# 预期看到：
# "Temporal server started"
# "Listening on: 7233"
```

### 3.3 访问Temporal Web UI

**URL**：http://localhost:8080

**功能**：
- 查看Workflow执行历史
- 查看Event History详情
- 查看Activity执行结果
- 发送Signal和Query
- 调试Workflow

---

## 四、启动所有服务（可选）

**启动完整开发环境**：
```bash
cd deployments

# 启动所有服务
docker compose up -d

# 包含服务：
# - temporal-server (Temporal工作流引擎)
# - temporal-postgres (Temporal持久化)
# - postgres (业务数据库)
# - redis (缓存和状态)
# - milvus-standalone (向量数据库)
# - etcd, minio (Milvus依赖)
# - clickhouse (时序数据库)
# - prometheus (监控)
# - grafana (可视化)
```

**检查所有服务状态**：
```bash
docker compose ps

# 预期10个服务全部running
```

---

## 五、Temporal Server架构

### 5.1 Temporal Server组件

```
┌────────────────────────────────────┐
│      Temporal Server               │
│                                    │
│  ┌──────────────┐  ┌─────────────┐│
│  │ Frontend     │  │ History     ││
│  │ Service      │  │ Service     ││
│  │ (API入口)    │  │ (Event存储) ││
│  └──────────────┘  └─────────────┘│
│                                    │
│  ┌──────────────┐  ┌─────────────┐│
│  │ Matching     │  │ Worker      ││
│  │ Service      │  │ Service     ││
│  │ (任务匹配)   │  │ (系统任务)  ││
│  └──────────────┘  └─────────────┘│
│                                    │
│  ┌──────────────────────────────┐│
│  │     PostgreSQL持久化         ││
│  │   (Event History存储)        ││
│  └──────────────────────────────┘│
└────────────────────────────────────┘
```

### 5.2 Temporal端口说明

| 端口 | 服务 | 说明 |
|------|------|------|
| **7233** | Temporal Server | Worker连接端口 |
| **8080** | Temporal Web UI | Web界面 |
| **7234** | Frontend Service | API入口 |
| **7235** | History Service | Event历史 |
| **7236** | Matching Service | 任务匹配 |

---

## 六、Temporal Web UI使用指南

### 6.1 界面导航

**首页**：
- Workflows：查看所有Workflow执行历史
- Schedules：定时任务管理
- Search：搜索Workflow

**Workflow详情页**：
- **Summary**：Workflow概览
  - Workflow ID
  - Run ID
  - Status（Running/Completed/Failed）
  - Duration
  
- **History**：Event History详情
  - 查看每个事件的执行细节
  - 查看Activity输入输出
  - 查看执行时间
  
- **JSON**：原始JSON数据
  
- **Queries**：发送Query
  
- **Signals**：发送Signal（人机交互）

### 6.2 查看Workflow示例

**前提**：需要有Workflow执行记录

**步骤**：
1. 访问http://localhost:8080
2. 点击左侧"Workflows"
3. 点击某个Workflow查看详情
4. 点击"History"查看Event History
5. 点击某个Event查看详细信息

### 6.3 发送Signal示例

**场景**：等待用户确认的Workflow

**步骤**：
1. 找到运行中的Workflow（Status: Running）
2. 点击"Signals"标签
3. 输入Signal名称：`approval`
4. 输入Signal值：`{"approved": true}`
5. 点击"Send"
6. Workflow继续执行

---

## 七、开发环境配置

### 7.1 Temporal客户端连接

**Go客户端连接**：
```go
package main

import (
    "go.temporal.io/sdk/client"
)

func main() {
    // 连接Temporal Server
    c, err := client.Dial(client.Options{
        HostPort:  "localhost:7233",
        Namespace: "default",
    })
    if err != nil {
        panic(err)
    }
    defer c.Close()
    
    // 使用客户端
    // ...
}
```

### 7.2 Temporal Worker启动

**Go Worker启动**：
```go
package main

import (
    "go.temporal.io/sdk/worker"
)

func main() {
    // 连接Temporal Server
    c, _ := client.Dial(client.Options{
        HostPort: "localhost:7233",
    })
    
    // 创建Worker
    w := worker.New(c, "aiops-task-queue", worker.Options{})
    
    // 注册Workflow和Activity
    w.RegisterWorkflow(IncidentHandlingWorkflow)
    w.RegisterActivity(MonitorAgentActivity)
    
    // 启动Worker
    w.Run(worker.InterruptCh())
}
```

---

## 八、常见问题排查

### 8.1 Temporal Server启动失败

**症状**：容器状态为Exit

**排查步骤**：
```bash
# 1. 查看日志
docker compose logs temporal-server

# 2. 检查PostgreSQL连接
docker compose logs temporal-postgres

# 3. 重启服务
docker compose restart temporal-server

# 4. 检查端口占用
lsof -i:7233
lsof -i:8080
```

### 8.2 Temporal Web UI无法访问

**症状**：浏览器无法打开http://localhost:8080

**排查步骤**：
```bash
# 1. 检查容器运行状态
docker compose ps temporal-server

# 2. 检查端口映射
docker compose port temporal-server 8080

# 3. 等待启动完成（Temporal启动较慢）
sleep 60
docker compose ps

# 4. 检查防火墙设置
# macOS通常无防火墙问题
```

### 8.3 Worker连接失败

**症状**：Go Worker无法连接Temporal Server

**排查步骤**：
```bash
# 1. 确认Temporal Server运行
docker compose ps temporal-server

# 2. 测试端口连通性
nc -zv localhost 7233

# 3. 检查Namespace配置
# Temporal默认Namespace: default

# 4. 查看Temporal Server日志
docker compose logs temporal-server | grep -i error
```

---

## 九、停止和清理

### 9.1 停止Temporal服务

```bash
cd deployments

# 停止所有服务
docker compose down

# 仅停止Temporal
docker compose stop temporal-server temporal-postgres
```

### 9.2 清理数据（重新初始化）

```bash
# 停止并删除容器、网络、卷
docker compose down -v

# 删除Temporal PostgreSQL数据
docker volume rm deployments_temporal-postgres-data
```

### 9.3 完全清理（慎用）

```bash
# 停止所有服务并删除所有数据
docker compose down -v --remove-orphans

# 清理所有未使用的Docker资源
docker system prune -a --volumes
```

---

## 十、生产环境部署（Kubernetes）

### 10.1 Temporal Helm Chart

**使用Helm部署Temporal**：
```bash
# 添加Temporal Helm仓库
helm repo add temporal https://helm.temporal.io

# 更新仓库
helm repo update

# 部署Temporal（开发环境）
helm install temporal temporal/temporal \
  --set server.replicaCount=1 \
  --set postgresql.enabled=true

# 部署Temporal（生产环境）
helm install temporal temporal/temporal \
  --set server.replicaCount=3 \
  --set postgresql.enabled=true \
  --set postgresql.postgresqlReplicaCount=2
```

### 10.2 Temporal Kubernetes配置

**yaml配置**：
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: temporal-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: temporal-server
  template:
    spec:
      containers:
      - name: temporal-server
        image: temporalio/server:latest
        ports:
        - containerPort: 7233
        env:
        - name: DB
          value: "postgresql"
        - name: POSTGRES_SEEDS
          value: "temporal-postgres"
```

---

## 十一、下一步操作

### 11.1 安装Docker Desktop后

**验证Temporal部署**：
```bash
cd deployments
docker compose up -d temporal-server temporal-postgres

# 等待启动完成
sleep 60

# 检查状态
docker compose ps

# 访问Web UI
open http://localhost:8080
```

### 11.2 体验Temporal Workflow

**推荐步骤**：
1. 先部署Temporal Server
2. 访问Temporal Web UI了解界面
3. 运行Temporal官方示例：
   ```bash
   git clone https://github.com/temporalio/samples-go
   cd samples-go/hello-world
   
   # 启动Worker
   go run worker/main.go
   
   # 启动Workflow
   go run starter/main.go
   
   # 在Web UI查看Workflow执行
   ```
4. 体验完成后，继续AiOpsHub开发

---

## 十二、参考资源

- [Temporal官方文档](https://docs.temporal.io/)
- [Temporal部署文档](https://docs.temporal.io/server/quick-install)
- [Temporal Docker Hub](https://hub.docker.com/r/temporalio/server)
- [Temporal Helm Chart](https://github.com/temporalio/helm-charts)
- [Temporal GitHub](https://github.com/temporalio/temporal)

---

## 十三、更新记录

| 版本 | 日期 | 更新内容 |
|------|------|----------|
| v1.0.0 | 2026-06-24 | 初稿（包含Docker Desktop安装指南） |
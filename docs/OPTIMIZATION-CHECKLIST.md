# AiOpsHub 优化执行清单

**快速执行指南** - 按优先级完成关键优化任务

---

## 第一周任务（P0 - 立即执行）

### 安全加固（必须完成）

#### 1. 配置文件敏感信息移除

**执行步骤**：
```bash
# 1. 创建配置文件模板
cd backend/configs
cp config.yaml config.yaml.example

# 2. 移除敏感信息，替换为环境变量
vim config.yaml.example
# 修改:
#   database.password: ${DATABASE_PASSWORD}
#   redis.password: ${REDIS_PASSWORD}
#   llm.api_key: ${LLM_API_KEY}
#   jwt.secret: ${JWT_SECRET_KEY}

# 3. 添加真实配置到.gitignore
echo "backend/configs/config.yaml" >> .gitignore
git add .gitignore

# 4. 提交模板文件
git add backend/configs/config.yaml.example
git commit -m "Add config template without sensitive data"

# 5. 创建环境变量文件（不提交）
vim backend/.env
# 内容:
DATABASE_PASSWORD=aiops123
REDIS_PASSWORD=1qaz!QAZ
LLM_API_KEY=your_api_key_here
JWT_SECRET_KEY=your_strong_secret_here

# 6. 添加.env到.gitignore
echo "backend/.env" >> .gitignore
```

**验证**：
- 检查 git status，确认 config.yaml 不会被提交
- 检查 config.yaml.example 无敏感信息

**工作量**: 1小时

---

#### 2. JWT密钥强化

**执行步骤**：
```bash
# 1. 生成强密钥（256位以上）
cd backend/scripts
go run generate_jwt.go
# 输出: Generated JWT Secret: <strong-secret>

# 2. 更新.env文件
vim ../.env
# JWT_SECRET_KEY=<strong-secret>

# 3. 实现密钥轮换机制
vim ../pkg/jwt/jwt.go
# 添加 KeyManager 和轮换逻辑

# 4. 添加密钥轮换定时任务
vim ../internal/service/jwt_rotation_service.go
```

**验证**：
- 密钥长度 ≥ 256位
- 轮换机制测试通过

**工作量**: 4小时

---

#### 3. Docker配置修复

**执行步骤**：
```bash
# 1. 修改docker-compose.yml使用环境变量
cd deployments
vim docker-compose.yml

# 修改:
#   POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-changeme}
#   GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD:-admin123}

# 2. 创建.env文件
vim .env
# POSTGRES_PASSWORD=aiops123
# GRAFANA_PASSWORD=admin123

# 3. 添加到.gitignore
echo "deployments/.env" >> ../.gitignore

# 4. 创建docker-compose.yml.example
cp docker-compose.yml docker-compose.yml.example
# 移除实际密码
```

**验证**：
- docker-compose.yml 使用环境变量
- .env 文件不在 git 仓库中

**工作量**: 1小时

---

#### 4. CORS配置优化

**执行步骤**：
```bash
# 1. 修改CORS中间件
vim backend/internal/middleware/middleware.go

# 添加:
#   allowedOrigins := viper.GetStringSlice("cors.allowed_origins")
#   检查请求来源是否在允许列表

# 2. 添加CORS配置到config.yaml.example
vim backend/configs/config.yaml.example
# 添加:
# cors:
#   allowed_origins:
#     - "https://aiops.example.com"
#     - "http://localhost:5173"
```

**验证**：
- CORS仅允许配置的域名
- 未配置域名返回403

**工作量**: 2小时

---

#### 5. 前端Token存储安全

**执行步骤**：
```bash
# 1. 后端使用HttpOnly Cookie
vim backend/internal/handler/handler.go
# Login方法修改:
#   c.SetCookie("auth_token", token, 86400, "/", "", true, true)

# 2. 前端移除localStorage存储
vim frontend/src/stores/auth.ts
# 移除:
#   localStorage.setItem('auth_token', token)

# 3. 前端自动携带Cookie
# 无需手动处理，浏览器自动携带HttpOnly Cookie
```

**验证**：
- Token在Cookie中，不在localStorage
- Cookie属性: HttpOnly=true, Secure=true

**工作量**: 4小时

---

### CI/CD流程建立

#### 6. GitHub Actions配置

**执行步骤**：
```bash
# 1. 创建CI配置
mkdir -p .github/workflows
vim .github/workflows/ci.yml

# 内容见: docs/OPTIMIZATION-RECOMMENDATIONS.md > 五、CI/CD和部署优化

# 2. 测试CI流程
git add .github/workflows/ci.yml
git commit -m "Add CI workflow"
git push

# 3. 检查GitHub Actions运行状态
# https://github.com/your-org/AiOpsHub/actions
```

**验证**：
- CI自动运行测试
- Lint检查通过
- 安全扫描通过

**工作量**: 4小时

---

## 本月任务（P1 - 本周开始）

### 测试覆盖率提升

#### 7. Handler层单元测试

**执行步骤**：
```bash
# 1. 为关键Handler创建测试文件
cd backend/internal/handler

# Agent Handler测试
vim agent_handler_test.go
# 内容见: docs/OPTIMIZATION-RECOMMENDATIONS.md > 问题6

# Tool Handler测试
vim tool_handler_test.go

# Chat Handler测试
vim chat_handler_test.go

# 2. 运行测试
go test ./... -v -cover

# 3. 查看覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**验证**：
- Handler层覆盖率 ≥ 70%
- 关键Handler测试通过

**工作量**: 20小时（分批次）

---

#### 8. Service层单元测试

**执行步骤**：
```bash
# 1. 为关键Service创建测试文件
cd backend/internal/service

# Agent Service测试
vim agent_service_test.go

# Tool Service测试
vim tool_service_test.go

# Chat Service测试
vim chat_service_test.go

# 2. 运行测试
go test ./... -v -cover
```

**验证**：
- Service层覆盖率 ≥ 80%
- Mock对象隔离依赖

**工作量**: 20小时（分批次）

---

#### 9. 前端组件测试

**执行步骤**：
```bash
# 1. 安装测试框架
cd frontend
npm install -D vitest @vue/test-utils @vitest/coverage-v8

# 2. 创建vitest配置
vim vitest.config.ts

# 3. 为关键组件创建测试
vim src/components/chat/__tests__/MessageList.test.ts
vim src/views/__tests__/AIAssistant.test.ts

# 4. 运行测试
npm run test

# 5. 查看覆盖率
npm run test -- --coverage
```

**验证**：
- 关键组件测试覆盖率 ≥ 60%
- Store测试覆盖率 ≥ 90%

**工作量**: 30小时（分批次）

---

## 本月任务（P2 - 本月完成）

### 代码质量改进

#### 10. 统一错误处理

**执行步骤**：
```bash
# 1. 定义统一错误响应格式
vim backend/internal/model/response.go

# 2. 创建错误处理中间件
vim backend/internal/middleware/error_handler.go

# 3. Handler使用统一格式
vim backend/internal/handler/base_handler.go
# 统一 Success 和 Error 方法
```

**验证**：
- 所有Handler返回统一格式
- 前端可统一处理错误

**工作量**: 6小时

---

#### 11. 数据库查询优化

**执行步骤**：
```bash
# 1. 实现游标分页
vim backend/internal/repository/agent_repo.go
# 修改 List 方法使用游标分页

# 2. 添加数据库索引
vim backend/migrations/add_indexes.sql
# 内容:
# CREATE INDEX idx_agent_type ON agents(type);
# CREATE INDEX idx_agent_enabled ON agents(enabled);
# CREATE INDEX idx_chat_session_created ON chat_messages(session_id, created_at);

# 3. 执行索引迁移
psql -h localhost -U aiops -d aiopsdb -f migrations/add_indexes.sql
```

**验证**：
- 大数据量查询性能提升
- 索引创建成功

**工作量**: 6小时

---

#### 12. 缓存策略实现

**执行步骤**：
```bash
# 1. 实现缓存Service
vim backend/internal/service/cache_service.go

# 2. Repository集成缓存
vim backend/internal/repository/cached_agent_repo.go

# 3. 配置缓存TTL
vim backend/configs/config.yaml.example
# 添加:
# cache:
#   ttl: 300s  # 5分钟
#   enabled: true
```

**验证**：
- 高频查询使用缓存
- 缓存命中率监控

**工作量**: 8小时

---

## 每日检查清单

### 开发人员

- [ ] 代码提交前运行测试: `go test ./...`
- [ ] 代码提交前运行lint: `golangci-lint run`
- [ ] 无敏感信息提交（检查 git diff）
- [ ] 添加必要的单元测试
- [ ] 更新相关文档

### 团队负责人

- [ ] 检查CI/CD运行状态
- [ ] 检查测试覆盖率趋势
- [ ] 检查安全扫描结果
- [ ] 审查代码质量报告
- [ ] 更新项目进度

---

## 快速命令参考

### 测试命令

```bash
# 后端测试
cd backend
go test ./... -v                   # 运行所有测试
go test ./... -cover               # 查看覆盖率
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out  # 覆盖率报告

# 前端测试
cd frontend
npm run test                       # 运行测试
npm run test -- --coverage         # 覆盖率报告
```

### 质量检查

```bash
# Go代码检查
cd backend
golangci-lint run                  # 运行lint
golangci-lint run --fix            # 自动修复

# 前端代码检查
cd frontend
npm run lint                       # 运行lint
npm run lint -- --fix              # 自动修复
```

### 安全检查

```bash
# 检查敏感信息泄露
trufflehog git file://. --branch=main

# 容器安全扫描
trivy fs .

# 依赖安全扫描
cd backend
go list -m all | nancy audit       # Go依赖检查
cd frontend
npm audit                          # npm依赖检查
```

### 性能分析

```bash
# Go性能分析
cd backend
go tool pprof cpu.prof             # CPU分析
go tool pprof mem.prof             # 内存分析

# 数据库性能分析
psql -h localhost -U aiops -d aiopsdb
# 运行: EXPLAIN ANALYZE <query>
```

---

## 问题和帮助

遇到问题？请查看：
- 详细文档: `docs/OPTIMIZATION-RECOMMENDATIONS.md`
- 问题追踪: GitHub Issues
- 团队讨论: 项目Slack频道

---

**更新时间**: 2026-07-07
**下次更新**: 每周一审查进度
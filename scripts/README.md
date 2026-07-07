# 🛠️ 项目脚本说明

本目录包含项目的各类脚本工具，按用途分类管理。

## 📁 目录结构

```
scripts/
├── setup/          # 环境设置脚本
├── dev/            # 开发脚本
├── debug/          # 调试脚本
├── deployment/     # 部署脚本
└── archive/        # 旧脚本归档
```

## 🚀 Setup脚本

环境初始化和设置：

| 脚本 | 说明 | 使用方式 |
|------|------|---------|
| `init-db.sql` | 数据库初始化SQL | `psql -f setup/init-db.sql` |
| `init-temporal-db.sh` | Temporal数据库初始化 | `./setup/init-temporal-db.sh` |
| `start-temporal.sh` | Temporal Server启动 | `./setup/start-temporal.sh` |

## 💻 Dev脚本

开发辅助工具：

| 脚本 | 说明 | 使用方式 |
|------|------|---------|
| `start-dev.sh` | 启动开发环境 | `./dev/start-dev.sh` |
| `api-test.sh` | API接口测试 | `./dev/api-test.sh` |
| `quick-test.sh` | 快速测试 | `./dev/quick-test.sh` |
| `quick-start-demo.sh` | 快速演示启动 | `./dev/quick-start-demo.sh` |

## 🔧 Debug脚本

调试和问题排查工具：

| 脚本 | 说明 | 使用方式 |
|------|------|---------|
| `test_tool_execution.sh` | 工具执行测试 | `./debug/test_tool_execution.sh` |
| `verify_tool_binding.sh` | 工具绑定验证 | `./debug/verify_tool_binding.sh` |
| `debug_frontend_rag.sh` | 前端RAG调试 | `./debug/debug_frontend_rag.sh` |
| `fix_milvus_length.sh` | Milvus长度修复 | `./debug/fix_milvus_length.sh` |
| `fix-category-model.sh` | 分类模型修复 | `./debug/fix-category-model.sh` |
| `test_long_document.sh` | 长文档测试 | `./debug/test_long_document.sh` |
| `cleanup_wrong_data.go` | 数据清理工具 | `go run debug/cleanup_wrong_data.go` |

## 📦 Deployment脚本

生产环境部署：

| 脚本 | 说明 | 使用方式 |
|------|------|---------|
| `system.sh` | 系统部署脚本 | `./deployment/system.sh` |
| `start-temporal-dev.sh` | Temporal开发环境 | `./deployment/start-temporal-dev.sh` |

## 🔙 Backend专用脚本

后端专用脚本位于 `backend/scripts/` 目录，包含：
- 数据库迁移工具
- 知识库初始化
- RAG测试工具
- JWT生成工具
- Milvus操作工具

详见 `backend/scripts/` 目录。

## 📝 脚本命名规范

### Shell脚本
- 使用 `.sh` 扩展名
- 使用小写连字符命名：`fix-milvus-length.sh`
- 添加可执行权限：`chmod +x script.sh`

### Go工具
- 使用 `.go` 扩展名
- 使用小写下划线命名：`check_docs.go`
- 直接运行：`go run tool.go`

### SQL脚本
- 使用 `.sql` 扩展名
- 使用小写连字符命名：`init-db.sql`

## ⚠️ 使用注意事项

1. **权限检查**: 运行shell脚本前确保有可执行权限
2. **路径问题**: 脚本中的路径相对于项目根目录
3. **环境变量**: 检查脚本中是否需要特定的环境变量
4. **依赖检查**: 运行前确保所需依赖服务已启动

## 🎯 快速使用

### 初始化开发环境
```bash
# 1. 初始化数据库
./scripts/setup/init-db.sql

# 2. 启动Temporal
./scripts/setup/start-temporal.sh

# 3. 启动开发环境
./scripts/dev/start-dev.sh
```

### 测试功能
```bash
# API测试
./scripts/dev/api-test.sh

# 工具测试
./scripts/debug/test_tool_execution.sh
```

### 问题排查
```bash
# 前端RAG调试
./scripts/debug/debug_frontend_rag.sh

# 数据清理
go run ./scripts/debug/cleanup_wrong_data.go
```

---

**脚本持续更新中，如有疑问请查看脚本内容或联系开发团队。**
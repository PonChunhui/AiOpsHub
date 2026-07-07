# 🔧 Backend专用脚本

本目录包含后端专用的工具和脚本。

## 📚 知识库管理

| 工具 | 说明 | 使用方式 |
|------|------|---------|
| `init-knowledge.sh` | 知识库初始化 | `./init-knowledge.sh` |
| `import_docs.go` | 文档导入工具 | `go run import_docs.go` |
| `check_docs.go` | 文档检查工具 | `go run check_docs.go` |

## 🔐 认证工具

| 工具 | 说明 | 使用方式 |
|------|------|---------|
| `generate_jwt.go` | JWT生成工具 | `go run generate_jwt.go` |
| `generate_basic_auth.go` | Basic Auth生成 | `go run generate_basic_auth.go` |

## 🗄️ 数据库迁移

| 工具 | 说明 | 使用方式 |
|------|------|---------|
| `migrate_tools_db.go` | 工具数据库迁移 | `go run migrate_tools_db.go` |
| `apply_history_indexes.sh` | 历史索引应用 | `./apply_history_indexes.sh` |
| `verify_migration.go` | 迁移验证工具 | `go run verify_migration.go` |

## 🔍 Milvus操作

| 工具 | 说明 | 使用方式 |
|------|------|---------|
| `rebuild_milvus_collection.go` |重建Milvus集合 | `go run rebuild_milvus_collection.go` |
| `rebuild_milvus_index.go` |重建Milvus索引 | `go run rebuild_milvus_index.go` |
| `sync_milvus_to_pg.go` | Milvus到PG同步 | `go run sync_milvus_to_pg.go` |
| `sync_pg_to_milvus.go` | PG到Milvus同步 | `go run sync_pg_to_milvus.go` |
| `check_cosine_scores.go` | Cosine评分检查 | `go run check_cosine_scores.go` |

## 🧪 RAG测试

| 工具 | 说明 | 使用方式 |
|------|------|---------|
| `test_embedding.sh` | Embedding测试 | `./test_embedding.sh` |
| `test_rag_full.sh` | RAG完整测试 | `./test_rag_full.sh` |
| `test_rag_complete.sh` | RAG完整流程测试 | `./test_rag_complete.sh` |
| `test_rag_sync.go` | RAG同步测试 | `go run test_rag_sync.go` |
| `quick_rag_verify.go` | RAG快速验证 | `go run quick_rag_verify.go` |
| `verify_rag_optimization.sh` | RAG优化验证 | `./verify_rag_optimization.sh` |

## 🤖 工具初始化

| 工具 | 说明 | 使用方式 |
|------|------|---------|
| `init_preset_tools.go` | 预设工具初始化 | `go run init_preset_tools.go` |
| `init_preset_bindings.go` | 预设绑定初始化 | `go run init_preset_bindings.go` |

## 📊 历史上下文测试

| 工具 | 说明 | 使用方式 |
|------|------|---------|
| `test_history_context.go` | 历史上下文测试 | `go run test_history_context.go` |
| `test_history_context_integration.sh` | 历史集成测试 | `./test_history_context_integration.sh` |

## 🚀 快速使用

### 初始化知识库
```bash
# 1. 初始化知识库
./init-knowledge.sh

# 2. 导入文档
go run import_docs.go
```

### 测试RAG功能
```bash
# 快速验证
go run quick_rag_verify.go

# 完整测试
./test_rag_complete.sh
```

### 数据库迁移
```bash
# 迁移工具数据库
go run migrate_tools_db.go

# 验证迁移
go run verify_migration.go
```

### 生成JWT
```bash
# 生成JWT token
go run generate_jwt.go
```

---

**注意**: 运行Go工具前确保已编译backend或在backend目录下运行。
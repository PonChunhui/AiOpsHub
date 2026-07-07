# 📚 AiOpsHub 文档导航

本目录包含项目的所有文档，按功能分类管理。

## 📖 核心文档

| 文档 | 说明 | 状态 |
|------|------|------|
| [PRD.md](./PRD.md) | 产品需求文档 | ✅ 核心文档 |
| [architecture.md](./architecture.md) | 系统架构设计 | ✅ 核心文档 |
| [PROJECT-SUMMARY.md](./PROJECT-SUMMARY.md) | 项目功能总结 | ✅ 核心文档 |
| [FEATURES.md](./FEATURES.md) | 功能特性说明 | ✅ 核心文档 |
| [README.md](./README.md) | 文档导航（本文件） | ✅ 导航文档 |

## 🎨 设计文档

详细设计文档位于 `design/` 目录：

| 文档 | 说明 |
|------|------|
| [design/database-design.md](./design/database-design.md) | 数据库设计 |
| [design/temporal-workflow-design.md](./design/temporal-workflow-design.md) | Temporal工作流设计 |
| [design/langchaingo-agent-design.md](./design/langchaingo-agent-design.md) | LangChainGo Agent设计 |
| [design/coordinator-agent-quick-start.md](./design/coordinator-agent-quick-start.md) | Coordinator Agent快速开始 |
| [design/implementation-summary.md](./design/implementation-summary.md) | 实现总结 |

## 🔌 API文档

API相关文档位于 `api/` 目录：

| 文档 | 说明 |
|------|------|
| [api/api-reference.md](./api/api-reference.md) | API接口参考 |
| [api/backend-api.md](./api/backend-api.md) | 后端API文档 |

## 📘 使用指南

使用指南位于 `guides/` 目录：

| 文档 | 说明 |
|------|------|
| [guides/quick-start.md](./guides/quick-start.md) | 快速开始指南 |
| [guides/deployment.md](./guides/deployment.md) | 部署指南 |
| [guides/frontend-development-guide.md](./guides/frontend-development-guide.md) | 前端开发指南 |

## 📦 归档文档

历史文档和实施记录归档在 `archive/` 目录：

### 2024-06 实施文档
包含2024年6月的实施和调试文档：
- `archive/2024-06/agent-tool-integration-status.md` - Agent工具集成状态
- `archive/2024-06/auto-rag-implementation-summary.md` - RAG自动检索实施总结
- `archive/2024-06/auto-rag-usage.md` - RAG使用说明
- `archive/2024-06/chat-rag-diagnosis.md` - Chat RAG诊断
- `archive/2024-06/chat-style-optimization.md` - Chat样式优化
- `archive/2024-06/chat-window-optimization.md` - Chat窗口优化
- `archive/2024-06/knowledge-model-change.md` - 知识模型变更
- `archive/2024-06/KNOWLEDGE-MODEL-CHANGE-COMPLETE.md` - 知识模型变更完成报告
- `archive/2024-06/markdown-render-fix.md` - Markdown渲染修复
- `archive/2024-06/rag-reference-display-issue.md` - RAG引用显示问题
- `archive/2024-06/rag-test-report.md` - RAG测试报告
- `archive/2024-06/RAG_OPTIMIZATION_DEPLOYMENT.md` - RAG优化部署
- `archive/2024-06/tool-management-implementation.md` - 工具管理实施
- `archive/2024-06/tool-testing-guide.md` - 工具测试指南
- `archive/2024-06/FINAL-SUMMARY.md` - 最终总结
- `archive/2024-06/方案C-实施完成报告.md` - 方案C实施报告
- `archive/2024-06/embedding-fix-validation-report.md` - Embedding修复验证

### 问题排查文档
排查和调试文档：
- `archive/troubleshooting/embedding-error-troubleshooting.md` - Embedding错误排查
- `archive/troubleshooting/frontend-500-error-debug.md` - 前端500错误调试

### 旧版文档
- `archive/old-project-readme.md` - 旧版项目说明文档

## 📝 文档命名规范

### 核心文档
使用大写连字符命名：
- `PRD.md` - 产品需求文档
- `PROJECT-SUMMARY.md` - 项目总结
- `FEATURES.md` - 功能特性

### 设计文档
使用小写连字符命名：
- `database-design.md`
- `temporal-workflow-design.md`

### 归档文档
归档文档按月份组织，添加前缀说明：
- `archive/2024-06/rag-implementation-summary.md`

### 排查文档
排查文档归档在 troubleshooting 目录：
- `archive/troubleshooting/error-troubleshooting.md`

## 🔍 文档查找指南

### 新手入门
1. 先读 [PRD.md](./PRD.md) 了解产品定位
2. 再读 [architecture.md](./architecture.md) 了解系统架构
3. 最后读 [guides/quick-start.md](./guides/quick-start.md) 快速上手

### 开发者
1. 查看 [design/](./design/) 目录了解详细设计
2. 参考 [api/](./api/) 目录了解API接口
3. 需要调试时查看 [archive/troubleshooting/](./archive/troubleshooting/)

### 产品经理
1. 阅读 [PRD.md](./PRD.md) 了解需求
2. 查看 [PROJECT-SUMMARY.md](./PROJECT-SUMMARY.md) 了解功能实现
3. 参考 [FEATURES.md](./FEATURES.md) 了解功能特性

## 📌 文档维护

### 添加新文档
1. 核心文档：直接放在 `docs/` 根目录
2. 设计文档：放在 `docs/design/`
3. API文档：放在 `docs/api/`
4. 使用指南：放在 `docs/guides/`

### 归档文档
实施完成后，将实施文档归档到：
- 一般实施文档：`docs/archive/YYYY-MM/`
- 排查文档：`docs/archive/troubleshooting/`

### 更新文档导航
添加新文档后，更新本文件的目录列表。

---

**文档持续更新中，如有疑问请查看对应目录的详细文档。**
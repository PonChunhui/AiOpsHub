# 知识库模型修改总结

## 修改内容

将知识库模型中的`Category`字段改为两个字段：
- `DocType` - 文档类型：sop / faq / alert  
- `Component` - 组件名：mysql / k8s / redis

## 已完成的修改

### 1. 数据模型 ✅
**文件**: `backend/internal/model/models.go`
```go
type RAGDocument struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Title     string    `json:"title" gorm:"not null"`
    Content   string    `json:"content" gorm:"type:text"`
    DocType   string    `json:"doc_type" gorm:"type:varchar(50)"`  // 文档类型：sop / faq / alert
    Component string    `json:"component" gorm:"type:varchar(50)"` // 组件名：mysql / k8s / redis
    Tags      string    `json:"tags" gorm:"type:varchar(500)"`
    CreatedBy string    `json:"created_by" gorm:"type:varchar(100)"`
    UpdatedBy string    `json:"updated_by" gorm:"type:varchar(100)"`
    CreatedAt time.Time `json:"created_at" gorm:"index"`
    UpdatedAt time.Time `json:"updated_at" gorm:"index"`
}
```

**文件**: `backend/internal/service/rag_service.go`
```go
type KnowledgeDocument struct {
    ID        string                 `json:"id"`
    Title     string                 `json:"title"`
    Content   string                 `json:"content"`
    DocType   string                 `json:"doc_type"`  // 文档类型：sop / faq / alert
    Component string                 `json:"component"` // 组件名：mysql / k8s / redis
    Tags      []string               `json:"tags"`
    Metadata  map[string]interface{} `json:"metadata"`
}
```

### 2. Handler层 ✅
**文件**: `backend/internal/handler/service_handler.go`

**修改点**:
- `ListRAGDocuments`: 参数从`category`改为`doc_type`和`component`
- `AddRAGDocument`: 请求体增加`doc_type`和`component`字段
- `UpdateRAGDocument`: 请求体增加`doc_type`和`component`字段

### 3. 数据库迁移脚本 ✅
**文件**: `backend/migrations/add_doc_type_component.sql`

**内容**:
- 添加新字段`doc_type`和`component`
- 从旧`category`字段迁移数据
- 创建索引
- 添加字段注释

## 待修改的文件

### Service层 (需继续修改)
**文件**: `backend/internal/service/rag_service.go`

**需要修改的方法**:
- `UpdateDocument(ctx, docID, title, content, docType, component, tags, metadata)` - 参数签名已改，内部实现待修改
- `ListDocuments(ctx, docType, component, search, page, pageSize)` - 参数签名已改，内部实现待修改
- `matchQuery()`方法中使用Category的地方
- 示例知识库文档(hardcoded知识库)

**文件**: `backend/internal/service/milvus_service.go`
- Milvus schema定义
- Insert和Search操作
- 字段映射

### Repository层
**文件**: `backend/internal/repository/rag_repo.go`
- `List()`方法参数和查询逻辑

### 前端代码
**文件列表**:
- `frontend/src/views/KnowledgeBase.vue`
- `frontend/src/views/DocumentEditor.vue`
- `frontend/src/views/AIAssistant.vue`
- `frontend/src/components/chat/RagReferences.vue`
- `frontend/src/api/index.ts`

### 其他脚本
**文件列表**:
- `backend/scripts/check_docs.go`
- `backend/scripts/import_docs.go`
- `backend/scripts/sync_pg_to_milvus.go`
- `backend/scripts/sync_milvus_to_pg.go`
- `backend/scripts/test_rag_sync.go`

## 下一步操作

1. **修改rag_service.go内部实现**
   - 修改所有使用Category的代码
   - 更新示例知识库文档

2. **修改milvus_service.go**
   - 更新schema定义
   - 修改字段映射

3. **修改rag_repo.go**
   - 更新查询逻辑

4. **修改前端代码**
   - 更新表单字段
   - 更新显示逻辑

5. **执行数据库迁移**
   ```bash
   # 连接到PostgreSQL执行迁移脚本
   psql -U aiops -d aiops -f backend/migrations/add_doc_type_component.sql
   ```

6. **修改初始化知识库数据**
   - 更新示例文档的分类数据

## 数据迁移逻辑说明

### DocType映射规则
| 原Category | 新DocType |
|-----------|-----------|
| troubleshooting, 排查 | sop |
| optimization, 优化 | sop |
| 配置 | sop |
| faq, 问答 | faq |
| alert, 告警 | alert |
| 其他 | sop |

### Component映射规则
| 原Category关键词 | 新Component |
|-----------------|-------------|
| mysql, 数据库, database | mysql |
| k8s, kubernetes | k8s |
| redis, 缓存 | redis |
| docker | docker |
| nginx | nginx |
| java, go | application |
| 其他 | general |

## API参数变化

### 查询文档列表
**旧接口**:
```
GET /api/v1/rag/documents?category=troubleshooting&search=xxx&page=1&pageSize=10
```

**新接口**:
```
GET /api/v1/rag/documents?doc_type=sop&component=mysql&search=xxx&page=1&pageSize=10
```

### 创建文档
**旧请求体**:
```json
{
  "title": "xxx",
  "content": "xxx",
  "category": "troubleshooting",
  "tags": ["性能"]
}
```

**新请求体**:
```json
{
  "title": "xxx",
  "content": "xxx",
  "doc_type": "sop",
  "component": "mysql",
  "tags": ["性能"]
}
```

---

**创建时间**: 2026-06-28  
**状态**: 进行中，已完成模型定义和Handler层修改
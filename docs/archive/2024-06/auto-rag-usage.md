# Eino对话自动RAG使用说明

## 功能概述

在Eino对话系统中,已实现自动RAG(检索增强生成)功能。当用户发送消息时,系统会自动从Milvus向量库检索相关知识,并作为上下文注入到对话中,增强AI的回答能力。

## 核心特性

- **自动知识检索**: 在每次对话时,系统自动根据用户问题检索相关知识
- **智能上下文注入**: 将检索到的知识无缝融入对话上下文
- **可配置开关**: 支持通过配置文件控制是否启用RAG功能
- **Milvus集成**: 使用已集成的Milvus向量数据库进行高效检索

## 配置说明

### 1. 启用/禁用RAG

在配置文件 `backend/configs/config.yaml` 中设置:

```yaml
llm:
  enable_rag: true  # 启用RAG功能
  # enable_rag: false # 禁用RAG功能
```

### 2. Milvus配置

确保Milvus向量库已正确配置:

```yaml
milvus:
  host: localhost
  port: 19530
  database: default
  collection: aiops_knowledge
```

### 3. Embedding配置

配置文本embedding服务(用于将查询转换为向量):

```yaml
embedding:
  provider: openai  # 或 aliyun_bailian
  model: text-embedding-ada-002
  api_key: your-api-key
  base_url: https://api.openai.com/v1
```

## 工作流程

1. **用户发送消息**: 用户通过API发送对话消息
2. **RAG检索**: 系统自动将用户问题转换为向量,从Milvus检索相关知识(topK=3)
3. **上下文构建**: 将检索到的知识作为背景信息,与历史对话合并
4. **LLM生成**: 将完整上下文发送给Eino LLM生成回答
5. **返回结果**: 返回AI回复给用户

## API使用示例

### 创建会话

```bash
POST /api/v1/chat/sessions
Authorization: Bearer <token>

{
  "title": "运维问题咨询",
  "model": "gpt-3.5-turbo"
}
```

### 发送消息(自动RAG)

```bash
POST /api/v1/chat/messages
Authorization: Bearer <token>

{
  "session_id": "session-001",
  "content": "订单服务响应很慢,帮我分析原因"
}
```

系统会自动:
1. 从Milvus检索与"订单服务响应慢"相关的知识
2. 将知识注入到对话上下文
3. 基于知识背景生成回答

### 示例响应

```json
{
  "message": "消息发送成功",
  "ai_response": "根据知识库分析,订单服务响应慢可能的原因有:\n1. CPU使用率过高\n2. 内存不足或内存泄漏\n3. 数据库慢查询\n4. 连接池配置不当\n\n建议排查步骤:\n- 使用top命令查看CPU占用\n- 检查内存使用情况\n- 分析数据库慢查询日志\n- 检查连接池配置...",
  "user_message": {...},
  "ai_message": {...}
}
```

## 知识库管理

### 添加知识文档

```bash
POST /api/v1/rag/documents
Authorization: Bearer <token>

{
  "title": "CPU使用率高排查方法",
  "content": "# CPU使用率高排查\n\n## 排查步骤\n...",
  "category": "troubleshooting",
  "tags": ["CPU", "性能", "排查"]
}
```

### 查询知识文档

```bash
GET /api/v1/rag/documents?category=troubleshooting&page=1&pageSize=10
Authorization: Bearer <token>
```

### 搜索知识

```bash
POST /api/v1/rag/search
Authorization: Bearer <token>

{
  "query": "服务响应慢",
  "top_k": 5
}
```

## 技术实现

### 核心代码位置

- **ChatService**: `backend/internal/service/chat_service.go`
- **RAGService**: `backend/internal/service/rag_service.go`
- **MilvusService**: `backend/internal/service/milvus_service.go`
- **EmbeddingService**: `backend/internal/service/embedding_service.go`
- **ChatHandler**: `backend/internal/handler/chat_handler.go`

### 关键流程

```go
// chat_service.go中的SendMessage方法
func (s *ChatService) SendMessage(ctx context.Context, sessionID, content string) {
    // 1. 如果启用RAG,先检索相关知识
    if s.enableRAG && s.ragSvc != nil {
        knowledgeContext, err := s.ragSvc.GetContextForQuery(ctx, content, 1000)
        if err == nil && knowledgeContext != "" {
            promptBuilder.WriteString(knowledgeContext)
        }
    }
    
    // 2. 添加历史对话
    historyMessages := s.repo.GetRecentMessages(sessionID, s.maxCtx)
    // ...
    
    // 3. 构建完整prompt
    fullPrompt := promptBuilder.String()
    
    // 4. 调用LLM生成回复
    aiResponse := s.llm.Generate(ctx, fullPrompt)
    
    // 5. 返回结果
    return aiResponse
}
```

## 性能优化建议

1. **合理设置topK**: 建议3-5个,避免过多知识影响响应速度
2. **优化embedding**: 使用高质量的embedding模型提高检索精度
3. **定期更新知识库**: 保持知识库内容新鲜和准确
4. **监控检索耗时**: 监控RAG检索的耗时,避免影响用户体验

## 故障排查

### RAG未生效

检查以下配置:
1. `llm.enable_rag` 是否为 `true`
2. Milvus是否正常运行并可连接
3. Embedding服务是否正常配置
4. 知识库是否有相关文档

### 检索结果不准确

优化方案:
1. 检查embedding模型配置
2. 优化知识文档的标签和分类
3. 调整topK参数
4. 提高知识文档质量

### 性能问题

排查步骤:
1. 检查Milvus查询性能
2. 检查embedding服务响应时间
3. 检查LLM调用耗时
4. 优化索引和查询参数

## 最佳实践

1. **知识库分类**: 按领域分类管理知识文档
2. **标签管理**: 为知识文档添加准确标签
3. **定期维护**: 定期更新和清理知识库
4. **监控分析**: 监控RAG效果和用户满意度
5. **迭代优化**: 根据反馈持续优化知识库内容

## 总结

通过Eino对话自动RAG功能,系统能够自动从Milvus向量库检索相关知识并增强对话质量,大大提升了AI回答的专业性和准确性。结合知识库管理功能,可以构建强大的智能运维对话系统。
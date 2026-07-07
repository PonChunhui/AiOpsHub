# Embedding API 400错误排查指南

## 问题现象
添加RAG文档时报错：`{"error":"embedding request failed: status 400"}`

## 可能原因

### 1. 阿里云Embedding API URL配置错误

**当前配置**（可能有问题）：
```yaml
embedding:
  provider: "aliyun_bailian"
  model: "text-embedding-v2"
  api_key: "sk-086920878fb641a3bea1ce785eacb200"
  base_url: "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding"
```

**正确的配置选项**：

#### 方案1: 使用OpenAI兼容模式（推荐）
```yaml
embedding:
  provider: "aliyun_bailian"
  model: "text-embedding-v2"  # 或 text-embedding-v1
  api_key: "your-api-key"
  base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
```

#### 方案2: 使用阿里云原生API（需要修改代码）
阿里云原生embedding API的URL和格式不同，需要检查官方文档。

### 2. 模型名称错误

阿里云支持的embedding模型：
- `text-embedding-v1`
- `text-embedding-v2`

确保model名称正确。

### 3. API Key问题

检查API Key是否：
- 有效
- 有embedding服务权限
- 没有过期

### 4. 文本长度问题

如果文本过长可能超出API限制，阿里云embedding单次请求文本长度有限制。

## 排查步骤

### 1. 查看详细错误日志

我已经添加了详细日志，现在会输出：
- 请求URL
- 请求Body
- 响应Body

重启服务后再次尝试添加文档，查看日志输出：
```bash
# 查看API服务日志
tail -f logs/api-server.log
```

日志会显示：
```
Alibaba embedding request: URL=xxx, Model=xxx, Body=xxx
Alibaba embedding response: status=xxx, body=xxx
```

### 2. 测试Embedding API

使用curl直接测试API：

**OpenAI兼容模式**：
```bash
curl -X POST https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings \
  -H "Authorization: Bearer sk-086920878fb641a3bea1ce785eacb200" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "text-embedding-v2",
    "input": "测试文本"
  }'
```

**阿里云原生模式**：
```bash
curl -X POST https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding \
  -H "Authorization: Bearer sk-086920878fb641a3bea1ce785eacb200" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "text-embedding-v2",
    "input": {
      "texts": ["测试文本"]
    },
    "parameters": {
      "text_type": "document"
    }
  }'
```

### 3. 检查响应内容

查看API返回的具体错误信息，常见的400错误原因：
- `{"error": "invalid model"}` - 模型名称错误
- `{"error": "invalid api_key"}` - API Key无效
- `{"error": "text too long"}` - 文本超长
- `{"error": "rate limit exceeded"}` - 频率限制

## 解决方案

### 快速修复（推荐）

修改配置文件使用OpenAI兼容模式：

```yaml
embedding:
  provider: "aliyun_bailian"
  model: "text-embedding-v2"
  api_key: "your-api-key"
  base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"  # 修改为OpenAI兼容模式
```

同时需要修改embedding_service.go中的阿里云embedding实现，使其支持OpenAI兼容格式：

```go
func (e *EmbeddingService) getAlibabaEmbedding(ctx context.Context, text string) ([]float32, error) {
	// 使用OpenAI兼容模式
	return e.getOpenAIEmbedding(ctx, text)
}
```

或者修改provider判断逻辑：

```go
func (e *EmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	switch e.Provider {
	case "openai":
		return e.getOpenAIEmbedding(ctx, text)
	case "aliyun_bailian", "qwen":
		// 阿里云使用OpenAI兼容模式
		return e.getOpenAIEmbedding(ctx, text)
	default:
		return e.getMockEmbedding(text), nil
	}
}
```

### 验证修复

1. 修改配置后重启服务
2. 再次尝试添加RAG文档
3. 查看日志确认embedding生成成功
4. 检查文档是否成功添加到Milvus

## 临时方案：使用Mock Embedding

如果embedding服务暂时无法修复，可以使用mock embedding：

修改配置：
```yaml
embedding:
  provider: "mock"  # 使用mock provider
  model: ""
  api_key: ""
  base_url: ""
```

这样会生成模拟的embedding向量，虽然检索质量较低，但可以验证整个流程是否正常工作。

## 完整配置示例

### 推荐配置（阿里云OpenAI兼容模式）

```yaml
llm:
  provider: "aliyun_bailian"
  model: "qwen-turbo"
  api_key: "sk-086920878fb641a3bea1ce785eacb200"
  base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
  temperature: 0.7
  max_tokens: 4000
  enable_rag: true

embedding:
  provider: "aliyun_bailian"
  model: "text-embedding-v2"
  api_key: "sk-086920878fb641a3bea1ce785eacb200"
  base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"  # 与llm使用相同的base_url

milvus:
  host: "192.168.100.10"
  port: 19530
  collection: "aiops_knowledge"
```

注意：阿里云的LLM和Embedding都支持OpenAI兼容模式，使用相同的base_url。

## 获取帮助

如果以上方案都无法解决：
1. 查看完整错误日志
2. 联系阿里云技术支持确认embedding API格式
3. 提交Issue到项目GitHub仓库

## 相关文档

- 阿里云DashScope API文档：https://help.aliyun.com/zh/dashscope/
- OpenAI Embedding API文档：https://platform.openai.com/docs/api-reference/embeddings
- 项目RAG使用指南：`docs/auto-rag-usage.md`
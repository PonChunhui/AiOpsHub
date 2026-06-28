# 阿里云百炼LLM集成

## 支持的模型

阿里云百炼提供以下模型：

| 模型 | 说明 | 适用场景 |
|------|------|----------|
| qwen-turbo | 快速响应 | 实时监控、快速分析 |
| qwen-plus | 平衡性能 | 常规运维任务 |
| qwen-max | 最强能力 | 复杂故障诊断 |
| qwen-long | 长文本处理 | 日志分析、报告生成 |

## API配置

### 1. 获取API Key

访问阿里云百炼平台：https://bailian.console.aliyun.com/

1. 创建应用
2. 获取API Key

### 2. 环境变量配置

```bash
export ALIYUN_BAILIAN_API_KEY="your-api-key-here"
```

### 3. 配置文件

编辑 `config/config.yaml`：

```yaml
llm:
  provider: "aliyun_bailian"
  model: "qwen-turbo"
  api_key: "${ALIYUN_BAILIAN_API_KEY}"
  base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
  temperature: 0.7
  max_tokens: 4000
```

## API使用

### 创建AI Agent

```bash
curl -X POST http://localhost:8080/api/v1/ai-agents \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer token" \
  -d '{
    "id": "my-agent-001",
    "name": "MyMonitorAgent",
    "type": "monitor",
    "description": "系统监控Agent",
    "provider": "aliyun_bailian",
    "model": "qwen-turbo",
    "temperature": 0.7,
    "max_tokens": 2000
  }'
```

### 执行Agent任务

```bash
curl -X POST http://localhost:8080/api/v1/ai-agents/my-agent-001/execute \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer token" \
  -d '{
    "task_type": "alert_analysis",
    "input": {
      "alert": "CPU使用率超过90%",
      "source": "prometheus",
      "timestamp": "2024-01-01T10:00:00Z"
    }
  }'
```

### 任务类型

支持的TaskType：

1. **alert_analysis** - 告警分析
   - 评估告警严重性
   - 分析根本原因
   - 提供处理建议

2. **incident_diagnosis** - 故障诊断
   - 根因分析
   - 影响范围评估
   - 修复建议

3. **auto_remediation** - 自动修复
   - 可执行步骤
   - 风险评估
   - 验证方法

4. **自定义任务** - 其他运维任务

## OpenAI兼容

阿里云百炼使用OpenAI兼容的API格式，因此可以直接使用langchaingo的openai包：

- BaseURL: `https://dashscope.aliyuncs.com/compatible-mode/v1`
- 支持OpenAI标准接口
- 支持Temperature、MaxTokens等参数

## 多Provider支持

系统支持多个LLM Provider：

### OpenAI

```yaml
llm:
  provider: "openai"
  model: "gpt-3.5-turbo"
  api_key: "${OPENAI_API_KEY}"
```

### 阿里云百炼

```yaml
llm:
  provider: "aliyun_bailian"
  model: "qwen-turbo"
  api_key: "${ALIYUN_BAILIAN_API_KEY}"
```

### 智谱AI（GLM）

```yaml
llm:
  provider: "zhipu"
  model: "glm-4"
  api_key: "${ZHIPU_API_KEY}"
  base_url: "https://open.bigmodel.cn/api/paas/v4"
```

## 性能对比

| Provider | 模型 | 响应速度 | 成本 | 中文支持 |
|----------|------|----------|------|----------|
| OpenAI | gpt-3.5-turbo | 快 | 中等 | 一般 |
| OpenAI | gpt-4 | 中等 | 高 | 良好 |
| 阿里云百炼 | qwen-turbo | 极快 | 低 | 优秀 |
| 阿里云百炼 | qwen-max | 中等 | 中等 | 优秀 |

## 使用建议

### 开发测试
使用 `qwen-turbo` - 响应快，成本低

### 生产环境
使用 `qwen-plus` 或 `qwen-max` - 性能更强

### 长文本处理
使用 `qwen-long` - 支持长文本分析

## Temporal集成

Agent执行通过Temporal Workflow编排：

```go
workflow.ExecuteActivity(ctx, "ExecuteAgentTask", ActivityInput{
    AgentID:  "analysis-agent-001",
    TaskType: "alert_analysis",
    Input:    alertData,
})
```

## 错误处理

如果API Key未配置，会返回错误：

```json
{
  "code": 500,
  "message": "failed to create agent: Aliyun Bailian API key not configured"
}
```

解决方法：
1. 设置环境变量
2. 在创建Agent时传入API Key参数
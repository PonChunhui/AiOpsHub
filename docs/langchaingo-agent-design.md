# langchaingo Agent设计文档

## 文档信息

| 项目 | 内容 |
|------|------|
| 文档名称 | langchaingo Agent设计 |
| 版本 | v1.0.0 |
| 创建日期 | 2026-06-24 |

## 一、langchaingo概述

### 1.1 什么是langchaingo？

**langchaingo是LangChain的Go实现版本**，由tmc维护的开源项目。

**GitHub**：https://github.com/tmc/langchaingo
- Stars：9.4k+
- 最新版本：v0.1.14
- License：MIT

### 1.2 为什么选择langchaingo？

**优势**：
- ✅ **完整Agent能力**：Agent、Tools、Memory、Chains
- ✅ **多LLM支持**：OpenAI、Gemini、Ollama、Anthropic等15+提供商
- ✅ **Go原生**：与Go生态无缝集成，高性能
- ✅ **轻量级**：相比Python LangChain更轻量

### 1.3 与Python LangChain对比

| 特性 | Python LangChain | langchaingo |
|------|-----------------|-------------|
| **Agent** | ✅ 完整支持 | ✅ 完整支持 |
| **Memory** | ✅ 完整支持 | ✅ 完整支持 |
| **Tools** | ✅ 丰富内置工具 | ✅ 基础工具 + 自定义 |
| **Chains** | ✅ 丰富Chain类型 | ✅ 丰富Chain类型 |
| **LLM支持** | ✅ 30+ providers | ✅ 15+ providers（主流覆盖） |
| **Vector Store** | ✅ 20+ | ✅ 15+（主流覆盖） |
| **LangGraph** | ✅ 工作流编排 | ❌ 不支持（用Temporal替代） |
| **LangSmith** | ✅ 监控平台 | ❌ 不支持（用Temporal Web UI） |

## 二、核心组件

### 2.1 Agent类型

#### 2.1.1 OneShotAgent

**用途**：单次任务执行

**示例**：

```go
package agent

import (
    "github.com/tmc/langchaingo/agents"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/tools"
)

func CreateOneShotAgent() (*agents.Executor, error) {
    // 初始化LLM
    llm, err := openai.New(openai.WithModel("gpt-3.5-turbo"))
    if err != nil {
        return nil, err
    }
    
    // 定义工具
    tools := []tools.Tool{
        NewPrometheusTool(),
        NewKubernetesTool(),
    }
    
    // 创建Agent
    agent := agents.NewOneShotAgent(llm, tools)
    
    // 创建Executor
    executor := agents.NewExecutor(
        agent,
        agents.WithMaxIterations(10),
    )
    
    return executor, nil
}

// 使用Agent
func ExecuteAgent(ctx context.Context, query string) (string, error) {
    executor, err := CreateOneShotAgent()
    if err != nil {
        return "", err
    }
    
    result, err := executor.Call(ctx, map[string]any{
        "input": query,
    })
    
    return result["output"].(string), err
}
```

#### 2.1.2 ConversationalAgent

**用途**：多轮对话Agent

**示例**：

```go
func CreateConversationalAgent() (*agents.Executor, error) {
    llm, _ := openai.New()
    
    tools := []tools.Tool{NewCalculatorTool()}
    
    agent := agents.NewConversationalAgent(llm, tools)
    
    executor := agents.NewExecutor(
        agent,
        agents.WithMemory(memory.NewConversationBuffer()),
    )
    
    return executor, nil
}
```

### 2.2 Tool（工具）

#### 2.2.1 自定义Tool实现

**Tool接口**：

```go
type Tool interface {
    Name() string
    Description() string
    ArgsSchema() any
    Call(ctx context.Context, input string) (string, error)
}
```

**Prometheus查询工具示例**：

```go
package tools

import (
    "context"
    "fmt"
    "github.com/tmc/langchaingo/tools"
    "github.com/tmc/langchaingo/pydantic"
)

type PrometheusQueryInput struct {
    Query    string `json:"query" pydantic:"description=PromQL查询语句"`
    TimeRange string `json:"time_range" pydantic:"description=时间范围，如-1h"`
}

type PrometheusTool struct {
    URL string
}

func (t *PrometheusTool) Name() string {
    return "prometheus_query"
}

func (t *PrometheusTool) Description() string {
    return "查询Prometheus监控指标。输入PromQL查询语句，返回指标数据。"
}

func (t *PrometheusTool) ArgsSchema() any {
    return PrometheusQueryInput{}
}

func (t *PrometheusTool) Call(ctx context.Context, input string) (string, error) {
    // 解析输入参数
    params := ParseToolInput(input)
    
    // 执行Prometheus查询
    result, err := QueryPrometheus(t.URL, params.Query)
    if err != nil {
        return "", err
    }
    
    // 格式化结果
    return FormatPrometheusResult(result), nil
}

func NewPrometheusTool() tools.Tool {
    return &PrometheusTool{
        URL: "http://localhost:9090",
    }
}
```

#### 2.2.2 内置工具

**langchaingo内置工具**：
- Calculator：数学计算
- SerpAPI：搜索引擎
- DuckDuckGo Search：搜索
- Wikipedia：维基百科查询
- SQLDatabase：数据库查询

### 2.3 Memory（记忆）

#### 2.3.1 ConversationBuffer

**用途**：缓冲所有对话历史

```go
import "github.com/tmc/langchaingo/memory"

// 创建ConversationBuffer
mem := memory.NewConversationBuffer()

// 添加消息
mem.AddUserMessage(ctx, "查询CPU使用率")
mem.AddAIMessage(ctx, "当前CPU使用率45%")

// 获取历史
messages := mem.GetMessages()
```

#### 2.3.2 ConversationBufferWindow

**用途**：保留最近N轮对话

```go
// 保留最近5轮对话
mem := memory.NewConversationBufferWindow(5)
```

#### 2.3.3 ConversationTokenBuffer

**用途**：Token限制的记忆

```go
// 限制最多4000 tokens
mem := memory.NewConversationTokenBuffer(
    llm,
    4000,
)
```

#### 2.3.4 PostgreSQL持久化Memory

```go
import "github.com/tmc/langchaingo/memory/sqlite"

// SQLite持久化
store, _ := sqlite.New("aiops.db")
mem := memory.NewConversationBuffer(
    memory.WithChatHistory(store),
)
```

### 2.4 Chain（链）

#### 2.4.1 LLMChain

**基础LLM调用链**：

```go
import (
    "github.com/tmc/langchaingo/chains"
    "github.com/tmc/langchaingo/prompts"
)

// 创建Prompt
prompt := prompts.NewPromptTemplate(
    "请分析以下异常：{{.input}}",
    []string{"input"},
)

// 创建LLMChain
llmChain := chains.NewLLMChain(llm, prompt)

// 执行Chain
result, err := chains.Call(ctx, llmChain, map[string]any{
    "input": "CPU使用率95%",
})
```

#### 2.4.2 SequentialChain

**顺序执行多个Chain**：

```go
// 创建多个Chain
chain1 := chains.NewLLMChain(llm, prompt1)
chain2 := chains.NewLLMChain(llm, prompt2)

// 顺序组合
sequentialChain := chains.NewSequentialChain(
    []chains.Chain{chain1, chain2},
    []string{"input"},
    []string{"output"},
)

// 执行
result, err := chains.Call(ctx, sequentialChain, map[string]any{
    "input": "查询CPU使用率",
})
```

#### 2.4.3 RetrievalQA（RAG问答）

```go
import (
    "github.com/tmc/langchaingo/vectorstores/milvus"
    "github.com/tmc/langchaingo/embeddings"
)

// 创建Embedder
embedder, _ := embeddings.NewEmbedder(llm)

// 创建Vector Store
store, _ := milvus.New(
    milvus.WithHost("localhost"),
    milvus.WithPort(19530),
    milvus.WithEmbedder(embedder),
)

// 创建RetrievalQA Chain
qaChain := chains.NewRetrievalQAFromLLM(
    llm,
    vectorstores.ToRetriever(store, 3),
)

// 执行问答
result, err := chains.Call(ctx, qaChain, map[string]any{
    "query": "如何处理MySQL慢查询？",
})
```

### 2.5 LLM客户端

#### 2.5.1 OpenAI

```go
import "github.com/tmc/langchaingo/llms/openai"

// OpenAI客户端
llm, err := openai.New(
    openai.WithModel("gpt-4-turbo-preview"),
    openai.WithToken(os.Getenv("OPENAI_API_KEY")),
    openai.WithTemperature(0.7),
    openai.WithMaxTokens(4000),
)
```

#### 2.5.2 Ollama（本地模型）

```go
import "github.com/tmc/langchaingo/llms/ollama"

// Ollama客户端（本地LLaMA/Qwen）
llm, err := ollama.New(
    ollama.WithModel("llama2"),
    ollama.WithServerURL("http://localhost:11434"),
)
```

#### 2.5.3 Google Gemini

```go
import "github.com/tmc/langchaingo/llms/googleai"

// Gemini客户端
llm, err := googleai.New(
    googleai.WithAPIKey(os.Getenv("GOOGLE_API_KEY")),
    googleai.WithModel("gemini-pro"),
)
```

### 2.6 Vector Store（向量存储）

#### 2.6.1 Milvus

```go
import "github.com/tmc/langchaingo/vectorstores/milvus"

// Milvus客户端
store, err := milvus.New(
    milvus.WithHost("localhost"),
    milvus.WithPort(19530),
    milvus.WithCollectionName("knowledge_base"),
    milvus.WithEmbedder(embedder),
)

// 添加文档
store.AddDocuments(ctx, []schema.Document{
    {PageContent: "故障案例1：MySQL慢查询"},
})

// 检索
results, _ := store.SimilaritySearch(ctx, "MySQL慢查询", 3)
```

#### 2.6.2 Chroma

```go
import "github.com/tmc/langchaingo/vectorstores/chroma"

store, _ := chroma.New(
    chroma.WithChromaURL("http://localhost:8000"),
    chroma.WithEmbedder(embedder),
)
```

### 2.7 Embeddings（向量化）

#### 2.7.1 OpenAI Embeddings

```go
import "github.com/tmc/langchaingo/embeddings"

// OpenAI Embeddings
embedder, err := embeddings.NewEmbedder(
    llm,
    embeddings.WithModel("text-embedding-3-large"),
)

// 生成向量
embedding, err := embedder.EmbedDocuments(ctx, []string{
    "故障案例1：MySQL慢查询",
})
```

## 三、AiOpsHub Agent实现

### 3.1 Monitor Agent（监控采集）

**职责**：采集监控数据

**完整实现**：

```go
package agents

import (
    "context"
    
    "github.com/tmc/langchaingo/agents"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/memory"
)

type MonitorAgent struct {
    LLM       *openai.LLM
    Tools     []tools.Tool
    Memory    memory.Memory
    Executor  *agents.Executor
}

func NewMonitorAgent() (*MonitorAgent, error) {
    // 初始化LLM
    llm, err := openai.New(
        openai.WithModel("gpt-3.5-turbo"),
        openai.WithTemperature(0.7),
    )
    if err != nil {
        return nil, err
    }
    
    // 创建工具
    tools := []tools.Tool{
        NewPrometheusTool(),
        NewKubernetesTool(),
        NewZabbixTool(),
    }
    
    // 创建记忆
    mem := memory.NewConversationBuffer()
    
    // 创建Agent
    agent := agents.NewOneShotAgent(llm, tools)
    executor := agents.NewExecutor(
        agent,
        agents.WithMemory(mem),
        agents.WithMaxIterations(10),
    )
    
    return &MonitorAgent{
        LLM:      llm,
        Tools:    tools,
        Memory:   mem,
        Executor: executor,
    }, nil
}

func (a *MonitorAgent) Execute(ctx context.Context, query string) (string, error) {
    result, err := a.Executor.Call(ctx, map[string]any{
        "input": query,
    })
    
    if err != nil {
        return "", err
    }
    
    return result["output"].(string), nil
}
```

### 3.2 Analysis Agent（根因分析）

**职责**：分析根因，使用RAG检索历史案例

```go
func NewAnalysisAgent() (*AnalysisAgent, error) {
    llm, _ := openai.New(openai.WithModel("gpt-4-turbo-preview"))
    
    // RAG检索器
    retriever := NewKnowledgeRetriever()
    
    // 使用RetrievalQA Chain
    qaChain := chains.NewRetrievalQAFromLLM(llm, retriever)
    
    return &AnalysisAgent{
        LLM:     llm,
        Chain:   qaChain,
    }, nil
}

func (a *AnalysisAgent) Analyze(ctx context.Context, anomaly string) (string, error) {
    result, err := chains.Call(ctx, a.Chain, map[string]any{
        "query": fmt.Sprintf("分析异常：%s", anomaly),
    })
    
    return result["answer"].(string), err
}
```

### 3.3 Alert Agent（告警处理）

**职责**：告警去重、聚合、分派

```go
func NewAlertAgent() (*AlertAgent, error) {
    llm, _ := openai.New(openai.WithModel("glm-4")) // 使用智谱中文模型
    
    // Prompt模板
    alertPrompt := prompts.NewPromptTemplate(
        `分析以下告警并进行降噪处理：
告警列表：{{.alerts}}
请：
1. 识别重复告警（语义相似度>0.8）
2. 聚合相关告警（同一服务、同一根因）
3. 评估严重性（P0/P1/P2/P3）
4. 推荐处理人`,
        []string{"alerts"},
    )
    
    chain := chains.NewLLMChain(llm, alertPrompt)
    
    return &AlertAgent{Chain: chain}, nil
}

func (a *AlertAgent) ProcessAlerts(ctx context.Context, alerts []Alert) (*AlertProcessResult, error) {
    alertsJSON := json.Marshal(alerts)
    
    result, err := chains.Call(ctx, a.Chain, map[string]any{
        "alerts": alertsJSON,
    })
    
    return ParseAlertResult(result), err
}
```

### 3.4 Decision Agent（决策执行）

**职责**：制定执行方案，执行自动化操作

```go
func NewDecisionAgent() (*DecisionAgent, error) {
    llm, _ := openai.New(openai.WithModel("gpt-4-turbo-preview"))
    
    // 执行工具
    tools := []tools.Tool{
        NewKubernetesOperatorTool(),
        NewSSHExecutorTool(),
        NewSQLExecutorTool(),
    }
    
    agent := agents.NewOneShotAgent(llm, tools)
    executor := agents.NewExecutor(agent)
    
    return &DecisionAgent{Executor: executor}, nil
}

func (a *DecisionAgent) Decide(ctx context.Context, rootCause string) (*DecisionResult, error) {
    prompt := fmt.Sprintf(
        "根据根因分析结果，制定自动化修复方案。根因：%s。请评估风险，生成执行计划。",
        rootCause,
    )
    
    result, err := a.Executor.Call(ctx, map[string]any{
        "input": prompt,
    })
    
    return ParseDecisionResult(result), err
}
```

## 四、工具实现示例

### 4.1 Prometheus查询工具

```go
package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    
    "github.com/tmc/langchaingo/tools"
)

type PrometheusTool struct {
    URL string
}

func (t *PrometheusTool) Name() string {
    return "prometheus_query"
}

func (t *PrometheusTool) Description() string {
    return `查询Prometheus监控指标。
输入格式：{"query": "cpu_usage{service='order-service'}"}
返回：指标数据，包括值和时间戳`
}

func (t *PrometheusTool) ArgsSchema() any {
    return struct {
        Query string `json:"query"`
    }{}
}

func (t *PrometheusTool) Call(ctx context.Context, input string) (string, error) {
    // 解析输入
    var params struct {
        Query string `json:"query"`
    }
    json.Unmarshal([]byte(input), &params)
    
    // 执行查询
    u := fmt.Sprintf("%s/api/v1/query?query=%s", t.URL, url.QueryEscape(params.Query))
    
    resp, err := http.Get(u)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    // 格式化返回
    data := result["data"].(map[string]interface{})
    results := data["result"].([]interface{})
    
    output := fmt.Sprintf("查询：%s\n结果数量：%d\n", params.Query, len(results))
    for _, r := range results {
        item := r.(map[string]interface{})
        metric := item["metric"].(map[string]interface{})
        value := item["value"].([]interface{})
        
        output += fmt.Sprintf("Metric: %v, Value: %v\n", metric, value[1])
    }
    
    return output, nil
}
```

### 4.2 Kubernetes操作工具

```go
package tools

import (
    "context"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

type KubernetesTool struct {
    Client *kubernetes.Clientset
}

func (t *KubernetesTool) Name() string {
    return "kubernetes_api"
}

func (t *KubernetesTool) Description() string {
    return `查询和操作Kubernetes资源。
支持的命令：
- get pods: 获取pod列表
- get services: 获取服务列表
- describe pod <name>: 查看pod详情`
}

func (t *KubernetesTool) Call(ctx context.Context, input string) (string, error) {
    // 解析命令
    cmd := ParseKubectlCommand(input)
    
    switch cmd.Action {
    case "get pods":
        pods, _ := t.Client.CoreV1().Pods(cmd.Namespace).List(ctx, metav1.ListOptions{})
        return FormatPodList(pods), nil
        
    case "get services":
        services, _ := t.Client.CoreV1().Services(cmd.Namespace).List(ctx, metav1.ListOptions{})
        return FormatServiceList(services), nil
        
    default:
        return "未知命令", fmt.Errorf("unknown action: %s", cmd.Action)
    }
}
```

## 五、最佳实践

### 5.1 Prompt工程

**System Prompt模板**：

```go
monitorPrompt := prompts.NewPromptTemplate(
    `你是监控采集Agent，负责从Prometheus、Kubernetes等数据源采集监控数据。

可用工具：
- prometheus_query: 查询Prometheus指标
- kubernetes_api: 查询K8s资源

用户需求：{{.input}}

请：
1. 理解用户需求
2. 选择合适的工具
3. 采集数据
4. 总结结果`,
    []string{"input"},
)
```

### 5.2 错误处理

```go
func ExecuteAgentWithErrorHandling(ctx context.Context, agent *agents.Executor, query string) (string, error) {
    // 重试机制
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        result, err := agent.Call(ctx, map[string]any{"input": query})
        
        if err == nil {
            return result["output"].(string), nil
        }
        
        // 等待后重试
        time.Sleep(time.Second * 2 * (i + 1))
    }
    
    return "", fmt.Errorf("agent execution failed after %d retries", maxRetries)
}
```

### 5.3 性能优化

**并行工具调用**：

```go
// 并行查询多个指标
metrics := []string{"cpu_usage", "memory_usage", "network_io"}

results := make(chan string, len(metrics))

for _, metric := range metrics {
    go func(m string) {
        result, _ := prometheusTool.Call(ctx, m)
        results <- result
    }(metric)
}

// 收集结果
outputs := []string{}
for i := 0; i < len(metrics); i++ {
    outputs = append(outputs, <-results)
}
```

## 六、与Temporal集成

### 6.1 Agent作为Temporal Activity

```go
// Monitor Agent作为Temporal Activity
func MonitorAgentActivity(ctx context.Context, input AgentInput) (*AgentOutput, error) {
    agent, err := NewMonitorAgent()
    if err != nil {
        return nil, err
    }
    
    result, err := agent.Execute(ctx, input.Query)
    
    return &AgentOutput{
        Response: result,
    }, err
}
```

## 七、参考资源

- [langchaingo GitHub](https://github.com/tmc/langchaingo)
- [langchaingo文档](https://tmc.github.io/langchaingo/docs/)
- [langchaingo示例](https://github.com/tmc/langchaingo/tree/main/examples)
- [LangChain Python对比](https://python.langchain.com/docs/)

## 八、更新记录

| 版本 | 日期 | 更新内容 |
|------|------|----------|
| v1.0.0 | 2026-06-24 | 初稿（纯Go + langchaingo） |
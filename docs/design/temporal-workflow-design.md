# Temporal工作流设计文档

## 文档信息

| 项目 | 内容 |
|------|------|
| 文档名称 | Temporal工作流设计 |
| 版本 | v1.0.0 |
| 创建日期 | 2026-06-24 |
| 最后更新 | 2026-06-24 |

## 一、Temporal概述

### 1.1 什么是Temporal？

**Temporal是一个持久执行平台**，专为构建可靠的应用而设计。

**核心特性**：
- **持久执行**：Workflow可运行数月甚至数年，自动捕获状态
- **确定性重放**：Event History是完整执行日志，失败时从历史重建状态
- **自动恢复**：Workflow失败后自动恢复到失败点继续执行
- **多语言支持**：Go、Python、TypeScript、Java等7种语言SDK

### 1.2 为什么选择Temporal？

**生产级案例**：
- OpenAI（使用Temporal运行AI Agent）
- Hashicorp
- Replit（代码执行平台）
- Lovable（AI代码生成）
- Coinbase（支付流程）

**AiOpsHub使用理由**：
1. ✅ **替代LangGraph**：Temporal提供完整的DAG工作流能力
2. ✅ **持久化可靠**：运维场景需要强可靠性（故障处理）
3. ✅ **Go原生支持**：完整Go SDK，无Python依赖
4. ✅ **长运行支持**：Agent协作可能耗时数小时甚至数天
5. ✅ **可视化调试**：Temporal Web UI查看Workflow执行过程

### 1.3 Temporal vs LangGraph对比

| 特性 | Temporal | LangGraph |
|------|----------|-----------|
| **持久执行** | ✅ Event History重放 | ✅ Checkpointer |
| **失败恢复** | ✅ 自动恢复到失败点 | ✅ 从checkpoint恢复 |
| **长运行** | ✅ 数月/数年 | ✅ 支持长运行 |
| **多语言** | ✅ 7种语言 | ❌ 仅Python/JS |
| **可视化** | ✅ Temporal Web UI | ✅ LangSmith |
| **生产级** | ✅ OpenAI等案例 | ✅ LangSmith Cloud |
| **部署方式** | ✅ Self-hosted免费 | ⚠️ LangSmith收费 |

## 二、Temporal核心概念

### 2.1 Workflow（工作流）

**定义**：业务逻辑代码，持久执行

**特性**：
- 确定性执行（不能有随机数、时间等）
- 自动状态捕获
- 可运行数月甚至数年
- 支持人机交互（Signals/Queries）

**Workflow示例**：

```go
func IncidentHandlingWorkflow(ctx workflow.Context, input AgentInput) (*AgentOutput, error) {
    // Workflow逻辑
    // 自动持久化状态
    // 失败后自动恢复
}
```

### 2.2 Activity（活动）

**定义**：外部操作（非确定性部分）

**用途**：
- LLM调用
- 工具执行（API调用、数据库操作）
- 文件操作
- HTTP请求

**Activity示例**：

```go
func MonitorAgentActivity(ctx context.Context, input AgentInput) (*MonitorResult, error) {
    // 使用langchaingo执行Agent
    // 可以有随机操作
    // 失败后自动重试
}
```

### 2.3 Worker（工作进程）

**定义**：执行Workflow和Activity的进程

**特性**：
- 轮询Temporal Server获取任务
- 执行Workflow和Activity代码
- 支持水平扩展

### 2.4 Temporal Server

**定义**：持久化状态和协调执行的服务端

**组件**：
- History Service：存储Event History
- Matching Service：任务匹配和分发
- Frontend Service：API入口
- Web UI：可视化界面

### 2.5 Signal和Query

**Signal**：运行时向Workflow发送消息（人机交互）

```go
// 等待用户确认
var approval bool
workflow.GetSignalChannel(ctx, "approval").Receive(ctx, &approval)
```

**Query**：查询Workflow状态（不改变状态）

```go
// 查询当前Agent状态
workflow.SetQueryHandler(ctx, "current_agent", func() (string, error) {
    return currentAgent, nil
})
```

## 三、AiOpsHub Workflow设计

### 3.1 核心Workflow定义

#### 3.1.1 IncidentHandlingWorkflow（故障处理工作流）

**流程图**：

```
用户发起请求 → Monitor Agent采集 → Analysis Agent分析 → Decision Agent决策 → 执行修复
     ↓               ↓                    ↓                    ↓
  自然语言        指标数据             根因定位            自动化方案
                                      ↓
                                等待用户确认
```

**Workflow定义**：

```go
package workflow

import (
    "context"
    "time"
    
    "go.temporal.io/sdk/workflow"
)

// Workflow输入
type IncidentHandlingInput struct {
    SessionID string
    Query     string                 // 用户自然语言描述
    Context   map[string]interface{} // 上下文信息
}

// Workflow输出
type IncidentHandlingOutput struct {
    Response    string
    RootCause   string
    Solution    string
    TokensUsed  int
    Duration    int64
}

// IncidentHandlingWorkflow
func IncidentHandlingWorkflow(ctx workflow.Context, input IncidentHandlingInput) (*IncidentHandlingOutput, error) {
    // Activity选项（自动重试）
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: time.Minute * 5,
        RetryOptions: workflow.RetryOptions{
            InitialInterval:    time.Second * 10,
            BackoffCoefficient: 2.0,
            MaximumAttempts:    3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // 1. Monitor Agent采集数据
    var monitorResult MonitorActivityResult
    err := workflow.ExecuteActivity(ctx, MonitorAgentActivity, input).Get(ctx, &monitorResult)
    if err != nil {
        return nil, err
    }
    
    // 2. Analysis Agent分析根因
    var analysisResult AnalysisActivityResult
    err = workflow.ExecuteActivity(ctx, AnalysisAgentActivity, monitorResult).Get(ctx, &analysisResult)
    if err != nil {
        return nil, err
    }
    
    // 3. Decision Agent制定方案
    var decisionResult DecisionActivityResult
    err = workflow.ExecuteActivity(ctx, DecisionAgentActivity, analysisResult).Get(ctx, &decisionResult)
    if err != nil {
        return nil, err
    }
    
    // 4. 等待用户确认（人机交互）
    if decisionResult.RequiresApproval {
        // 发送通知给用户
        err = workflow.ExecuteActivity(ctx, NotifyUserActivity, decisionResult).Get(ctx, nil)
        if err != nil {
            return nil, err
        }
        
        // 等待用户Signal
        var approvalSignal ApprovalSignal
        workflow.GetSignalChannel(ctx, "approval").Receive(ctx, &approvalSignal)
        
        if !approvalSignal.Approved {
            return &IncidentHandlingOutput{
                Response: "用户拒绝执行",
            }, nil
        }
    }
    
    // 5. 执行修复方案
    if decisionResult.ShouldExecute {
        var executionResult ExecutionActivityResult
        err = workflow.ExecuteActivity(ctx, ExecutionAgentActivity, decisionResult).Get(ctx, &executionResult)
        if err != nil {
            return nil, err
        }
        
        return &IncidentHandlingOutput{
            Response:  executionResult.Result,
            RootCause: analysisResult.RootCause,
            Solution:  decisionResult.Solution,
        }, nil
    }
    
    return &IncidentHandlingOutput{
        Response: decisionResult.Response,
    }, nil
}
```

#### 3.1.2 AlertDedupWorkflow（告警降噪工作流）

**流程图**：

```
接收100条告警 → Alert Agent去重 → Alert Agent聚合 → Alert Agent分派 → 通知值班人员
                  ↓                    ↓
               语义去重              相关告警合并
```

**Workflow定义**：

```go
func AlertDedupWorkflow(ctx workflow.Context, alerts []Alert) (*AlertDedupOutput, error) {
    // 1. Alert Agent去重（语义理解）
    var dedupResult DedupActivityResult
    err := workflow.ExecuteActivity(ctx, AlertDedupActivity, alerts).Get(ctx, &dedupResult)
    
    // 2. Alert Agent聚合（相关告警合并）
    var aggregationResult AggregationActivityResult
    err = workflow.ExecuteActivity(ctx, AlertAggregationActivity, dedupResult.Alerts).Get(ctx, &aggregationResult)
    
    // 3. Alert Agent分派（智能分配）
    var dispatchResult DispatchActivityResult
    err = workflow.ExecuteActivity(ctx, AlertDispatchActivity, aggregationResult.AggregatedAlerts).Get(ctx, &dispatchResult)
    
    // 4. 通知值班人员
    err = workflow.ExecuteActivity(ctx, NotifyOnCallActivity, dispatchResult).Get(ctx, nil)
    
    return &AlertDedupOutput{
        OriginalCount:  len(alerts),
        DedupedCount:    len(dedupResult.Alerts),
        FinalCount:      len(aggregationResult.AggregatedAlerts),
        DedupRate:       float64(len(alerts)-len(aggregationResult.AggregatedAlerts)) / float64(len(alerts)) * 100,
    }, nil
}
```

#### 3.1.3 RootCauseAnalysisWorkflow（根因分析工作流）

**流程图**：

```
接收异常 → Monitor Agent采集指标 → Analysis Agent分析 → RAG检索历史案例 → LLM推理根因
                                        ↓
                                   拓扑图分析
```

### 3.2 Activity设计

#### 3.2.1 MonitorAgentActivity

**职责**：采集监控数据

**实现**（使用langchaingo）：

```go
package activity

import (
    "context"
    
    "github.com/tmc/langchaingo/agents"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/tools"
)

type MonitorActivityResult struct {
    Data       map[string]interface{}
    HasAnomaly bool
    Metrics    []Metric
    TokensUsed int
}

func MonitorAgentActivity(ctx context.Context, input AgentInput) (*MonitorActivityResult, error) {
    // 1. 初始化LLM
    llm, err := openai.New(openai.WithModel("gpt-3.5-turbo"))
    if err != nil {
        return nil, err
    }
    
    // 2. 创建工具
    prometheusTool := NewPrometheusTool()
    k8sTool := NewKubernetesTool()
    
    // 3. 创建Agent（使用langchaingo）
    agent := agents.NewOneShotAgent(llm, []tools.Tool{prometheusTool, k8sTool})
    executor := agents.NewExecutor(agent)
    
    // 4. 执行Agent
    result, err := executor.Call(ctx, map[string]interface{}{
        "input": input.Query,
    })
    
    return &MonitorActivityResult{
        Data:       result,
        HasAnomaly: result["has_anomaly"].(bool),
        TokensUsed: result["tokens_used"].(int),
    }, err
}
```

#### 3.2.2 AnalysisAgentActivity

**职责**：分析根因

**实现**：

```go
func AnalysisAgentActivity(ctx context.Context, input MonitorActivityResult) (*AnalysisActivityResult, error) {
    llm, err := openai.New(openai.WithModel("gpt-4-turbo-preview"))
    
    // RAG知识检索
    retriever := NewMilvusRetriever()
    qaChain := chains.NewRetrievalQAFromLLM(llm, retriever)
    
    // 分析根因
    result, err := qaChain.Call(ctx, map[string]interface{}{
        "query": fmt.Sprintf("分析异常：%v", input.Data),
    })
    
    return &AnalysisActivityResult{
        RootCause: result["answer"],
        Evidence:  result["evidence"],
    }, err
}
```

#### 3.2.3 ToolExecutionActivity

**职责**：执行具体工具（Prometheus查询、K8s操作等）

**实现**：

```go
func ToolExecutionActivity(ctx context.Context, input ToolExecutionInput) (*ToolExecutionResult, error) {
    switch input.ToolName {
    case "prometheus_query":
        return ExecutePrometheusQuery(ctx, input.Parameters)
    case "kubernetes_api":
        return ExecuteKubernetesAPI(ctx, input.Parameters)
    case "ssh_command":
        return ExecuteSSHCommand(ctx, input.Parameters)
    default:
        return nil, fmt.Errorf("unknown tool: %s", input.ToolName)
    }
}
```

## 四、人机交互设计

### 4.1 用户确认机制

**场景**：高风险操作需要人工确认

**实现**：

```go
// Workflow等待用户确认
var approval ApprovalSignal
workflow.GetSignalChannel(ctx, "approval").Receive(ctx, &approval)

if approval.Approved {
    // 执行操作
} else {
    // 拒绝操作
}
```

**Web前端发送Signal**：

```javascript
// 用户点击"确认"按钮
async function approveWorkflow(workflowId) {
    const response = await fetch(`/api/workflow/${workflowId}/signal`, {
        method: 'POST',
        body: JSON.stringify({
            signal_name: 'approval',
            signal_value: { approved: true }
        })
    });
}
```

### 4.2 实时状态查询

**场景**：前端实时显示Workflow进度

**实现**：

```go
// Workflow内部设置QueryHandler
workflow.SetQueryHandler(ctx, "progress", func() (map[string]interface{}, error) {
    return map[string]interface{}{
        "current_agent": currentAgent,
        "progress":      progressPercent,
        "status":        "running",
    }, nil
})
```

**Web前端查询状态**：

```javascript
// 定时查询Workflow进度
setInterval(async () => {
    const progress = await fetch(`/api/workflow/${workflowId}/query?name=progress`);
    // 更新UI显示进度
}, 2000);
```

## 五、失败恢复策略

### 5.1 Activity自动重试

**配置**：

```go
ao := workflow.ActivityOptions{
    StartToCloseTimeout: time.Minute * 5,
    RetryOptions: workflow.RetryOptions{
        InitialInterval:    time.Second * 10,  // 第一次重试间隔10秒
        BackoffCoefficient: 2.0,               // 间隔翻倍
        MaximumAttempts:    3,                 // 最多重试3次
        MaximumInterval:    time.Minute * 5,   // 最大间隔5分钟
    },
}
```

### 5.2 Workflow状态持久化

**Temporal自动持久化Event History**：

```
Event 1: WorkflowStarted
Event 2: ActivityTaskScheduled (MonitorAgentActivity)
Event 3: ActivityTaskCompleted (MonitorAgentActivity result)
Event 4: ActivityTaskScheduled (AnalysisAgentActivity)
Event 5: WorkflowCompleted
```

**失败后重放**：
- Temporal读取Event History
- 重放所有事件
- 恢复到失败点继续执行

### 5.3 超时控制

**多层超时设置**：

```go
ao := workflow.ActivityOptions{
    ScheduleToCloseTimeout:  time.Hour,        // 总超时1小时
    ScheduleToStartTimeout:  time.Minute * 5,  // 等待分配超时5分钟
    StartToCloseTimeout:     time.Minute * 10, // Activity执行超时10分钟
    HeartbeatTimeout:        time.Second * 30, // 心跳超时30秒
}
```

## 六、Temporal部署

### 6.1 开发环境部署（Docker Compose）

**配置文件**：

```yaml
version: '3.8'
services:
  temporal-server:
    image: temporalio/server:latest
    ports:
      - "7233:7233"  # Temporal Server
      - "8080:8080"  # Temporal Web UI
    environment:
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=temporal
      - POSTGRES_PWD=temporal
      - POSTGRES_SEEDS=postgres
    depends_on:
      - postgres
  
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: temporal
      POSTGRES_PASSWORD: temporal
    ports:
      - "5432:5432"
```

### 6.2 生产环境部署（Kubernetes）

**Temporal Server部署**：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: temporal-server
spec:
  replicas: 3
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
```

**Temporal Worker部署**：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: temporal-worker
spec:
  replicas: 5
  template:
    spec:
      containers:
      - name: temporal-worker
        image: aiops/temporal-worker:latest
        env:
        - name: TEMPORAL_ADDRESS
          value: "temporal-server:7233"
```

### 6.3 Temporal Web UI

**访问地址**：http://localhost:8080

**功能**：
- 查看Workflow执行历史
- 查看Event History详情
- 查看Activity执行结果
- 发送Signal和Query
- 调试Workflow

## 七、监控和运维

### 7.1 Temporal监控

**指标**：
- Workflow执行成功率
- Workflow平均执行时间
- Activity执行成功率
- Worker处理速率

**监控工具**：
- Temporal Server内置Prometheus metrics
- Grafana可视化

### 7.2 日志管理

**Temporal日志**：
- Event History（持久化到PostgreSQL）
- Workflow日志（应用层）
- Activity日志（应用层）

### 7.3 运维操作

**常见操作**：
- 查看Workflow状态：Temporal Web UI
- 重试失败的Workflow：Web UI或CLI
- 取消运行中的Workflow：CLI
- 搜索历史Workflow：Web UI搜索

## 八、最佳实践

### 8.1 Workflow设计原则

1. **确定性执行**：Workflow内不能有随机数、当前时间等非确定性操作
2. **幂等性**：Activity应该幂等，重试时不会产生副作用
3. **原子性**：一个Activity做一件事
4. **超时设置**：合理设置超时，避免长时间阻塞

### 8.2 Activity设计原则

1. **重试友好**：Activity失败后可以安全重试
2. **幂等性**：重复执行不会产生副作用
3. **心跳机制**：长时间Activity发送心跳
4. **错误处理**：返回明确的错误信息

### 8.3 性能优化

1. **并行Activity**：使用`workflow.ExecuteActivity`并行执行
2. **LocalActivity**：小任务使用LocalActivity（不持久化）
3. **Continue-As-New**：长运行Workflow定期Continue-As-New
4. **Child Workflow**：复杂流程拆分为Child Workflow

## 九、参考资源

### 9.1 官方文档

- [Temporal官方文档](https://docs.temporal.io/)
- [Temporal Go SDK文档](https://docs.temporal.io/dev/go)
- [Temporal GitHub](https://github.com/temporalio/temporal)

### 9.2 学习资源

- [Temporal教程](https://learn.temporal.io/)
- [Temporal示例代码](https://github.com/temporalio/samples-go)
- [Temporal社区Discord](https://discord.gg/temporal)

### 9.3 生产案例

- [OpenAI使用Temporal](https://temporal.io/customers/openai)
- [Temporal案例集](https://temporal.io/customers)

## 十、更新记录

| 版本 | 日期 | 更新内容 |
|------|------|----------|
| v1.0.0 | 2026-06-24 | 初稿（纯Go + Temporal架构） |
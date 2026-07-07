# Agent自主决策工具选择架构重构方案

## 一、当前架构问题分析

### 1.1 当前流程

```
用户消息 
  → ChatService.SendMessage()
    → AgentRouter.RouteAgent() (关键词规则匹配)
    → 加载预绑定工具 (静态绑定)
    → LLM生成回复 (包含工具调用标记)
    → ChatService解析```tool_call
    → ChatService执行工具
    → LLM处理结果
```

### 1.2 核心问题

| 问题 | 当前实现 | 影响 |
|------|---------|------|
| **Agent选择** | 关键词规则匹配 (硬编码) | 无法理解复杂意图，不够智能 |
| **工具绑定** | 数据库静态绑定 (AgentTool表) | Agent无法动态决策使用哪些工具 |
| **工具调用** | ChatService硬编码解析 | Agent没有真正的工具执行能力 |
| **Agent角色** | 只是SystemPrompt容器 | 缺乏自主性和决策能力 |
| **执行流程** | 线性流程，ChatService主导 | 无法支持多Agent协作 |

---

## 二、新架构设计

### 2.1 核心理念

```
主LLM (路由器) → Agent实例 (决策者) → 工具 (执行者)
```

**三大原则：**
1. **LLM作为主路由器** - 理解用户意图，智能选择Agent
2. **Agent作为决策者** - 根据任务自主选择和调用工具
3. **工具作为执行者** - 提供原子化能力，由Agent调度

### 2.2 架构层次

```
┌─────────────────────────────────────────────────────────┐
│                  ChatService (协调层)                   │
│  - 接收用户消息                                          │
│  - 调用MasterRouter选择Agent                            │
│  - 返回结果                                              │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│              MasterRouter (智能路由层)                   │
│  - LLM理解用户意图                                       │
│  - 智能选择最合适的Agent                                 │
│  - 构建Agent实例                                         │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│            AgentRuntime (Agent运行时环境)               │
│  - Agent实例管理                                         │
│  - 工具注册表                                            │
│  - 执行上下文                                            │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│               AgentInstance (Agent实例)                  │
│  - 维护Agent配置 (SystemPrompt, 可用工具池)             │
│  - LLM决策：根据任务选择工具                             │
│  - 执行工具并处理结果                                    │
│  - 支持多轮对话和工具调用                                │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                  ToolRegistry (工具注册表)               │
│  - 工具定义和元数据                                      │
│  - 工具实例化                                            │
│  - 工具执行                                              │
└─────────────────────────────────────────────────────────┘
```

---

## 三、详细设计

### 3.1 MasterRouter (智能路由器)

**职责：**
- 使用LLM理解用户意图
- 从Agent池中选择最合适的Agent
- 返回Agent实例

**实现：**

```go
// MasterRouter - 智能路由器
type MasterRouter struct {
    agentSvc      *AgentService
    agentRuntime  *AgentRuntime
    llm           *llm.EinoLLM
}

type RoutingDecision struct {
    SelectedAgentID   string  `json:"selected_agent_id"`
    Confidence        float64 `json:"confidence"`
    Reasoning         string  `json:"reasoning"`
    AlternativeAgents []string `json:"alternative_agents"`
}

// Route - 智能选择Agent
func (r *MasterRouter) Route(ctx context.Context, userMessage string, sessionContext string) (*AgentInstance, *RoutingDecision, error) {
    // 1. 获取所有启用的Agent
    agents, err := r.agentSvc.ListEnabled()
    if err != nil {
        return nil, nil, err
    }
    
    // 2. 构建Agent选择Prompt
    prompt := r.buildRoutingPrompt(agents, userMessage, sessionContext)
    
    // 3. 调用LLM决策
    response, err := r.llm.Generate(ctx, prompt)
    if err != nil {
        return nil, nil, err
    }
    
    // 4. 解析LLM决策
    decision := r.parseRoutingDecision(response)
    
    // 5. 构建Agent实例
    agentInstance, err := r.agentRuntime.CreateAgentInstance(ctx, decision.SelectedAgentID)
    if err != nil {
        return nil, nil, err
    }
    
    return agentInstance, decision, nil
}

func (r *MasterRouter) buildRoutingPrompt(agents []model.Agent, userMsg string, ctx string) string {
    return fmt.Sprintf(`
你是一个智能路由助手，需要根据用户问题选择最合适的Agent来处理。

## 可用的Agent列表：
%s

## 会话上下文：
%s

## 用户问题：
%s

## 任务：
1. 分析用户问题的意图和需求
2. 根据Agent的描述和能力选择最合适的Agent
3. 给出选择的理由和置信度

## 输出格式（JSON）：
{
  "selected_agent_id": "agent-id",
  "confidence": 0.95,
  "reasoning": "选择理由...",
  "alternative_agents": ["agent-id-2", "agent-id-3"]
}

请直接输出JSON，不要包含其他内容。
`, r.formatAgentList(agents), ctx, userMsg)
}
```

### 3.2 AgentRuntime (Agent运行时)

**职责：**
- 管理Agent实例的生命周期
- 维护工具注册表
- 提供执行上下文

**实现：**

```go
// AgentRuntime - Agent运行时环境
type AgentRuntime struct {
    toolRegistry  *ToolRegistry
    agentSvc      *AgentService
    toolSvc       *ToolService
    einoToolSvc   *EinoToolService
    llm           *llm.EinoLLM
    mcpSvc        *MCPService
}

// CreateAgentInstance - 创建Agent实例
func (r *AgentRuntime) CreateAgentInstance(ctx context.Context, agentID string) (*AgentInstance, error) {
    // 1. 获取Agent配置
    agentModel, err := r.agentSvc.GetByID(agentID)
    if err != nil {
        return nil, err
    }
    
    // 2. 获取Agent可用工具池（工具定义，不是实例）
    toolPool, err := r.toolSvc.GetAgentToolPool(agentID)
    if err != nil {
        return nil, err
    }
    
    // 3. 获取MCP工具池
    var mcpToolPool []model.Tool
    if agentModel.MCPServerIDs != "" {
        mcpToolPool, err = r.loadMCPToolPool(ctx, agentModel.MCPServerIDs)
        if err != nil {
            return nil, err
        }
    }
    
    // 4. 合并工具池
    allTools := append(toolPool, mcpToolPool...)
    
    // 5. 创建Agent实例
    instance := &AgentInstance{
        AgentModel:    agentModel,
        AvailableTools: allTools,
        toolRegistry:   r.toolRegistry,
        llm:            r.llm,
        maxToolCalls:   5, // 最大工具调用次数
    }
    
    return instance, nil
}
```

### 3.3 AgentInstance (Agent实例)

**职责：**
- 维护Agent配置
- LLM决策选择工具
- 执行工具并处理结果
- 支持多轮对话和工具调用循环

**实现：**

```go
// AgentInstance - Agent实例
type AgentInstance struct {
    AgentModel     *model.Agent
    AvailableTools []model.Tool  // 可用工具池
    toolRegistry   *ToolRegistry
    llm            *llm.EinoLLM
    maxToolCalls   int
    callHistory    []ToolCallRecord
}

type ToolCallRecord struct {
    ToolName   string
    Arguments  map[string]interface{}
    Result     string
    Timestamp  time.Time
}

// Execute - 执行Agent任务
func (a *AgentInstance) Execute(ctx context.Context, userMessage string, history []model.ChatMessage) (string, []ToolCallRecord, error) {
    // 1. 构建初始Prompt
    prompt := a.buildInitialPrompt(userMessage, history)
    
    // 2. LLM生成回复
    response, err := a.llm.Generate(ctx, prompt)
    if err != nil {
        return "", nil, err
    }
    
    // 3. 检查是否需要工具调用
    toolCalls := a.parseToolCalls(response)
    
    // 4. 如果没有工具调用，直接返回
    if len(toolCalls) == 0 {
        return response, nil, nil
    }
    
    // 5. 工具调用循环
    callCount := 0
    currentResponse := response
    
    for len(toolCalls) > 0 && callCount < a.maxToolCalls {
        // 执行工具调用
        toolResults := a.executeTools(ctx, toolCalls)
        
        // 记录调用历史
        a.callHistory = append(a.callHistory, toolResults...)
        
        // 让LLM处理工具结果
        processPrompt := a.buildToolProcessPrompt(userMessage, toolResults, currentResponse)
        currentResponse, err = a.llm.Generate(ctx, processPrompt)
        if err != nil {
            return "", a.callHistory, err
        }
        
        // 检查是否还需要工具调用
        toolCalls = a.parseToolCalls(currentResponse)
        callCount++
    }
    
    return currentResponse, a.callHistory, nil
}

// buildInitialPrompt - 构建初始Prompt
func (a *AgentInstance) buildInitialPrompt(userMsg string, history []model.ChatMessage) string {
    var promptBuilder strings.Builder
    
    // 1. Agent的SystemPrompt
    promptBuilder.WriteString(a.AgentModel.SystemPrompt)
    promptBuilder.WriteString("\n\n")
    
    // 2. 可用工具说明
    promptBuilder.WriteString("## 你可以使用的工具：\n\n")
    for _, tool := range a.AvailableTools {
        promptBuilder.WriteString(fmt.Sprintf("### %s\n", tool.Name))
        promptBuilder.WriteString(fmt.Sprintf("描述: %s\n", tool.Description))
        promptBuilder.WriteString(fmt.Sprintf("参数: %s\n\n", tool.ParametersSchema))
    }
    
    // 3. 工具调用格式说明
    promptBuilder.WriteString(`
## 工具调用格式：
当你需要使用工具时，请按以下格式输出：

\`\`\`tool_call
{
  "tool": "工具名称",
  "arguments": {
    "参数名": "参数值"
  }
}
\`\`\`

你可以根据需要多次调用不同的工具。

`)
    
    // 4. 历史对话
    if len(history) > 0 {
        promptBuilder.WriteString("## 历史对话：\n\n")
        for _, msg := range history {
            if msg.Role == "user" {
                promptBuilder.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
            } else {
                promptBuilder.WriteString(fmt.Sprintf("助手: %s\n", msg.Content))
            }
        }
    }
    
    // 5. 当前用户问题
    promptBuilder.WriteString(fmt.Sprintf("\n## 用户问题：\n%s\n\n请回答：", userMsg))
    
    return promptBuilder.String()
}

// executeTools - 执行工具调用
func (a *AgentInstance) executeTools(ctx context.Context, calls []ToolCall) []ToolCallRecord {
    var records []ToolCallRecord
    
    for _, call := range calls {
        // 从注册表获取工具
        tool, err := a.toolRegistry.GetTool(call.ToolName)
        if err != nil {
            records = append(records, ToolCallRecord{
                ToolName:  call.ToolName,
                Arguments: call.Arguments,
                Result:    fmt.Sprintf("错误: 工具未找到 - %v", err),
                Timestamp: time.Now(),
            })
            continue
        }
        
        // 执行工具
        result, err := tool.Execute(ctx, call.Arguments)
        if err != nil {
            result = fmt.Sprintf("执行失败: %v", err)
        }
        
        records = append(records, ToolCallRecord{
            ToolName:  call.ToolName,
            Arguments: call.Arguments,
            Result:    result,
            Timestamp: time.Now(),
        })
    }
    
    return records
}

// parseToolCalls - 解析工具调用
func (a *AgentInstance) parseToolCalls(text string) []ToolCall {
    var calls []ToolCall
    
    // 解析 ```tool_call 块
    start := 0
    for {
        idx := strings.Index(text[start:], "```tool_call\n")
        if idx == -1 {
            break
        }
        start += idx + len("```tool_call\n")
        
        endIdx := strings.Index(text[start:], "```")
        if endIdx == -1 {
            break
        }
        
        jsonStr := text[start : start+endIdx]
        start += endIdx + 3
        
        var call ToolCall
        if err := json.Unmarshal([]byte(jsonStr), &call); err != nil {
            continue
        }
        
        calls = append(calls, call)
    }
    
    return calls
}
```

### 3.4 ToolRegistry (工具注册表)

**职责：**
- 管理所有工具定义
- 工具实例化
- 工具执行

**实现：**

```go
// ToolRegistry - 工具注册表
type ToolRegistry struct {
    tools      map[string]*ToolWrapper
    toolSvc    *ToolService
    einoToolSvc *EinoToolService
    mcpSvc     *MCPService
    hostRepo   *repository.HostRepository
}

type ToolWrapper struct {
    ToolModel  *model.Tool
    Instance   tool.InvokableTool
}

// RegisterTool - 注册工具
func (r *ToolRegistry) RegisterTool(toolModel *model.Tool) error {
    wrapper := &ToolWrapper{
        ToolModel: toolModel,
    }
    
    // 根据工具类型创建实例
    switch toolModel.Name {
    case "ssh_exec":
        wrapper.Instance = eino_tools.NewSSHTool(toolModel, nil, r.hostRepo)
    case "prometheus_query":
        wrapper.Instance = eino_tools.NewPrometheusTool(toolModel, nil)
    case "kubernetes_query":
        wrapper.Instance = eino_tools.NewKubernetesTool(toolModel, nil)
    case "log_query":
        wrapper.Instance = eino_tools.NewLogQueryTool(toolModel, nil)
    default:
        return fmt.Errorf("未知工具类型: %s", toolModel.Name)
    }
    
    r.tools[toolModel.Name] = wrapper
    return nil
}

// GetTool - 获取工具
func (r *ToolRegistry) GetTool(name string) (*ToolWrapper, error) {
    tool, ok := r.tools[name]
    if !ok {
        return nil, fmt.Errorf("工具未注册: %s", name)
    }
    return tool, nil
}

// Execute - 执行工具
func (w *ToolWrapper) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    result, err := w.Instance.Invoke(ctx, args)
    if err != nil {
        return "", err
    }
    return fmt.Sprintf("%v", result), nil
}
```

### 3.5 重构后的ChatService

```go
// ChatService - 重构后
type ChatService struct {
    repo         *repository.ChatRepository
    masterRouter *MasterRouter  // 智能路由器
    agentRuntime *AgentRuntime   // Agent运行时
    ragSvc       *RAGService
    tokenSvc     *TokenService
    maxCtx       int
    enableRAG    bool
}

// SendMessage - 重构后的实现
func (s *ChatService) SendMessage(ctx context.Context, sessionID, content string) (string, *model.ChatMessage, *model.ChatMessage, []map[string]interface{}, error) {
    // 1. 获取会话和历史消息
    session, err := s.repo.GetSessionByID(sessionID)
    if err != nil {
        return "", nil, nil, nil, err
    }
    
    history, err := s.repo.GetRecentMessages(sessionID, s.maxCtx)
    if err != nil {
        return "", nil, nil, nil, err
    }
    
    // 2. RAG检索（可选）
    var ragReferences []map[string]interface{}
    var knowledgeContext string
    if s.enableRAG && s.ragSvc != nil {
        knowledgeContext, ragReferences = s.retrieveKnowledge(ctx, content)
    }
    
    // 3. 构建会话上下文
    sessionContext := s.buildSessionContext(history, knowledgeContext)
    
    // 4. 智能路由选择Agent
    agentInstance, routingDecision, err := s.masterRouter.Route(ctx, content, sessionContext)
    if err != nil {
        return "", nil, nil, nil, err
    }
    
    // 5. Agent执行任务
    response, toolCalls, err := agentInstance.Execute(ctx, content, history)
    if err != nil {
        return "", nil, nil, nil, err
    }
    
    // 6. 保存消息
    userMsg, aiMsg, err := s.saveMessages(sessionID, content, response, ragReferences, routingDecision, toolCalls)
    if err != nil {
        return "", nil, nil, nil, err
    }
    
    return response, userMsg, aiMsg, ragReferences, nil
}
```

---

## 四、数据模型变更

### 4.1 Agent模型增强

```go
// model/agent.go - 增强字段
type Agent struct {
    ID              string    `json:"id" gorm:"primaryKey"`
    Name            string    `json:"name"`
    Type            string    `json:"type"`
    Avatar          string    `json:"avatar"`
    Role            string    `json:"role"`
    Category        string    `json:"category"`
    Description     string    `json:"description"`
    SystemPrompt    string    `json:"system_prompt"`
    Model           string    `json:"model"`
    Temperature     float64   `json:"temperature"`
    
    // 新增字段
    MaxToolCalls    int       `json:"max_tool_calls" gorm:"default:5"`     // 最大工具调用次数
    ToolSelectionStrategy string `json:"tool_selection_strategy"`         // 工具选择策略: auto/manual/llm
    Capability      string    `json:"capability"`                        // 能力标签(JSON)
    Priority        int       `json:"priority"`                           // 路由优先级
    
    IsPreset        bool      `json:"is_preset"`
    Enabled         bool      `json:"enabled"`
    MCPServerIDs    string    `json:"mcp_server_ids"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

### 4.2 新增ToolSelection模型

```go
// model/tool_selection.go - 记录Agent的工具选择决策
type ToolSelection struct {
    ID            string    `json:"id" gorm:"primaryKey"`
    SessionID     string    `json:"session_id" gorm:"index"`
    AgentID       string    `json:"agent_id" gorm:"index"`
    UserMessage   string    `json:"user_message"`
    SelectedTools string    `json:"selected_tools"` // JSON: ["tool1", "tool2"]
    Reasoning     string    `json:"reasoning"`       // LLM的选择理由
    Confidence    float64   `json:"confidence"`
    CreatedAt     time.Time `json:"created_at"`
}
```

### 4.3 新增RoutingLog模型

```go
// model/routing_log.go - 记录路由决策日志
type RoutingLog struct {
    ID              string    `json:"id" gorm:"primaryKey"`
    SessionID       string    `json:"session_id" gorm:"index"`
    UserMessage     string    `json:"user_message"`
    SelectedAgentID string    `json:"selected_agent_id"`
    Confidence      float64   `json:"confidence"`
    Reasoning       string    `json:"reasoning"`
    AlternativeAgents string  `json:"alternative_agents"` // JSON
    RoutingMethod   string    `json:"routing_method"`     // llm/rule/fallback
    CreatedAt       time.Time `json:"created_at"`
}
```

---

## 五、迁移计划

### 5.1 阶段一：基础设施（1-2周）

**目标：** 建立新架构的基础组件

**任务：**
1. ✅ 创建 `ToolRegistry` 工具注册表
2. ✅ 创建 `AgentInstance` Agent实例接口
3. ✅ 创建 `AgentRuntime` Agent运行时
4. ✅ 创建 `MasterRouter` 智能路由器
5. ✅ 添加数据模型字段和迁移脚本

**文件清单：**
```
backend/internal/service/
  ├── tool_registry.go        (新增)
  ├── agent_instance.go       (新增)
  ├── agent_runtime.go         (新增)
  ├── master_router.go         (新增)
  
backend/internal/model/
  ├── routing_log.go           (新增)
  ├── tool_selection.go        (新增)
  ├── agent.go                 (修改)
```

### 5.2 阶段二：并行运行（2-3周）

**目标：** 新旧架构并行运行，验证新架构

**任务：**
1. ✅ 实现新的 `ChatService` 方法（`SendMessageV2`）
2. ✅ 添加特性开关，控制使用新旧架构
3. ✅ 修改前端，支持调用新接口
4. ✅ 收集日志和指标对比新旧架构

**配置开关：**
```yaml
agent:
  use_new_architecture: true  # 开启新架构
  fallback_to_old: true       # 失败时回退旧架构
```

**API端点：**
```
POST /api/v2/chat/send      # 新架构
POST /api/v1/chat/send      # 旧架构（保留）
```

### 5.3 阶段三：全面切换（1-2周）

**目标：** 完全切换到新架构

**任务：**
1. ✅ 监控新架构稳定性和性能
2. ✅ 修复新架构的问题
3. ✅ 迁移所有用户到新架构
4. ✅ 删除旧架构代码

**监控指标：**
- Agent选择准确率
- 工具调用成功率
- 平均响应时间
- 用户满意度

### 5.4 阶段四：优化迭代（持续）

**目标：** 持续优化新架构

**优化方向：**
1. Agent选择策略优化（强化学习）
2. 工具选择策略优化
3. 多Agent协作支持
4. Agent能力热更新

---

## 六、关键代码示例

### 6.1 MasterRouter完整实现

```go
// master_router.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    
    "github.com/aiops/AiOpsHub/backend/internal/model"
    "github.com/aiops/AiOpsHub/backend/pkg/llm"
    "github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type MasterRouter struct {
    agentSvc     *AgentService
    agentRuntime *AgentRuntime
    llm          *llm.EinoLLM
}

func NewMasterRouter(agentSvc *AgentService, runtime *AgentRuntime, llm *llm.EinoLLM) *MasterRouter {
    return &MasterRouter{
        agentSvc:     agentSvc,
        agentRuntime: runtime,
        llm:          llm,
    }
}

func (r *MasterRouter) Route(ctx context.Context, userMessage string, sessionContext string) (*AgentInstance, *model.RoutingLog, error) {
    // 1. 获取所有启用的Agent
    agents, err := r.agentSvc.ListEnabled()
    if err != nil {
        return nil, nil, fmt.Errorf("获取Agent列表失败: %w", err)
    }
    
    if len(agents) == 0 {
        return nil, nil, fmt.Errorf("没有可用的Agent")
    }
    
    // 2. 快速路由：关键词匹配（作为预筛选）
    candidateAgents := r.quickFilter(userMessage, agents)
    if len(candidateAgents) == 0 {
        candidateAgents = agents // 如果没有匹配，使用全部
    }
    
    // 3. LLM智能路由
    routingPrompt := r.buildRoutingPrompt(candidateAgents, userMessage, sessionContext)
    
    response, err := r.llm.Generate(ctx, routingPrompt)
    if err != nil {
        logger.Error(fmt.Sprintf("LLM路由失败: %v", err))
        // 降级：使用第一个候选Agent
        if len(candidateAgents) > 0 {
            instance, err := r.agentRuntime.CreateAgentInstance(ctx, candidateAgents[0].ID)
            return instance, r.buildFallbackLog(userMessage, candidateAgents[0].ID, "llm_error"), err
        }
        return nil, nil, err
    }
    
    // 4. 解析LLM决策
    decision := r.parseRoutingDecision(response)
    if decision.SelectedAgentID == "" {
        // 未找到合适Agent，使用默认
        if len(candidateAgents) > 0 {
            decision.SelectedAgentID = candidateAgents[0].ID
        }
    }
    
    // 5. 创建Agent实例
    agentInstance, err := r.agentRuntime.CreateAgentInstance(ctx, decision.SelectedAgentID)
    if err != nil {
        return nil, nil, fmt.Errorf("创建Agent实例失败: %w", err)
    }
    
    logger.Info(fmt.Sprintf("✅ 路由决策: Agent=%s, 置信度=%.2f, 理由=%s", 
        decision.SelectedAgentID, decision.Confidence, decision.Reasoning))
    
    return agentInstance, r.buildRoutingLog(userMessage, decision), nil
}

func (r *MasterRouter) quickFilter(userMessage string, agents []model.Agent) []model.Agent {
    // 简单的关键词预筛选，减少LLM调用次数
    var candidates []model.Agent
    
    messageLower := strings.ToLower(userMessage)
    
    for _, agent := range agents {
        // 检查Agent的名称、类别、描述是否与消息相关
        if strings.Contains(messageLower, strings.ToLower(agent.Name)) ||
           strings.Contains(messageLower, strings.ToLower(agent.Category)) ||
           strings.Contains(messageLower, strings.ToLower(agent.Role)) {
            candidates = append(candidates, agent)
        }
    }
    
    return candidates
}

func (r *MasterRouter) buildRoutingPrompt(agents []model.Agent, userMsg string, ctx string) string {
    var agentList strings.Builder
    for i, agent := range agents {
        agentList.WriteString(fmt.Sprintf("%d. **%s** (ID: %s)\n", i+1, agent.Name, agent.ID))
        agentList.WriteString(fmt.Sprintf("   - 角色: %s\n", agent.Role))
        agentList.WriteString(fmt.Sprintf("   - 类别: %s\n", agent.Category))
        agentList.WriteString(fmt.Sprintf("   - 描述: %s\n", agent.Description))
        agentList.WriteString(fmt.Sprintf("   - 能力: %s\n\n", agent.Capability))
    }
    
    return fmt.Sprintf(`
# Agent智能路由任务

你是一个智能路由助手，需要根据用户问题选择最合适的Agent来处理。

## 可用的Agent：
%s

## 会话上下文：
%s

## 用户问题：
%s

## 任务：
1. 分析用户问题的意图和需求
2. 根据Agent的角色、类别、描述和能力，选择最合适的Agent
3. 给出选择的理由和置信度

## 输出格式（严格JSON，不要包含其他内容）：
{
  "selected_agent_id": "agent-id",
  "confidence": 0.95,
  "reasoning": "选择理由...",
  "alternative_agents": ["agent-id-2", "agent-id-3"]
}

请直接输出JSON：
`, agentList.String(), ctx, userMsg)
}

func (r *MasterRouter) parseRoutingDecision(response string) *RoutingDecision {
    // 提取JSON部分
    start := strings.Index(response, "{")
    end := strings.LastIndex(response, "}")
    
    if start == -1 || end == -1 {
        return &RoutingDecision{
            SelectedAgentID: "",
            Confidence:      0.0,
            Reasoning:       "无法解析LLM响应",
        }
    }
    
    jsonStr := response[start : end+1]
    
    var decision RoutingDecision
    if err := json.Unmarshal([]byte(jsonStr), &decision); err != nil {
        logger.Error(fmt.Sprintf("解析路由决策失败: %v", err))
        return &RoutingDecision{
            SelectedAgentID: "",
            Confidence:      0.0,
            Reasoning:       fmt.Sprintf("解析失败: %v", err),
        }
    }
    
    return &decision
}

type RoutingDecision struct {
    SelectedAgentID   string   `json:"selected_agent_id"`
    Confidence        float64  `json:"confidence"`
    Reasoning         string   `json:"reasoning"`
    AlternativeAgents []string `json:"alternative_agents"`
}

func (r *MasterRouter) buildRoutingLog(userMsg string, decision *RoutingDecision) *model.RoutingLog {
    alternatives, _ := json.Marshal(decision.AlternativeAgents)
    
    return &model.RoutingLog{
        UserMessage:       userMsg,
        SelectedAgentID:   decision.SelectedAgentID,
        Confidence:        decision.Confidence,
        Reasoning:         decision.Reasoning,
        AlternativeAgents: string(alternatives),
        RoutingMethod:     "llm",
    }
}

func (r *MasterRouter) buildFallbackLog(userMsg string, agentID string, reason string) *model.RoutingLog {
    return &model.RoutingLog{
        UserMessage:     userMsg,
        SelectedAgentID: agentID,
        Confidence:      0.5,
        Reasoning:       reason,
        RoutingMethod:   "fallback",
    }
}
```

### 6.2 AgentInstance完整实现

见上文3.3节。

---

## 七、性能优化

### 7.1 Agent缓存

```go
// AgentRuntime中添加缓存
type AgentRuntime struct {
    // ...
    cache *lru.Cache  // Agent实例缓存
}

func (r *AgentRuntime) CreateAgentInstance(ctx context.Context, agentID string) (*AgentInstance, error) {
    // 尝试从缓存获取
    if cached, ok := r.cache.Get(agentID); ok {
        return cached.(*AgentInstance), nil
    }
    
    // 创建新实例
    instance, err := r.createInstance(ctx, agentID)
    if err != nil {
        return nil, err
    }
    
    // 加入缓存
    r.cache.Add(agentID, instance)
    
    return instance, nil
}
```

### 7.2 工具预加载

```go
// 启动时预加载所有工具
func (r *ToolRegistry) PreloadTools(ctx context.Context) error {
    tools, err := r.toolSvc.ListAll()
    if err != nil {
        return err
    }
    
    for _, tool := range tools {
        if err := r.RegisterTool(&tool); err != nil {
            logger.Error(fmt.Sprintf("预加载工具失败: %s, %v", tool.Name, err))
            continue
        }
    }
    
    logger.Info(fmt.Sprintf("预加载完成，共加载 %d 个工具", len(tools)))
    return nil
}
```

### 7.3 LLM调用优化

```go
// 减少LLM调用次数的策略
func (r *MasterRouter) Route(ctx context.Context, userMessage string, sessionContext string) (*AgentInstance, *model.RoutingLog, error) {
    // 1. 先尝试快速规则匹配
    if agent := r.matchByRules(userMessage); agent != nil {
        instance, err := r.agentRuntime.CreateAgentInstance(ctx, agent.ID)
        return instance, r.buildRuleLog(userMessage, agent.ID), err
    }
    
    // 2. 再使用LLM智能路由
    return r.routeByLLM(ctx, userMessage, sessionContext)
}
```

---

## 八、监控与可观测性

### 8.1 关键指标

```go
// 监控指标
type Metrics struct {
    // 路由指标
    RoutingAccuracy    float64  // Agent选择准确率
    RoutingLatency     float64  // 路由延迟
    
    // Agent指标
    AgentSuccessRate   float64  // Agent执行成功率
    AgentLatency       float64  // Agent执行延迟
    
    // 工具指标
    ToolCallSuccessRate float64 // 工具调用成功率
    ToolCallLatency     float64 // 工具调用延迟
    AvgToolCallsPerMsg  float64 // 平均每条消息的工具调用次数
}
```

### 8.2 日志记录

```go
// 记录路由决策
func (s *ChatService) logRouting(log *model.RoutingLog) {
    if err := s.routingLogRepo.Create(log); err != nil {
        logger.Error(fmt.Sprintf("保存路由日志失败: %v", err))
    }
}

// 记录工具选择
func (a *AgentInstance) logToolSelection(sessionID string, tools []string, reasoning string) {
    selection := &model.ToolSelection{
        SessionID:     sessionID,
        AgentID:       a.AgentModel.ID,
        UserMessage:   userMsg,
        SelectedTools: tools,
        Reasoning:     reasoning,
    }
    // 保存到数据库
}
```

---

## 九、测试策略

### 9.1 单元测试

```go
// master_router_test.go
func TestMasterRouter_Route(t *testing.T) {
    tests := []struct {
        name          string
        userMessage   string
        expectedAgent string
    }{
        {
            name:          "告警处理",
            userMessage:   "收到一条严重告警，需要分析",
            expectedAgent: "preset-alert-handler",
        },
        {
            name:          "日志分析",
            userMessage:   "查看应用错误日志",
            expectedAgent: "preset-log-analyzer",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            router := NewMasterRouter(mockAgentSvc, mockRuntime, mockLLM)
            instance, log, err := router.Route(context.Background(), tt.userMessage, "")
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expectedAgent, instance.AgentModel.ID)
            assert.Greater(t, log.Confidence, 0.5)
        })
    }
}
```

### 9.2 集成测试

```go
// agent_instance_test.go
func TestAgentInstance_Execute(t *testing.T) {
    // 创建测试Agent实例
    instance := createTestAgentInstance()
    
    // 测试无工具调用
    response, calls, err := instance.Execute(context.Background(), "你好", nil)
    assert.NoError(t, err)
    assert.Empty(t, calls)
    
    // 测试工具调用
    response, calls, err = instance.Execute(context.Background(), "查询服务器状态", nil)
    assert.NoError(t, err)
    assert.NotEmpty(t, calls)
    assert.Contains(t, calls[0].ToolName, "ssh_exec")
}
```

### 9.3 压力测试

```go
// 并发测试
func TestChatService_ConcurrentRequests(t *testing.T) {
    var wg sync.WaitGroup
    errors := make(chan error, 100)
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _, _, _, _, err := chatSvc.SendMessage(context.Background(), "test-session", "测试消息")
            if err != nil {
                errors <- err
            }
        }()
    }
    
    wg.Wait()
    close(errors)
    
    for err := range errors {
        t.Errorf("并发请求失败: %v", err)
    }
}
```

---

## 十、回滚方案

### 10.1 快速回滚

```yaml
# config.yaml
agent:
  use_new_architecture: false  # 切换回旧架构
  fallback_to_old: true
```

### 10.2 数据兼容

- 新旧架构共享数据库
- AgentTool绑定表继续使用
- 新增的路由日志表不影响旧架构

---

## 十一、总结

### 核心改进

1. **智能路由** - LLM理解意图，自主选择Agent
2. **Agent实例化** - Agent成为真正的执行主体
3. **工具自主选择** - Agent根据任务动态选择工具
4. **可扩展性** - 支持多Agent协作

### 预期收益

- Agent选择准确率提升 30%+
- 工具调用成功率提升 20%+
- 响应延迟降低 15%+
- 代码可维护性显著提升

### 风险控制

- 灰度发布，逐步迁移
- 新旧架构并行运行
- 完善的监控和告警
- 快速回滚机制
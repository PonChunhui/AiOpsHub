package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/llm"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type CoordinatorAgent struct {
	ID             string
	Name           string
	Type           string
	Description    string
	Config         AgentConfig
	llm            *llm.EinoLLM
	DecisionEngine *DecisionEngine
	MessageBus     MessageBusInterface
	StateManager   StateManagerInterface
}

type CoordinatorInput struct {
	SessionID string                 `json:"session_id"`
	UserQuery string                 `json:"user_query"`
	Context   map[string]interface{} `json:"context"`
	Timestamp time.Time              `json:"timestamp"`
}

type CoordinatorOutput struct {
	SessionID        string            `json:"session_id"`
	Intent           string            `json:"intent"`
	TaskType         string            `json:"task_type"`
	SubTasks         []SubTask         `json:"sub_tasks"`
	Orchestration    OrchestrationPlan `json:"orchestration"`
	RequiresApproval bool              `json:"requires_approval"`
	Response         string            `json:"response"`
}

type SubTask struct {
	TaskID       string                 `json:"task_id"`
	TaskType     string                 `json:"task_type"`
	Description  string                 `json:"description"`
	AgentID      string                 `json:"agent_id"`
	Parameters   map[string]interface{} `json:"parameters"`
	Priority     int                    `json:"priority"`
	Dependencies []string               `json:"dependencies"`
}

type OrchestrationPlan struct {
	Strategy      string     `json:"strategy"` // "sequential", "parallel", "hybrid"
	TaskSequence  []string   `json:"task_sequence"`
	ParallelTasks [][]string `json:"parallel_tasks"`
	EstimatedTime int        `json:"estimated_time"`
}

func NewCoordinatorAgent(id, name string, config AgentConfig, decisionEngine *DecisionEngine) (*CoordinatorAgent, error) {
	coordinator := &CoordinatorAgent{
		ID:             id,
		Name:           name,
		Type:           "coordinator",
		Description:    "Coordinator Agent，负责任务分解、Agent调度、结果整合",
		Config:         config,
		DecisionEngine: decisionEngine,
	}

	llm, err := coordinator.createLLM(config)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create LLM for Coordinator: %v", err))
		return nil, fmt.Errorf("failed to create LLM: %w", err)
	}

	coordinator.llm = llm
	logger.Info(fmt.Sprintf("Created Coordinator Agent: %s (%s)", name, id))

	return coordinator, nil
}

func (c *CoordinatorAgent) createLLM(config AgentConfig) (*llm.EinoLLM, error) {
	llmConfig := llm.EinoLLMConfig{
		Model:       config.Model,
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
		Provider:    config.Provider,
		APIKey:      config.APIKey,
		BaseURL:     config.BaseURL,
	}

	einoLLM, err := llm.NewEinoLLM(llmConfig)
	if err != nil {
		return nil, err
	}

	return einoLLM, nil
}

func (c *CoordinatorAgent) UnderstandIntent(ctx context.Context, input CoordinatorInput) (string, string, error) {
	logger.Info(fmt.Sprintf("Coordinator understanding intent for query: %s", input.UserQuery))

	prompt := fmt.Sprintf(`分析以下用户请求，识别用户意图和任务类型：

用户请求：%s
上下文：%v

请分析：
1. 用户意图（如：故障诊断、监控查询、告警处理、自动修复等）
2. 任务类型（如：incident_handling、alert_dedup、monitoring、auto_remediation等）
3. 任务复杂度（简单/中等/复杂）

请以JSON格式返回：
{
  "intent": "...",
  "task_type": "...",
  "complexity": "..."
}`, input.UserQuery, input.Context)

	resp, err := c.llm.Generate(ctx, prompt)
	if err != nil {
		logger.Error(fmt.Sprintf("Intent understanding failed: %v", err))
		return "", "", err
	}

	var intentResult struct {
		Intent     string `json:"intent"`
		TaskType   string `json:"task_type"`
		Complexity string `json:"complexity"`
	}

	err = json.Unmarshal([]byte(resp), &intentResult)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse intent result: %v", err))
		return "", "", err
	}

	logger.Info(fmt.Sprintf("Intent: %s, TaskType: %s, Complexity: %s",
		intentResult.Intent, intentResult.TaskType, intentResult.Complexity))

	return intentResult.Intent, intentResult.TaskType, nil
}

func (c *CoordinatorAgent) DecomposeTask(ctx context.Context, input CoordinatorInput, taskType string) ([]SubTask, error) {
	logger.Info(fmt.Sprintf("Coordinator decomposing task: %s", taskType))

	prompt := fmt.Sprintf(`将以下任务分解为子任务序列：

用户请求：%s
任务类型：%s

请分解为具体的子任务，每个子任务包含：
1. 子任务类型（如：monitor_collect、analysis_diagnosis、alert_process、decision_execute、learning_optimize）
2. 子任务描述
3. 推荐的Agent ID（如：monitor-agent-001、analysis-agent-001等）
4. 子任务参数
5. 优先级（1-5，1最高）
6. 依赖关系（依赖哪些前置子任务）

请以JSON数组格式返回：
[
  {
    "task_id": "task-001",
    "task_type": "monitor_collect",
    "description": "采集订单服务CPU指标",
    "agent_id": "monitor-agent-001",
    "parameters": {"service": "order-service", "metric": "cpu_usage"},
    "priority": 1,
    "dependencies": []
  },
  ...
]`, input.UserQuery, taskType)

	resp, err := c.llm.Generate(ctx, prompt)
	if err != nil {
		logger.Error(fmt.Sprintf("Task decomposition failed: %v", err))
		return nil, err
	}

	var subTasks []SubTask
	err = json.Unmarshal([]byte(resp), &subTasks)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse subtasks: %v", err))
		return nil, err
	}

	logger.Info(fmt.Sprintf("Decomposed into %d subtasks", len(subTasks)))
	for _, task := range subTasks {
		logger.Info(fmt.Sprintf("Subtask: %s - %s (Agent: %s)", task.TaskID, task.Description, task.AgentID))
	}

	return subTasks, nil
}

func (c *CoordinatorAgent) OrchestrateCollaboration(subTasks []SubTask) (OrchestrationPlan, error) {
	logger.Info(fmt.Sprintf("Coordinator orchestrating collaboration for %d tasks", len(subTasks)))

	plan := c.DecisionEngine.DetermineOrchestrationStrategy(subTasks)

	logger.Info(fmt.Sprintf("Orchestration strategy: %s", plan.Strategy))
	logger.Info(fmt.Sprintf("Task sequence: %v", plan.TaskSequence))
	logger.Info(fmt.Sprintf("Parallel groups: %v", plan.ParallelTasks))

	return plan, nil
}

func (c *CoordinatorAgent) IntegrateResults(ctx context.Context, results []AgentResult) (string, error) {
	logger.Info(fmt.Sprintf("Coordinator integrating %d agent results", len(results)))

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to marshal results: %v", err))
		return "", err
	}

	prompt := fmt.Sprintf(`整合以下多个Agent的执行结果，生成综合报告：

Agent执行结果：%s

请整合结果：
1. 汇总各Agent的关键发现
2. 提炼核心结论和根因
3. 给出综合建议和下一步行动
4. 如果有冲突结果，请说明并给出最佳判断

请生成简洁的综合报告。`, string(resultsJSON))

	resp, err := c.llm.Generate(ctx, prompt)
	if err != nil {
		logger.Error(fmt.Sprintf("Result integration failed: %v", err))
		return "", err
	}

	logger.Info(fmt.Sprintf("Integrated result: %s", resp))

	return resp, nil
}

func (c *CoordinatorAgent) ResolveConflict(ctx context.Context, conflicts []ConflictInput) (string, error) {
	logger.Info(fmt.Sprintf("Coordinator resolving %d conflicts", len(conflicts)))

	if len(conflicts) == 0 {
		return "", nil
	}

	conflictsJSON, err := json.Marshal(conflicts)
	if err != nil {
		return "", err
	}

	conflictPrompt := fmt.Sprintf(`解决以下Agent协作冲突：

冲突情况：%s

请分析冲突：
1. 冲突类型（资源竞争、结果冲突、执行冲突）
2. 冲突原因
3. 解决策略建议（投票、优先级、人工决策）

请给出冲突解决方案。`, string(conflictsJSON))

	resp, err := c.llm.Generate(ctx, conflictPrompt)
	if err != nil {
		return "", err
	}

	return resp, nil
}

func (c *CoordinatorAgent) Execute(ctx context.Context, input CoordinatorInput) (CoordinatorOutput, error) {
	logger.Info(fmt.Sprintf("Coordinator executing for session: %s", input.SessionID))

	intent, taskType, err := c.UnderstandIntent(ctx, input)
	if err != nil {
		return CoordinatorOutput{}, err
	}

	subTasks, err := c.DecomposeTask(ctx, input, taskType)
	if err != nil {
		return CoordinatorOutput{}, err
	}

	orchestration, err := c.OrchestrateCollaboration(subTasks)
	if err != nil {
		return CoordinatorOutput{}, err
	}

	requiresApproval := c.DecisionEngine.RequiresHumanApproval(taskType, subTasks)

	output := CoordinatorOutput{
		SessionID:        input.SessionID,
		Intent:           intent,
		TaskType:         taskType,
		SubTasks:         subTasks,
		Orchestration:    orchestration,
		RequiresApproval: requiresApproval,
		Response:         fmt.Sprintf("任务已分解为%d个子任务，编排策略：%s", len(subTasks), orchestration.Strategy),
	}

	logger.Info(fmt.Sprintf("Coordinator execution completed for session: %s", input.SessionID))

	return output, nil
}

type AgentResult struct {
	AgentID   string                 `json:"agent_id"`
	TaskID    string                 `json:"task_id"`
	Result    map[string]interface{} `json:"result"`
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
}

type ConflictInput struct {
	ConflictType string                   `json:"conflict_type"`
	Agents       []string                 `json:"agents"`
	Resource     string                   `json:"resource"`
	Results      []map[string]interface{} `json:"results"`
	Description  string                   `json:"description"`
}

type MessageBusInterface interface {
	Publish(channel string, message interface{}) error
	Subscribe(channel string) (<-chan interface{}, error)
}

type StateManagerInterface interface {
	SetAgentState(agentID, sessionID, status string, progress int) error
	GetAgentState(agentID, sessionID string) (*AgentState, error)
	SetIntermediateResult(sessionID, agentID string, result map[string]interface{}) error
	GetIntermediateResult(sessionID, agentID string) (map[string]interface{}, error)
}

type AgentState struct {
	AgentID            string                 `json:"agent_id"`
	SessionID          string                 `json:"session_id"`
	Status             string                 `json:"status"`
	Progress           int                    `json:"progress"`
	CurrentTask        string                 `json:"current_task"`
	StartTime          time.Time              `json:"start_time"`
	UpdateTime         time.Time              `json:"update_time"`
	IntermediateResult map[string]interface{} `json:"intermediate_result"`
}

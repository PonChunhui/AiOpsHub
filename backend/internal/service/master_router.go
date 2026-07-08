package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/aiops/AiOpsHub/backend/pkg/llm"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/google/uuid"
)

type MasterRouter struct {
	agentSvc       *AgentService
	agentRuntime   *AgentRuntime
	llm            *llm.EinoLLM
	routingLogRepo *repository.RoutingLogRepository
}

type RoutingDecision struct {
	SelectedAgentID   string   `json:"selected_agent_id"`
	Confidence        float64  `json:"confidence"`
	Reasoning         string   `json:"reasoning"`
	AlternativeAgents []string `json:"alternative_agents,omitempty"`
}

func NewMasterRouter(agentSvc *AgentService, runtime *AgentRuntime, llm *llm.EinoLLM) *MasterRouter {
	return &MasterRouter{
		agentSvc:       agentSvc,
		agentRuntime:   runtime,
		llm:            llm,
		routingLogRepo: repository.NewRoutingLogRepository(),
	}
}

func (r *MasterRouter) Route(ctx context.Context, userMessage string, sessionContext string) (*AgentInstance, *model.RoutingLog, error) {
	logger.Info("=== MasterRouter: 开始路由决策 ===")

	agents, err := r.agentSvc.ListEnabled()
	if err != nil {
		return nil, nil, fmt.Errorf("获取Agent列表失败: %w", err)
	}

	if len(agents) == 0 {
		return nil, nil, fmt.Errorf("没有可用的Agent")
	}

	candidateAgents := r.quickFilter(userMessage, agents)
	if len(candidateAgents) == 0 {
		candidateAgents = agents
	}

	if len(candidateAgents) == 1 {
		instance, err := r.agentRuntime.CreateAgentInstance(ctx, candidateAgents[0].ID)
		log := r.buildQuickMatchLog(userMessage, candidateAgents[0].ID, "唯一候选")
		return instance, log, err
	}

	decision, err := r.llmRouting(ctx, candidateAgents, userMessage, sessionContext)
	if err != nil {
		logger.Error(fmt.Sprintf("LLM路由失败: %v，降级使用默认Agent", err))
		instance, err := r.agentRuntime.CreateAgentInstance(ctx, candidateAgents[0].ID)
		log := r.buildFallbackLog(userMessage, candidateAgents[0].ID, fmt.Sprintf("LLM失败: %v", err))
		return instance, log, err
	}

	instance, err := r.agentRuntime.CreateAgentInstance(ctx, decision.SelectedAgentID)
	if err != nil {
		return nil, nil, fmt.Errorf("创建Agent实例失败: %w", err)
	}

	log := r.buildRoutingLog(userMessage, decision, "llm")

	if err := r.routingLogRepo.Create(log); err != nil {
		logger.Error(fmt.Sprintf("保存路由日志失败: %v", err))
	}

	logger.Info(fmt.Sprintf("✅ 路由决策完成: Agent=%s, 置信度=%.2f",
		decision.SelectedAgentID, decision.Confidence))

	return instance, log, nil
}

func (r *MasterRouter) quickFilter(userMessage string, agents []model.Agent) []model.Agent {
	var candidates []model.Agent
	messageLower := strings.ToLower(userMessage)

	for _, agent := range agents {
		if strings.Contains(messageLower, strings.ToLower(agent.Name)) ||
			strings.Contains(messageLower, strings.ToLower(agent.Category)) ||
			strings.Contains(messageLower, strings.ToLower(agent.Role)) {
			candidates = append(candidates, agent)
		}
	}

	return candidates
}

func (r *MasterRouter) llmRouting(ctx context.Context, agents []model.Agent, userMsg string, sessionCtx string) (*RoutingDecision, error) {
	prompt := r.buildRoutingPrompt(agents, userMsg, sessionCtx)

	response, err := r.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM生成失败: %w", err)
	}

	decision := r.parseRoutingDecision(response)

	if decision.SelectedAgentID == "" {
		return nil, fmt.Errorf("LLM未返回有效的Agent ID")
	}

	found := false
	for _, agent := range agents {
		if agent.ID == decision.SelectedAgentID {
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("LLM选择的Agent不存在: %s", decision.SelectedAgentID)
	}

	return decision, nil
}

func (r *MasterRouter) buildRoutingPrompt(agents []model.Agent, userMsg string, ctx string) string {
	var agentList strings.Builder
	for i, agent := range agents {
		agentList.WriteString(fmt.Sprintf("%d. **%s** (ID: `%s`)\n", i+1, agent.Name, agent.ID))
		agentList.WriteString(fmt.Sprintf("   - 角色: %s\n", agent.Role))
		agentList.WriteString(fmt.Sprintf("   - 类别: %s\n", agent.Category))
		agentList.WriteString(fmt.Sprintf("   - 描述: %s\n\n", agent.Description))
	}

	sessionContext := ""
	if ctx != "" {
		sessionContext = fmt.Sprintf("\n## 会话上下文：\n%s\n", ctx)
	}

	return fmt.Sprintf(`
# Agent智能路由任务

你是一个智能路由助手，需要根据用户问题选择最合适的Agent来处理。

## 可用的Agent：
%s
%s
## 用户问题：
%s

## 任务：
1. 分析用户问题的意图和需求
2. 根据Agent的角色、类别、描述，选择最合适的Agent
3. 给出选择的理由和置信度（0-1之间）

## 输出格式（严格JSON）：
{
  "selected_agent_id": "agent-id",
  "confidence": 0.95,
  "reasoning": "选择理由..."
}

**重要提示**：
- 只输出JSON，不要包含任何其他内容
- selected_agent_id必须是上面列出的Agent ID之一

请输出JSON：
`, agentList.String(), sessionContext, userMsg)
}

func (r *MasterRouter) parseRoutingDecision(response string) *RoutingDecision {
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start == -1 || end == -1 || end <= start {
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

func (r *MasterRouter) buildRoutingLog(userMsg string, decision *RoutingDecision, method string) *model.RoutingLog {
	alternativesJSON := "[]"
	if len(decision.AlternativeAgents) > 0 {
		arr, _ := json.Marshal(decision.AlternativeAgents)
		alternativesJSON = string(arr)
	}

	return &model.RoutingLog{
		ID:                uuid.New().String(),
		UserMessage:       userMsg,
		SelectedAgentID:   decision.SelectedAgentID,
		Confidence:        decision.Confidence,
		Reasoning:         decision.Reasoning,
		AlternativeAgents: alternativesJSON,
		RoutingMethod:     method,
		CreatedAt:         time.Now(),
	}
}

func (r *MasterRouter) buildFallbackLog(userMsg, agentID, reason string) *model.RoutingLog {
	return &model.RoutingLog{
		ID:              uuid.New().String(),
		UserMessage:     userMsg,
		SelectedAgentID: agentID,
		Confidence:      0.5,
		Reasoning:       reason,
		RoutingMethod:   "fallback",
		CreatedAt:       time.Now(),
	}
}

func (r *MasterRouter) buildQuickMatchLog(userMsg, agentID, reason string) *model.RoutingLog {
	return &model.RoutingLog{
		ID:              uuid.New().String(),
		UserMessage:     userMsg,
		SelectedAgentID: agentID,
		Confidence:      0.9,
		Reasoning:       reason,
		RoutingMethod:   "quick_match",
		CreatedAt:       time.Now(),
	}
}

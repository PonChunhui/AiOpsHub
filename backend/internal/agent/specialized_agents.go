package agent

import (
	"context"
	"fmt"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type MonitorAgent struct {
	*BaseAgent
}

func NewMonitorAgent(config AgentConfig) (*MonitorAgent, error) {
	baseAgent, err := NewBaseAgent(
		"monitor-agent-001",
		"MonitorAgent",
		"monitor",
		"监控采集Agent，负责采集监控数据、检测异常",
		config,
	)
	if err != nil {
		return nil, err
	}
	return &MonitorAgent{BaseAgent: baseAgent}, nil
}

func (m *MonitorAgent) Execute(ctx context.Context, input AgentInput) (AgentOutput, error) {
	logger.Info(fmt.Sprintf("MonitorAgent executing: %s", input.TaskType))
	return m.BaseAgent.Execute(ctx, input)
}

type AnalysisAgent struct {
	*BaseAgent
}

func NewAnalysisAgent(config AgentConfig) (*AnalysisAgent, error) {
	baseAgent, err := NewBaseAgent(
		"analysis-agent-001",
		"AnalysisAgent",
		"analysis",
		"分析诊断Agent，负责根因分析和故障诊断",
		config,
	)
	if err != nil {
		return nil, err
	}
	return &AnalysisAgent{BaseAgent: baseAgent}, nil
}

func (a *AnalysisAgent) Execute(ctx context.Context, input AgentInput) (AgentOutput, error) {
	logger.Info(fmt.Sprintf("AnalysisAgent executing: %s", input.TaskType))

	prompt := fmt.Sprintf(`分析以下监控数据或故障信息，定位根本原因：

输入：%v

请分析：
1. 识别关键异常指标
2. 分析可能的根本原因
3. 检查系统依赖关系
4. 给出根因定位结果和证据

返回根因分析报告（JSON格式）。`, input.Input)

	resp, err := a.llm.Generate(ctx, prompt)
	if err != nil {
		return AgentOutput{Status: "error"}, err
	}

	return AgentOutput{
		Result:  map[string]interface{}{"root_cause_analysis": resp},
		Status:  "completed",
		Message: "根因分析完成",
	}, nil
}

type AlertAgent struct {
	*BaseAgent
}

func NewAlertAgent(config AgentConfig) (*AlertAgent, error) {
	baseAgent, err := NewBaseAgent(
		"alert-agent-001",
		"AlertAgent",
		"alert",
		"告警处理Agent，负责告警降噪、智能聚合、智能分派",
		config,
	)
	if err != nil {
		return nil, err
	}
	return &AlertAgent{BaseAgent: baseAgent}, nil
}

func (a *AlertAgent) Execute(ctx context.Context, input AgentInput) (AgentOutput, error) {
	logger.Info(fmt.Sprintf("AlertAgent executing: %s", input.TaskType))

	prompt := fmt.Sprintf(`处理以下告警，进行智能降噪和聚合：

告警列表：%v

请处理：
1. 语义去重：识别相同或相似的告警
2. 智能聚合：将相关告警合并
3. 严重性评估：P0/P1/P2/P3分级
4. 智能分派：推荐合适的处理人

返回降噪后的告警列表和处理建议（JSON格式）。`, input.Input)

	resp, err := a.llm.Generate(ctx, prompt)
	if err != nil {
		return AgentOutput{Status: "error"}, err
	}

	return AgentOutput{
		Result:  map[string]interface{}{"alert_processing": resp},
		Status:  "completed",
		Message: "告警处理完成",
	}, nil
}

type DecisionAgent struct {
	*BaseAgent
}

func NewDecisionAgent(config AgentConfig) (*DecisionAgent, error) {
	baseAgent, err := NewBaseAgent(
		"decision-agent-001",
		"DecisionAgent",
		"decision",
		"决策执行Agent，负责风险评估和执行计划生成",
		config,
	)
	if err != nil {
		return nil, err
	}
	return &DecisionAgent{BaseAgent: baseAgent}, nil
}

func (d *DecisionAgent) Execute(ctx context.Context, input AgentInput) (AgentOutput, error) {
	logger.Info(fmt.Sprintf("DecisionAgent executing: %s", input.TaskType))

	prompt := fmt.Sprintf(`根据根因分析结果，制定自动化修复方案：

根因：%v

请制定：
1. 修复方案和执行步骤
2. 风险评估（低/中/高）
3. 影响范围分析
4. 是否需要人工确认
5. 验证方法

返回执行计划和风险评估（JSON格式）。`, input.Input)

	resp, err := d.llm.Generate(ctx, prompt)
	if err != nil {
		return AgentOutput{Status: "error"}, err
	}

	return AgentOutput{
		Result:  map[string]interface{}{"decision_plan": resp},
		Status:  "completed",
		Message: "决策方案生成完成",
	}, nil
}

type LearningAgent struct {
	*BaseAgent
}

func NewLearningAgent(config AgentConfig) (*LearningAgent, error) {
	baseAgent, err := NewBaseAgent(
		"learning-agent-001",
		"LearningAgent",
		"learning",
		"学习优化Agent，负责知识沉淀和最佳实践生成",
		config,
	)
	if err != nil {
		return nil, err
	}
	return &LearningAgent{BaseAgent: baseAgent}, nil
}

func (l *LearningAgent) Execute(ctx context.Context, input AgentInput) (AgentOutput, error) {
	logger.Info(fmt.Sprintf("LearningAgent executing: %s", input.TaskType))
	return l.BaseAgent.Execute(ctx, input)
}

type InteractionAgent struct {
	*BaseAgent
}

func NewInteractionAgent(config AgentConfig) (*InteractionAgent, error) {
	baseAgent, err := NewBaseAgent(
		"interaction-agent-001",
		"InteractionAgent",
		"interaction",
		"交互服务Agent，负责自然语言交互和报告生成",
		config,
	)
	if err != nil {
		return nil, err
	}
	return &InteractionAgent{BaseAgent: baseAgent}, nil
}

func (i *InteractionAgent) Execute(ctx context.Context, input AgentInput) (AgentOutput, error) {
	logger.Info(fmt.Sprintf("InteractionAgent executing: %s", input.TaskType))
	return i.BaseAgent.Execute(ctx, input)
}

func CreateDefaultAgentsWithTools(llmConfig AgentConfig) error {
	monitorAgent, err := NewMonitorAgent(llmConfig)
	if err != nil {
		return err
	}
	RegisterAgent(monitorAgent.BaseAgent)

	analysisAgent, err := NewAnalysisAgent(llmConfig)
	if err != nil {
		return err
	}
	RegisterAgent(analysisAgent.BaseAgent)

	alertAgent, err := NewAlertAgent(llmConfig)
	if err != nil {
		return err
	}
	RegisterAgent(alertAgent.BaseAgent)

	decisionAgent, err := NewDecisionAgent(llmConfig)
	if err != nil {
		return err
	}
	RegisterAgent(decisionAgent.BaseAgent)

	learningAgent, err := NewLearningAgent(llmConfig)
	if err != nil {
		return err
	}
	RegisterAgent(learningAgent.BaseAgent)

	interactionAgent, err := NewInteractionAgent(llmConfig)
	if err != nil {
		return err
	}
	RegisterAgent(interactionAgent.BaseAgent)

	logger.Info("Created and registered 6 specialized agents")
	return nil
}

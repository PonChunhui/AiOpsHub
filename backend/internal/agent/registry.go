package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

var (
	agents      map[string]*BaseAgent
	agentsMutex sync.RWMutex
)

func init() {
	agents = make(map[string]*BaseAgent)
}

func RegisterAgent(agent *BaseAgent) {
	agentsMutex.Lock()
	defer agentsMutex.Unlock()
	agents[agent.ID] = agent
	logger.Info(fmt.Sprintf("Registered agent: %s (%s)", agent.Name, agent.ID))
}

func GetAgent(id string) (*BaseAgent, error) {
	agentsMutex.RLock()
	defer agentsMutex.RUnlock()

	agent, exists := agents[id]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", id)
	}
	return agent, nil
}

func ListAgents() []map[string]interface{} {
	agentsMutex.RLock()
	defer agentsMutex.RUnlock()

	var list []map[string]interface{}
	for _, agent := range agents {
		list = append(list, map[string]interface{}{
			"id":          agent.ID,
			"name":        agent.Name,
			"type":        agent.Type,
			"description": agent.Description,
			"provider":    agent.Config.Provider,
			"model":       agent.Config.Model,
		})
	}
	return list
}

func CreateDefaultAgents(llmConfig AgentConfig) error {
	monitorAgent, err := NewBaseAgent(
		"monitor-agent-001",
		"MonitorAgent",
		"monitor",
		"系统监控Agent，负责监控指标分析和异常检测",
		llmConfig,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create MonitorAgent: %v", err))
		return err
	}
	RegisterAgent(monitorAgent)

	analysisAgent, err := NewBaseAgent(
		"analysis-agent-001",
		"AnalysisAgent",
		"analysis",
		"故障分析Agent，负责根因分析和诊断建议",
		llmConfig,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create AnalysisAgent: %v", err))
		return err
	}
	RegisterAgent(analysisAgent)

	remediationAgent, err := NewBaseAgent(
		"remediation-agent-001",
		"RemediationAgent",
		"remediation",
		"自动修复Agent，负责执行自动化修复操作",
		llmConfig,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create RemediationAgent: %v", err))
		return err
	}
	RegisterAgent(remediationAgent)

	logger.Info("Created and registered 3 default agents with Aliyun Bailian")
	return nil
}

func ExecuteAgent(ctx context.Context, agentID string, input AgentInput) (AgentOutput, error) {
	agent, err := GetAgent(agentID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get agent: %v", err))
		return AgentOutput{
			Status:  "error",
			Message: err.Error(),
		}, err
	}

	return agent.Execute(ctx, input)
}

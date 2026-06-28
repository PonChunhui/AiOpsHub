package integration

import (
	"testing"

	"github.com/aiops/AiOpsHub/backend/internal/agent"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

func TestDecisionEngineIntegration(t *testing.T) {
	logger.Init()
	engine := agent.NewDecisionEngine()

	t.Run("TaskRouting", func(t *testing.T) {
		tests := []struct {
			name       string
			taskType   string
			expectedID string
		}{
			{
				name:       "Monitor collect",
				taskType:   "monitor_collect",
				expectedID: "monitor-agent-001",
			},
			{
				name:       "Analysis diagnosis",
				taskType:   "analysis_diagnosis",
				expectedID: "analysis-agent-001",
			},
			{
				name:       "Decision execute",
				taskType:   "decision_execute",
				expectedID: "decision-agent-001",
			},
		}

		for _, tt := range tests {
			agentID, err := engine.RouteToAgent(tt.taskType)
			if err != nil {
				t.Errorf("Failed to route task %s: %v", tt.taskType, err)
			}
			if agentID != tt.expectedID {
				t.Errorf("Expected agent %s, got %s", tt.expectedID, agentID)
			}
		}
	})

	t.Run("OrchestrationStrategy", func(t *testing.T) {
		subTasks := []agent.SubTask{
			{
				TaskID:       "task-001",
				TaskType:     "monitor_collect",
				Description:  "采集CPU指标",
				AgentID:      "monitor-agent-001",
				Dependencies: []string{},
			},
			{
				TaskID:       "task-002",
				TaskType:     "analysis_diagnosis",
				Description:  "分析根因",
				AgentID:      "analysis-agent-001",
				Dependencies: []string{"task-001"},
			},
		}

		plan := engine.DetermineOrchestrationStrategy(subTasks)

		if plan.Strategy == "" {
			t.Error("Strategy should not be empty")
		}

		if len(plan.TaskSequence) == 0 {
			t.Error("Task sequence should not be empty")
		}

		if len(plan.ParallelTasks) == 0 {
			t.Error("Parallel tasks should not be empty")
		}
	})
}

func TestParallelGroupsIntegration(t *testing.T) {
	logger.Init()
	engine := agent.NewDecisionEngine()

	subTasks := []agent.SubTask{
		{
			TaskID:       "task-001",
			TaskType:     "monitor_collect",
			Description:  "采集CPU指标",
			AgentID:      "monitor-agent-001",
			Dependencies: []string{},
		},
		{
			TaskID:       "task-002",
			TaskType:     "monitor_collect",
			Description:  "采集内存指标",
			AgentID:      "monitor-agent-001",
			Dependencies: []string{},
		},
		{
			TaskID:       "task-003",
			TaskType:     "analysis_diagnosis",
			Description:  "分析根因",
			AgentID:      "analysis-agent-001",
			Dependencies: []string{"task-001", "task-002"},
		},
		{
			TaskID:       "task-004",
			TaskType:     "decision_execute",
			Description:  "制定修复方案",
			AgentID:      "decision-agent-001",
			Dependencies: []string{"task-003"},
		},
	}

	plan := engine.DetermineOrchestrationStrategy(subTasks)

	t.Logf("Strategy: %s", plan.Strategy)
	t.Logf("Parallel groups: %v", plan.ParallelTasks)
	t.Logf("Task sequence: %v", plan.TaskSequence)
	t.Logf("Estimated time: %d seconds", plan.EstimatedTime)

	expectedParallelGroups := 3
	if len(plan.ParallelTasks) != expectedParallelGroups {
		t.Errorf("Expected %d parallel groups, got %d", expectedParallelGroups, len(plan.ParallelTasks))
	}

	firstGroup := plan.ParallelTasks[0]
	if len(firstGroup) != 2 {
		t.Errorf("First group should have 2 parallel tasks (task-001, task-002), got %d", len(firstGroup))
	}

	if plan.Strategy != "hybrid" {
		t.Errorf("Expected hybrid strategy for complex dependencies, got %s", plan.Strategy)
	}
}

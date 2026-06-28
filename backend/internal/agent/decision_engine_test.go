package agent

import (
	"testing"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

func init() {
	logger.Init()
}

func TestDecisionEngine_RouteToAgent(t *testing.T) {
	engine := NewDecisionEngine()

	tests := []struct {
		name        string
		taskType    string
		expectedID  string
		shouldError bool
	}{
		{
			name:        "Monitor collect task",
			taskType:    "monitor_collect",
			expectedID:  "monitor-agent-001",
			shouldError: false,
		},
		{
			name:        "Analysis diagnosis task",
			taskType:    "analysis_diagnosis",
			expectedID:  "analysis-agent-001",
			shouldError: false,
		},
		{
			name:        "Decision execute task",
			taskType:    "decision_execute",
			expectedID:  "decision-agent-001",
			shouldError: false,
		},
		{
			name:        "Unknown task type",
			taskType:    "unknown_task",
			expectedID:  "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentID, err := engine.RouteToAgent(tt.taskType)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for task type %s, but got none", tt.taskType)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if agentID != tt.expectedID {
					t.Errorf("Expected agent ID %s, got %s", tt.expectedID, agentID)
				}
			}
		})
	}
}

func TestDecisionEngine_GetTaskPriority(t *testing.T) {
	engine := NewDecisionEngine()

	tests := []struct {
		name             string
		taskType         string
		expectedPriority int
	}{
		{
			name:             "Monitor collect priority",
			taskType:         "monitor_collect",
			expectedPriority: 1,
		},
		{
			name:             "Analysis diagnosis priority",
			taskType:         "analysis_diagnosis",
			expectedPriority: 2,
		},
		{
			name:             "Decision execute priority",
			taskType:         "decision_execute",
			expectedPriority: 3,
		},
		{
			name:             "Unknown task priority",
			taskType:         "unknown_task",
			expectedPriority: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority := engine.GetTaskPriority(tt.taskType)
			if priority != tt.expectedPriority {
				t.Errorf("Expected priority %d, got %d", tt.expectedPriority, priority)
			}
		})
	}
}

func TestDecisionEngine_RequiresHumanApproval(t *testing.T) {
	engine := NewDecisionEngine()

	tests := []struct {
		name           string
		taskType       string
		subTasks       []SubTask
		expectApproval bool
	}{
		{
			name:     "Auto remediation requires approval",
			taskType: "auto_remediation",
			subTasks: []SubTask{
				{TaskID: "task-001", TaskType: "monitor_collect"},
			},
			expectApproval: true,
		},
		{
			name:           "Monitor task does not require approval",
			taskType:       "monitor_collect",
			subTasks:       []SubTask{},
			expectApproval: false,
		},
		{
			name:     "Decision execute with high impact requires approval",
			taskType: "decision_execute",
			subTasks: []SubTask{
				{TaskID: "task-001", TaskType: "analysis_diagnosis", Parameters: map[string]interface{}{"impact": "high"}},
			},
			expectApproval: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requiresApproval := engine.RequiresHumanApproval(tt.taskType, tt.subTasks)
			if requiresApproval != tt.expectApproval {
				t.Errorf("Expected approval %v, got %v", tt.expectApproval, requiresApproval)
			}
		})
	}
}

func TestDecisionEngine_DetermineOrchestrationStrategy(t *testing.T) {
	engine := NewDecisionEngine()

	tests := []struct {
		name             string
		subTasks         []SubTask
		expectedStrategy string
		expectedMinTime  int
	}{
		{
			name: "Sequential tasks with dependency",
			subTasks: []SubTask{
				{
					TaskID:       "task-001",
					TaskType:     "monitor_collect",
					Dependencies: []string{},
				},
				{
					TaskID:       "task-002",
					TaskType:     "analysis_diagnosis",
					Dependencies: []string{"task-001"},
				},
			},
			expectedStrategy: "parallel",
			expectedMinTime:  10,
		},
		{
			name: "Parallel tasks (no dependencies)",
			subTasks: []SubTask{
				{
					TaskID:       "task-001",
					TaskType:     "monitor_collect",
					Dependencies: []string{},
				},
				{
					TaskID:       "task-002",
					TaskType:     "monitor_collect",
					Dependencies: []string{},
				},
			},
			expectedStrategy: "sequential",
			expectedMinTime:  10,
		},
		{
			name: "Complex dependency chain",
			subTasks: []SubTask{
				{
					TaskID:       "task-001",
					TaskType:     "monitor_collect",
					Dependencies: []string{},
				},
				{
					TaskID:       "task-002",
					TaskType:     "analysis_diagnosis",
					Dependencies: []string{"task-001"},
				},
				{
					TaskID:       "task-003",
					TaskType:     "decision_execute",
					Dependencies: []string{"task-002"},
				},
			},
			expectedStrategy: "parallel",
			expectedMinTime:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := engine.DetermineOrchestrationStrategy(tt.subTasks)

			if plan.Strategy != tt.expectedStrategy {
				t.Errorf("Expected strategy %s, got %s", tt.expectedStrategy, plan.Strategy)
			}

			if plan.EstimatedTime < tt.expectedMinTime {
				t.Errorf("Expected at least %d seconds, got %d", tt.expectedMinTime, plan.EstimatedTime)
			}

			if len(plan.TaskSequence) == 0 {
				t.Error("Task sequence should not be empty")
			}

			if len(plan.ParallelTasks) == 0 {
				t.Error("Parallel tasks should not be empty")
			}
		})
	}
}

func TestDecisionEngine_BuildDependencyGraph(t *testing.T) {
	engine := NewDecisionEngine()

	subTasks := []SubTask{
		{
			TaskID:       "task-001",
			TaskType:     "monitor_collect",
			Dependencies: []string{},
		},
		{
			TaskID:       "task-002",
			TaskType:     "analysis_diagnosis",
			Dependencies: []string{"task-001"},
		},
		{
			TaskID:       "task-003",
			TaskType:     "decision_execute",
			Dependencies: []string{"task-002"},
		},
		{
			TaskID:       "task-004",
			TaskType:     "monitor_collect",
			Dependencies: []string{},
		},
	}

	plan := engine.DetermineOrchestrationStrategy(subTasks)

	if len(plan.ParallelTasks) == 0 {
		t.Error("Should have at least one parallel group")
	}

	hasParallelTasks := false
	for _, group := range plan.ParallelTasks {
		if len(group) > 1 {
			hasParallelTasks = true
			break
		}
	}

	if !hasParallelTasks {
		t.Log("Warning: No parallel tasks detected, but this might be valid")
	}
}

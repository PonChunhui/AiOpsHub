package agent

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	logger.Init()

	viper.SetConfigName("config.test")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../../configs")

	viper.SetDefault("llm.provider", "openai")
	viper.SetDefault("llm.model", "gpt-4")
	viper.SetDefault("llm.temperature", 0.7)
	viper.SetDefault("llm.max_tokens", 500)
	viper.SetDefault("llm.api_key", "test-api-key")
	viper.SetDefault("llm.base_url", "")

	if err := viper.ReadInConfig(); err != nil {
	}

	os.Exit(m.Run())
}

func TestCoordinatorAgent(t *testing.T) {
	decisionEngine := NewDecisionEngine()

	agentConfig := AgentConfig{
		Provider:    viper.GetString("llm.provider"),
		Model:       viper.GetString("llm.model"),
		Temperature: viper.GetFloat64("llm.temperature"),
		MaxTokens:   viper.GetInt("llm.max_tokens"),
		APIKey:      viper.GetString("llm.api_key"),
		BaseURL:     viper.GetString("llm.base_url"),
	}

	coordinator, err := NewCoordinatorAgent("test-coordinator-001", "TestCoordinator", agentConfig, decisionEngine)
	if err != nil {
		t.Fatalf("Failed to create Coordinator Agent: %v", err)
	}

	t.Run("IntentUnderstanding", func(t *testing.T) {
		input := CoordinatorInput{
			SessionID: "test-session-001",
			UserQuery: "订单服务响应很慢，帮我分析原因",
			Context:   map[string]interface{}{},
			Timestamp: time.Now(),
		}

		intent, taskType, err := coordinator.UnderstandIntent(context.Background(), input)
		if err != nil {
			t.Fatalf("Intent understanding failed: %v", err)
		}

		t.Logf("Intent: %s, TaskType: %s", intent, taskType)

		if intent == "" || taskType == "" {
			t.Error("Intent or TaskType is empty")
		}
	})

	t.Run("TaskDecomposition", func(t *testing.T) {
		input := CoordinatorInput{
			SessionID: "test-session-002",
			UserQuery: "监控订单服务的CPU和内存使用情况",
			Context:   map[string]interface{}{},
			Timestamp: time.Now(),
		}

		subTasks, err := coordinator.DecomposeTask(context.Background(), input, "monitoring")
		if err != nil {
			t.Fatalf("Task decomposition failed: %v", err)
		}

		t.Logf("Decomposed into %d subtasks", len(subTasks))

		if len(subTasks) == 0 {
			t.Error("No subtasks generated")
		}

		for _, task := range subTasks {
			t.Logf("Subtask: %s - %s (Agent: %s)", task.TaskID, task.Description, task.AgentID)
		}
	})

	t.Run("OrchestrationPlanning", func(t *testing.T) {
		subTasks := []SubTask{
			{
				TaskID:       "task-001",
				TaskType:     "monitor_collect",
				Description:  "采集订单服务CPU指标",
				AgentID:      "monitor-agent-001",
				Parameters:   map[string]interface{}{"service": "order-service"},
				Priority:     1,
				Dependencies: []string{},
			},
			{
				TaskID:       "task-002",
				TaskType:     "analysis_diagnosis",
				Description:  "分析根因",
				AgentID:      "analysis-agent-001",
				Parameters:   map[string]interface{}{},
				Priority:     2,
				Dependencies: []string{"task-001"},
			},
			{
				TaskID:       "task-003",
				TaskType:     "decision_execute",
				Description:  "制定修复方案",
				AgentID:      "decision-agent-001",
				Parameters:   map[string]interface{}{},
				Priority:     3,
				Dependencies: []string{"task-002"},
			},
		}

		plan := decisionEngine.DetermineOrchestrationStrategy(subTasks)

		t.Logf("Orchestration strategy: %s", plan.Strategy)
		t.Logf("Task sequence: %v", plan.TaskSequence)
		t.Logf("Parallel groups: %v", plan.ParallelTasks)
		t.Logf("Estimated time: %d seconds", plan.EstimatedTime)

		if plan.Strategy == "" {
			t.Error("Strategy is empty")
		}
	})

	t.Run("ResultIntegration", func(t *testing.T) {
		results := []AgentResult{
			{
				AgentID:   "monitor-agent-001",
				TaskID:    "task-001",
				Result:    map[string]interface{}{"cpu_usage": "85%", "memory_usage": "62%"},
				Status:    "completed",
				Timestamp: time.Now(),
			},
			{
				AgentID:   "analysis-agent-001",
				TaskID:    "task-002",
				Result:    map[string]interface{}{"root_cause": "MySQL慢查询"},
				Status:    "completed",
				Timestamp: time.Now(),
			},
			{
				AgentID:   "decision-agent-001",
				TaskID:    "task-003",
				Result:    map[string]interface{}{"solution": "添加索引"},
				Status:    "completed",
				Timestamp: time.Now(),
			},
		}

		report, err := coordinator.IntegrateResults(context.Background(), results)
		if err != nil {
			t.Fatalf("Result integration failed: %v", err)
		}

		t.Logf("Integrated report: %s", report)

		if report == "" {
			t.Error("Report is empty")
		}
	})
}

func TestDecisionEngine(t *testing.T) {
	engine := NewDecisionEngine()

	t.Run("AgentRouting", func(t *testing.T) {
		agentID, err := engine.RouteToAgent("monitor_collect")
		if err != nil {
			t.Fatalf("Agent routing failed: %v", err)
		}

		t.Logf("Routed task 'monitor_collect' to agent: %s", agentID)

		if agentID != "monitor-agent-001" {
			t.Errorf("Expected 'monitor-agent-001', got '%s'", agentID)
		}
	})

	t.Run("PriorityCheck", func(t *testing.T) {
		priority := engine.GetTaskPriority("monitor_collect")
		t.Logf("Task 'monitor_collect' priority: %d", priority)

		if priority == 0 {
			t.Error("Priority is 0")
		}
	})

	t.Run("HumanApprovalCheck", func(t *testing.T) {
		subTasks := []SubTask{
			{
				TaskID:   "task-001",
				TaskType: "decision_execute",
			},
		}

		requiresApproval := engine.RequiresHumanApproval("auto_remediation", subTasks)
		t.Logf("Task 'auto_remediation' requires approval: %v", requiresApproval)

		if !requiresApproval {
			t.Error("Auto remediation should require approval")
		}
	})
}

func TestCoordinatorFullExecution(t *testing.T) {
	decisionEngine := NewDecisionEngine()

	agentConfig := AgentConfig{
		Provider:    viper.GetString("llm.provider"),
		Model:       viper.GetString("llm.model"),
		Temperature: viper.GetFloat64("llm.temperature"),
		MaxTokens:   viper.GetInt("llm.max_tokens"),
		APIKey:      viper.GetString("llm.api_key"),
		BaseURL:     viper.GetString("llm.base_url"),
	}

	coordinator, err := NewCoordinatorAgent("test-coordinator-001", "TestCoordinator", agentConfig, decisionEngine)
	if err != nil {
		t.Fatalf("Failed to create Coordinator Agent: %v", err)
	}

	input := CoordinatorInput{
		SessionID: "test-session-full",
		UserQuery: "订单服务响应很慢，帮我分析原因并给出解决方案",
		Context:   map[string]interface{}{"service": "order-service"},
		Timestamp: time.Now(),
	}

	output, err := coordinator.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Coordinator execution failed: %v", err)
	}

	t.Logf("Coordinator Output:")
	t.Logf("  SessionID: %s", output.SessionID)
	t.Logf("  Intent: %s", output.Intent)
	t.Logf("  TaskType: %s", output.TaskType)
	t.Logf("  SubTasks: %d", len(output.SubTasks))
	t.Logf("  Strategy: %s", output.Orchestration.Strategy)
	t.Logf("  RequiresApproval: %v", output.RequiresApproval)

	if output.Intent == "" {
		t.Error("Intent is empty")
	}

	if len(output.SubTasks) == 0 {
		t.Error("No subtasks")
	}

	if output.Orchestration.Strategy == "" {
		t.Error("Strategy is empty")
	}
}

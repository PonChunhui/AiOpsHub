package state_sync

import (
	"testing"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

func init() {
	logger.Init()
}

func TestAgentStateCreation(t *testing.T) {
	state := NewAgentState("monitor-agent-001", "session-001", "workflow-001")

	if state.AgentID != "monitor-agent-001" {
		t.Errorf("Expected agent ID 'monitor-agent-001', got '%s'", state.AgentID)
	}

	if state.SessionID != "session-001" {
		t.Errorf("Expected session ID 'session-001', got '%s'", state.SessionID)
	}

	if state.Status != StatusPending {
		t.Errorf("Expected initial status '%s', got '%s'", StatusPending, state.Status)
	}

	if state.Progress != 0 {
		t.Errorf("Expected initial progress 0, got %d", state.Progress)
	}

	t.Logf("AgentState created: %s (status=%s)", state.AgentID, state.Status)
}

func TestAgentStateStatusChecks(t *testing.T) {
	state := NewAgentState("test-agent", "test-session", "test-workflow")

	tests := []struct {
		name     string
		status   string
		expected bool
		method   func() bool
	}{
		{"Pending check", StatusPending, true, state.IsPending},
		{"Running check", StatusRunning, true, state.IsRunning},
		{"Completed check", StatusCompleted, true, state.IsCompleted},
		{"Failed check", StatusFailed, true, state.IsFailed},
		{"Timeout check", StatusTimeout, true, state.IsTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state.Status = tt.status
			result := tt.method()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestAgentStateTransitions(t *testing.T) {
	state := NewAgentState("transition-agent", "transition-session", "transition-workflow")

	state.SetRunning("task-001")
	if state.Status != StatusRunning {
		t.Errorf("Expected status RUNNING after SetRunning, got '%s'", state.Status)
	}
	if state.CurrentTask != "task-001" {
		t.Errorf("Expected current task 'task-001', got '%s'", state.CurrentTask)
	}

	state.UpdateProgress(50)
	if state.Progress != 50 {
		t.Errorf("Expected progress 50, got %d", state.Progress)
	}

	state.SetCompleted(map[string]interface{}{"result": "success"})
	if state.Status != StatusCompleted {
		t.Errorf("Expected status COMPLETED, got '%s'", state.Status)
	}
	if state.Progress != 100 {
		t.Errorf("Expected progress 100 after completion, got %d", state.Progress)
	}

	state.SetFailed("test error")
	if state.Status != StatusFailed {
		t.Errorf("Expected status FAILED, got '%s'", state.Status)
	}
	if state.Error != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", state.Error)
	}

	state.SetTimeout()
	if state.Status != StatusTimeout {
		t.Errorf("Expected status TIMEOUT, got '%s'", state.Status)
	}

	t.Logf("All transitions tested successfully")
}

func TestAgentStateIntermediateResult(t *testing.T) {
	state := NewAgentState("result-agent", "result-session", "result-workflow")

	result := map[string]interface{}{
		"cpu_usage":    85.5,
		"memory_usage": 62.3,
	}

	state.SetIntermediateResult(result)
	if len(state.IntermediateResult) == 0 {
		t.Error("Intermediate result should not be empty")
	}

	_, ok := state.IntermediateResult["cpu_usage"]
	if !ok {
		t.Error("cpu_usage should exist in intermediate result")
	}

	t.Logf("Intermediate result set: %d values", len(state.IntermediateResult))
}

func TestSessionStateCreation(t *testing.T) {
	session := NewSessionState("session-001", "workflow-001")

	if session.SessionID != "session-001" {
		t.Errorf("Expected session ID 'session-001', got '%s'", session.SessionID)
	}

	if session.WorkflowID != "workflow-001" {
		t.Errorf("Expected workflow ID 'workflow-001', got '%s'", session.WorkflowID)
	}

	if session.Status != StatusPending {
		t.Errorf("Expected initial status '%s', got '%s'", StatusPending, session.Status)
	}

	if len(session.AgentStates) != 0 {
		t.Errorf("Expected empty agent states map, got %d", len(session.AgentStates))
	}

	t.Logf("SessionState created: %s", session.SessionID)
}

func TestSessionStateAgentManagement(t *testing.T) {
	session := NewSessionState("session-agents", "workflow-agents")

	agent1 := NewAgentState("agent-001", session.SessionID, session.WorkflowID)
	agent2 := NewAgentState("agent-002", session.SessionID, session.WorkflowID)

	session.AddAgentState(agent1)
	session.AddAgentState(agent2)

	if len(session.AgentStates) != 2 {
		t.Errorf("Expected 2 agent states, got %d", len(session.AgentStates))
	}

	retrieved := session.GetAgentState("agent-001")
	if retrieved == nil {
		t.Error("Agent-001 should exist")
	}

	session.UpdateAgentState("agent-001", StatusRunning, 50)
	if retrieved.Status != StatusRunning {
		t.Errorf("Expected status RUNNING, got '%s'", retrieved.Status)
	}

	t.Logf("Session has %d agents", len(session.AgentStates))
}

func TestSessionStateCompletionChecks(t *testing.T) {
	session := NewSessionState("session-completion", "workflow-completion")

	agent1 := NewAgentState("agent-001", session.SessionID, session.WorkflowID)
	agent2 := NewAgentState("agent-002", session.SessionID, session.WorkflowID)

	session.AddAgentState(agent1)
	session.AddAgentState(agent2)

	if session.AllCompleted() {
		t.Error("Session should not be all completed initially")
	}

	agent1.SetCompleted(map[string]interface{}{})
	agent2.SetCompleted(map[string]interface{}{})

	if !session.AllCompleted() {
		t.Error("Session should be all completed after all agents complete")
	}

	agent1.SetFailed("error")
	if !session.AnyFailed() {
		t.Error("Session should have failed agents")
	}

	agent2.SetTimeout()
	if !session.AnyTimeout() {
		t.Error("Session should have timeout agents")
	}

	t.Logf("Completion checks verified")
}

func TestSessionStateRunningAgents(t *testing.T) {
	session := NewSessionState("session-running", "workflow-running")

	agent1 := NewAgentState("agent-001", session.SessionID, session.WorkflowID)
	agent2 := NewAgentState("agent-002", session.SessionID, session.WorkflowID)
	agent3 := NewAgentState("agent-003", session.SessionID, session.WorkflowID)

	agent1.SetRunning("task-001")
	agent2.SetRunning("task-002")
	agent3.SetCompleted(map[string]interface{}{})

	session.AddAgentState(agent1)
	session.AddAgentState(agent2)
	session.AddAgentState(agent3)

	running := session.GetRunningAgents()
	if len(running) != 2 {
		t.Errorf("Expected 2 running agents, got %d", len(running))
	}

	t.Logf("Running agents: %v", running)
}

func TestSessionStateProgressSummary(t *testing.T) {
	session := NewSessionState("session-progress", "workflow-progress")

	agent1 := NewAgentState("agent-001", session.SessionID, session.WorkflowID)
	agent2 := NewAgentState("agent-002", session.SessionID, session.WorkflowID)
	agent3 := NewAgentState("agent-003", session.SessionID, session.WorkflowID)
	agent4 := NewAgentState("agent-004", session.SessionID, session.WorkflowID)

	agent1.SetCompleted(map[string]interface{}{})
	agent2.SetRunning("task-002")
	agent2.UpdateProgress(50)
	agent3.Status = StatusPending
	agent4.SetFailed("error")

	session.AddAgentState(agent1)
	session.AddAgentState(agent2)
	session.AddAgentState(agent3)
	session.AddAgentState(agent4)

	summary := session.GetProgressSummary()

	if summary["total"] != 4 {
		t.Errorf("Expected total 4, got %d", summary["total"])
	}

	if summary["completed"] != 1 {
		t.Errorf("Expected completed 1, got %d", summary["completed"])
	}

	if summary["running"] != 1 {
		t.Errorf("Expected running 1, got %d", summary["running"])
	}

	if summary["pending"] != 1 {
		t.Errorf("Expected pending 1, got %d", summary["pending"])
	}

	if summary["failed"] != 1 {
		t.Errorf("Expected failed 1, got %d", summary["failed"])
	}

	t.Logf("Progress summary: total=%d, completed=%d, running=%d, pending=%d, failed=%d",
		summary["total"], summary["completed"], summary["running"], summary["pending"], summary["failed"])
}

func TestSessionStateFinalResult(t *testing.T) {
	session := NewSessionState("session-result", "workflow-result")

	result := map[string]interface{}{
		"analysis":   "root cause identified",
		"solution":   "apply fix",
		"confidence": 0.95,
	}

	session.SetFinalResult(result)

	if len(session.FinalResult) == 0 {
		t.Error("Final result should not be empty")
	}

	_, ok := session.FinalResult["confidence"]
	if !ok {
		t.Error("confidence should exist in final result")
	}

	t.Logf("Final result set: %d keys", len(session.FinalResult))
}

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"Pending", StatusPending, "PENDING"},
		{"Running", StatusRunning, "RUNNING"},
		{"Completed", StatusCompleted, "COMPLETED"},
		{"Failed", StatusFailed, "FAILED"},
		{"Timeout", StatusTimeout, "TIMEOUT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, tt.constant)
			}
		})
	}
}

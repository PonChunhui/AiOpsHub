package conflict_resolver

import (
	"testing"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

func init() {
	logger.Init()
}

func TestResultResolverCreation(t *testing.T) {
	resolver := NewResultResolver()

	if len(resolver.AgentPriority) == 0 {
		t.Error("AgentPriority should not be empty")
	}

	t.Logf("ResultResolver created with %d agent priorities", len(resolver.AgentPriority))
}

func TestVoteResults(t *testing.T) {
	resolver := NewResultResolver()

	tests := []struct {
		name         string
		results      []ConflictResult
		expectWinner string
	}{
		{
			name: "Clear majority",
			results: []ConflictResult{
				{AgentID: "agent-001", Value: "solution-A"},
				{AgentID: "agent-002", Value: "solution-A"},
				{AgentID: "agent-003", Value: "solution-B"},
			},
			expectWinner: "solution-A",
		},
		{
			name: "Tie (first wins)",
			results: []ConflictResult{
				{AgentID: "agent-001", Value: "solution-A"},
				{AgentID: "agent-002", Value: "solution-B"},
			},
			expectWinner: "solution-A",
		},
		{
			name: "Single result",
			results: []ConflictResult{
				{AgentID: "agent-001", Value: "solution-X"},
			},
			expectWinner: "solution-X",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			winner := resolver.VoteResults(tt.results)
			if winner != tt.expectWinner {
				t.Errorf("Expected winner '%s', got '%s'", tt.expectWinner, winner)
			}
		})
	}
}

func TestSelectByPriority(t *testing.T) {
	resolver := NewResultResolver()

	tests := []struct {
		name         string
		results      []ConflictResult
		expectWinner string
	}{
		{
			name: "Analysis agent highest priority",
			results: []ConflictResult{
				{AgentID: "monitor-agent-001", Value: "solution-A"},
				{AgentID: "analysis-agent-001", Value: "solution-B"},
			},
			expectWinner: "solution-B",
		},
		{
			name: "Decision agent priority",
			results: []ConflictResult{
				{AgentID: "interaction-agent-001", Value: "solution-A"},
				{AgentID: "decision-agent-001", Value: "solution-B"},
			},
			expectWinner: "solution-B",
		},
		{
			name: "Unknown agent default priority",
			results: []ConflictResult{
				{AgentID: "unknown-agent", Value: "solution-A"},
				{AgentID: "learning-agent-001", Value: "solution-B"},
			},
			expectWinner: "solution-B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			winner := resolver.SelectByPriority(tt.results)
			if winner != tt.expectWinner {
				t.Errorf("Expected winner '%s', got '%s'", tt.expectWinner, winner)
			}
		})
	}
}

func TestResolveConflict(t *testing.T) {
	resolver := NewResultResolver()

	tests := []struct {
		name         string
		conflictType string
		results      []ConflictResult
		expectMethod string
	}{
		{
			name:         "Result conflict - vote",
			conflictType: "result_conflict",
			results: []ConflictResult{
				{AgentID: "agent-001", Value: "A"},
				{AgentID: "agent-002", Value: "A"},
				{AgentID: "agent-003", Value: "B"},
			},
			expectMethod: "vote",
		},
		{
			name:         "Priority conflict",
			conflictType: "priority_conflict",
			results: []ConflictResult{
				{AgentID: "monitor-agent-001", Value: "A"},
				{AgentID: "analysis-agent-001", Value: "B"},
			},
			expectMethod: "priority",
		},
		{
			name:         "Mixed conflict with clear majority",
			conflictType: "mixed_conflict",
			results: []ConflictResult{
				{AgentID: "agent-001", Value: "A"},
				{AgentID: "agent-002", Value: "A"},
				{AgentID: "agent-003", Value: "B"},
			},
			expectMethod: "vote",
		},
		{
			name:         "Single result - no conflict",
			conflictType: "result_conflict",
			results: []ConflictResult{
				{AgentID: "agent-001", Value: "A"},
			},
			expectMethod: "single_result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, resolution := resolver.ResolveConflict(tt.conflictType, tt.results)
			if resolution.Method != tt.expectMethod {
				t.Errorf("Expected method '%s', got '%s'", tt.expectMethod, resolution.Method)
			}
		})
	}
}

func TestHasConflict(t *testing.T) {
	resolver := NewResultResolver()

	tests := []struct {
		name           string
		results        []ConflictResult
		expectConflict bool
	}{
		{
			name: "No conflict - same values",
			results: []ConflictResult{
				{AgentID: "agent-001", Value: "solution-A"},
				{AgentID: "agent-002", Value: "solution-A"},
			},
			expectConflict: false,
		},
		{
			name: "Conflict - different values",
			results: []ConflictResult{
				{AgentID: "agent-001", Value: "solution-A"},
				{AgentID: "agent-002", Value: "solution-B"},
			},
			expectConflict: true,
		},
		{
			name: "Single result - no conflict",
			results: []ConflictResult{
				{AgentID: "agent-001", Value: "solution-A"},
			},
			expectConflict: false,
		},
		{
			name:           "Empty results - no conflict",
			results:        []ConflictResult{},
			expectConflict: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasConflict := resolver.HasConflict(tt.results)
			if hasConflict != tt.expectConflict {
				t.Errorf("Expected conflict=%v, got %v", tt.expectConflict, hasConflict)
			}
		})
	}
}

func TestSetAndGetAgentPriority(t *testing.T) {
	resolver := NewResultResolver()

	resolver.SetAgentPriority("test-agent-001", 7)

	priority := resolver.GetAgentPriority("test-agent-001")
	if priority != 7 {
		t.Errorf("Expected priority 7, got %d", priority)
	}

	unknownPriority := resolver.GetAgentPriority("unknown-agent")
	if unknownPriority != 10 {
		t.Errorf("Expected default priority 10, got %d", unknownPriority)
	}

	t.Logf("Agent priorities managed successfully")
}

func TestConflictResultCreation(t *testing.T) {
	result := NewConflictResult("agent-001", "solution-A", 0.95, "2024-01-01T00:00:00Z")

	if result.AgentID != "agent-001" {
		t.Errorf("Expected agent ID 'agent-001', got '%s'", result.AgentID)
	}

	if result.Value != "solution-A" {
		t.Errorf("Expected value 'solution-A', got '%s'", result.Value)
	}

	if result.Confidence != 0.95 {
		t.Errorf("Expected confidence 0.95, got %f", result.Confidence)
	}

	t.Logf("ConflictResult created: agent=%s, value=%s, confidence=%f", result.AgentID, result.Value, result.Confidence)
}

func TestConflictRequestCreation(t *testing.T) {
	results := []ConflictResult{
		{AgentID: "agent-001", Value: "A"},
		{AgentID: "agent-002", Value: "B"},
	}

	request := NewConflictRequest("conflict-001", "result_conflict", results, "Test conflict")

	if request.ConflictID != "conflict-001" {
		t.Errorf("Expected conflict ID 'conflict-001', got '%s'", request.ConflictID)
	}

	if request.ConflictType != "result_conflict" {
		t.Errorf("Expected conflict type 'result_conflict', got '%s'", request.ConflictType)
	}

	if len(request.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(request.Results))
	}

	t.Logf("ConflictRequest created: %s (%s)", request.ConflictID, request.ConflictType)
}

func TestHumanDecisionRequest(t *testing.T) {
	resolver := NewResultResolver()

	conflict := ConflictRequest{
		ConflictID:   "human-conflict-001",
		ConflictType: "critical_decision",
		Results: []ConflictResult{
			{AgentID: "agent-001", Value: "critical-action-A"},
			{AgentID: "agent-002", Value: "critical-action-B"},
		},
		Description: "Critical decision requires human approval",
		Timestamp:   "2024-01-01T00:00:00Z",
	}

	request := resolver.RequestHumanDecision(conflict)

	if !request.RequiresAction {
		t.Error("RequiresAction should be true")
	}

	if request.ConflictID != conflict.ConflictID {
		t.Errorf("Expected conflict ID '%s', got '%s'", conflict.ConflictID, request.ConflictID)
	}

	t.Logf("HumanDecisionRequest created for conflict: %s", request.ConflictID)
}

func TestResourceLockCreation(t *testing.T) {
	lock := NewResourceLock("resource-001", "agent-001", 30*time.Second)

	if lock.ResourceID != "resource-001" {
		t.Errorf("Expected resource ID 'resource-001', got '%s'", lock.ResourceID)
	}

	if lock.AgentID != "agent-001" {
		t.Errorf("Expected agent ID 'agent-001', got '%s'", lock.AgentID)
	}

	if lock.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", lock.Timeout)
	}

	t.Logf("ResourceLock created: resource=%s, agent=%s, timeout=%v", lock.ResourceID, lock.AgentID, lock.Timeout)
}

func TestResourceLockExpiration(t *testing.T) {
	tests := []struct {
		name          string
		acquiredAt    time.Time
		timeout       time.Duration
		expectExpired bool
	}{
		{
			name:          "Not expired",
			acquiredAt:    time.Now(),
			timeout:       30 * time.Second,
			expectExpired: false,
		},
		{
			name:          "Expired",
			acquiredAt:    time.Now().Add(-31 * time.Second),
			timeout:       30 * time.Second,
			expectExpired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lock := &ResourceLock{
				ResourceID: "test-resource",
				AgentID:    "test-agent",
				AcquiredAt: tt.acquiredAt,
				Timeout:    tt.timeout,
			}

			isExpired := lock.IsExpired()
			if isExpired != tt.expectExpired {
				t.Errorf("Expected expired=%v, got %v", tt.expectExpired, isExpired)
			}
		})
	}
}

func TestResourceLockRemainingTime(t *testing.T) {
	lock := NewResourceLock("resource-001", "agent-001", 30*time.Second)

	remaining := lock.RemainingTime()

	if remaining < 0 {
		t.Error("Remaining time should not be negative")
	}

	if remaining > 30*time.Second {
		t.Error("Remaining time should not exceed timeout")
	}

	t.Logf("ResourceLock remaining time: %v", remaining)
}

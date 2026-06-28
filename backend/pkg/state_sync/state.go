package state_sync

import (
	"time"
)

type AgentState struct {
	AgentID            string                 `json:"agent_id"`
	SessionID          string                 `json:"session_id"`
	WorkflowID         string                 `json:"workflow_id"`
	Status             string                 `json:"status"`
	Progress           int                    `json:"progress"`
	CurrentTask        string                 `json:"current_task"`
	StartTime          time.Time              `json:"start_time"`
	UpdateTime         time.Time              `json:"update_time"`
	IntermediateResult map[string]interface{} `json:"intermediate_result"`
	Error              string                 `json:"error"`
}

type SessionState struct {
	SessionID   string                 `json:"session_id"`
	WorkflowID  string                 `json:"workflow_id"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	AgentStates map[string]*AgentState `json:"agent_states"`
	FinalResult map[string]interface{} `json:"final_result"`
}

const (
	StatusPending   = "PENDING"
	StatusRunning   = "RUNNING"
	StatusCompleted = "COMPLETED"
	StatusFailed    = "FAILED"
	StatusTimeout   = "TIMEOUT"
)

func NewAgentState(agentID, sessionID, workflowID string) *AgentState {
	now := time.Now()
	return &AgentState{
		AgentID:            agentID,
		SessionID:          sessionID,
		WorkflowID:         workflowID,
		Status:             StatusPending,
		Progress:           0,
		CurrentTask:        "",
		StartTime:          now,
		UpdateTime:         now,
		IntermediateResult: make(map[string]interface{}),
		Error:              "",
	}
}

func NewSessionState(sessionID, workflowID string) *SessionState {
	now := time.Now()
	return &SessionState{
		SessionID:   sessionID,
		WorkflowID:  workflowID,
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
		AgentStates: make(map[string]*AgentState),
		FinalResult: make(map[string]interface{}),
	}
}

func (s *AgentState) IsPending() bool {
	return s.Status == StatusPending
}

func (s *AgentState) IsRunning() bool {
	return s.Status == StatusRunning
}

func (s *AgentState) IsCompleted() bool {
	return s.Status == StatusCompleted
}

func (s *AgentState) IsFailed() bool {
	return s.Status == StatusFailed
}

func (s *AgentState) IsTimeout() bool {
	return s.Status == StatusTimeout
}

func (s *AgentState) SetRunning(currentTask string) {
	s.Status = StatusRunning
	s.CurrentTask = currentTask
	s.Progress = 0
	s.UpdateTime = time.Now()
}

func (s *AgentState) UpdateProgress(progress int) {
	s.Progress = progress
	s.UpdateTime = time.Now()
}

func (s *AgentState) SetCompleted(result map[string]interface{}) {
	s.Status = StatusCompleted
	s.Progress = 100
	s.IntermediateResult = result
	s.UpdateTime = time.Now()
}

func (s *AgentState) SetFailed(error string) {
	s.Status = StatusFailed
	s.Error = error
	s.UpdateTime = time.Now()
}

func (s *AgentState) SetTimeout() {
	s.Status = StatusTimeout
	s.UpdateTime = time.Now()
}

func (s *AgentState) SetIntermediateResult(result map[string]interface{}) {
	s.IntermediateResult = result
	s.UpdateTime = time.Now()
}

func (s *SessionState) AddAgentState(agentState *AgentState) {
	s.AgentStates[agentState.AgentID] = agentState
	s.UpdatedAt = time.Now()
}

func (s *SessionState) GetAgentState(agentID string) *AgentState {
	return s.AgentStates[agentID]
}

func (s *SessionState) UpdateAgentState(agentID string, status string, progress int) {
	if agentState, exists := s.AgentStates[agentID]; exists {
		agentState.Status = status
		agentState.Progress = progress
		agentState.UpdateTime = time.Now()
		s.UpdatedAt = time.Now()
	}
}

func (s *SessionState) SetFinalResult(result map[string]interface{}) {
	s.FinalResult = result
	s.UpdatedAt = time.Now()
}

func (s *SessionState) AllCompleted() bool {
	for _, agentState := range s.AgentStates {
		if !agentState.IsCompleted() {
			return false
		}
	}
	return true
}

func (s *SessionState) AnyFailed() bool {
	for _, agentState := range s.AgentStates {
		if agentState.IsFailed() {
			return true
		}
	}
	return false
}

func (s *SessionState) AnyTimeout() bool {
	for _, agentState := range s.AgentStates {
		if agentState.IsTimeout() {
			return true
		}
	}
	return false
}

func (s *SessionState) GetRunningAgents() []string {
	running := []string{}
	for agentID, agentState := range s.AgentStates {
		if agentState.IsRunning() {
			running = append(running, agentID)
		}
	}
	return running
}

func (s *SessionState) GetProgressSummary() map[string]int {
	summary := make(map[string]int)
	total := len(s.AgentStates)

	if total == 0 {
		return summary
	}

	completedCount := 0
	runningCount := 0
	pendingCount := 0
	failedCount := 0

	for _, agentState := range s.AgentStates {
		switch agentState.Status {
		case StatusCompleted:
			completedCount++
		case StatusRunning:
			runningCount++
		case StatusPending:
			pendingCount++
		case StatusFailed:
			failedCount++
		}
	}

	summary["total"] = total
	summary["completed"] = completedCount
	summary["running"] = runningCount
	summary["pending"] = pendingCount
	summary["failed"] = failedCount

	return summary
}

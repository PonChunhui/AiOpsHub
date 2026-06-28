package state_sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/aiops/AiOpsHub/backend/pkg/redis"
)

type StateManager struct {
	RedisClient *redis.RedisClient
}

func NewStateManager(redisClient *redis.RedisClient) *StateManager {
	manager := &StateManager{
		RedisClient: redisClient,
	}

	logger.Info("Created State Manager")
	return manager
}

func (sm *StateManager) SetAgentState(agentID, sessionID, workflowID, status string, progress int) error {
	stateKey := fmt.Sprintf("agent:state:%s:%s", sessionID, agentID)

	state, err := sm.GetAgentState(agentID, sessionID)
	if err != nil || state == nil {
		state = NewAgentState(agentID, sessionID, workflowID)
	}

	state.Status = status
	state.Progress = progress
	state.UpdateTime = time.Now()

	stateJSON, err := json.Marshal(state)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to marshal agent state: %v", err))
		return err
	}

	err = sm.RedisClient.Set(context.Background(), stateKey, stateJSON, 1*time.Hour)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to set agent state: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Set agent state: %s (session: %s, status: %s, progress: %d)",
		agentID, sessionID, status, progress))

	return nil
}

func (sm *StateManager) GetAgentState(agentID, sessionID string) (*AgentState, error) {
	stateKey := fmt.Sprintf("agent:state:%s:%s", sessionID, agentID)

	stateJSON, err := sm.RedisClient.Get(context.Background(), stateKey)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get agent state: %v", err))
		return nil, err
	}

	var state AgentState
	err = json.Unmarshal([]byte(stateJSON), &state)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to unmarshal agent state: %v", err))
		return nil, err
	}

	return &state, nil
}

func (sm *StateManager) UpdateAgentProgress(agentID, sessionID string, progress int) error {
	state, err := sm.GetAgentState(agentID, sessionID)
	if err != nil {
		return err
	}

	state.UpdateProgress(progress)

	return sm.SetAgentState(agentID, sessionID, state.WorkflowID, state.Status, progress)
}

func (sm *StateManager) SetAgentRunning(agentID, sessionID, currentTask string) error {
	state, err := sm.GetAgentState(agentID, sessionID)
	if err != nil {
		return err
	}

	state.SetRunning(currentTask)

	return sm.SetAgentState(agentID, sessionID, state.WorkflowID, StatusRunning, 0)
}

func (sm *StateManager) SetAgentCompleted(agentID, sessionID string, result map[string]interface{}) error {
	state, err := sm.GetAgentState(agentID, sessionID)
	if err != nil {
		return err
	}

	state.SetCompleted(result)

	return sm.SetAgentState(agentID, sessionID, state.WorkflowID, StatusCompleted, 100)
}

func (sm *StateManager) SetAgentFailed(agentID, sessionID string, error string) error {
	state, err := sm.GetAgentState(agentID, sessionID)
	if err != nil {
		return err
	}

	state.SetFailed(error)

	return sm.SetAgentState(agentID, sessionID, state.WorkflowID, StatusFailed, state.Progress)
}

func (sm *StateManager) SetAgentTimeout(agentID, sessionID string) error {
	state, err := sm.GetAgentState(agentID, sessionID)
	if err != nil {
		return err
	}

	state.SetTimeout()

	return sm.SetAgentState(agentID, sessionID, state.WorkflowID, StatusTimeout, state.Progress)
}

func (sm *StateManager) SetIntermediateResult(sessionID, agentID string, result map[string]interface{}) error {
	resultKey := fmt.Sprintf("collaboration:result:%s:%s", sessionID, agentID)

	resultJSON, err := json.Marshal(result)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to marshal intermediate result: %v", err))
		return err
	}

	err = sm.RedisClient.Set(context.Background(), resultKey, resultJSON, 2*time.Hour)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to set intermediate result: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Set intermediate result for agent %s (session: %s)", agentID, sessionID))

	return nil
}

func (sm *StateManager) GetIntermediateResult(sessionID, agentID string) (map[string]interface{}, error) {
	resultKey := fmt.Sprintf("collaboration:result:%s:%s", sessionID, agentID)

	resultJSON, err := sm.RedisClient.Get(context.Background(), resultKey)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get intermediate result: %v", err))
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(resultJSON), &result)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to unmarshal intermediate result: %v", err))
		return nil, err
	}

	return result, nil
}

func (sm *StateManager) GetAllIntermediateResults(sessionID string, agentIDs []string) (map[string]map[string]interface{}, error) {
	results := make(map[string]map[string]interface{})

	for _, agentID := range agentIDs {
		result, err := sm.GetIntermediateResult(sessionID, agentID)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get result for agent %s: %v", agentID, err))
			continue
		}

		results[agentID] = result
	}

	logger.Info(fmt.Sprintf("Got intermediate results for %d agents (session: %s)", len(results), sessionID))

	return results, nil
}

func (sm *StateManager) MonitorAgents(sessionID string, agentIDs []string) (map[string]*AgentState, error) {
	states := make(map[string]*AgentState)

	for _, agentID := range agentIDs {
		state, err := sm.GetAgentState(agentID, sessionID)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get state for agent %s: %v", agentID, err))
			continue
		}

		states[agentID] = state
	}

	logger.Info(fmt.Sprintf("Monitored %d agents (session: %s)", len(states), sessionID))

	return states, nil
}

func (sm *StateManager) ClearAgentState(agentID, sessionID string) error {
	stateKey := fmt.Sprintf("agent:state:%s:%s", sessionID, agentID)

	err := sm.RedisClient.Del(context.Background(), stateKey)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to clear agent state: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Cleared agent state: %s (session: %s)", agentID, sessionID))

	return nil
}

func (sm *StateManager) ClearIntermediateResult(sessionID, agentID string) error {
	resultKey := fmt.Sprintf("collaboration:result:%s:%s", sessionID, agentID)

	err := sm.RedisClient.Del(context.Background(), resultKey)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to clear intermediate result: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Cleared intermediate result for agent %s (session: %s)", agentID, sessionID))

	return nil
}

func (sm *StateManager) ClearSession(sessionID string) error {
	statePattern := fmt.Sprintf("agent:state:%s:*", sessionID)
	resultPattern := fmt.Sprintf("collaboration:result:%s:*", sessionID)

	ctx := context.Background()

	stateKeys := sm.scanKeys(ctx, statePattern)
	for _, key := range stateKeys {
		sm.RedisClient.Del(ctx, key)
	}

	resultKeys := sm.scanKeys(ctx, resultPattern)
	for _, key := range resultKeys {
		sm.RedisClient.Del(ctx, key)
	}

	logger.Info(fmt.Sprintf("Cleared session: %s (state keys: %d, result keys: %d)",
		sessionID, len(stateKeys), len(resultKeys)))

	return nil
}

func (sm *StateManager) scanKeys(ctx context.Context, pattern string) []string {
	keys := []string{}
	iter := sm.RedisClient.Client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	return keys
}

func (sm *StateManager) GetSessionProgress(sessionID string, agentIDs []string) (int, error) {
	states, err := sm.MonitorAgents(sessionID, agentIDs)
	if err != nil {
		return 0, err
	}

	if len(states) == 0 {
		return 0, nil
	}

	totalProgress := 0
	for _, state := range states {
		totalProgress += state.Progress
	}

	averageProgress := totalProgress / len(states)

	logger.Info(fmt.Sprintf("Session progress: %d%% (session: %s)", averageProgress, sessionID))

	return averageProgress, nil
}

func (sm *StateManager) CheckAgentTimeout(sessionID string, agentIDs []string, timeoutDuration time.Duration) ([]string, error) {
	timeoutAgents := []string{}

	states, err := sm.MonitorAgents(sessionID, agentIDs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for agentID, state := range states {
		if state.IsRunning() {
			elapsed := now.Sub(state.StartTime)
			if elapsed > timeoutDuration {
				logger.Info(fmt.Sprintf("Agent %s timeout detected (elapsed: %v)", agentID, elapsed))
				timeoutAgents = append(timeoutAgents, agentID)
			}
		}
	}

	return timeoutAgents, nil
}

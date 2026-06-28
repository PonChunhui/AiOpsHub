package conflict_resolver

import (
	"fmt"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type ResultResolver struct {
	AgentPriority map[string]int
}

func NewResultResolver() *ResultResolver {
	resolver := &ResultResolver{
		AgentPriority: map[string]int{
			"analysis-agent-001":    1,
			"monitor-agent-001":     2,
			"alert-agent-001":       3,
			"decision-agent-001":    4,
			"learning-agent-001":    5,
			"interaction-agent-001": 6,
		},
	}

	logger.Info("Created Result Resolver")
	return resolver
}

func (rr *ResultResolver) VoteResults(results []ConflictResult) string {
	votes := make(map[string]int)

	for _, result := range results {
		votes[result.Value]++
	}

	maxVotes := 0
	winner := ""
	for value, count := range votes {
		if count > maxVotes {
			maxVotes = count
			winner = value
		}
	}

	logger.Info(fmt.Sprintf("Vote results: winner '%s' with %d votes (total results: %d)", winner, maxVotes, len(results)))

	return winner
}

func (rr *ResultResolver) SelectByPriority(results []ConflictResult) string {
	if len(results) == 0 {
		return ""
	}

	highestPriority := 999
	selectedResult := ""

	for _, result := range results {
		priority, exists := rr.AgentPriority[result.AgentID]
		if !exists {
			priority = 10
		}

		if priority < highestPriority {
			highestPriority = priority
			selectedResult = result.Value
		}
	}

	logger.Info(fmt.Sprintf("Selected by priority: '%s' (priority: %d)", selectedResult, highestPriority))

	return selectedResult
}

func (rr *ResultResolver) ResolveConflict(conflictType string, results []ConflictResult) (string, ConflictResolution) {
	if len(results) == 0 {
		return "", ConflictResolution{
			Method:      "none",
			Description: "No results to resolve",
		}
	}

	if len(results) == 1 {
		return results[0].Value, ConflictResolution{
			Method:      "single_result",
			Description: "Only one result, no conflict",
		}
	}

	switch conflictType {
	case "result_conflict":
		winner := rr.VoteResults(results)
		return winner, ConflictResolution{
			Method:      "vote",
			Description: fmt.Sprintf("Selected by voting (winner: %s)", winner),
		}

	case "priority_conflict":
		winner := rr.SelectByPriority(results)
		return winner, ConflictResolution{
			Method:      "priority",
			Description: fmt.Sprintf("Selected by priority (winner: %s)", winner),
		}

	case "mixed_conflict":
		voteWinner := rr.VoteResults(results)
		votes := rr.countVotes(results, voteWinner)

		if votes > len(results)/2 {
			return voteWinner, ConflictResolution{
				Method:      "vote",
				Description: fmt.Sprintf("Clear majority (winner: %s with %d votes)", voteWinner, votes),
			}
		}

		priorityWinner := rr.SelectByPriority(results)
		return priorityWinner, ConflictResolution{
			Method:      "priority",
			Description: fmt.Sprintf("No clear majority, selected by priority (winner: %s)", priorityWinner),
		}

	default:
		winner := rr.VoteResults(results)
		return winner, ConflictResolution{
			Method:      "vote",
			Description: "Default resolution method: vote",
		}
	}
}

func (rr *ResultResolver) countVotes(results []ConflictResult, value string) int {
	count := 0
	for _, result := range results {
		if result.Value == value {
			count++
		}
	}
	return count
}

func (rr *ResultResolver) HasConflict(results []ConflictResult) bool {
	if len(results) <= 1 {
		return false
	}

	firstValue := results[0].Value
	for _, result := range results {
		if result.Value != firstValue {
			logger.Info(fmt.Sprintf("Conflict detected: different values found (first: '%s', other: '%s')", firstValue, result.Value))
			return true
		}
	}

	return false
}

func (rr *ResultResolver) RequestHumanDecision(conflict ConflictRequest) HumanDecisionRequest {
	return HumanDecisionRequest{
		ConflictID:     conflict.ConflictID,
		ConflictType:   conflict.ConflictType,
		Results:        conflict.Results,
		Description:    conflict.Description,
		Timestamp:      conflict.Timestamp,
		RequiresAction: true,
	}
}

func (rr *ResultResolver) SetAgentPriority(agentID string, priority int) {
	rr.AgentPriority[agentID] = priority
	logger.Info(fmt.Sprintf("Set priority for agent %s: %d", agentID, priority))
}

func (rr *ResultResolver) GetAgentPriority(agentID string) int {
	priority, exists := rr.AgentPriority[agentID]
	if !exists {
		return 10
	}
	return priority
}

type ConflictResult struct {
	AgentID    string
	Value      string
	Confidence float64
	Timestamp  string
}

type ConflictResolution struct {
	Method      string
	Description string
	Winner      string
}

type ConflictRequest struct {
	ConflictID   string
	ConflictType string
	Results      []ConflictResult
	Description  string
	Timestamp    string
}

type HumanDecisionRequest struct {
	ConflictID     string
	ConflictType   string
	Results        []ConflictResult
	Description    string
	Timestamp      string
	RequiresAction bool
}

func NewConflictResult(agentID, value string, confidence float64, timestamp string) *ConflictResult {
	return &ConflictResult{
		AgentID:    agentID,
		Value:      value,
		Confidence: confidence,
		Timestamp:  timestamp,
	}
}

func NewConflictRequest(conflictID, conflictType string, results []ConflictResult, description string) *ConflictRequest {
	return &ConflictRequest{
		ConflictID:   conflictID,
		ConflictType: conflictType,
		Results:      results,
		Description:  description,
		Timestamp:    "",
	}
}

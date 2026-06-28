package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type TokenUsage struct {
	SessionID    string    `json:"session_id"`
	WorkflowID   string    `json:"workflow_id"`
	AgentID      string    `json:"agent_id"`
	Model        string    `json:"model"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	TotalTokens  int       `json:"total_tokens"`
	Timestamp    time.Time `json:"timestamp"`
}

type TokenStats struct {
	TotalInputTokens  int64             `json:"total_input_tokens"`
	TotalOutputTokens int64             `json:"total_output_tokens"`
	TotalTokens       int64             `json:"total_tokens"`
	TotalCost         float64           `json:"total_cost"`
	SessionCount      int64             `json:"session_count"`
	AgentCount        int64             `json:"agent_count"`
	TopAgents         []AgentTokenUsage `json:"top_agents"`
	TopModels         []ModelTokenUsage `json:"top_models"`
	LastUpdated       time.Time         `json:"last_updated"`
}

type AgentTokenUsage struct {
	AgentID     string  `json:"agent_id"`
	TotalTokens int64   `json:"total_tokens"`
	Cost        float64 `json:"cost"`
}

type ModelTokenUsage struct {
	Model       string  `json:"model"`
	TotalTokens int64   `json:"total_tokens"`
	Cost        float64 `json:"cost"`
}

type TokenService struct {
	usages      []TokenUsage
	stats       TokenStats
	mu          sync.RWMutex
	modelPrices map[string]ModelPrice
}

type ModelPrice struct {
	InputPrice  float64
	OutputPrice float64
}

func NewTokenService() *TokenService {
	service := &TokenService{
		usages: []TokenUsage{},
		stats:  TokenStats{},
		modelPrices: map[string]ModelPrice{
			"gpt-4":         {InputPrice: 0.03, OutputPrice: 0.06},
			"gpt-4-turbo":   {InputPrice: 0.01, OutputPrice: 0.03},
			"gpt-3.5-turbo": {InputPrice: 0.0015, OutputPrice: 0.002},
		},
	}

	logger.Info("Token Service created")
	return service
}

func (t *TokenService) RecordUsage(ctx context.Context, usage TokenUsage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	usage.Timestamp = time.Now()
	t.usages = append(t.usages, usage)

	t.stats.TotalInputTokens += int64(usage.InputTokens)
	t.stats.TotalOutputTokens += int64(usage.OutputTokens)
	t.stats.TotalTokens += int64(usage.TotalTokens)

	cost := t.calculateCost(usage.Model, usage.InputTokens, usage.OutputTokens)
	t.stats.TotalCost += cost

	t.stats.SessionCount++
	t.stats.AgentCount++

	t.stats.LastUpdated = time.Now()

	logger.Info(fmt.Sprintf("Recorded token usage: session=%s, agent=%s, total=%d, cost=$%.4f",
		usage.SessionID, usage.AgentID, usage.TotalTokens, cost))

	return nil
}

func (t *TokenService) calculateCost(model string, inputTokens, outputTokens int) float64 {
	price, exists := t.modelPrices[model]
	if !exists {
		price = ModelPrice{InputPrice: 0.01, OutputPrice: 0.03}
	}

	inputCost := float64(inputTokens) / 1000 * price.InputPrice
	outputCost := float64(outputTokens) / 1000 * price.OutputPrice

	return inputCost + outputCost
}

func (t *TokenService) GetStats(ctx context.Context) (*TokenStats, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	stats := t.stats

	stats.TopAgents = t.getTopAgents()
	stats.TopModels = t.getTopModels()

	logger.Info(fmt.Sprintf("Token stats: total=%d, cost=$%.2f", stats.TotalTokens, stats.TotalCost))

	return &stats, nil
}

func (t *TokenService) getTopAgents() []AgentTokenUsage {
	agentMap := make(map[string]int64)

	for _, usage := range t.usages {
		agentMap[usage.AgentID] += int64(usage.TotalTokens)
	}

	topAgents := []AgentTokenUsage{}
	for agentID, tokens := range agentMap {
		cost := float64(tokens) / 1000 * 0.03
		topAgents = append(topAgents, AgentTokenUsage{
			AgentID:     agentID,
			TotalTokens: tokens,
			Cost:        cost,
		})
	}

	return topAgents
}

func (t *TokenService) getTopModels() []ModelTokenUsage {
	modelMap := make(map[string]int64)

	for _, usage := range t.usages {
		modelMap[usage.Model] += int64(usage.TotalTokens)
	}

	topModels := []ModelTokenUsage{}
	for model, tokens := range modelMap {
		cost := float64(tokens) / 1000 * 0.03
		topModels = append(topModels, ModelTokenUsage{
			Model:       model,
			TotalTokens: tokens,
			Cost:        cost,
		})
	}

	return topModels
}

func (t *TokenService) GetSessionUsage(ctx context.Context, sessionID string) ([]TokenUsage, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var sessionUsages []TokenUsage
	for _, usage := range t.usages {
		if usage.SessionID == sessionID {
			sessionUsages = append(sessionUsages, usage)
		}
	}

	logger.Info(fmt.Sprintf("Session usage: session=%s, count=%d", sessionID, len(sessionUsages)))

	return sessionUsages, nil
}

func (t *TokenService) GetAgentUsage(ctx context.Context, agentID string) ([]TokenUsage, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var agentUsages []TokenUsage
	for _, usage := range t.usages {
		if usage.AgentID == agentID {
			agentUsages = append(agentUsages, usage)
		}
	}

	return agentUsages, nil
}

func (t *TokenService) EstimateCost(model string, estimatedTokens int) float64 {
	return t.calculateCost(model, estimatedTokens/2, estimatedTokens/2)
}

func (t *TokenService) GetCostBreakdown(ctx context.Context) (map[string]interface{}, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	breakdown := map[string]interface{}{
		"total_cost":          t.stats.TotalCost,
		"input_cost_ratio":    0.4,
		"output_cost_ratio":   0.6,
		"by_agent":            t.getTopAgents(),
		"by_model":            t.getTopModels(),
		"average_per_session": t.stats.TotalCost / float64(t.stats.SessionCount),
		"average_per_agent":   t.stats.TotalCost / float64(t.stats.AgentCount),
	}

	return breakdown, nil
}

func (t *TokenService) ClearHistory(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.usages = []TokenUsage{}
	t.stats = TokenStats{}

	logger.Info("Token history cleared")

	return nil
}

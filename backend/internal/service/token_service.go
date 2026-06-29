package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/aiops/AiOpsHub/backend/pkg/llm"
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
	TotalTokens       int64             `json:"total_tokens"`
	TotalCost         float64           `json:"total_cost"`
	TodayInputTokens  int64             `json:"today_input_tokens"`
	TodayOutputTokens int64             `json:"today_output_tokens"`
	TodayTotalTokens  int64             `json:"today_total_tokens"`
	TodayCost         float64           `json:"today_cost"`
	MonthInputTokens  int64             `json:"month_input_tokens"`
	MonthOutputTokens int64             `json:"month_output_tokens"`
	MonthTotalTokens  int64             `json:"month_total_tokens"`
	MonthCost         float64           `json:"month_cost"`
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
	repo        *repository.TokenRepository
	usages      []TokenUsage
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
		modelPrices: map[string]ModelPrice{
			"gpt-4":         {InputPrice: 0.03, OutputPrice: 0.06},
			"gpt-4-turbo":   {InputPrice: 0.01, OutputPrice: 0.03},
			"gpt-3.5-turbo": {InputPrice: 0.0015, OutputPrice: 0.002},
			"qwen3.7-max":   {InputPrice: 0.002, OutputPrice: 0.006},
			"doubao-4o":     {InputPrice: 0.0008, OutputPrice: 0.002},
		},
	}

	logger.Debug("Token Service created")
	return service
}

func NewTokenServiceWithRepo(repo *repository.TokenRepository) *TokenService {
	service := NewTokenService()
	service.repo = repo
	return service
}

func (t *TokenService) RecordUsage(ctx context.Context, usage TokenUsage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	usage.Timestamp = time.Now()
	t.usages = append(t.usages, usage)

	if t.repo != nil {
		cost := t.calculateCost(usage.Model, usage.InputTokens, usage.OutputTokens)
		record := &model.TokenUsageRecord{
			SessionID:    usage.SessionID,
			AgentID:      usage.AgentID,
			Model:        usage.Model,
			InputTokens:  usage.InputTokens,
			OutputTokens: usage.OutputTokens,
			TotalTokens:  usage.TotalTokens,
			Cost:         cost,
			CreatedAt:    usage.Timestamp,
		}
		if err := t.repo.Create(record); err != nil {
			logger.Error(fmt.Sprintf("Failed to save token usage to database: %v", err))
		}
	}

	logger.Debug(fmt.Sprintf("Recorded token usage: session=%s, agent=%s, total=%d",
		usage.SessionID, usage.AgentID, usage.TotalTokens))

	return nil
}

func (t *TokenService) RecordUsageFromData(ctx context.Context, data llm.TokenUsageData) error {
	usage := TokenUsage{
		SessionID:    data.SessionID,
		AgentID:      data.AgentID,
		Model:        data.Model,
		InputTokens:  data.InputTokens,
		OutputTokens: data.OutputTokens,
		TotalTokens:  data.TotalTokens,
		Timestamp:    time.Now(),
	}
	return t.RecordUsage(ctx, usage)
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
	stats := &TokenStats{
		LastUpdated: time.Now(),
	}

	if t.repo != nil {
		_, _, totalTokens, totalCost, err := t.repo.GetTotalStats()
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get total stats: %v", err))
		} else {
			stats.TotalTokens = totalTokens
			stats.TotalCost = totalCost
		}

		todayInput, todayOutput, todayTotal, todayCost, err := t.repo.GetTodayStats()
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get today stats: %v", err))
		} else {
			stats.TodayInputTokens = todayInput
			stats.TodayOutputTokens = todayOutput
			stats.TodayTotalTokens = todayTotal
			stats.TodayCost = todayCost
		}

		monthInput, monthOutput, monthTotal, monthCost, err := t.repo.GetMonthStats()
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get month stats: %v", err))
		} else {
			stats.MonthInputTokens = monthInput
			stats.MonthOutputTokens = monthOutput
			stats.MonthTotalTokens = monthTotal
			stats.MonthCost = monthCost
		}

		topAgents, err := t.repo.GetTopAgents(10)
		if err == nil {
			for _, agent := range topAgents {
				agentID, ok := agent["agent_id"].(string)
				if !ok {
					agentID = "unknown"
				}
				totalTokens, ok := agent["total_tokens"].(int64)
				if !ok {
					totalTokens = 0
				}
				cost, ok := agent["cost"].(float64)
				if !ok {
					cost = 0
				}
				stats.TopAgents = append(stats.TopAgents, AgentTokenUsage{
					AgentID:     agentID,
					TotalTokens: totalTokens,
					Cost:        cost,
				})
			}
		}

		topModels, err := t.repo.GetTopModels(10)
		if err == nil {
			for _, m := range topModels {
				modelName, ok := m["model"].(string)
				if !ok {
					modelName = "unknown"
				}
				totalTokens, ok := m["total_tokens"].(int64)
				if !ok {
					totalTokens = 0
				}
				cost, ok := m["cost"].(float64)
				if !ok {
					cost = 0
				}
				stats.TopModels = append(stats.TopModels, ModelTokenUsage{
					Model:       modelName,
					TotalTokens: totalTokens,
					Cost:        cost,
				})
			}
		}
	} else {
		t.mu.RLock()
		defer t.mu.RUnlock()

		var totalInput, totalOutput int64
		for _, usage := range t.usages {
			totalInput += int64(usage.InputTokens)
			totalOutput += int64(usage.OutputTokens)
		}
		stats.TotalTokens = totalInput + totalOutput
	}

	logger.Debug(fmt.Sprintf("Token stats: total=%d, cost=$%.2f", stats.TotalTokens, stats.TotalCost))

	return stats, nil
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

	logger.Debug(fmt.Sprintf("Session usage: session=%s, count=%d", sessionID, len(sessionUsages)))

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
	stats, err := t.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	breakdown := map[string]interface{}{
		"total_cost": stats.TotalCost,
		"today_cost": stats.TodayCost,
		"month_cost": stats.MonthCost,
		"by_agent":   stats.TopAgents,
		"by_model":   stats.TopModels,
	}

	return breakdown, nil
}

func (t *TokenService) ClearHistory(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.usages = []TokenUsage{}

	logger.Debug("Token history cleared (memory only)")

	return nil
}

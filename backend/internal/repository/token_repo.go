package repository

import (
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(record *model.TokenUsageRecord) error {
	if record.ID == "" {
		record.ID = uuid.New().String()
	}
	return r.db.Create(record).Error
}

func (r *TokenRepository) GetTotalStats() (totalInput, totalOutput, totalTokens int64, totalCost float64, err error) {
	var stats struct {
		TotalInput  int64
		TotalOutput int64
		TotalTokens int64
		TotalCost   float64
	}

	err = r.db.Model(&model.TokenUsageRecord{}).
		Select("COALESCE(SUM(input_tokens), 0) as total_input, COALESCE(SUM(output_tokens), 0) as total_output, COALESCE(SUM(total_tokens), 0) as total_tokens, COALESCE(SUM(cost), 0) as total_cost").
		Scan(&stats).Error

	return stats.TotalInput, stats.TotalOutput, stats.TotalTokens, stats.TotalCost, err
}

func (r *TokenRepository) GetTodayStats() (input, output, total int64, cost float64, err error) {
	today := time.Now().Format("2006-01-02")
	todayStart, _ := time.Parse("2006-01-02", today)
	todayEnd := todayStart.Add(24 * time.Hour)

	var stats struct {
		Input  int64
		Output int64
		Total  int64
		Cost   float64
	}

	err = r.db.Model(&model.TokenUsageRecord{}).
		Select("COALESCE(SUM(input_tokens), 0) as input, COALESCE(SUM(output_tokens), 0) as output, COALESCE(SUM(total_tokens), 0) as total, COALESCE(SUM(cost), 0) as cost").
		Where("created_at >= ? AND created_at < ?", todayStart, todayEnd).
		Scan(&stats).Error

	return stats.Input, stats.Output, stats.Total, stats.Cost, err
}

func (r *TokenRepository) GetMonthStats() (input, output, total int64, cost float64, err error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	var stats struct {
		Input  int64
		Output int64
		Total  int64
		Cost   float64
	}

	err = r.db.Model(&model.TokenUsageRecord{}).
		Select("COALESCE(SUM(input_tokens), 0) as input, COALESCE(SUM(output_tokens), 0) as output, COALESCE(SUM(total_tokens), 0) as total, COALESCE(SUM(cost), 0) as cost").
		Where("created_at >= ? AND created_at < ?", monthStart, monthEnd).
		Scan(&stats).Error

	return stats.Input, stats.Output, stats.Total, stats.Cost, err
}

func (r *TokenRepository) GetTopAgents(limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := r.db.Model(&model.TokenUsageRecord{}).
		Select("agent_id, SUM(total_tokens) as total_tokens, SUM(cost) as cost").
		Where("agent_id != ''").
		Group("agent_id").
		Order("total_tokens DESC").
		Limit(limit).
		Find(&results).Error

	return results, err
}

func (r *TokenRepository) GetTopModels(limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := r.db.Model(&model.TokenUsageRecord{}).
		Select("model, SUM(total_tokens) as total_tokens, SUM(cost) as cost").
		Group("model").
		Order("total_tokens DESC").
		Limit(limit).
		Find(&results).Error

	return results, err
}

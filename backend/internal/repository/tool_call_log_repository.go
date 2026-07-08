package repository

import (
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"gorm.io/gorm"
)

type ToolCallLogRepository struct {
	db *gorm.DB
}

func NewToolCallLogRepository() *ToolCallLogRepository {
	return &ToolCallLogRepository{
		db: database.DB,
	}
}

func (r *ToolCallLogRepository) Create(log *model.ToolCallLog) error {
	return r.db.Create(log).Error
}

func (r *ToolCallLogRepository) GetBySessionID(sessionID string) ([]model.ToolCallLog, error) {
	var logs []model.ToolCallLog
	err := r.db.Where("session_id = ?", sessionID).Order("created_at DESC").Find(&logs).Error
	return logs, err
}

func (r *ToolCallLogRepository) GetByAgentID(agentID string, limit int) ([]model.ToolCallLog, error) {
	var logs []model.ToolCallLog
	err := r.db.Where("agent_id = ?", agentID).Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *ToolCallLogRepository) GetSuccessRate(agentID string) (float64, error) {
	var total, success int64
	r.db.Model(&model.ToolCallLog{}).Where("agent_id = ?", agentID).Count(&total)
	r.db.Model(&model.ToolCallLog{}).Where("agent_id = ? AND success = ?", agentID, true).Count(&success)

	if total == 0 {
		return 0, nil
	}
	return float64(success) / float64(total), nil
}

func (r *ToolCallLogRepository) GetRecent(limit int) ([]model.ToolCallLog, error) {
	var logs []model.ToolCallLog
	err := r.db.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

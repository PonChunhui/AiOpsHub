package repository

import (
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"gorm.io/gorm"
)

type RoutingLogRepository struct {
	db *gorm.DB
}

func NewRoutingLogRepository() *RoutingLogRepository {
	return &RoutingLogRepository{
		db: database.DB,
	}
}

func (r *RoutingLogRepository) Create(log *model.RoutingLog) error {
	return r.db.Create(log).Error
}

func (r *RoutingLogRepository) GetBySessionID(sessionID string) ([]model.RoutingLog, error) {
	var logs []model.RoutingLog
	err := r.db.Where("session_id = ?", sessionID).Order("created_at DESC").Find(&logs).Error
	return logs, err
}

func (r *RoutingLogRepository) GetRecent(limit int) ([]model.RoutingLog, error) {
	var logs []model.RoutingLog
	err := r.db.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *RoutingLogRepository) GetByAgentID(agentID string, limit int) ([]model.RoutingLog, error) {
	var logs []model.RoutingLog
	err := r.db.Where("selected_agent_id = ?", agentID).Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

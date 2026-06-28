package repository

import (
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
)

type AlertAnalysisRepository struct{}

func NewAlertAnalysisRepository() *AlertAnalysisRepository {
	return &AlertAnalysisRepository{}
}

func (r *AlertAnalysisRepository) Create(result *model.AlertAnalysisResult) error {
	return database.DB.Create(result).Error
}

func (r *AlertAnalysisRepository) GetByAlertID(alertID string) (*model.AlertAnalysisResult, error) {
	var result model.AlertAnalysisResult
	err := database.DB.Where("alert_id = ?", alertID).Order("created_at DESC").First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *AlertAnalysisRepository) Update(result *model.AlertAnalysisResult) error {
	return database.DB.Save(result).Error
}

func (r *AlertAnalysisRepository) List(limit, offset int) ([]model.AlertAnalysisResult, error) {
	var results []model.AlertAnalysisResult
	err := database.DB.Order("created_at DESC").Limit(limit).Offset(offset).Find(&results).Error
	return results, err
}

func (r *AlertAnalysisRepository) Delete(id string) error {
	return database.DB.Delete(&model.AlertAnalysisResult{}, "id = ?", id).Error
}

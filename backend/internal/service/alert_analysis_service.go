package service

import (
	"encoding/json"

	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/google/uuid"
)

type AlertAnalysisService struct {
	repo *repository.AlertAnalysisRepository
}

func NewAlertAnalysisService() *AlertAnalysisService {
	return &AlertAnalysisService{
		repo: repository.NewAlertAnalysisRepository(),
	}
}

func (s *AlertAnalysisService) SaveResult(alertID, status string, result map[string]interface{}, analysisText string) (*model.AlertAnalysisResult, error) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	analysisResult := &model.AlertAnalysisResult{
		ID:           uuid.New().String(),
		AlertID:      alertID,
		Status:       status,
		Result:       string(resultJSON),
		AnalysisText: analysisText,
	}

	err = s.repo.Create(analysisResult)
	if err != nil {
		return nil, err
	}

	return analysisResult, nil
}

func (s *AlertAnalysisService) GetByAlertID(alertID string) (*model.AlertAnalysisResult, error) {
	return s.repo.GetByAlertID(alertID)
}

func (s *AlertAnalysisService) UpdateStatus(id, status string) error {
	var result model.AlertAnalysisResult
	err := database.DB.First(&result, "id = ?", id).Error
	if err != nil {
		return err
	}

	result.Status = status
	return s.repo.Update(&result)
}

func (s *AlertAnalysisService) List(limit, offset int) ([]model.AlertAnalysisResult, error) {
	return s.repo.List(limit, offset)
}

func (s *AlertAnalysisService) Delete(id string) error {
	return s.repo.Delete(id)
}

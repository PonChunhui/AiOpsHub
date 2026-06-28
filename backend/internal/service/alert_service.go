package service

import (
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/google/uuid"
)

type AlertService struct {
	repo *repository.AlertRepository
}

func NewAlertService() *AlertService {
	return &AlertService{
		repo: repository.NewAlertRepository(),
	}
}

func (s *AlertService) Create(source, severity, title, description, rawData string) (*model.Alert, error) {
	alert := &model.Alert{
		ID:          uuid.New().String(),
		Source:      source,
		Severity:    severity,
		Title:       title,
		Description: description,
		Status:      "open",
		RawData:     rawData,
	}

	if err := s.repo.Create(alert); err != nil {
		return nil, err
	}

	return alert, nil
}

func (s *AlertService) GetByID(id string) (*model.Alert, error) {
	return s.repo.GetByID(id)
}

func (s *AlertService) List(limit, offset int) ([]model.Alert, error) {
	return s.repo.List(limit, offset)
}

func (s *AlertService) UpdateStatus(id, status string) (*model.Alert, error) {
	alert, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	alert.Status = status

	if err := s.repo.Update(alert); err != nil {
		return nil, err
	}

	return alert, nil
}

func (s *AlertService) GetByStatus(status string) ([]model.Alert, error) {
	return s.repo.GetByStatus(status)
}

func (s *AlertService) GetBySeverity(severity string) ([]model.Alert, error) {
	return s.repo.GetBySeverity(severity)
}

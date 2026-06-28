package repository

import (
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"gorm.io/gorm"
)

type AlertRepository struct {
	db *gorm.DB
}

func NewAlertRepository() *AlertRepository {
	return &AlertRepository{db: database.DB}
}

func (r *AlertRepository) Create(alert *model.Alert) error {
	return r.db.Create(alert).Error
}

func (r *AlertRepository) GetByID(id string) (*model.Alert, error) {
	var alert model.Alert
	err := r.db.Where("id = ?", id).First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *AlertRepository) List(limit, offset int) ([]model.Alert, error) {
	var alerts []model.Alert
	err := r.db.Limit(limit).Offset(offset).
		Order("created_at DESC").
		Find(&alerts).Error
	return alerts, err
}

func (r *AlertRepository) Update(alert *model.Alert) error {
	return r.db.Save(alert).Error
}

func (r *AlertRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.Alert{}).Error
}

func (r *AlertRepository) GetByStatus(status string) ([]model.Alert, error) {
	var alerts []model.Alert
	err := r.db.Where("status = ?", status).
		Order("created_at DESC").
		Find(&alerts).Error
	return alerts, err
}

func (r *AlertRepository) GetBySeverity(severity string) ([]model.Alert, error) {
	var alerts []model.Alert
	err := r.db.Where("severity = ?", severity).
		Order("created_at DESC").
		Find(&alerts).Error
	return alerts, err
}

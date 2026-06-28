package repository

import (
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"gorm.io/gorm"
)

type AgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository() *AgentRepository {
	return &AgentRepository{db: database.DB}
}

func (r *AgentRepository) Create(agent *model.Agent) error {
	return r.db.Create(agent).Error
}

func (r *AgentRepository) GetByID(id string) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.Where("id = ?", id).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AgentRepository) List(page, pageSize int) ([]model.Agent, int64, error) {
	var agents []model.Agent
	var total int64

	offset := (page - 1) * pageSize
	err := r.db.Model(&model.Agent{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&agents).Error
	return agents, total, err
}

func (r *AgentRepository) Update(agent *model.Agent) error {
	return r.db.Save(agent).Error
}

func (r *AgentRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.Agent{}).Error
}

func (r *AgentRepository) GetByCategory(category string) ([]model.Agent, error) {
	var agents []model.Agent
	err := r.db.Where("category = ? AND enabled = ?", category, true).Find(&agents).Error
	return agents, err
}

func (r *AgentRepository) ListEnabled() ([]model.Agent, error) {
	var agents []model.Agent
	err := r.db.Where("enabled = ?", true).Order("created_at DESC").Find(&agents).Error
	return agents, err
}

func (r *AgentRepository) ListPresets() ([]model.Agent, error) {
	var agents []model.Agent
	err := r.db.Where("is_preset = ? AND enabled = ?", true, true).Order("category, name").Find(&agents).Error
	return agents, err
}

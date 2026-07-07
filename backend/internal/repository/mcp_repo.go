package repository

import (
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"gorm.io/gorm"
)

type MCPRepository struct {
	db *gorm.DB
}

func NewMCPRepository() *MCPRepository {
	return &MCPRepository{db: database.DB}
}

func (r *MCPRepository) Create(server *model.MCPServer) error {
	return r.db.Create(server).Error
}

func (r *MCPRepository) GetByID(id string) (*model.MCPServer, error) {
	var server model.MCPServer
	err := r.db.Where("id = ?", id).First(&server).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (r *MCPRepository) List(page, pageSize int) ([]model.MCPServer, int64, error) {
	var servers []model.MCPServer
	var total int64

	query := r.db.Model(&model.MCPServer{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("updated_at DESC, created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&servers).Error

	return servers, total, err
}

func (r *MCPRepository) ListActive() ([]model.MCPServer, error) {
	var servers []model.MCPServer
	err := r.db.Where("status = ?", "active").Order("name").Find(&servers).Error
	return servers, err
}

func (r *MCPRepository) Update(server *model.MCPServer) error {
	return r.db.Save(server).Error
}

func (r *MCPRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.MCPServer{}).Error
}

func (r *MCPRepository) GetByIDs(ids []string) ([]model.MCPServer, error) {
	var servers []model.MCPServer
	err := r.db.Where("id IN ?", ids).Find(&servers).Error
	return servers, err
}

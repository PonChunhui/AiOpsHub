package repository

import (
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"gorm.io/gorm"
)

type RAGRepository struct {
	db *gorm.DB
}

func NewRAGRepository() *RAGRepository {
	return &RAGRepository{db: database.DB}
}

func (r *RAGRepository) Create(doc *model.RAGDocument) error {
	return r.db.Create(doc).Error
}

func (r *RAGRepository) GetByID(id string) (*model.RAGDocument, error) {
	var doc model.RAGDocument
	err := r.db.Where("id = ?", id).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *RAGRepository) List(category string, search string, page, pageSize int) ([]model.RAGDocument, int64, error) {
	var docs []model.RAGDocument
	var total int64

	query := r.db.Model(&model.RAGDocument{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if search != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("updated_at DESC, created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&docs).Error

	return docs, total, err
}

func (r *RAGRepository) Update(doc *model.RAGDocument) error {
	return r.db.Save(doc).Error
}

func (r *RAGRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.RAGDocument{}).Error
}

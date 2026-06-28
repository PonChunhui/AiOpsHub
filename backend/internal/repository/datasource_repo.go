package repository

import (
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"gorm.io/gorm"
)

type DatasourceRepository struct {
	db *gorm.DB
}

func NewDatasourceRepository() *DatasourceRepository {
	return &DatasourceRepository{db: database.DB}
}

func (r *DatasourceRepository) Create(datasource *model.Datasource) error {
	return r.db.Create(datasource).Error
}

func (r *DatasourceRepository) GetByID(id string) (*model.Datasource, error) {
	var datasource model.Datasource
	err := r.db.Where("id = ?", id).First(&datasource).Error
	if err != nil {
		return nil, err
	}
	return &datasource, nil
}

func (r *DatasourceRepository) List(limit, offset int) ([]model.Datasource, error) {
	var datasources []model.Datasource
	err := r.db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&datasources).Error
	return datasources, err
}

func (r *DatasourceRepository) Update(datasource *model.Datasource) error {
	return r.db.Save(datasource).Error
}

func (r *DatasourceRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.Datasource{}).Error
}

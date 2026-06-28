package service

import (
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/google/uuid"
)

type DatasourceService struct {
	repo *repository.DatasourceRepository
}

func NewDatasourceService() *DatasourceService {
	return &DatasourceService{
		repo: repository.NewDatasourceRepository(),
	}
}

func (s *DatasourceService) Create(name, datasourceType, config string) (*model.Datasource, error) {
	datasource := &model.Datasource{
		ID:     uuid.New().String(),
		Name:   name,
		Type:   datasourceType,
		Config: config,
		Status: "active",
	}

	if err := s.repo.Create(datasource); err != nil {
		return nil, err
	}

	return datasource, nil
}

func (s *DatasourceService) GetByID(id string) (*model.Datasource, error) {
	return s.repo.GetByID(id)
}

func (s *DatasourceService) List(limit, offset int) ([]model.Datasource, error) {
	return s.repo.List(limit, offset)
}

func (s *DatasourceService) Update(id, name, datasourceType, config, status string) (*model.Datasource, error) {
	datasource, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		datasource.Name = name
	}
	if datasourceType != "" {
		datasource.Type = datasourceType
	}
	if config != "" {
		datasource.Config = config
	}
	if status != "" {
		datasource.Status = status
	}

	if err := s.repo.Update(datasource); err != nil {
		return nil, err
	}

	return datasource, nil
}

func (s *DatasourceService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *DatasourceService) Test(id string) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	return nil
}

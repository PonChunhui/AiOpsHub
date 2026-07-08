package repository

import (
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/pkg/repository"
	"gorm.io/gorm"
)

type ToolRepository struct {
	db *gorm.DB
}

func NewToolRepository() *ToolRepository {
	return &ToolRepository{db: database.DB}
}

func (r *ToolRepository) Create(tool *model.Tool) error {
	return r.db.Create(tool).Error
}

func (r *ToolRepository) GetByID(id string) (*model.Tool, error) {
	var tool model.Tool
	err := r.db.Where("id = ?", id).First(&tool).Error
	if err != nil {
		return nil, err
	}
	return &tool, nil
}

func (r *ToolRepository) GetByName(name string) (*model.Tool, error) {
	var tool model.Tool
	err := r.db.Where("name = ?", name).First(&tool).Error
	if err != nil {
		return nil, err
	}
	return &tool, nil
}

func (r *ToolRepository) List(page, pageSize int) ([]model.Tool, int64, error) {
	var tools []model.Tool
	var total int64

	total, err := repository.Count(r.db, &model.Tool{})
	if err != nil {
		return nil, 0, err
	}

	err = repository.Paginate(r.db.Order("created_at DESC"), page, pageSize).Find(&tools).Error
	return tools, total, err
}

func (r *ToolRepository) Update(tool *model.Tool) error {
	return r.db.Save(tool).Error
}

func (r *ToolRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.Tool{}).Error
}

func (r *ToolRepository) ListEnabled() ([]model.Tool, error) {
	var tools []model.Tool
	err := r.db.Where("enabled = ?", true).Find(&tools).Error
	return tools, err
}

func (r *ToolRepository) BindToolToAgent(binding *model.AgentTool) error {
	return r.db.Create(binding).Error
}

func (r *ToolRepository) UnbindToolFromAgent(agentID, toolID string) error {
	return r.db.Where("agent_id = ? AND tool_id = ?", agentID, toolID).Delete(&model.AgentTool{}).Error
}

func (r *ToolRepository) GetAgentTools(agentID string) ([]model.AgentTool, error) {
	var bindings []model.AgentTool
	err := r.db.Where("agent_id = ? AND enabled = ?", agentID, true).Order("priority DESC").Find(&bindings).Error
	return bindings, err
}

func (r *ToolRepository) GetToolBindings(toolID string) ([]model.AgentTool, error) {
	var bindings []model.AgentTool
	err := r.db.Where("tool_id = ?", toolID).Find(&bindings).Error
	return bindings, err
}

func (r *ToolRepository) UpdateAgentToolBinding(binding *model.AgentTool) error {
	return r.db.Save(binding).Error
}

func (r *ToolRepository) SaveSSHAuditLog(log *model.SSHAuditLog) error {
	return r.db.Create(log).Error
}

func (r *ToolRepository) GetAgentToolBinding(agentID, toolID string) (*model.AgentTool, error) {
	var binding model.AgentTool
	err := r.db.Where("agent_id = ? AND tool_id = ?", agentID, toolID).First(&binding).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

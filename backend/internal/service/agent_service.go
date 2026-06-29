package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/google/uuid"
)

type AgentService struct {
	BaseService
	repo     *repository.AgentRepository
	toolRepo *repository.ToolRepository
}

func NewAgentService() *AgentService {
	return &AgentService{
		repo:     repository.NewAgentRepository(),
		toolRepo: repository.NewToolRepository(),
	}
}

func (s *AgentService) Create(name, avatar, role, category, description, systemPrompt, modelName string, temperature float64, isPreset bool) (*model.Agent, error) {
	agent := &model.Agent{
		ID:           uuid.New().String(),
		Name:         name,
		Avatar:       avatar,
		Role:         role,
		Category:     category,
		Description:  description,
		SystemPrompt: systemPrompt,
		Model:        modelName,
		Temperature:  temperature,
		IsPreset:     isPreset,
		Enabled:      true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(agent); err != nil {
		return nil, s.HandleError(err, "创建Agent失败")
	}

	s.LogInfo("Agent创建成功: %s (%s)", name, agent.ID)
	return agent, nil
}

func (s *AgentService) GetByID(id string) (*model.Agent, error) {
	return s.repo.GetByID(id)
}

func (s *AgentService) List(page, pageSize int) ([]model.Agent, int64, error) {
	return s.repo.List(page, pageSize)
}

func (s *AgentService) Update(id string, updates map[string]interface{}) (*model.Agent, error) {
	agent, err := s.repo.GetByID(id)
	if err != nil {
		return nil, NewServiceError(AgentNotFound, "Agent不存在", err)
	}

	if name, ok := updates["name"].(string); ok && name != "" {
		agent.Name = name
	}
	if avatar, ok := updates["avatar"].(string); ok {
		agent.Avatar = avatar
	}
	if role, ok := updates["role"].(string); ok {
		agent.Role = role
	}
	if category, ok := updates["category"].(string); ok {
		agent.Category = category
	}
	if description, ok := updates["description"].(string); ok {
		agent.Description = description
	}
	if systemPrompt, ok := updates["system_prompt"].(string); ok {
		agent.SystemPrompt = systemPrompt
	}
	if model, ok := updates["model"].(string); ok {
		agent.Model = model
	}
	if temperature, ok := updates["temperature"].(float64); ok {
		agent.Temperature = temperature
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		agent.Enabled = enabled
	}

	agent.UpdatedAt = time.Now()

	if err := s.repo.Update(agent); err != nil {
		return nil, s.HandleError(err, "更新Agent失败")
	}

	s.LogInfo("Agent更新成功: %s", id)
	return agent, nil
}

func (s *AgentService) Delete(id string) error {
	err := s.repo.Delete(id)
	if err != nil {
		return s.HandleError(err, "删除Agent失败")
	}
	s.LogInfo("Agent删除成功: %s", id)
	return nil
}

func (s *AgentService) ListEnabled() ([]model.Agent, error) {
	return s.repo.ListEnabled()
}

func (s *AgentService) ListPresets() ([]model.Agent, error) {
	return s.repo.ListPresets()
}

func (s *AgentService) GetByCategory(category string) ([]model.Agent, error) {
	return s.repo.GetByCategory(category)
}

func (s *AgentService) ToggleEnabled(id string) (*model.Agent, error) {
	agent, err := s.repo.GetByID(id)
	if err != nil {
		return nil, NewServiceError(AgentNotFound, "Agent不存在", err)
	}

	agent.Enabled = !agent.Enabled
	agent.UpdatedAt = time.Now()

	if err := s.repo.Update(agent); err != nil {
		return nil, s.HandleError(err, "切换Agent状态失败")
	}

	s.LogInfo("Agent状态切换: %s, enabled: %v", id, agent.Enabled)
	return agent, nil
}

func (s *AgentService) InitializePresets() error {
	presetAgents := GetPresetAgents()

	s.LogInfo("开始初始化 %d 个预设Agent...", len(presetAgents))

	for _, preset := range presetAgents {
		_, err := s.repo.GetByID(preset.ID)
		if err == nil {
			s.LogInfo("预设Agent %s (%s) 已存在，跳过", preset.Name, preset.Avatar)
		} else {
			if err := s.repo.Create(&preset); err != nil {
				s.LogError("创建预设Agent %s 失败: %v", preset.Name, err)
			} else {
				s.LogInfo("创建预设Agent %s (%s) 成功", preset.Name, preset.Avatar)
			}
		}
	}

	s.LogInfo("预设Agent初始化完成")
	return nil
}

func (s *AgentService) ForceResetPresets() error {
	s.LogInfo("开始强制重置预设Agent...")

	presetAgents := GetPresetAgents()

	for _, preset := range presetAgents {
		existing, err := s.repo.GetByID(preset.ID)
		if err == nil {
			preset.Enabled = existing.Enabled
			preset.UpdatedAt = time.Now()

			if err := s.repo.Update(&preset); err != nil {
				s.LogError("重置预设Agent %s 失败: %v", preset.Name, err)
			} else {
				s.LogInfo("重置预设Agent %s (%s) 成功", preset.Name, preset.Avatar)
			}
		} else {
			if err := s.repo.Create(&preset); err != nil {
				s.LogError("创建预设Agent %s 失败: %v", preset.Name, err)
			} else {
				s.LogInfo("创建预设Agent %s (%s) 成功", preset.Name, preset.Avatar)
			}
		}
	}

	s.LogInfo("预设Agent强制重置完成")
	return nil
}

func (s *AgentService) BindToolToAgent(ctx context.Context, agentID, toolID string, configOverride map[string]interface{}) error {
	_, err := s.repo.GetByID(agentID)
	if err != nil {
		return NewServiceError(AgentNotFound, "Agent不存在", err)
	}

	_, err = s.toolRepo.GetByID(toolID)
	if err != nil {
		return NewServiceError(ToolNotFound, "Tool不存在", err)
	}

	configJSON, _ := json.Marshal(configOverride)

	binding := &model.AgentTool{
		ID:             uuid.New().String(),
		AgentID:        agentID,
		ToolID:         toolID,
		ConfigOverride: string(configJSON),
		Enabled:        true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.toolRepo.BindToolToAgent(binding); err != nil {
		return s.HandleError(err, "绑定工具到Agent失败")
	}

	s.LogInfo("工具绑定成功: Agent %s -> Tool %s", agentID, toolID)
	return nil
}

func (s *AgentService) UnbindToolFromAgent(ctx context.Context, agentID, toolID string) error {
	if err := s.toolRepo.UnbindToolFromAgent(agentID, toolID); err != nil {
		return s.HandleError(err, "解绑工具失败")
	}

	s.LogInfo("工具解绑成功: Agent %s - Tool %s", agentID, toolID)
	return nil
}

func (s *AgentService) GetAgentTools(ctx context.Context, agentID string) ([]model.Tool, []model.AgentTool, error) {
	bindings, err := s.toolRepo.GetAgentTools(agentID)
	if err != nil {
		return nil, nil, s.HandleError(err, "获取Agent工具列表失败")
	}

	tools := []model.Tool{}
	for _, binding := range bindings {
		tool, err := s.toolRepo.GetByID(binding.ToolID)
		if err != nil {
			s.LogError("获取工具 %s 失败: %v", binding.ToolID, err)
			continue
		}

		if tool.Enabled {
			tools = append(tools, *tool)
		}
	}

	return tools, bindings, nil
}

func (s *AgentService) UpdateAgentToolConfig(ctx context.Context, agentID, toolID string, configOverride map[string]interface{}) error {
	bindings, err := s.toolRepo.GetAgentTools(agentID)
	if err != nil {
		return s.HandleError(err, "获取Agent工具绑定失败")
	}

	for _, binding := range bindings {
		if binding.ToolID == toolID {
			configJSON, _ := json.Marshal(configOverride)
			binding.ConfigOverride = string(configJSON)
			binding.UpdatedAt = time.Now()

			if err := s.toolRepo.UpdateAgentToolBinding(&binding); err != nil {
				return s.HandleError(err, "更新工具配置失败")
			}

			s.LogInfo("Agent工具配置更新成功: Agent %s -> Tool %s", agentID, toolID)
			return nil
		}
	}

	return NewServiceError(EntityNotFound, "工具绑定不存在", nil)
}

func (s *AgentService) ToggleAgentToolEnabled(ctx context.Context, agentID, toolID string) error {
	bindings, err := s.toolRepo.GetAgentTools(agentID)
	if err != nil {
		return s.HandleError(err, "获取Agent工具绑定失败")
	}

	for _, binding := range bindings {
		if binding.ToolID == toolID {
			binding.Enabled = !binding.Enabled
			binding.UpdatedAt = time.Now()

			if err := s.toolRepo.UpdateAgentToolBinding(&binding); err != nil {
				return s.HandleError(err, "更新工具状态失败")
			}

			s.LogInfo("Agent工具状态切换: Agent %s -> Tool %s, Enabled: %v", agentID, toolID, binding.Enabled)
			return nil
		}
	}

	return NewServiceError(EntityNotFound, "工具绑定不存在", nil)
}

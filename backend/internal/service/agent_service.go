package service

import (
	"fmt"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/google/uuid"
)

type AgentService struct {
	repo *repository.AgentRepository
}

func NewAgentService() *AgentService {
	return &AgentService{
		repo: repository.NewAgentRepository(),
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
		return nil, err
	}

	logger.Info("Agent created: " + name)
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
		return nil, err
	}

	// 应用更新
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
		return nil, err
	}

	logger.Info("Agent updated: " + id)
	return agent, nil
}

func (s *AgentService) Delete(id string) error {
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	logger.Info("Agent deleted: " + id)
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
		return nil, err
	}

	agent.Enabled = !agent.Enabled
	agent.UpdatedAt = time.Now()

	if err := s.repo.Update(agent); err != nil {
		return nil, err
	}

	logger.Info("Agent toggled: " + id + ", enabled: " + boolToString(agent.Enabled))
	return agent, nil
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func (s *AgentService) InitializePresets() error {
	presetAgents := GetPresetAgents()

	logger.Info(fmt.Sprintf("开始初始化 %d 个预设 Agent...", len(presetAgents)))

	for _, preset := range presetAgents {
		// 检查是否已存在
		_, err := s.repo.GetByID(preset.ID)
		if err == nil {
			// 已存在，跳过更新（保留用户的自定义配置）
			logger.Info(fmt.Sprintf("⏭️  预设 Agent %s (%s) 已存在，跳过更新", preset.Name, preset.Avatar))
		} else {
			// 不存在，创建
			if err := s.repo.Create(&preset); err != nil {
				logger.Error(fmt.Sprintf("创建预设 Agent %s 失败: %v", preset.Name, err))
			} else {
				logger.Info(fmt.Sprintf("✅ 创建预设 Agent %s (%s) 成功", preset.Name, preset.Avatar))
			}
		}
	}

	logger.Info("🎉 预设 Agent 初始化完成")
	return nil
}

// ForceResetPresets 强制重置所有预设 Agent 到默认配置
// 用于系统维护或恢复默认设置
func (s *AgentService) ForceResetPresets() error {
	logger.Info("⚠️  开始强制重置预设 Agent...")

	presetAgents := GetPresetAgents()

	for _, preset := range presetAgents {
		// 检查是否已存在
		existing, err := s.repo.GetByID(preset.ID)
		if err == nil {
			// 已存在，强制更新到默认配置（保留 enabled 状态）
			preset.Enabled = existing.Enabled
			preset.UpdatedAt = time.Now()

			if err := s.repo.Update(&preset); err != nil {
				logger.Error(fmt.Sprintf("重置预设 Agent %s 失败: %v", preset.Name, err))
			} else {
				logger.Info(fmt.Sprintf("🔄 重置预设 Agent %s (%s) 成功", preset.Name, preset.Avatar))
			}
		} else {
			// 不存在，创建
			if err := s.repo.Create(&preset); err != nil {
				logger.Error(fmt.Sprintf("创建预设 Agent %s 失败: %v", preset.Name, err))
			} else {
				logger.Info(fmt.Sprintf("✅ 创建预设 Agent %s (%s) 成功", preset.Name, preset.Avatar))
			}
		}
	}

	logger.Info("🎉 预设 Agent 强制重置完成")
	return nil
}

package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/google/uuid"
)

type ToolService struct {
	BaseService
	repo *repository.ToolRepository
}

func NewToolService(repo *repository.ToolRepository) *ToolService {
	return &ToolService{
		repo: repo,
	}
}

func ValidateToolConfig(tool *model.Tool) error {
	if tool.Name == "" {
		return fmt.Errorf("tool name is required")
	}

	if tool.ParametersSchema != "" {
		var schema map[string]interface{}
		if err := json.Unmarshal([]byte(tool.ParametersSchema), &schema); err != nil {
			return fmt.Errorf("invalid parameters schema: %w", err)
		}
	}

	return nil
}

func (s *ToolService) Create(tool *model.Tool) (*model.Tool, error) {
	if err := ValidateToolConfig(tool); err != nil {
		return nil, NewServiceError(InvalidParameter, "Tool 配置验证失败", err)
	}

	tool.ID = uuid.New().String()
	tool.CreatedAt = time.Now()
	tool.UpdatedAt = time.Now()

	if err := s.repo.Create(tool); err != nil {
		return nil, s.HandleError(err, "创建工具失败")
	}

	s.LogInfo("工具创建成功: %s (%s)", tool.Name, tool.ID)
	return tool, nil
}

func (s *ToolService) GetByID(id string) (*model.Tool, error) {
	tool, err := s.repo.GetByID(id)
	if err != nil {
		return nil, NewServiceError(ToolNotFound, "工具不存在", err)
	}
	return tool, nil
}

func (s *ToolService) GetByName(name string) (*model.Tool, error) {
	tool, err := s.repo.GetByName(name)
	if err != nil {
		return nil, NewServiceError(ToolNotFound, "工具不存在", err)
	}
	return tool, nil
}

func (s *ToolService) List(page, pageSize int) ([]model.Tool, int64, error) {
	return s.repo.List(page, pageSize)
}

func (s *ToolService) Update(tool *model.Tool) (*model.Tool, error) {
	if err := ValidateToolConfig(tool); err != nil {
		return nil, NewServiceError(InvalidParameter, "Tool 配置验证失败", err)
	}

	tool.UpdatedAt = time.Now()

	if err := s.repo.Update(tool); err != nil {
		return nil, s.HandleError(err, "更新工具失败")
	}

	s.LogInfo("工具更新成功: %s", tool.ID)
	return tool, nil
}

func (s *ToolService) Delete(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return s.HandleError(err, "删除工具失败")
	}

	s.LogInfo("工具删除成功: %s", id)
	return nil
}

func (s *ToolService) BindToAgent(agentID, toolID string, configOverride map[string]interface{}) error {
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

	if err := s.repo.BindToolToAgent(binding); err != nil {
		return s.HandleError(err, "绑定工具到Agent失败")
	}

	s.LogInfo("工具绑定成功: Agent %s -> Tool %s", agentID, toolID)
	return nil
}

func (s *ToolService) UnbindFromAgent(agentID, toolID string) error {
	if err := s.repo.UnbindToolFromAgent(agentID, toolID); err != nil {
		return s.HandleError(err, "解绑工具失败")
	}

	s.LogInfo("工具解绑成功: Agent %s - Tool %s", agentID, toolID)
	return nil
}

func (s *ToolService) GetAgentTools(agentID string) ([]model.Tool, []model.AgentTool, error) {
	bindings, err := s.repo.GetAgentTools(agentID)
	if err != nil {
		return nil, nil, s.HandleError(err, "获取Agent工具列表失败")
	}

	tools := []model.Tool{}
	for _, binding := range bindings {
		tool, err := s.repo.GetByID(binding.ToolID)
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

func (s *ToolService) UpdateAgentToolConfig(agentID, toolID string, configOverride map[string]interface{}) error {
	bindings, err := s.repo.GetAgentTools(agentID)
	if err != nil {
		return s.HandleError(err, "获取Agent工具绑定失败")
	}

	for _, binding := range bindings {
		if binding.ToolID == toolID {
			configJSON, _ := json.Marshal(configOverride)
			binding.ConfigOverride = string(configJSON)
			binding.UpdatedAt = time.Now()

			if err := s.repo.UpdateAgentToolBinding(&binding); err != nil {
				return s.HandleError(err, "更新工具配置失败")
			}

			s.LogInfo("Agent工具配置更新成功: Agent %s -> Tool %s", agentID, toolID)
			return nil
		}
	}

	return NewServiceError(EntityNotFound, "工具绑定不存在", nil)
}

func (s *ToolService) ToggleAgentToolEnabled(agentID, toolID string) error {
	bindings, err := s.repo.GetAgentTools(agentID)
	if err != nil {
		return s.HandleError(err, "获取Agent工具绑定失败")
	}

	for _, binding := range bindings {
		if binding.ToolID == toolID {
			binding.Enabled = !binding.Enabled
			binding.UpdatedAt = time.Now()

			if err := s.repo.UpdateAgentToolBinding(&binding); err != nil {
				return s.HandleError(err, "更新工具状态失败")
			}

			s.LogInfo("Agent工具状态切换: Agent %s -> Tool %s, Enabled: %v", agentID, toolID, binding.Enabled)
			return nil
		}
	}

	return NewServiceError(EntityNotFound, "工具绑定不存在", nil)
}

func (s *ToolService) InitializePresetTools() error {
	presetTools := GetPresetTools()

	s.LogInfo("开始初始化 %d 个预设工具...", len(presetTools))

	for _, preset := range presetTools {
		_, err := s.repo.GetByID(preset.ID)
		if err == nil {
			s.LogInfo("预设工具 %s (%s) 已存在，跳过", preset.Name, preset.Icon)
		} else {
			if err := s.repo.Create(&preset); err != nil {
				s.LogError("创建预设工具 %s 失败: %v", preset.Name, err)
			} else {
				s.LogInfo("创建预设工具 %s (%s) 成功", preset.Name, preset.Icon)
			}
		}
	}

	s.LogInfo("预设工具初始化完成")
	return nil
}

func (s *ToolService) GetAgentToolPool(agentID string) ([]model.Tool, error) {
	bindings, err := s.repo.GetAgentTools(agentID)
	if err != nil {
		return nil, s.HandleError(err, "获取Agent工具池失败")
	}

	tools := []model.Tool{}
	for _, binding := range bindings {
		tool, err := s.repo.GetByID(binding.ToolID)
		if err != nil {
			s.LogError("获取工具 %s 失败: %v", binding.ToolID, err)
			continue
		}

		if tool.Enabled && binding.Enabled {
			tools = append(tools, *tool)
		}
	}

	return tools, nil
}

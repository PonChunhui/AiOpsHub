package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/aiops/AiOpsHub/backend/internal/agent/eino_tools"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/cloudwego/eino/components/tool"
)

type ToolRegistry struct {
	tools    map[string]*ToolWrapper
	toolRepo *repository.ToolRepository
	hostRepo *repository.HostRepository

	toolInstanceCache map[string]tool.InvokableTool
	mu                sync.RWMutex
}

type ToolWrapper struct {
	ToolModel *model.Tool
}

var globalToolRegistry *ToolRegistry
var toolRegistryOnce sync.Once

func GetToolRegistry() *ToolRegistry {
	toolRegistryOnce.Do(func() {
		globalToolRegistry = &ToolRegistry{
			tools:             make(map[string]*ToolWrapper),
			toolRepo:          repository.NewToolRepository(),
			hostRepo:          repository.NewHostRepository(),
			toolInstanceCache: make(map[string]tool.InvokableTool),
		}
		if err := globalToolRegistry.PreloadTools(); err != nil {
			logger.Error(fmt.Sprintf("预加载工具失败: %v", err))
		}
	})
	return globalToolRegistry
}

func (r *ToolRegistry) PreloadTools() error {
	tools, err := r.toolRepo.ListEnabled()
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, toolModel := range tools {
		wrapper := &ToolWrapper{ToolModel: &toolModel}
		r.tools[toolModel.Name] = wrapper
	}

	logger.Info(fmt.Sprintf("预加载完成，共加载 %d 个工具定义", len(r.tools)))
	return nil
}

func (r *ToolRegistry) GetTool(name string) (*ToolWrapper, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	wrapper, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("工具未注册: %s", name)
	}
	return wrapper, nil
}

func (r *ToolRegistry) ExecuteTool(ctx context.Context, agentID string, toolName string, args map[string]interface{}) (string, error) {
	cacheKey := agentID + "_" + toolName

	r.mu.RLock()
	if cachedInstance, ok := r.toolInstanceCache[cacheKey]; ok {
		r.mu.RUnlock()
		argsJSON, err := json.Marshal(args)
		if err != nil {
			return "", fmt.Errorf("参数序列化失败: %w", err)
		}
		result, err := cachedInstance.InvokableRun(ctx, string(argsJSON))
		if err != nil {
			return "", fmt.Errorf("工具执行失败: %w", err)
		}
		return result, nil
	}
	r.mu.RUnlock()

	wrapper, err := r.GetTool(toolName)
	if err != nil {
		return "", err
	}

	configOverride := r.getAgentToolConfig(agentID, wrapper.ToolModel.ID)

	toolInstance := r.createToolInstance(wrapper.ToolModel, configOverride)

	r.mu.Lock()
	r.toolInstanceCache[cacheKey] = toolInstance
	r.mu.Unlock()

	argsJSON, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("参数序列化失败: %w", err)
	}

	result, err := toolInstance.InvokableRun(ctx, string(argsJSON))
	if err != nil {
		return "", fmt.Errorf("工具执行失败: %w", err)
	}

	return result, nil
}

func (r *ToolRegistry) getAgentToolConfig(agentID, toolID string) map[string]interface{} {
	binding, err := r.toolRepo.GetAgentToolBinding(agentID, toolID)
	if err != nil {
		logger.Debug(fmt.Sprintf("获取Agent工具绑定失败: %v, 使用默认配置", err))
		return nil
	}

	if binding.ConfigOverride == "" {
		return nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(binding.ConfigOverride), &config); err != nil {
		logger.Error(fmt.Sprintf("解析config_override失败: %v", err))
		return nil
	}

	logger.Debug(fmt.Sprintf("Agent %s 工具 %s 配置覆盖: %v", agentID, toolID, config))
	return config
}

func (r *ToolRegistry) createToolInstance(toolModel *model.Tool, configOverride map[string]interface{}) tool.InvokableTool {
	config := map[string]interface{}{}
	if toolModel.DefaultConfig != "" {
		json.Unmarshal([]byte(toolModel.DefaultConfig), &config)
	}

	for k, v := range configOverride {
		config[k] = v
	}

	switch toolModel.Name {
	case "ssh_exec":
		return eino_tools.NewSSHTool(toolModel, config, r.hostRepo)
	case "prometheus_query":
		return eino_tools.NewPrometheusTool(toolModel, config)
	case "kubernetes_query":
		return eino_tools.NewKubernetesTool(toolModel, config)
	case "log_query":
		return eino_tools.NewLogQueryTool(toolModel, config)
	default:
		logger.Error(fmt.Sprintf("未知工具类型: %s", toolModel.Name))
		return nil
	}
}

func (r *ToolRegistry) ClearCache() {
	r.mu.Lock()
	r.toolInstanceCache = make(map[string]tool.InvokableTool)
	r.mu.Unlock()
	logger.Info("工具实例缓存已清空")
}

func (r *ToolRegistry) ClearAgentCache(agentID string) {
	r.mu.Lock()
	for key := range r.toolInstanceCache {
		if strings.HasPrefix(key, agentID+"_") {
			delete(r.toolInstanceCache, key)
		}
	}
	r.mu.Unlock()
	logger.Info(fmt.Sprintf("Agent %s 工具实例缓存已清除", agentID))
}

func (r *ToolRegistry) ListAllTools() []model.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := []model.Tool{}
	for _, wrapper := range r.tools {
		tools = append(tools, *wrapper.ToolModel)
	}
	return tools
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/pkg/llm"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type AgentRuntime struct {
	agentSvc     *AgentService
	toolSvc      *ToolService
	mcpSvc       *MCPService
	toolRegistry *ToolRegistry
	llm          *llm.EinoLLM

	cache     *lru.Cache[string, *AgentInstance]
	mu        sync.RWMutex
	cacheSize int
}

func NewAgentRuntime(agentSvc *AgentService, toolSvc *ToolService, mcpSvc *MCPService, llm *llm.EinoLLM, cacheSize int) *AgentRuntime {
	if cacheSize <= 0 {
		cacheSize = 100
	}

	cache, err := lru.New[string, *AgentInstance](cacheSize)
	if err != nil {
		logger.Error(fmt.Sprintf("创建LRU缓存失败: %v", err))
		cache, _ = lru.New[string, *AgentInstance](10)
	}

	return &AgentRuntime{
		agentSvc:     agentSvc,
		toolSvc:      toolSvc,
		mcpSvc:       mcpSvc,
		toolRegistry: GetToolRegistry(),
		llm:          llm,
		cache:        cache,
		cacheSize:    cacheSize,
	}
}

func (r *AgentRuntime) CreateAgentInstance(ctx context.Context, agentID string) (*AgentInstance, error) {
	r.mu.RLock()
	if cached, ok := r.cache.Get(agentID); ok {
		r.mu.RUnlock()
		logger.Info(fmt.Sprintf("从缓存获取Agent实例: %s", agentID))
		return cached, nil
	}
	r.mu.RUnlock()

	instance, err := r.createInstance(ctx, agentID)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.cache.Add(agentID, instance)
	r.mu.Unlock()

	logger.Info(fmt.Sprintf("创建并缓存Agent实例: %s", agentID))

	return instance, nil
}

func (r *AgentRuntime) createInstance(ctx context.Context, agentID string) (*AgentInstance, error) {
	agentModel, err := r.agentSvc.GetByID(agentID)
	if err != nil {
		return nil, fmt.Errorf("获取Agent失败: %w", err)
	}

	toolPool, err := r.toolSvc.GetAgentToolPool(agentID)
	if err != nil {
		logger.Error(fmt.Sprintf("获取Agent工具池失败: %v", err))
		toolPool = []model.Tool{}
	}

	var mcpToolPool []model.Tool
	if agentModel.MCPServerIDs != "" && r.mcpSvc != nil {
		var mcpServerIDs []string
		if err := json.Unmarshal([]byte(agentModel.MCPServerIDs), &mcpServerIDs); err == nil {
			mcpToolPool, err = r.loadMCPToolPool(ctx, mcpServerIDs)
			if err != nil {
				logger.Error(fmt.Sprintf("加载MCP工具池失败: %v", err))
			}
		}
	}

	allTools := append(toolPool, mcpToolPool...)

	logger.Info(fmt.Sprintf("Agent %s 可用工具池: %d 内置 + %d MCP = %d 总计",
		agentModel.Name, len(toolPool), len(mcpToolPool), len(allTools)))

	maxToolCalls := agentModel.MaxToolCalls
	if maxToolCalls <= 0 {
		maxToolCalls = 5
	}

	instance := &AgentInstance{
		AgentModel:     agentModel,
		AvailableTools: allTools,
		toolRegistry:   r.toolRegistry,
		llm:            r.llm,
		maxToolCalls:   maxToolCalls,
		callHistory:    []ToolCallRecord{},
		agentID:        agentID,
	}

	return instance, nil
}

func (r *AgentRuntime) loadMCPToolPool(ctx context.Context, serverIDs []string) ([]model.Tool, error) {
	var tools []model.Tool

	for _, serverID := range serverIDs {
		mcpTools, err := r.mcpSvc.GetTools(ctx, serverID)
		if err != nil {
			logger.Error(fmt.Sprintf("获取MCP服务器 %s 工具失败: %v", serverID, err))
			continue
		}

		for _, mcpTool := range mcpTools {
			tool := model.Tool{
				ID:               fmt.Sprintf("mcp-%s-%s", serverID, mcpTool.Name),
				Name:             mcpTool.Name,
				Type:             "mcp",
				Category:         "MCP工具",
				Description:      mcpTool.Description,
				ParametersSchema: fmt.Sprintf("%v", mcpTool.InputSchema),
				Enabled:          true,
			}
			tools = append(tools, tool)
		}
	}

	return tools, nil
}

func (r *AgentRuntime) ClearCache() {
	r.mu.Lock()
	r.cache.Purge()
	r.mu.Unlock()
	logger.Info("Agent实例缓存已清空")
}

func (r *AgentRuntime) ClearAgentCache(agentID string) {
	r.mu.Lock()
	r.cache.Remove(agentID)
	r.mu.Unlock()
	logger.Info(fmt.Sprintf("Agent %s 缓存已清除", agentID))
}

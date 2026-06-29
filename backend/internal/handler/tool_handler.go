package handler

import (
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ToolHandler struct {
	BaseHandler
	toolSvc *service.ToolService
}

func NewToolHandler(toolSvc *service.ToolService) *ToolHandler {
	return &ToolHandler{
		toolSvc: toolSvc,
	}
}

func (h *ToolHandler) Create(c *gin.Context) {
	var req model.Tool
	if err := h.BindJSON(c, &req); err != nil {
		h.Error(c, err)
		return
	}

	tool, err := h.toolSvc.Create(&req)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, gin.H{
		"message": "工具创建成功",
		"tool":    tool,
	})
}

func (h *ToolHandler) List(c *gin.Context) {
	page, pageSize := h.GetPageParams(c)

	tools, total, err := h.toolSvc.List(page, pageSize)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, gin.H{
		"tools":    tools,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *ToolHandler) GetByID(c *gin.Context) {
	id := h.GetIDParam(c)

	tool, err := h.toolSvc.GetByID(id)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, tool)
}

func (h *ToolHandler) Update(c *gin.Context) {
	id := h.GetIDParam(c)

	var req model.Tool
	if err := h.BindJSON(c, &req); err != nil {
		h.Error(c, err)
		return
	}

	req.ID = id
	tool, err := h.toolSvc.Update(&req)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, gin.H{
		"message": "工具更新成功",
		"tool":    tool,
	})
}

func (h *ToolHandler) Delete(c *gin.Context) {
	id := h.GetIDParam(c)

	err := h.toolSvc.Delete(id)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, gin.H{
		"message": "工具删除成功",
	})
}

func (h *ToolHandler) BindToAgent(c *gin.Context) {
	agentID := c.Param("id")
	toolID := c.Param("tool_id")

	var req struct {
		ConfigOverride map[string]interface{} `json:"config_override"`
	}

	if err := h.BindJSON(c, &req); err != nil {
		h.Error(c, err)
		return
	}

	err := h.toolSvc.BindToAgent(agentID, toolID, req.ConfigOverride)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, gin.H{
		"message": "工具绑定成功",
	})
}

func (h *ToolHandler) GetAgentTools(c *gin.Context) {
	agentID := c.Param("id")

	tools, bindings, err := h.toolSvc.GetAgentTools(agentID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, gin.H{
		"tools":    tools,
		"bindings": bindings,
	})
}

func (h *ToolHandler) UnbindFromAgent(c *gin.Context) {
	agentID := c.Param("id")
	toolID := c.Param("tool_id")

	err := h.toolSvc.UnbindFromAgent(agentID, toolID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, gin.H{
		"message": "工具解绑成功",
	})
}

func (h *ToolHandler) UpdateAgentToolConfig(c *gin.Context) {
	agentID := c.Param("id")
	toolID := c.Param("tool_id")

	var req struct {
		ConfigOverride map[string]interface{} `json:"config_override"`
	}

	if err := h.BindJSON(c, &req); err != nil {
		h.Error(c, err)
		return
	}

	err := h.toolSvc.UpdateAgentToolConfig(agentID, toolID, req.ConfigOverride)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, gin.H{
		"message": "工具配置更新成功",
	})
}

func (h *ToolHandler) ToggleAgentToolEnabled(c *gin.Context) {
	agentID := c.Param("id")
	toolID := c.Param("tool_id")

	err := h.toolSvc.ToggleAgentToolEnabled(agentID, toolID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, gin.H{
		"message": "工具状态切换成功",
	})
}

func (h *ToolHandler) InitializePresets(c *gin.Context) {
	err := h.toolSvc.InitializePresetTools()
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, gin.H{
		"message": "预设工具初始化成功",
	})
}

func CreateTool(c *gin.Context) {
	handler := NewToolHandler(toolService)
	handler.Create(c)
}

func UpdateTool(c *gin.Context) {
	handler := NewToolHandler(toolService)
	handler.Update(c)
}

func DeleteTool(c *gin.Context) {
	handler := NewToolHandler(toolService)
	handler.Delete(c)
}

func InitPresets(c *gin.Context) {
	handler := NewToolHandler(toolService)
	handler.InitializePresets(c)
}

func BindToolToAgent(c *gin.Context) {
	handler := NewToolHandler(toolService)
	handler.BindToAgent(c)
}

func GetAgentTools(c *gin.Context) {
	handler := NewToolHandler(toolService)
	handler.GetAgentTools(c)
}

func UnbindToolFromAgent(c *gin.Context) {
	handler := NewToolHandler(toolService)
	handler.UnbindFromAgent(c)
}

func UpdateAgentToolConfig(c *gin.Context) {
	handler := NewToolHandler(toolService)
	handler.UpdateAgentToolConfig(c)
}

func ToggleAgentToolEnabled(c *gin.Context) {
	handler := NewToolHandler(toolService)
	handler.ToggleAgentToolEnabled(c)
}

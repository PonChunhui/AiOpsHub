package handler

import (
	"net/http"

	"github.com/aiops/AiOpsHub/backend/internal/agent"
	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/gin-gonic/gin"
)

func ExecuteAgent(c *gin.Context) {
	agentID := c.Param("id")

	var req struct {
		Input map[string]interface{} `json:"input"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	var dbAgent model.Agent

	if err := database.DB.Where("id = ?", agentID).First(&dbAgent).Error; err != nil {
		ErrorResponse(c, http.StatusNotFound, "agent not found in database")
		return
	}

	// 使用新的 Agent 字段构建配置
	config := agent.AgentConfig{
		Provider:     "aliyun_bailian",
		Model:        dbAgent.Model,
		Temperature:  dbAgent.Temperature,
		SystemPrompt: dbAgent.SystemPrompt,
		MaxTokens:    2000,
	}

	// 检查 Agent 是否启用
	if !dbAgent.Enabled {
		ErrorResponse(c, http.StatusBadRequest, "Agent 未启用")
		return
	}

	runtimeAgent, err := agent.NewBaseAgent(dbAgent.ID, dbAgent.Name, dbAgent.Category, dbAgent.Description, config)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to create runtime agent: "+err.Error())
		return
	}

	input := agent.AgentInput{
		Input: req.Input,
	}

	output, err := runtimeAgent.Execute(c.Request.Context(), input)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to execute agent: "+err.Error())
		return
	}

	SuccessResponse(c, output)
}

func getString(m map[string]interface{}, key string, defaultVal string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultVal
}

func getFloat(m map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return defaultVal
}

func getInt(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return defaultVal
}

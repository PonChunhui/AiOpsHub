package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/aiops/AiOpsHub/backend/internal/service"
	jwtutil "github.com/aiops/AiOpsHub/backend/pkg/jwt"
	"github.com/gin-gonic/gin"
)

var (
	alertService      *service.AlertService
	userService       *service.UserService
	datasourceService *service.DatasourceService
	serviceInitMutex  sync.Once
)

func initServices() {
	serviceInitMutex.Do(func() {
		alertService = service.NewAlertService()
		userService = service.NewUserService()
		datasourceService = service.NewDatasourceService()
	})
}

func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": message,
	})
}

func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": data,
	})
}

// ==================== Auth Handlers ====================

func Login(c *gin.Context) {
	initServices()

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "invalid request")
		return
	}

	user, err := userService.Login(req.Username, req.Password)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := jwtutil.GenerateToken(c.Request.Context(), user.ID, user.Username, user.Role)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	SuccessResponse(c, gin.H{
		"token":    token,
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

func Logout(c *gin.Context) {
	SuccessResponse(c, gin.H{"message": "logged out successfully"})
}

func Register(c *gin.Context) {
	initServices()

	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Role == "" {
		req.Role = "user"
	}

	user, err := userService.Register(req.Username, req.Email, req.Password, req.Role)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

func generateToken(userID string) string {
	return "token-" + userID
}

// ==================== Agent Handlers ====================

func ListAgents(c *gin.Context) {
	initServices()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	agents, total, err := agentService.List(page, pageSize)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "获取 Agent 列表失败")
		return
	}

	SuccessResponse(c, gin.H{
		"agents": agents,
		"total":  total,
	})
}

func GetAgent(c *gin.Context) {
	initServices()

	id := c.Param("id")

	agent, err := agentService.GetByID(id)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, "Agent 不存在")
		return
	}

	SuccessResponse(c, agent)
}

func CreateAgent(c *gin.Context) {
	initServices()

	var req struct {
		Name         string  `json:"name" binding:"required"`
		Avatar       string  `json:"avatar"`
		Role         string  `json:"role"`
		Category     string  `json:"category"`
		Description  string  `json:"description"`
		SystemPrompt string  `json:"system_prompt"`
		Model        string  `json:"model"`
		Temperature  float64 `json:"temperature"`
		IsPreset     bool    `json:"is_preset"`
		MCPServerIDs string  `json:"mcp_server_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 默认值
	if req.Avatar == "" {
		req.Avatar = "🤖"
	}
	if req.Model == "" {
		req.Model = "qwen-turbo"
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	agent, err := agentService.Create(
		req.Name,
		req.Avatar,
		req.Role,
		req.Category,
		req.Description,
		req.SystemPrompt,
		req.Model,
		req.Temperature,
		req.IsPreset,
		req.MCPServerIDs,
	)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "创建 Agent 失败")
		return
	}

	SuccessResponse(c, gin.H{
		"message": "Agent 创建成功",
		"agent":   agent,
	})
}

func UpdateAgent(c *gin.Context) {
	initServices()

	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	agent, err := agentService.Update(id, req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "更新 Agent 失败")
		return
	}

	SuccessResponse(c, gin.H{
		"message": "Agent 更新成功",
		"agent":   agent,
	})
}

func DeleteAgent(c *gin.Context) {
	initServices()

	id := c.Param("id")

	err := agentService.Delete(id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "删除 Agent 失败")
		return
	}

	SuccessResponse(c, gin.H{
		"message": "Agent 删除成功",
	})
}

func ToggleAgentEnabled(c *gin.Context) {
	initServices()

	id := c.Param("id")

	agent, err := agentService.ToggleEnabled(id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "切换 Agent 状态失败")
		return
	}

	SuccessResponse(c, gin.H{
		"message": "Agent 状态已切换",
		"agent":   agent,
	})
}

func ListPresetAgents(c *gin.Context) {
	initServices()

	agents, err := agentService.ListPresets()
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "获取预设 Agent 失败")
		return
	}

	SuccessResponse(c, gin.H{
		"agents": agents,
	})
}

func ListEnabledAgents(c *gin.Context) {
	initServices()

	agents, err := agentService.ListEnabled()
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "获取启用的 Agent 失败")
		return
	}

	SuccessResponse(c, gin.H{
		"agents": agents,
	})
}

// ==================== Alert Handlers ====================

func ListAlerts(c *gin.Context) {
	initServices()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status")
	severity := c.Query("severity")

	var alerts []interface{}
	var err error

	if status != "" {
		alertList, e := alertService.GetByStatus(status)
		err = e
		for _, a := range alertList {
			alerts = append(alerts, a)
		}
	} else if severity != "" {
		alertList, e := alertService.GetBySeverity(severity)
		err = e
		for _, a := range alertList {
			alerts = append(alerts, a)
		}
	} else {
		alertList, e := alertService.List(limit, offset)
		err = e
		for _, a := range alertList {
			alerts = append(alerts, a)
		}
	}

	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to list alerts")
		return
	}

	SuccessResponse(c, alerts)
}

func GetAlert(c *gin.Context) {
	initServices()

	id := c.Param("id")

	alert, err := alertService.GetByID(id)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, "alert not found")
		return
	}

	SuccessResponse(c, gin.H{
		"id":          alert.ID,
		"source":      alert.Source,
		"severity":    alert.Severity,
		"title":       alert.Title,
		"description": alert.Description,
		"status":      alert.Status,
		"raw_data":    alert.RawData,
		"created_at":  alert.CreatedAt,
		"updated_at":  alert.UpdatedAt,
	})
}

func CreateAlert(c *gin.Context) {
	initServices()

	var req struct {
		Source      string `json:"source" binding:"required"`
		Severity    string `json:"severity" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		RawData     string `json:"raw_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "invalid request")
		return
	}

	alert, err := alertService.Create(req.Source, req.Severity, req.Title, req.Description, req.RawData)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to create alert")
		return
	}

	SuccessResponse(c, gin.H{
		"id":       alert.ID,
		"source":   alert.Source,
		"severity": alert.Severity,
		"title":    alert.Title,
		"status":   alert.Status,
	})
}

func AlertWebhook(c *gin.Context) {
	initServices()

	var req map[string]interface{}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "invalid request")
		return
	}

	source, _ := req["source"].(string)
	severity, _ := req["severity"].(string)
	title, _ := req["title"].(string)
	description, _ := req["description"].(string)
	rawData := ""

	if rawBytes, err := json.Marshal(req); err == nil {
		rawData = string(rawBytes)
	}

	alert, err := alertService.Create(source, severity, title, description, rawData)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to create alert from webhook")
		return
	}

	SuccessResponse(c, gin.H{
		"id":       alert.ID,
		"message":  "alert received via webhook",
		"title":    alert.Title,
		"severity": alert.Severity,
	})
}

// ==================== Datasource Handlers ====================

func ListDatasources(c *gin.Context) {
	initServices()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	datasources, err := datasourceService.List(limit, offset)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to list datasources")
		return
	}

	var result []gin.H
	for _, ds := range datasources {
		result = append(result, gin.H{
			"id":     ds.ID,
			"name":   ds.Name,
			"type":   ds.Type,
			"status": ds.Status,
		})
	}

	SuccessResponse(c, result)
}

func GetDatasource(c *gin.Context) {
	initServices()

	id := c.Param("id")

	datasource, err := datasourceService.GetByID(id)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, "datasource not found")
		return
	}

	SuccessResponse(c, gin.H{
		"id":         datasource.ID,
		"name":       datasource.Name,
		"type":       datasource.Type,
		"config":     datasource.Config,
		"status":     datasource.Status,
		"created_at": datasource.CreatedAt,
		"updated_at": datasource.UpdatedAt,
	})
}

func CreateDatasource(c *gin.Context) {
	initServices()

	var req struct {
		Name   string `json:"name" binding:"required"`
		Type   string `json:"type" binding:"required"`
		Config string `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "invalid request")
		return
	}

	datasource, err := datasourceService.Create(req.Name, req.Type, req.Config)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to create datasource")
		return
	}

	SuccessResponse(c, gin.H{
		"id":     datasource.ID,
		"name":   datasource.Name,
		"type":   datasource.Type,
		"status": datasource.Status,
	})
}

func UpdateDatasource(c *gin.Context) {
	initServices()

	id := c.Param("id")

	var req struct {
		Name   string `json:"name"`
		Type   string `json:"type"`
		Config string `json:"config"`
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "invalid request")
		return
	}

	datasource, err := datasourceService.Update(id, req.Name, req.Type, req.Config, req.Status)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to update datasource")
		return
	}

	SuccessResponse(c, gin.H{
		"id":     datasource.ID,
		"name":   datasource.Name,
		"status": datasource.Status,
	})
}

func DeleteDatasource(c *gin.Context) {
	initServices()

	id := c.Param("id")

	if err := datasourceService.Delete(id); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to delete datasource")
		return
	}

	SuccessResponse(c, gin.H{"message": "datasource deleted"})
}

func TestDatasource(c *gin.Context) {
	initServices()

	id := c.Param("id")

	if err := datasourceService.Test(id); err != nil {
		SuccessResponse(c, gin.H{
			"id":      id,
			"success": false,
			"message": "connection test failed: " + err.Error(),
		})
		return
	}

	SuccessResponse(c, gin.H{
		"id":      id,
		"success": true,
		"message": "connection test successful",
	})
}

// ==================== User Handlers ====================

func ListUsers(c *gin.Context) {
	initServices()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, err := userService.List(limit, offset)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to list users")
		return
	}

	var result []gin.H
	for _, user := range users {
		result = append(result, gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		})
	}

	SuccessResponse(c, result)
}

func GetUser(c *gin.Context) {
	initServices()

	id := c.Param("id")

	user, err := userService.GetByID(id)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, "user not found")
		return
	}

	SuccessResponse(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

func UpdateUser(c *gin.Context) {
	initServices()

	id := c.Param("id")

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "invalid request")
		return
	}

	user, err := userService.Update(id, req.Email, req.Role)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to update user")
		return
	}

	SuccessResponse(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

func DeleteUser(c *gin.Context) {
	initServices()

	id := c.Param("id")

	if err := userService.Delete(id); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to delete user")
		return
	}

	SuccessResponse(c, gin.H{"message": "user deleted"})
}

// ==================== Alert Rule Handlers ====================

func ListAlertRules(c *gin.Context) {
	SuccessResponse(c, []interface{}{})
}

func CreateAlertRule(c *gin.Context) {
	SuccessResponse(c, gin.H{"message": "alert rule created"})
}

func UpdateAlertRule(c *gin.Context) {
	SuccessResponse(c, gin.H{"message": "alert rule updated"})
}

func DeleteAlertRule(c *gin.Context) {
	SuccessResponse(c, gin.H{"message": "alert rule deleted"})
}

// ==================== Tool Handlers ====================

func ListTools(c *gin.Context) {
	handler := NewToolHandler(toolService)
	handler.List(c)
}

func GetTool(c *gin.Context) {
	handler := NewToolHandler(toolService)
	handler.GetByID(c)
}

func ExecuteTool(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Tool execution not implemented yet",
	})
}

// ==================== Stats Handler ====================

func GetStats(c *gin.Context) {
	SuccessResponse(c, gin.H{
		"agents":      0,
		"alerts":      0,
		"knowledge":   0,
		"datasources": 0,
		"users":       0,
	})
}

// ==================== Performance Handler ====================

func GetPerformance(c *gin.Context) {
	SuccessResponse(c, gin.H{
		"cpu_usage":    0,
		"memory_usage": 0,
		"goroutines":   0,
	})
}

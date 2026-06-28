package handler

import (
	"net/http"
	"strconv"

	"github.com/aiops/AiOpsHub/backend/internal/service"
	"github.com/gin-gonic/gin"
)

var (
	k8sService         *service.KubernetesService
	logService         *service.LogService
	remediationService *service.AutoRemediationService
)

func InitInfraServices() {
	k8sService = service.NewKubernetesService(true)
	logService = service.NewLogService()
	remediationService = service.NewAutoRemediationService()
}

// ==================== Kubernetes ====================

func ListPods(c *gin.Context) {
	namespace := c.Query("namespace")

	pods, err := k8sService.ListPods(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pods":  pods,
		"count": len(pods),
	})
}

func GetPod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	pod, err := k8sService.GetPod(c.Request.Context(), name, namespace)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pod)
}

func GetPodLogs(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	tailLines, _ := strconv.Atoi(c.DefaultQuery("tail", "100"))

	logs, err := k8sService.GetPodLogs(c.Request.Context(), name, namespace, tailLines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pod":       name,
		"namespace": namespace,
		"logs":      logs,
	})
}

func ListDeployments(c *gin.Context) {
	namespace := c.Query("namespace")

	deployments, err := k8sService.ListDeployments(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deployments": deployments,
		"count":       len(deployments),
	})
}

func GetDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	dep, err := k8sService.GetDeployment(c.Request.Context(), name, namespace)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dep)
}

func ScaleDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var req struct {
		Replicas int `json:"replicas" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := k8sService.ScaleDeployment(c.Request.Context(), name, namespace, req.Replicas)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "deployment scaled",
		"name":      name,
		"namespace": namespace,
		"replicas":  req.Replicas,
	})
}

func RestartDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	err := k8sService.RestartDeployment(c.Request.Context(), name, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "deployment restarted",
		"name":      name,
		"namespace": namespace,
	})
}

func GetK8sResourceUsage(c *gin.Context) {
	namespace := c.Query("namespace")

	usage, err := k8sService.GetResourceUsage(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// ==================== 日志查询 ====================

func QueryLogs(c *gin.Context) {
	var req service.LogQuery

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit == 0 {
		req.Limit = 50
	}

	logs, err := logService.QueryLogs(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query": req,
		"logs":  logs,
		"count": len(logs),
	})
}

func GetLogStats(c *gin.Context) {
	stats, err := logService.GetLogStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func GetServiceLogsHandler(c *gin.Context) {
	serviceName := c.Param("service")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	logs, err := logService.GetServiceLogs(c.Request.Context(), serviceName, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service": serviceName,
		"logs":    logs,
		"count":   len(logs),
	})
}

func GetErrorLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	logs, err := logService.GetErrorLogs(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"count": len(logs),
	})
}

func SearchLogs(c *gin.Context) {
	var req struct {
		Keywords []string `json:"keywords" binding:"required"`
		Limit    int      `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit == 0 {
		req.Limit = 50
	}

	logs, err := logService.SearchLogs(c.Request.Context(), req.Keywords, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keywords": req.Keywords,
		"logs":     logs,
		"count":    len(logs),
	})
}

func GetRecentLogs(c *gin.Context) {
	minutes, _ := strconv.Atoi(c.DefaultQuery("minutes", "30"))

	logs, err := logService.GetRecentLogs(c.Request.Context(), minutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"minutes": minutes,
		"logs":    logs,
		"count":   len(logs),
	})
}

func ExportLogs(c *gin.Context) {
	format := c.DefaultQuery("format", "json")

	output, err := logService.ExportLogs(c.Request.Context(), format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"format": format,
		"data":   output,
	})
}

// ==================== 自动修复 ====================

func CreateRemediationPlan(c *gin.Context) {
	var req struct {
		AlertID   string `json:"alert_id" binding:"required"`
		AlertName string `json:"alert_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan, err := remediationService.CreatePlan(c.Request.Context(), req.AlertID, req.AlertName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plan)
}

func ExecuteRemediationPlan(c *gin.Context) {
	planID := c.Param("plan_id")

	plan, err := remediationService.ExecutePlan(c.Request.Context(), planID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plan)
}

func GetRemediationPlan(c *gin.Context) {
	planID := c.Param("plan_id")

	plan, err := remediationService.GetPlan(c.Request.Context(), planID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plan)
}

func ListRemediationPlans(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	plans, err := remediationService.ListPlans(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plans": plans,
		"count": len(plans),
	})
}

func CancelRemediationPlan(c *gin.Context) {
	planID := c.Param("plan_id")

	err := remediationService.CancelPlan(c.Request.Context(), planID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "plan cancelled",
		"plan_id": planID,
	})
}

func ApproveRemediationAction(c *gin.Context) {
	actionID := c.Param("action_id")

	err := remediationService.ApproveAction(c.Request.Context(), actionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "action approved",
		"action_id": actionID,
	})
}

func GetRemediationStats(c *gin.Context) {
	stats, err := remediationService.GetStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

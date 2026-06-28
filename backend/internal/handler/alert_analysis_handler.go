package handler

import (
	"net/http"
	"strconv"

	"github.com/aiops/AiOpsHub/backend/internal/service"
	"github.com/gin-gonic/gin"
)

var alertAnalysisService *service.AlertAnalysisService

func initAlertAnalysisServices() {
	alertAnalysisService = service.NewAlertAnalysisService()
}

func SaveAlertAnalysis(c *gin.Context) {
	initAlertAnalysisServices()

	var req struct {
		AlertID      string                 `json:"alert_id" binding:"required"`
		Status       string                 `json:"status" binding:"required"`
		Result       map[string]interface{} `json:"result"`
		AnalysisText string                 `json:"analysis_text"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	result, err := alertAnalysisService.SaveResult(req.AlertID, req.Status, req.Result, req.AnalysisText)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to save analysis result")
		return
	}

	SuccessResponse(c, result)
}

func GetAlertAnalysis(c *gin.Context) {
	initAlertAnalysisServices()

	alertID := c.Param("alert_id")

	result, err := alertAnalysisService.GetByAlertID(alertID)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, "analysis result not found")
		return
	}

	SuccessResponse(c, result)
}

func ListAlertAnalysis(c *gin.Context) {
	initAlertAnalysisServices()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	results, err := alertAnalysisService.List(limit, offset)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to list analysis results")
		return
	}

	SuccessResponse(c, results)
}

func DeleteAlertAnalysis(c *gin.Context) {
	initAlertAnalysisServices()

	id := c.Param("id")

	err := alertAnalysisService.Delete(id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "failed to delete analysis result")
		return
	}

	SuccessResponse(c, gin.H{"message": "analysis result deleted"})
}

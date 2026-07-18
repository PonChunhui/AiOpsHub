package router

import (
	"github.com/aiops/AiOpsHub/backend/internal/handler"
	"github.com/aiops/AiOpsHub/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// AlertsRouter 告警路由
type AlertsRouter struct{}

func (r *AlertsRouter) Register(engine *gin.Engine) {
	v1 := engine.Group("/api/v1")
	alerts := v1.Group("/alerts")
	alerts.Use(middleware.Auth())
	{
		alerts.GET("", handler.ListAlerts)
		alerts.GET("/:id", handler.GetAlert)
		alerts.POST("", handler.CreateAlert)
		alerts.GET("/analysis/:alert_id", handler.GetAlertAnalysis)
		alerts.POST("/analysis", handler.SaveAlertAnalysis)
		alerts.GET("/analysis/list", handler.ListAlertAnalysis)
		alerts.DELETE("/analysis/:id", handler.DeleteAlertAnalysis)
	}
}

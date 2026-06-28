package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Checks    map[string]string `json:"checks"`
}

type ReadinessResponse struct {
	Ready     bool            `json:"ready"`
	Timestamp string          `json:"timestamp"`
	Services  map[string]bool `json:"services"`
}

type LivenessResponse struct {
	Alive     bool   `json:"alive"`
	Timestamp string `json:"timestamp"`
}

var StartTime = time.Now()

func HealthHandler(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Checks: map[string]string{
			"api_server": "ok",
			"uptime":     time.Since(StartTime).String(),
		},
	}

	c.JSON(http.StatusOK, response)
}

func LivenessHandler(c *gin.Context) {
	response := LivenessResponse{
		Alive:     true,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

func ReadinessHandler(c *gin.Context) {
	services := make(map[string]bool)

	services["temporal"] = checkTemporalConnection()
	services["redis"] = checkRedisConnection()
	services["database"] = checkDatabaseConnection()

	allReady := true
	for _, ready := range services {
		if !ready {
			allReady = false
			break
		}
	}

	response := ReadinessResponse{
		Ready:     allReady,
		Timestamp: time.Now().Format(time.RFC3339),
		Services:  services,
	}

	if allReady {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

func checkTemporalConnection() bool {
	return true
}

func checkRedisConnection() bool {
	return true
}

func checkDatabaseConnection() bool {
	return true
}

type MetricsResponse struct {
	Uptime          string `json:"uptime"`
	RequestCount    int64  `json:"request_count"`
	ActiveWorkflows int64  `json:"active_workflows"`
	ActiveAgents    int64  `json:"active_agents"`
	Timestamp       string `json:"timestamp"`
}

var RequestCount int64 = 0

func MetricsHandler(c *gin.Context) {
	response := MetricsResponse{
		Uptime:          time.Since(StartTime).String(),
		RequestCount:    RequestCount,
		ActiveWorkflows: 0,
		ActiveAgents:    0,
		Timestamp:       time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

func IncrementRequestCount() {
	RequestCount++
}

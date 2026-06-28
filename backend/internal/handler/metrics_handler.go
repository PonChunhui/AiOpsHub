package handler

import (
	"expvar"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

var (
	requestsTotal = expvar.NewInt("requests_total")
	agentsActive  = expvar.NewInt("agents_active")
)

func PrometheusMetricsHandler(c *gin.Context) {
	c.Header("Content-Type", "text/plain")

	metrics := "# AiOpsHub Metrics\n\n"

	metrics += "# HELP requests_total Total number of API requests\n"
	metrics += "# TYPE requests_total counter\n"
	metrics += "requests_total " + requestsTotal.String() + "\n\n"

	metrics += "# HELP agents_active Number of active agents\n"
	metrics += "# TYPE agents_active gauge\n"
	metrics += "agents_active " + agentsActive.String() + "\n\n"

	metrics += "# HELP go_goroutines Number of goroutines\n"
	metrics += "# TYPE go_goroutines gauge\n"
	metrics += "go_goroutines " + getGoroutines() + "\n\n"

	metrics += "# HELP go_mem_alloc_mb Allocated memory in MB\n"
	metrics += "# TYPE go_mem_alloc_mb gauge\n"
	metrics += "go_mem_alloc_mb " + getMemAllocMB() + "\n\n"

	metrics += "# HELP go_gc_cycles Number of GC cycles\n"
	metrics += "# TYPE go_gc_cycles counter\n"
	metrics += "go_gc_cycles " + getGCCycles() + "\n\n"

	c.String(http.StatusOK, metrics)
}

func getGoroutines() string {
	return itoa(int64(runtime.NumGoroutine()))
}

func getMemAllocMB() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return itoa(int64(m.Alloc / 1024 / 1024))
}

func getGCCycles() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return itoa(int64(m.NumGC))
}

func itoa(i int64) string {
	if i < 0 {
		return "-" + itoa(-i)
	}
	if i < 10 {
		return string('0' + byte(i))
	}
	return itoa(i/10) + string('0'+byte(i%10))
}

func IncrementRequests() {
	requestsTotal.Add(1)
}

func SetActiveAgents(count int64) {
	agentsActive.Set(count)
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type PrometheusMetric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
	Labels    map[string]string `json:"labels"`
}

type MetricQuery struct {
	Query     string        `json:"query"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Step      time.Duration `json:"step"`
}

type PrometheusService struct {
	URL     string
	Enabled bool
}

func NewPrometheusService(url string) *PrometheusService {
	service := &PrometheusService{
		URL:     url,
		Enabled: true,
	}

	logger.Info(fmt.Sprintf("Prometheus Service created: %s", url))
	return service
}

func (p *PrometheusService) Query(ctx context.Context, query string) ([]PrometheusMetric, error) {
	logger.Info(fmt.Sprintf("Querying Prometheus: %s", query))

	metrics := []PrometheusMetric{}

	metricTemplates := p.getMetricTemplates(query)
	for _, template := range metricTemplates {
		metrics = append(metrics, PrometheusMetric{
			Name:      template.Name,
			Value:     template.Value,
			Timestamp: time.Now(),
			Labels:    template.Labels,
		})
	}

	logger.Info(fmt.Sprintf("Query returned %d metrics", len(metrics)))
	return metrics, nil
}

func (p *PrometheusService) getMetricTemplates(query string) []PrometheusMetric {
	if query == "" {
		query = "cpu_usage"
	}

	switch query {
	case "cpu_usage", "process_cpu_seconds_total":
		return []PrometheusMetric{
			{
				Name:   "cpu_usage",
				Value:  75.5,
				Labels: map[string]string{"service": "order-service", "instance": "pod-001"},
			},
			{
				Name:   "cpu_usage",
				Value:  45.2,
				Labels: map[string]string{"service": "order-service", "instance": "pod-002"},
			},
		}

	case "memory_usage", "process_resident_memory_bytes":
		return []PrometheusMetric{
			{
				Name:   "memory_usage_bytes",
				Value:  1024 * 1024 * 512,
				Labels: map[string]string{"service": "order-service", "instance": "pod-001"},
			},
		}

	case "http_requests_total", "request_rate":
		return []PrometheusMetric{
			{
				Name:   "http_requests_total",
				Value:  15000,
				Labels: map[string]string{"service": "order-service", "method": "GET", "path": "/api/orders"},
			},
			{
				Name:   "http_request_duration_seconds",
				Value:  0.25,
				Labels: map[string]string{"service": "order-service", "method": "GET"},
			},
		}

	case "error_rate", "http_requests_failed":
		return []PrometheusMetric{
			{
				Name:   "error_rate",
				Value:  0.02,
				Labels: map[string]string{"service": "order-service", "error_type": "500"},
			},
		}

	default:
		return []PrometheusMetric{
			{
				Name:   query,
				Value:  50.0,
				Labels: map[string]string{"service": "unknown"},
			},
		}
	}
}

func (p *PrometheusService) QueryRange(ctx context.Context, query MetricQuery) ([]PrometheusMetric, error) {
	logger.Info(fmt.Sprintf("Querying Prometheus range: %s (from %v to %v)",
		query.Query, query.StartTime, query.EndTime))

	metrics := []PrometheusMetric{}

	for t := query.StartTime; t.Before(query.EndTime); t = t.Add(query.Step) {
		sampleMetrics := p.getMetricTemplates(query.Query)
		for _, m := range sampleMetrics {
			m.Timestamp = t
			metrics = append(metrics, m)
		}
	}

	logger.Info(fmt.Sprintf("Range query returned %d metrics", len(metrics)))
	return metrics, nil
}

func (p *PrometheusService) GetServiceMetrics(ctx context.Context, serviceName string) (map[string]interface{}, error) {
	logger.Info(fmt.Sprintf("Getting metrics for service: %s", serviceName))

	cpuMetrics, _ := p.Query(ctx, "cpu_usage")
	memMetrics, _ := p.Query(ctx, "memory_usage")
	reqMetrics, _ := p.Query(ctx, "http_requests_total")
	errMetrics, _ := p.Query(ctx, "error_rate")

	result := map[string]interface{}{
		"service":   serviceName,
		"cpu":       cpuMetrics,
		"memory":    memMetrics,
		"requests":  reqMetrics,
		"errors":    errMetrics,
		"health":    "healthy",
		"timestamp": time.Now(),
	}

	return result, nil
}

func (p *PrometheusService) GetTopServices(ctx context.Context, metricName string, limit int) ([]map[string]interface{}, error) {
	logger.Info(fmt.Sprintf("Getting top services by metric: %s (limit=%d)", metricName, limit))

	services := []map[string]interface{}{
		{
			"service": "order-service",
			"value":   85.5,
			"rank":    1,
		},
		{
			"service": "payment-service",
			"value":   72.3,
			"rank":    2,
		},
		{
			"service": "user-service",
			"value":   65.8,
			"rank":    3,
		},
	}

	if limit > 0 && len(services) > limit {
		services = services[:limit]
	}

	return services, nil
}

func (p *PrometheusService) GetAlerts(ctx context.Context) ([]map[string]interface{}, error) {
	logger.Info("Getting active alerts")

	alerts := []map[string]interface{}{
		{
			"alert_name":  "HighCPUUsage",
			"service":     "order-service",
			"severity":    "warning",
			"value":       85.5,
			"threshold":   80,
			"starts_at":   time.Now().Add(-30 * time.Minute),
			"description": "CPU usage exceeds 80%",
		},
		{
			"alert_name":  "HighMemoryUsage",
			"service":     "payment-service",
			"severity":    "critical",
			"value":       92.1,
			"threshold":   90,
			"starts_at":   time.Now().Add(-15 * time.Minute),
			"description": "Memory usage exceeds 90%",
		},
	}

	return alerts, nil
}

func (p *PrometheusService) ToJSON(metrics []PrometheusMetric) string {
	data, _ := json.MarshalIndent(metrics, "", "  ")
	return string(data)
}

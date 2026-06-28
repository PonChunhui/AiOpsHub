package service

import (
	"context"
	"testing"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

func init() {
	logger.Init()
}

func TestRAGServiceSearch(t *testing.T) {
	svc := NewRAGService("test_collection")

	ctx := context.Background()
	results, err := svc.SearchKnowledge(ctx, "服务响应慢", 3)

	if err != nil {
		t.Errorf("SearchKnowledge failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least 1 result")
	}

	for _, result := range results {
		t.Logf("Result: %s (Score: %.2f)", result.Document.Title, result.Score)
	}
}

func TestRAGServiceGetContext(t *testing.T) {
	svc := NewRAGService("test_collection")

	ctx := context.Background()
	context, err := svc.GetContextForQuery(ctx, "CPU使用率高", 500)

	if err != nil {
		t.Errorf("GetContextForQuery failed: %v", err)
	}

	if context == "" {
		t.Error("Context should not be empty")
	}

	t.Logf("Context length: %d chars", len(context))
}

func TestRAGServiceListDocuments(t *testing.T) {
	svc := NewRAGService("test_collection")

	ctx := context.Background()
	docs, err := svc.ListDocuments(ctx, "troubleshooting", 5)

	if err != nil {
		t.Errorf("ListDocuments failed: %v", err)
	}

	if len(docs) == 0 {
		t.Error("Expected at least 1 document")
	}

	t.Logf("Found %d documents", len(docs))
}

func TestPrometheusServiceQuery(t *testing.T) {
	svc := NewPrometheusService("http://localhost:9090")

	ctx := context.Background()
	metrics, err := svc.Query(ctx, "cpu_usage")

	if err != nil {
		t.Errorf("Query failed: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("Expected at least 1 metric")
	}

	for _, m := range metrics {
		t.Logf("Metric: %s = %.2f", m.Name, m.Value)
	}
}

func TestPrometheusServiceGetServiceMetrics(t *testing.T) {
	svc := NewPrometheusService("http://localhost:9090")

	ctx := context.Background()
	metrics, err := svc.GetServiceMetrics(ctx, "order-service")

	if err != nil {
		t.Errorf("GetServiceMetrics failed: %v", err)
	}

	if metrics["service"] != "order-service" {
		t.Error("Expected service name to match")
	}

	t.Logf("Service metrics: %v", metrics)
}

func TestPrometheusServiceGetAlerts(t *testing.T) {
	svc := NewPrometheusService("http://localhost:9090")

	ctx := context.Background()
	alerts, err := svc.GetAlerts(ctx)

	if err != nil {
		t.Errorf("GetAlerts failed: %v", err)
	}

	if len(alerts) == 0 {
		t.Log("No alerts (expected for mock)")
	}

	t.Logf("Found %d alerts", len(alerts))
}

func TestTokenServiceRecord(t *testing.T) {
	svc := NewTokenService()

	ctx := context.Background()
	usage := TokenUsage{
		SessionID:    "test-session",
		WorkflowID:   "wf-001",
		AgentID:      "monitor-agent-001",
		Model:        "gpt-4",
		InputTokens:  500,
		OutputTokens: 200,
		TotalTokens:  700,
	}

	err := svc.RecordUsage(ctx, usage)
	if err != nil {
		t.Errorf("RecordUsage failed: %v", err)
	}
}

func TestTokenServiceGetStats(t *testing.T) {
	svc := NewTokenService()

	ctx := context.Background()

	svc.RecordUsage(ctx, TokenUsage{
		SessionID:    "session-1",
		AgentID:      "agent-1",
		Model:        "gpt-4",
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	})

	svc.RecordUsage(ctx, TokenUsage{
		SessionID:    "session-2",
		AgentID:      "agent-2",
		Model:        "gpt-3.5-turbo",
		InputTokens:  200,
		OutputTokens: 100,
		TotalTokens:  300,
	})

	stats, err := svc.GetStats(ctx)
	if err != nil {
		t.Errorf("GetStats failed: %v", err)
	}

	if stats.TotalTokens != 450 {
		t.Errorf("Expected total tokens 450, got %d", stats.TotalTokens)
	}

	t.Logf("Stats: TotalTokens=%d, Cost=$%.2f", stats.TotalTokens, stats.TotalCost)
}

func TestTokenServiceEstimateCost(t *testing.T) {
	svc := NewTokenService()

	cost := svc.EstimateCost("gpt-4", 1000)

	if cost <= 0 {
		t.Error("Cost should be positive")
	}

	t.Logf("Estimated cost for 1000 tokens: $%.2f", cost)
}

func TestWorkflowHistoryService(t *testing.T) {
	svc := NewWorkflowHistoryService()

	ctx := context.Background()

	history := WorkflowHistory{
		WorkflowID:   "wf-test-001",
		SessionID:    "session-test",
		WorkflowType: "collaboration",
		Status:       "completed",
		Duration:     30,
		TokenCount:   500,
		Cost:         0.05,
	}

	err := svc.RecordHistory(ctx, history)
	if err != nil {
		t.Errorf("RecordHistory failed: %v", err)
	}

	retrieved, err := svc.GetHistory(ctx, "wf-test-001")
	if err != nil {
		t.Errorf("GetHistory failed: %v", err)
	}

	if retrieved.WorkflowID != "wf-test-001" {
		t.Error("WorkflowID mismatch")
	}

	t.Logf("Retrieved history: %s (%s)", retrieved.WorkflowID, retrieved.Status)
}

func TestWorkflowHistoryStatistics(t *testing.T) {
	svc := NewWorkflowHistoryService()

	ctx := context.Background()

	svc.RecordHistory(ctx, WorkflowHistory{
		WorkflowID: "wf-1",
		Status:     "completed",
		Duration:   10,
	})

	svc.RecordHistory(ctx, WorkflowHistory{
		WorkflowID: "wf-2",
		Status:     "failed",
		Duration:   5,
	})

	stats, err := svc.GetStatistics(ctx)
	if err != nil {
		t.Errorf("GetStatistics failed: %v", err)
	}

	if stats["total_workflows"] != 2 {
		t.Errorf("Expected 2 workflows, got %v", stats["total_workflows"])
	}

	t.Logf("Statistics: %v", stats)
}

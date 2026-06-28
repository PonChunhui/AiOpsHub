package service

import (
	"context"
	"testing"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

func init() {
	logger.Init()
}

func TestKubernetesServiceListPods(t *testing.T) {
	svc := NewKubernetesService(true)

	ctx := context.Background()
	pods, err := svc.ListPods(ctx, "production")

	if err != nil {
		t.Errorf("ListPods failed: %v", err)
	}

	if len(pods) == 0 {
		t.Error("Expected at least 1 pod")
	}

	for _, pod := range pods {
		t.Logf("Pod: %s (%s)", pod.Name, pod.Status)
	}
}

func TestKubernetesServiceGetPod(t *testing.T) {
	svc := NewKubernetesService(true)

	ctx := context.Background()
	pod, err := svc.GetPod(ctx, "order-service-pod-1", "production")

	if err != nil {
		t.Errorf("GetPod failed: %v", err)
	}

	if pod == nil {
		t.Error("Pod should not be nil")
	}

	t.Logf("Pod: %s, Status: %s", pod.Name, pod.Status)
}

func TestKubernetesServiceGetPodLogs(t *testing.T) {
	svc := NewKubernetesService(true)

	ctx := context.Background()
	logs, err := svc.GetPodLogs(ctx, "order-service-pod-1", "production", 10)

	if err != nil {
		t.Errorf("GetPodLogs failed: %v", err)
	}

	if logs == "" {
		t.Error("Logs should not be empty")
	}

	t.Logf("Logs length: %d chars", len(logs))
}

func TestKubernetesServiceListDeployments(t *testing.T) {
	svc := NewKubernetesService(true)

	ctx := context.Background()
	deployments, err := svc.ListDeployments(ctx, "production")

	if err != nil {
		t.Errorf("ListDeployments failed: %v", err)
	}

	if len(deployments) == 0 {
		t.Error("Expected at least 1 deployment")
	}

	t.Logf("Found %d deployments", len(deployments))
}

func TestKubernetesServiceScaleDeployment(t *testing.T) {
	svc := NewKubernetesService(true)

	ctx := context.Background()
	err := svc.ScaleDeployment(ctx, "order-service", "production", 5)

	if err != nil {
		t.Errorf("ScaleDeployment failed: %v", err)
	}

	t.Logf("Scaled deployment successfully")
}

func TestLogServiceQueryLogs(t *testing.T) {
	svc := NewLogService()

	ctx := context.Background()
	query := LogQuery{
		Service: "order-service",
		Level:   "ERROR",
		Limit:   10,
	}

	logs, err := svc.QueryLogs(ctx, query)
	if err != nil {
		t.Errorf("QueryLogs failed: %v", err)
	}

	if len(logs) == 0 {
		t.Log("No logs found (expected for filter)")
	}

	for _, log := range logs {
		t.Logf("Log: [%s] %s - %s", log.Level, log.Service, log.Message)
	}
}

func TestLogServiceGetStatistics(t *testing.T) {
	svc := NewLogService()

	ctx := context.Background()
	stats, err := svc.GetLogStatistics(ctx)

	if err != nil {
		t.Errorf("GetLogStatistics failed: %v", err)
	}

	if stats.TotalLogs == 0 {
		t.Error("Expected at least 1 log")
	}

	t.Logf("Stats: Total=%d, Errors=%d, Warns=%d", stats.TotalLogs, stats.ErrorCount, stats.WarnCount)
}

func TestLogServiceGetErrorLogs(t *testing.T) {
	svc := NewLogService()

	ctx := context.Background()
	logs, err := svc.GetErrorLogs(ctx, 10)

	if err != nil {
		t.Errorf("GetErrorLogs failed: %v", err)
	}

	for _, log := range logs {
		if log.Level != "ERROR" {
			t.Errorf("Expected ERROR level, got %s", log.Level)
		}
	}

	t.Logf("Found %d error logs", len(logs))
}

func TestLogServiceSearchLogs(t *testing.T) {
	svc := NewLogService()

	ctx := context.Background()
	logs, err := svc.SearchLogs(ctx, []string{"timeout"}, 10)

	if err != nil {
		t.Errorf("SearchLogs failed: %v", err)
	}

	t.Logf("Search returned %d logs", len(logs))
}

func TestLogServiceExportLogs(t *testing.T) {
	svc := NewLogService()

	ctx := context.Background()

	jsonOutput, err := svc.ExportLogs(ctx, "json")
	if err != nil {
		t.Errorf("ExportLogs (json) failed: %v", err)
	}

	textOutput, err := svc.ExportLogs(ctx, "text")
	if err != nil {
		t.Errorf("ExportLogs (text) failed: %v", err)
	}

	t.Logf("JSON output: %d chars", len(jsonOutput))
	t.Logf("Text output: %d chars", len(textOutput))
}

func TestRemediationServiceCreatePlan(t *testing.T) {
	svc := NewAutoRemediationService()

	ctx := context.Background()
	plan, err := svc.CreatePlan(ctx, "alert-001", "HighCPUUsage")

	if err != nil {
		t.Errorf("CreatePlan failed: %v", err)
	}

	if plan == nil {
		t.Error("Plan should not be nil")
	}

	if len(plan.Actions) == 0 {
		t.Error("Expected at least 1 action")
	}

	t.Logf("Plan: %s with %d actions", plan.PlanID, len(plan.Actions))
}

func TestRemediationServiceExecutePlan(t *testing.T) {
	svc := NewAutoRemediationService()

	ctx := context.Background()

	plan, _ := svc.CreatePlan(ctx, "alert-002", "HighMemoryUsage")

	executedPlan, err := svc.ExecutePlan(ctx, plan.PlanID)
	if err != nil {
		t.Errorf("ExecutePlan failed: %v", err)
	}

	if executedPlan.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", executedPlan.Status)
	}

	t.Logf("Plan executed: success rate %.2f%%", executedPlan.SuccessRate*100)
}

func TestRemediationServiceGetStatistics(t *testing.T) {
	svc := NewAutoRemediationService()

	ctx := context.Background()

	svc.CreatePlan(ctx, "alert-001", "HighCPUUsage")
	svc.CreatePlan(ctx, "alert-002", "HighMemoryUsage")

	stats, err := svc.GetStatistics(ctx)
	if err != nil {
		t.Errorf("GetStatistics failed: %v", err)
	}

	if stats["total_plans"] == 0 {
		t.Error("Expected at least 1 plan")
	}

	t.Logf("Stats: %v", stats)
}

func TestRemediationServiceApproveAction(t *testing.T) {
	svc := NewAutoRemediationService()

	ctx := context.Background()

	plan, _ := svc.CreatePlan(ctx, "alert-003", "DatabaseConnectionFailed")

	for _, action := range plan.Actions {
		if action.RequiresApproval {
			err := svc.ApproveAction(ctx, action.ID)
			if err != nil {
				t.Errorf("ApproveAction failed: %v", err)
			}
			t.Logf("Approved action: %s", action.ID)
		}
	}
}

func TestRemediationDifferentAlertTypes(t *testing.T) {
	svc := NewAutoRemediationService()

	ctx := context.Background()

	alertTypes := []string{"HighCPUUsage", "HighMemoryUsage", "DatabaseConnectionFailed", "UnknownAlert"}

	for _, alertType := range alertTypes {
		plan, err := svc.CreatePlan(ctx, "alert-"+alertType, alertType)
		if err != nil {
			t.Errorf("CreatePlan for %s failed: %v", alertType, err)
			continue
		}

		t.Logf("Alert %s -> Plan with %d actions", alertType, len(plan.Actions))

		for _, action := range plan.Actions {
			t.Logf("  - Action: %s (%s, risk=%s)", action.Type, action.Target, action.RiskLevel)
		}
	}
}

func TestKubernetesServiceGetResourceUsage(t *testing.T) {
	svc := NewKubernetesService(true)

	ctx := context.Background()
	usage, err := svc.GetResourceUsage(ctx, "production")

	if err != nil {
		t.Errorf("GetResourceUsage failed: %v", err)
	}

	t.Logf("Resource usage: %v", usage)
}

func TestKubernetesServiceGetPodEvents(t *testing.T) {
	svc := NewKubernetesService(true)

	ctx := context.Background()
	events, err := svc.GetPodEvents(ctx, "order-service-pod-1", "production")

	if err != nil {
		t.Errorf("GetPodEvents failed: %v", err)
	}

	t.Logf("Found %d events", len(events))
}

func TestLogServiceGetRecentLogs(t *testing.T) {
	svc := NewLogService()

	ctx := context.Background()
	logs, err := svc.GetRecentLogs(ctx, 60)

	if err != nil {
		t.Errorf("GetRecentLogs failed: %v", err)
	}

	t.Logf("Found %d recent logs", len(logs))
}

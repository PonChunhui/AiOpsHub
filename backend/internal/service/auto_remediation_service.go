package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type RemediationAction struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	Target           string                 `json:"target"`
	Description      string                 `json:"description"`
	Status           string                 `json:"status"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          time.Time              `json:"end_time"`
	Result           map[string]interface{} `json:"result"`
	Error            string                 `json:"error"`
	RiskLevel        string                 `json:"risk_level"`
	RequiresApproval bool                   `json:"requires_approval"`
}

type RemediationPlan struct {
	PlanID      string              `json:"plan_id"`
	AlertID     string              `json:"alert_id"`
	Actions     []RemediationAction `json:"actions"`
	Status      string              `json:"status"`
	CreatedAt   time.Time           `json:"created_at"`
	ExecutedAt  time.Time           `json:"executed_at"`
	CompletedAt time.Time           `json:"completed_at"`
	SuccessRate float64             `json:"success_rate"`
}

type AutoRemediationService struct {
	plans   map[string]RemediationPlan
	actions map[string]RemediationAction
	mu      sync.RWMutex
}

func NewAutoRemediationService() *AutoRemediationService {
	svc := &AutoRemediationService{
		plans:   make(map[string]RemediationPlan),
		actions: make(map[string]RemediationAction),
	}

	logger.Info("Auto Remediation Service created")
	return svc
}

func (a *AutoRemediationService) CreatePlan(ctx context.Context, alertID, alertName string) (*RemediationPlan, error) {
	actions := a.generateActions(alertName)

	plan := RemediationPlan{
		PlanID:    fmt.Sprintf("plan-%d", time.Now().Unix()),
		AlertID:   alertID,
		Actions:   actions,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	for _, action := range actions {
		a.actions[action.ID] = action
	}

	a.mu.Lock()
	a.plans[plan.PlanID] = plan
	a.mu.Unlock()

	logger.Info(fmt.Sprintf("Created remediation plan: %s for alert: %s", plan.PlanID, alertID))

	return &plan, nil
}

func (a *AutoRemediationService) generateActions(alertName string) []RemediationAction {
	var actions []RemediationAction

	switch alertName {
	case "HighCPUUsage":
		actions = append(actions,
			RemediationAction{
				ID:               "action-001",
				Type:             "scale_out",
				Target:           "order-service",
				Description:      "Scale out deployment to reduce CPU per pod",
				Status:           "pending",
				RiskLevel:        "low",
				RequiresApproval: false,
			},
			RemediationAction{
				ID:               "action-002",
				Type:             "restart",
				Target:           "order-service",
				Description:      "Restart high-CPU pods",
				Status:           "pending",
				RiskLevel:        "medium",
				RequiresApproval: true,
			},
		)

	case "HighMemoryUsage":
		actions = append(actions,
			RemediationAction{
				ID:               "action-003",
				Type:             "restart",
				Target:           "payment-service",
				Description:      "Restart pods to release memory",
				Status:           "pending",
				RiskLevel:        "medium",
				RequiresApproval: true,
			},
			RemediationAction{
				ID:               "action-004",
				Type:             "adjust_resources",
				Target:           "payment-service",
				Description:      "Increase memory limit",
				Status:           "pending",
				RiskLevel:        "low",
				RequiresApproval: false,
			},
		)

	case "DatabaseConnectionFailed":
		actions = append(actions,
			RemediationAction{
				ID:               "action-005",
				Type:             "restart",
				Target:           "database-proxy",
				Description:      "Restart database proxy",
				Status:           "pending",
				RiskLevel:        "high",
				RequiresApproval: true,
			},
		)

	default:
		actions = append(actions,
			RemediationAction{
				ID:               "action-default",
				Type:             "investigate",
				Target:           "unknown",
				Description:      "Manual investigation required",
				Status:           "pending",
				RiskLevel:        "unknown",
				RequiresApproval: true,
			},
		)
	}

	return actions
}

func (a *AutoRemediationService) ExecutePlan(ctx context.Context, planID string) (*RemediationPlan, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	plan, exists := a.plans[planID]
	if !exists {
		return nil, fmt.Errorf("plan not found: %s", planID)
	}

	plan.Status = "executing"
	plan.ExecutedAt = time.Now()

	successCount := 0
	for i := range plan.Actions {
		action := &plan.Actions[i]
		action.StartTime = time.Now()
		action.Status = "executing"

		err := a.executeAction(ctx, action)
		if err != nil {
			action.Status = "failed"
			action.Error = err.Error()
		} else {
			action.Status = "completed"
			action.EndTime = time.Now()
			action.Result = map[string]interface{}{"success": true}
			successCount++
		}
	}

	plan.SuccessRate = float64(successCount) / float64(len(plan.Actions))
	plan.Status = "completed"
	plan.CompletedAt = time.Now()

	a.plans[planID] = plan

	logger.Info(fmt.Sprintf("Executed plan %s: success rate %.2f%%", planID, plan.SuccessRate*100))

	return &plan, nil
}

func (a *AutoRemediationService) executeAction(ctx context.Context, action *RemediationAction) error {
	logger.Info(fmt.Sprintf("Executing action %s: %s on %s", action.ID, action.Type, action.Target))

	time.Sleep(100 * time.Millisecond)

	return nil
}

func (a *AutoRemediationService) GetPlan(ctx context.Context, planID string) (*RemediationPlan, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	plan, exists := a.plans[planID]
	if !exists {
		return nil, fmt.Errorf("plan not found: %s", planID)
	}

	return &plan, nil
}

func (a *AutoRemediationService) ListPlans(ctx context.Context, limit int) ([]RemediationPlan, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var plans []RemediationPlan
	for _, plan := range a.plans {
		plans = append(plans, plan)
	}

	if limit > 0 && len(plans) > limit {
		plans = plans[:limit]
	}

	return plans, nil
}

func (a *AutoRemediationService) CancelPlan(ctx context.Context, planID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	plan, exists := a.plans[planID]
	if !exists {
		return fmt.Errorf("plan not found: %s", planID)
	}

	plan.Status = "cancelled"
	a.plans[planID] = plan

	logger.Info(fmt.Sprintf("Cancelled plan: %s", planID))

	return nil
}

func (a *AutoRemediationService) ApproveAction(ctx context.Context, actionID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	action, exists := a.actions[actionID]
	if !exists {
		return fmt.Errorf("action not found: %s", actionID)
	}

	action.Status = "approved"
	a.actions[actionID] = action

	logger.Info(fmt.Sprintf("Approved action: %s", actionID))

	return nil
}

func (a *AutoRemediationService) GetAction(ctx context.Context, actionID string) (*RemediationAction, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	action, exists := a.actions[actionID]
	if !exists {
		return nil, fmt.Errorf("action not found: %s", actionID)
	}

	return &action, nil
}

func (a *AutoRemediationService) GetStatistics(ctx context.Context) (map[string]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	totalPlans := len(a.plans)
	completedPlans := 0
	failedPlans := 0

	var totalSuccessRate float64
	for _, plan := range a.plans {
		if plan.Status == "completed" {
			completedPlans++
			totalSuccessRate += plan.SuccessRate
		}
		if plan.Status == "failed" {
			failedPlans++
		}
	}

	avgSuccessRate := 0.0
	if completedPlans > 0 {
		avgSuccessRate = totalSuccessRate / float64(completedPlans)
	}

	stats := map[string]interface{}{
		"total_plans":      totalPlans,
		"completed_plans":  completedPlans,
		"failed_plans":     failedPlans,
		"avg_success_rate": avgSuccessRate,
		"pending_plans":    totalPlans - completedPlans - failedPlans,
	}

	return stats, nil
}

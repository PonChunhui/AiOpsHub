package agent

import (
	"fmt"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type DecisionEngine struct {
	TaskAgentMapping map[string]string
	PriorityRules    map[string]int
}

func NewDecisionEngine() *DecisionEngine {
	engine := &DecisionEngine{
		TaskAgentMapping: map[string]string{
			"monitor_collect":     "monitor-agent-001",
			"analysis_diagnosis":  "analysis-agent-001",
			"alert_process":       "alert-agent-001",
			"decision_execute":    "decision-agent-001",
			"learning_optimize":   "learning-agent-001",
			"interaction_service": "interaction-agent-001",
		},
		PriorityRules: map[string]int{
			"monitor_collect":     1,
			"analysis_diagnosis":  2,
			"alert_process":       1,
			"decision_execute":    3,
			"learning_optimize":   4,
			"interaction_service": 5,
		},
	}

	logger.Info("Created Decision Engine")
	return engine
}

func (d *DecisionEngine) DetermineOrchestrationStrategy(subTasks []SubTask) OrchestrationPlan {
	taskCount := len(subTasks)

	if taskCount == 0 {
		return OrchestrationPlan{
			Strategy:      "sequential",
			TaskSequence:  []string{},
			ParallelTasks: [][]string{},
			EstimatedTime: 0,
		}
	}

	dependencyGraph := d.BuildDependencyGraph(subTasks)
	taskSequence := d.GetSequentialOrder(subTasks, dependencyGraph)
	parallelTasks := d.GetParallelGroups(subTasks, dependencyGraph)

	strategy := "sequential"
	if len(parallelTasks) > 1 {
		strategy = "parallel"
	}
	if len(parallelTasks) > 1 && len(taskSequence) > len(parallelTasks) {
		strategy = "hybrid"
	}

	estimatedTime := d.EstimateExecutionTime(subTasks, strategy)

	plan := OrchestrationPlan{
		Strategy:      strategy,
		TaskSequence:  taskSequence,
		ParallelTasks: parallelTasks,
		EstimatedTime: estimatedTime,
	}

	logger.Info(fmt.Sprintf("Orchestration strategy determined: %s (tasks: %d, parallel groups: %d, estimated time: %d seconds)",
		strategy, taskCount, len(parallelTasks), estimatedTime))

	return plan
}

func (d *DecisionEngine) BuildDependencyGraph(subTasks []SubTask) map[string][]string {
	graph := make(map[string][]string)

	for _, task := range subTasks {
		graph[task.TaskID] = task.Dependencies
	}

	logger.Info(fmt.Sprintf("Dependency graph built with %d nodes", len(graph)))
	return graph
}

func (d *DecisionEngine) GetSequentialOrder(subTasks []SubTask, dependencyGraph map[string][]string) []string {
	visited := make(map[string]bool)
	sequence := []string{}

	for _, task := range subTasks {
		sequence = d.dfsTraversal(task.TaskID, dependencyGraph, visited, sequence)
	}

	logger.Info(fmt.Sprintf("Sequential order: %v", sequence))
	return sequence
}

func (d *DecisionEngine) dfsTraversal(taskID string, graph map[string][]string, visited map[string]bool, sequence []string) []string {
	if visited[taskID] {
		return sequence
	}

	visited[taskID] = true

	for _, dep := range graph[taskID] {
		sequence = d.dfsTraversal(dep, graph, visited, sequence)
	}

	sequence = append(sequence, taskID)
	return sequence
}

func (d *DecisionEngine) GetParallelGroups(subTasks []SubTask, dependencyGraph map[string][]string) [][]string {
	taskDepth := d.calculateTaskDepth(subTasks, dependencyGraph)

	groupedByDepth := make(map[int][]string)
	for _, task := range subTasks {
		depth := taskDepth[task.TaskID]
		groupedByDepth[depth] = append(groupedByDepth[depth], task.TaskID)
	}

	parallelGroups := [][]string{}
	for i := 0; i <= maxDepth(taskDepth); i++ {
		if groupedByDepth[i] != nil {
			parallelGroups = append(parallelGroups, groupedByDepth[i])
		}
	}

	logger.Info(fmt.Sprintf("Parallel groups: %v", parallelGroups))
	return parallelGroups
}

func (d *DecisionEngine) calculateTaskDepth(subTasks []SubTask, graph map[string][]string) map[string]int {
	depth := make(map[string]int)

	for _, task := range subTasks {
		depth[task.TaskID] = d.getTaskDepth(task.TaskID, graph, depth)
	}

	return depth
}

func (d *DecisionEngine) getTaskDepth(taskID string, graph map[string][]string, depth map[string]int) int {
	if depth[taskID] != 0 {
		return depth[taskID]
	}

	if len(graph[taskID]) == 0 {
		depth[taskID] = 0
		return 0
	}

	maxDepDepth := 0
	for _, dep := range graph[taskID] {
		depDepth := d.getTaskDepth(dep, graph, depth)
		if depDepth > maxDepDepth {
			maxDepDepth = depDepth
		}
	}

	depth[taskID] = maxDepDepth + 1
	return depth[taskID]
}

func maxDepth(depth map[string]int) int {
	max := 0
	for _, d := range depth {
		if d > max {
			max = d
		}
	}
	return max
}

func (d *DecisionEngine) EstimateExecutionTime(subTasks []SubTask, strategy string) int {
	baseTimePerTask := 5 // seconds
	taskCount := len(subTasks)

	switch strategy {
	case "sequential":
		return taskCount * baseTimePerTask
	case "parallel":
		return baseTimePerTask * 2
	case "hybrid":
		return (taskCount/2 + 1) * baseTimePerTask
	default:
		return taskCount * baseTimePerTask
	}
}

func (d *DecisionEngine) RequiresHumanApproval(taskType string, subTasks []SubTask) bool {
	riskyTaskTypes := []string{
		"auto_remediation",
		"decision_execute",
		"config_change",
		"system_restart",
	}

	for _, risky := range riskyTaskTypes {
		if taskType == risky {
			logger.Info(fmt.Sprintf("Task %s requires human approval", taskType))
			return true
		}
	}

	for _, task := range subTasks {
		if task.TaskType == "decision_execute" {
			logger.Info("Task contains decision_execute, requires human approval")
			return true
		}
	}

	return false
}

func (d *DecisionEngine) RouteToAgent(taskType string) (string, error) {
	agentID, exists := d.TaskAgentMapping[taskType]
	if !exists {
		logger.Error(fmt.Sprintf("No agent mapping for task type: %s", taskType))
		return "", fmt.Errorf("no agent mapping for task type: %s", taskType)
	}

	logger.Info(fmt.Sprintf("Routing task %s to agent %s", taskType, agentID))
	return agentID, nil
}

func (d *DecisionEngine) GetTaskPriority(taskType string) int {
	priority, exists := d.PriorityRules[taskType]
	if !exists {
		return 5 // default lowest priority
	}
	return priority
}

func (d *DecisionEngine) SelectAgentByCapability(requiredCapabilities []string, availableAgents []string) string {
	for _, agentID := range availableAgents {
		agentCapabilities := d.GetAgentCapabilities(agentID)

		matched := true
		for _, required := range requiredCapabilities {
			if !contains(agentCapabilities, required) {
				matched = false
				break
			}
		}

		if matched {
			logger.Info(fmt.Sprintf("Selected agent %s for capabilities %v", agentID, requiredCapabilities))
			return agentID
		}
	}

	logger.Info("No agent matches all capabilities, using fallback agent")
	return availableAgents[0]
}

func (d *DecisionEngine) GetAgentCapabilities(agentID string) []string {
	capabilities := map[string][]string{
		"monitor-agent-001":     []string{"monitor", "metrics", "prometheus", "kubernetes"},
		"analysis-agent-001":    []string{"analysis", "diagnosis", "rag", "root_cause"},
		"alert-agent-001":       []string{"alert", "dedup", "aggregation", "dispatch"},
		"decision-agent-001":    []string{"decision", "execution", "remediation", "risk_assessment"},
		"learning-agent-001":    []string{"learning", "optimization", "knowledge", "rag"},
		"interaction-agent-001": []string{"interaction", "conversation", "report", "visualization"},
	}

	return capabilities[agentID]
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

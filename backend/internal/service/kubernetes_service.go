package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type PodInfo struct {
	Name          string            `json:"name"`
	Namespace     string            `json:"namespace"`
	Status        string            `json:"status"`
	Ready         string            `json:"ready"`
	Restarts      int               `json:"restarts"`
	CPURequest    string            `json:"cpu_request"`
	MemoryRequest string            `json:"memory_request"`
	CPULimit      string            `json:"cpu_limit"`
	MemoryLimit   string            `json:"memory_limit"`
	CreateTime    time.Time         `json:"create_time"`
	Labels        map[string]string `json:"labels"`
	Node          string            `json:"node"`
}

type DeploymentInfo struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int               `json:"replicas"`
	ReadyReplicas     int               `json:"ready_replicas"`
	AvailableReplicas int               `json:"available_replicas"`
	UpdatedReplicas   int               `json:"updated_replicas"`
	Strategy          string            `json:"strategy"`
	CreateTime        time.Time         `json:"create_time"`
	Labels            map[string]string `json:"labels"`
}

type ServiceInfo struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Type       string            `json:"type"`
	ClusterIP  string            `json:"cluster_ip"`
	Ports      []PortInfo        `json:"ports"`
	Selectors  map[string]string `json:"selectors"`
	CreateTime time.Time         `json:"create_time"`
}

type PortInfo struct {
	Name       string `json:"name"`
	Port       int    `json:"port"`
	TargetPort string `json:"target_port"`
	Protocol   string `json:"protocol"`
}

type KubernetesService struct {
	Enabled     bool
	pods        map[string]PodInfo
	deployments map[string]DeploymentInfo
	services    map[string]ServiceInfo
	mu          sync.RWMutex
}

func NewKubernetesService(enabled bool) *KubernetesService {
	svc := &KubernetesService{
		Enabled:     enabled,
		pods:        make(map[string]PodInfo),
		deployments: make(map[string]DeploymentInfo),
		services:    make(map[string]ServiceInfo),
	}

	if enabled {
		svc.initMockData()
	}

	logger.Info(fmt.Sprintf("Kubernetes Service created (enabled=%v)", enabled))
	return svc
}

func (k *KubernetesService) initMockData() {
	k.pods["order-service-pod-1"] = PodInfo{
		Name:          "order-service-pod-1",
		Namespace:     "production",
		Status:        "Running",
		Ready:         "1/1",
		Restarts:      0,
		CPURequest:    "100m",
		MemoryRequest: "256Mi",
		CPULimit:      "500m",
		MemoryLimit:   "512Mi",
		CreateTime:    time.Now().Add(-24 * time.Hour),
		Labels:        map[string]string{"app": "order-service", "version": "v1.0"},
		Node:          "node-1",
	}

	k.deployments["order-service"] = DeploymentInfo{
		Name:              "order-service",
		Namespace:         "production",
		Replicas:          3,
		ReadyReplicas:     3,
		AvailableReplicas: 3,
		UpdatedReplicas:   3,
		Strategy:          "RollingUpdate",
		CreateTime:        time.Now().Add(-7 * 24 * time.Hour),
		Labels:            map[string]string{"app": "order-service"},
	}

	k.services["order-service-svc"] = ServiceInfo{
		Name:       "order-service-svc",
		Namespace:  "production",
		Type:       "ClusterIP",
		ClusterIP:  "10.96.100.1",
		Ports:      []PortInfo{{Name: "http", Port: 8080, TargetPort: "8080", Protocol: "TCP"}},
		Selectors:  map[string]string{"app": "order-service"},
		CreateTime: time.Now().Add(-7 * 24 * time.Hour),
	}
}

func (k *KubernetesService) ListPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var pods []PodInfo
	for _, pod := range k.pods {
		if namespace == "" || pod.Namespace == namespace {
			pods = append(pods, pod)
		}
	}

	logger.Info(fmt.Sprintf("Listed %d pods (namespace=%s)", len(pods), namespace))
	return pods, nil
}

func (k *KubernetesService) GetPod(ctx context.Context, name, namespace string) (*PodInfo, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key := fmt.Sprintf("%s/%s", namespace, name)
	pod, exists := k.pods[key]
	if !exists {
		for _, p := range k.pods {
			if p.Name == name && p.Namespace == namespace {
				return &p, nil
			}
		}
		return nil, fmt.Errorf("pod not found: %s/%s", namespace, name)
	}

	return &pod, nil
}

func (k *KubernetesService) GetPodLogs(ctx context.Context, name, namespace string, tailLines int) (string, error) {
	logger.Info(fmt.Sprintf("Getting logs for pod %s/%s (tail=%d)", namespace, name, tailLines))

	logs := fmt.Sprintf("[%s] Pod %s logs:\n", time.Now().Format(time.RFC3339), name)
	logs += "[INFO] Service started successfully\n"
	logs += "[INFO] Listening on port 8080\n"
	logs += "[WARN] High CPU usage detected (85%%)\n"
	logs += "[ERROR] Database connection timeout\n"
	logs += "[INFO] Retrying connection...\n"
	logs += "[INFO] Connection restored\n"

	return logs, nil
}

func (k *KubernetesService) ListDeployments(ctx context.Context, namespace string) ([]DeploymentInfo, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var deployments []DeploymentInfo
	for _, dep := range k.deployments {
		if namespace == "" || dep.Namespace == namespace {
			deployments = append(deployments, dep)
		}
	}

	return deployments, nil
}

func (k *KubernetesService) GetDeployment(ctx context.Context, name, namespace string) (*DeploymentInfo, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key := fmt.Sprintf("%s/%s", namespace, name)
	dep, exists := k.deployments[key]
	if !exists {
		for _, d := range k.deployments {
			if d.Name == name && d.Namespace == namespace {
				return &d, nil
			}
		}
		return nil, fmt.Errorf("deployment not found: %s/%s", namespace, name)
	}

	return &dep, nil
}

func (k *KubernetesService) ScaleDeployment(ctx context.Context, name, namespace string, replicas int) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	for key, dep := range k.deployments {
		if dep.Name == name && (namespace == "" || dep.Namespace == namespace) {
			dep.Replicas = replicas
			k.deployments[key] = dep
			logger.Info(fmt.Sprintf("Scaled deployment %s to %d replicas", name, replicas))
			return nil
		}
	}

	return fmt.Errorf("deployment not found: %s/%s", namespace, name)
}

func (k *KubernetesService) RestartDeployment(ctx context.Context, name, namespace string) error {
	logger.Info(fmt.Sprintf("Restarting deployment %s/%s", namespace, name))
	return nil
}

func (k *KubernetesService) ListServices(ctx context.Context, namespace string) ([]ServiceInfo, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var services []ServiceInfo
	for _, svc := range k.services {
		if namespace == "" || svc.Namespace == namespace {
			services = append(services, svc)
		}
	}

	return services, nil
}

func (k *KubernetesService) GetResourceUsage(ctx context.Context, namespace string) (map[string]interface{}, error) {
	pods, _ := k.ListPods(ctx, namespace)

	totalCPU := 0
	totalMemory := 0
	runningPods := 0

	for _, pod := range pods {
		if pod.Status == "Running" {
			runningPods++
		}
	}

	usage := map[string]interface{}{
		"namespace":       namespace,
		"total_pods":      len(pods),
		"running_pods":    runningPods,
		"cpu_requests":    totalCPU,
		"memory_requests": totalMemory,
		"timestamp":       time.Now(),
	}

	return usage, nil
}

func (k *KubernetesService) GetPodEvents(ctx context.Context, name, namespace string) ([]map[string]interface{}, error) {
	events := []map[string]interface{}{
		{
			"type":      "Normal",
			"reason":    "Started",
			"message":   "Container started",
			"timestamp": time.Now().Add(-1 * time.Hour),
		},
		{
			"type":      "Warning",
			"reason":    "CPUThrottling",
			"message":   "CPU usage exceeds limit",
			"timestamp": time.Now().Add(-30 * time.Minute),
		},
	}

	return events, nil
}

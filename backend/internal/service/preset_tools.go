package service

import (
	"time"

	"github.com/aiops/AiOpsHub/backend/internal/model"
)

func GetPresetTools() []model.Tool {
	now := time.Now()

	return []model.Tool{
		{
			ID:          "tool-ssh-exec",
			Name:        "ssh_exec",
			Type:        "builtin",
			Category:    "服务器操作",
			Icon:        "💻",
			Description: "在远程服务器执行巡检命令，支持命令白名单和主机范围限制",
			ParametersSchema: `{
				"type": "object",
				"properties": {
					"host": {"type": "string", "description": "服务器IP或主机名"},
					"command": {"type": "string", "description": "要执行的命令"}
				},
				"required": ["host", "command"]
			}`,
			DefaultConfig: `{
				"allowed_commands": ["ls", "top", "free", "df", "ps", "netstat", "cat /var/log/*"],
				"allowed_hosts": ["*"],
				"timeout": 30
			}`,
			Enabled:          true,
			IsPreset:         true,
			RiskLevel:        "medium",
			ExecutionTimeout: 30,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:          "tool-prometheus-query",
			Name:        "prometheus_query",
			Type:        "builtin",
			Category:    "监控查询",
			Icon:        "📊",
			Description: "查询Prometheus监控指标，获取系统性能数据",
			ParametersSchema: `{
				"type": "object",
				"properties": {
					"query": {"type": "string", "description": "PromQL查询语句"},
					"time_range": {"type": "string", "description": "时间范围，如-1h"}
				},
				"required": ["query"]
			}`,
			DefaultConfig: `{
				"url": "http://prometheus:9090",
				"timeout": 10
			}`,
			Enabled:          true,
			IsPreset:         true,
			RiskLevel:        "low",
			ExecutionTimeout: 10,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:          "tool-kubernetes-query",
			Name:        "kubernetes_query",
			Type:        "builtin",
			Category:    "容器管理",
			Icon:        "🚢",
			Description: "查询Kubernetes资源状态，如Pod、Service等",
			ParametersSchema: `{
				"type": "object",
				"properties": {
					"resource_type": {"type": "string", "description": "资源类型，如pods, services"},
					"namespace": {"type": "string", "description": "命名空间"},
					"name": {"type": "string", "description": "资源名称"}
				}
			}`,
			DefaultConfig: `{
				"kubeconfig": "",
				"timeout": 15
			}`,
			Enabled:          true,
			IsPreset:         true,
			RiskLevel:        "low",
			ExecutionTimeout: 15,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:          "tool-log-query",
			Name:        "log_query",
			Type:        "builtin",
			Category:    "日志分析",
			Icon:        "📝",
			Description: "查询系统日志，支持按服务、级别、时间过滤",
			ParametersSchema: `{
				"type": "object",
				"properties": {
					"service": {"type": "string", "description": "服务名称"},
					"level": {"type": "string", "description": "日志级别，如error, warning"},
					"time_range": {"type": "string", "description": "时间范围"}
				}
			}`,
			DefaultConfig: `{
				"datasource": "elasticsearch",
				"timeout": 20
			}`,
			Enabled:          true,
			IsPreset:         true,
			RiskLevel:        "low",
			ExecutionTimeout: 20,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}
}

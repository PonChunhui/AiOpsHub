package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type Tool interface {
	Name() string
	Description() string
	Call(ctx context.Context, input string) (string, error)
}

type PrometheusTool struct {
	URL string
}

func (t *PrometheusTool) Name() string {
	return "prometheus_query"
}

func (t *PrometheusTool) Description() string {
	return `查询Prometheus监控指标。输入JSON格式的查询参数，返回指标数据。

示例输入：
{"query": "cpu_usage{service='order-service'}", "time_range": "-1h"}

返回：指标数据，包括值和时间戳`
}

type PrometheusQueryInput struct {
	Query     string `json:"query"`
	TimeRange string `json:"time_range"`
	Step      string `json:"step"`
}

func NewPrometheusTool(url string) Tool {
	if url == "" {
		url = "http://localhost:9090"
	}
	return &PrometheusTool{URL: url}
}

func (t *PrometheusTool) Call(ctx context.Context, input string) (string, error) {
	logger.Info(fmt.Sprintf("PrometheusTool called with input: %s", input))

	var params PrometheusQueryInput
	err := json.Unmarshal([]byte(input), &params)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse Prometheus input: %v", err))
		return "", fmt.Errorf("invalid input format: %w", err)
	}

	if params.Query == "" {
		return "", fmt.Errorf("query parameter is required")
	}

	result, err := t.queryPrometheus(ctx, params)
	if err != nil {
		logger.Error(fmt.Sprintf("Prometheus query failed: %v", err))
		return "", err
	}

	logger.Info(fmt.Sprintf("PrometheusTool query successful"))
	return result, nil
}

func (t *PrometheusTool) queryPrometheus(ctx context.Context, params PrometheusQueryInput) (string, error) {
	queryURL := fmt.Sprintf("%s/api/v1/query", t.URL)

	urlParams := url.Values{}
	urlParams.Set("query", params.Query)

	if params.TimeRange != "" {
		endTime := time.Now()
		startTime := endTime.Add(-parseDuration(params.TimeRange))
		urlParams.Set("start", fmt.Sprintf("%d", startTime.Unix()))
		urlParams.Set("end", fmt.Sprintf("%d", endTime.Unix()))
		if params.Step == "" {
			params.Step = "60s"
		}
		urlParams.Set("step", params.Step)
	}

	fullURL := fmt.Sprintf("%s?%s", queryURL, urlParams.Encode())

	logger.Info(fmt.Sprintf("Querying Prometheus: %s", fullURL))

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to query Prometheus: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Prometheus returned status %d: %s", resp.StatusCode, string(body))
	}

	var promResult map[string]interface{}
	err = json.Unmarshal(body, &promResult)
	if err != nil {
		return "", fmt.Errorf("failed to parse Prometheus response: %w", err)
	}

	return t.formatResult(params.Query, promResult), nil
}

func (t *PrometheusTool) formatResult(query string, result map[string]interface{}) string {
	output := fmt.Sprintf("Prometheus查询结果：\n查询：%s\n\n", query)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return output + "响应格式错误\n"
	}

	resultType, ok := data["resultType"].(string)
	if !ok {
		return output + "结果类型未知\n"
	}

	output += fmt.Sprintf("结果类型：%s\n", resultType)

	results, ok := data["result"].([]interface{})
	if !ok {
		return output + "无结果数据\n"
	}

	output += fmt.Sprintf("结果数量：%d\n\n", len(results))

	for i, r := range results {
		if i >= 5 {
			output += "... (更多结果已省略)\n"
			break
		}

		item, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		metric, ok := item["metric"].(map[string]interface{})
		if ok {
			output += fmt.Sprintf("指标 %d:\n", i+1)
			for k, v := range metric {
				if k != "__name__" {
					output += fmt.Sprintf("  %s: %v\n", k, v)
				}
			}
		}

		value, ok := item["value"].([]interface{})
		if ok && len(value) >= 2 {
			output += fmt.Sprintf("  时间戳: %v\n", value[0])
			output += fmt.Sprintf("  值: %v\n\n", value[1])
		}

		values, ok := item["values"].([]interface{})
		if ok && len(values) > 0 {
			output += fmt.Sprintf("  时间序列 (%d 个点)\n", len(values))
			output += fmt.Sprintf("  最新值: %v\n\n", values[len(values)-1])
		}
	}

	return output
}

func parseDuration(duration string) time.Duration {
	if duration == "" {
		return time.Hour
	}

	d, err := time.ParseDuration(duration)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse duration %s: %v", duration, err))
		return time.Hour
	}

	return d
}

type KubernetesTool struct {
}

func (t *KubernetesTool) Name() string {
	return "kubernetes_query"
}

func (t *KubernetesTool) Description() string {
	return `查询Kubernetes资源信息。输入JSON格式的查询参数，返回K8s资源状态。

示例输入：
{"resource_type": "pods", "namespace": "default", "name": "order-service"}

返回：K8s资源详细信息`
}

type KubernetesQueryInput struct {
	ResourceType  string `json:"resource_type"`
	Namespace     string `json:"namespace"`
	Name          string `json:"name"`
	LabelSelector string `json:"label_selector"`
}

func NewKubernetesTool() Tool {
	return &KubernetesTool{}
}

func (t *KubernetesTool) Call(ctx context.Context, input string) (string, error) {
	logger.Info(fmt.Sprintf("KubernetesTool called with input: %s", input))

	var params KubernetesQueryInput
	err := json.Unmarshal([]byte(input), &params)
	if err != nil {
		return "", fmt.Errorf("invalid input format: %w", err)
	}

	if params.ResourceType == "" {
		params.ResourceType = "pods"
	}

	if params.Namespace == "" {
		params.Namespace = "default"
	}

	result := t.mockKubernetesQuery(params)

	logger.Info(fmt.Sprintf("KubernetesTool query successful"))
	return result, nil
}

func (t *KubernetesTool) mockKubernetesQuery(params KubernetesQueryInput) string {
	output := fmt.Sprintf("Kubernetes查询结果：\n")
	output += fmt.Sprintf("资源类型：%s\n", params.ResourceType)
	output += fmt.Sprintf("命名空间：%s\n", params.Namespace)

	if params.Name != "" {
		output += fmt.Sprintf("名称：%s\n\n", params.Name)
		output += fmt.Sprintf("状态：Running\n")
		output += fmt.Sprintf("Pod IP：10.0.0.1\n")
		output += fmt.Sprintf("节点：node-1\n")
		output += fmt.Sprintf("CPU使用：500m\n")
		output += fmt.Sprintf("内存使用：256Mi\n")
	} else {
		output += fmt.Sprintf("\n找到3个Pod：\n")
		output += fmt.Sprintf("1. order-service-pod-1 (Running)\n")
		output += fmt.Sprintf("2. order-service-pod-2 (Running)\n")
		output += fmt.Sprintf("3. order-service-pod-3 (Pending)\n")
	}

	return output
}

type LogQueryTool struct {
}

func (t *LogQueryTool) Name() string {
	return "log_query"
}

func (t *LogQueryTool) Description() string {
	return `查询系统日志。输入JSON格式的查询参数，返回日志内容。

示例输入：
{"service": "order-service", "level": "error", "time_range": "-1h"}

返回：匹配的日志内容`
}

type LogQueryInput struct {
	Service    string `json:"service"`
	Level      string `json:"level"`
	TimeRange  string `json:"time_range"`
	SearchTerm string `json:"search_term"`
	Limit      int    `json:"limit"`
}

func NewLogQueryTool() Tool {
	return &LogQueryTool{}
}

func (t *LogQueryTool) Call(ctx context.Context, input string) (string, error) {
	logger.Info(fmt.Sprintf("LogQueryTool called with input: %s", input))

	var params LogQueryInput
	err := json.Unmarshal([]byte(input), &params)
	if err != nil {
		return "", fmt.Errorf("invalid input format: %w", err)
	}

	if params.Limit == 0 {
		params.Limit = 10
	}

	result := t.mockLogQuery(params)

	logger.Info(fmt.Sprintf("LogQueryTool query successful"))
	return result, nil
}

func (t *LogQueryTool) mockLogQuery(params LogQueryInput) string {
	output := fmt.Sprintf("日志查询结果：\n")
	output += fmt.Sprintf("服务：%s\n", params.Service)
	if params.Level != "" {
		output += fmt.Sprintf("级别：%s\n", params.Level)
	}
	output += fmt.Sprintf("时间范围：%s\n\n", params.TimeRange)

	output += fmt.Sprintf("找到%d条日志：\n\n", params.Limit)

	for i := 1; i <= params.Limit && i <= 5; i++ {
		output += fmt.Sprintf("[%d] %s ERROR Database connection timeout\n", i, time.Now().Format(time.RFC3339))
		output += fmt.Sprintf("    服务：%s\n", params.Service)
		output += fmt.Sprintf("    错误：连接池耗尽\n\n")
	}

	return output
}

func GetDefaultTools() []Tool {
	return []Tool{
		NewPrometheusTool("http://localhost:9090"),
		NewKubernetesTool(),
		NewLogQueryTool(),
	}
}

package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Service   string                 `json:"service"`
	Message   string                 `json:"message"`
	TraceID   string                 `json:"trace_id"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type LogQuery struct {
	Service   string    `json:"service"`
	Level     string    `json:"level"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Keywords  []string  `json:"keywords"`
	Limit     int       `json:"limit"`
}

type LogStatistics struct {
	TotalLogs    int64            `json:"total_logs"`
	ErrorCount   int64            `json:"error_count"`
	WarnCount    int64            `json:"warn_count"`
	InfoCount    int64            `json:"info_count"`
	ServiceCount map[string]int64 `json:"service_count"`
	TopErrors    []string         `json:"top_errors"`
	TimeRange    string           `json:"time_range"`
}

type LogService struct {
	logs  []LogEntry
	stats LogStatistics
	mu    sync.RWMutex
}

func NewLogService() *LogService {
	svc := &LogService{
		logs: []LogEntry{},
		stats: LogStatistics{
			ServiceCount: make(map[string]int64),
			TopErrors:    []string{},
		},
	}

	svc.initMockLogs()

	logger.Info("Log Service created")
	return svc
}

func (l *LogService) initMockLogs() {
	l.logs = []LogEntry{
		{
			Timestamp: time.Now().Add(-2 * time.Hour),
			Level:     "ERROR",
			Service:   "order-service",
			Message:   "Database connection timeout",
			TraceID:   "trace-001",
			Metadata:  map[string]interface{}{"error_code": "DB_TIMEOUT"},
		},
		{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Level:     "WARN",
			Service:   "order-service",
			Message:   "High CPU usage detected",
			TraceID:   "trace-002",
			Metadata:  map[string]interface{}{"cpu_percent": 85},
		},
		{
			Timestamp: time.Now().Add(-30 * time.Minute),
			Level:     "INFO",
			Service:   "payment-service",
			Message:   "Transaction completed successfully",
			TraceID:   "trace-003",
			Metadata:  map[string]interface{}{"transaction_id": "txn-123"},
		},
		{
			Timestamp: time.Now().Add(-15 * time.Minute),
			Level:     "ERROR",
			Service:   "payment-service",
			Message:   "Payment gateway timeout",
			TraceID:   "trace-004",
			Metadata:  map[string]interface{}{"gateway": "stripe"},
		},
		{
			Timestamp: time.Now().Add(-5 * time.Minute),
			Level:     "INFO",
			Service:   "user-service",
			Message:   "User authentication successful",
			TraceID:   "trace-005",
			Metadata:  map[string]interface{}{"user_id": "user-001"},
		},
	}
}

func (l *LogService) QueryLogs(ctx context.Context, query LogQuery) ([]LogEntry, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var results []LogEntry

	for _, log := range l.logs {
		if query.Service != "" && log.Service != query.Service {
			continue
		}

		if query.Level != "" && log.Level != query.Level {
			continue
		}

		if !query.StartTime.IsZero() && log.Timestamp.Before(query.StartTime) {
			continue
		}

		if !query.EndTime.IsZero() && log.Timestamp.After(query.EndTime) {
			continue
		}

		if len(query.Keywords) > 0 {
			matched := false
			for _, keyword := range query.Keywords {
				if containsKeyword(log.Message, keyword) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		results = append(results, log)

		if query.Limit > 0 && len(results) >= query.Limit {
			break
		}
	}

	logger.Info(fmt.Sprintf("Query returned %d logs", len(results)))
	return results, nil
}

func containsKeyword(message, keyword string) bool {
	return len(message) > 0 && len(keyword) > 0
}

func (l *LogService) GetLogStatistics(ctx context.Context) (*LogStatistics, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	stats := LogStatistics{
		TotalLogs:    int64(len(l.logs)),
		ServiceCount: make(map[string]int64),
		TopErrors:    []string{},
	}

	for _, log := range l.logs {
		switch log.Level {
		case "ERROR":
			stats.ErrorCount++
			stats.TopErrors = append(stats.TopErrors, log.Message)
		case "WARN":
			stats.WarnCount++
		case "INFO":
			stats.InfoCount++
		}

		stats.ServiceCount[log.Service]++
	}

	stats.TimeRange = "Last 2 hours"

	return &stats, nil
}

func (l *LogService) GetServiceLogs(ctx context.Context, service string, limit int) ([]LogEntry, error) {
	query := LogQuery{
		Service: service,
		Limit:   limit,
	}

	return l.QueryLogs(ctx, query)
}

func (l *LogService) GetErrorLogs(ctx context.Context, limit int) ([]LogEntry, error) {
	query := LogQuery{
		Level: "ERROR",
		Limit: limit,
	}

	return l.QueryLogs(ctx, query)
}

func (l *LogService) SearchLogs(ctx context.Context, keywords []string, limit int) ([]LogEntry, error) {
	query := LogQuery{
		Keywords: keywords,
		Limit:    limit,
	}

	return l.QueryLogs(ctx, query)
}

func (l *LogService) GetLogByTraceID(ctx context.Context, traceID string) ([]LogEntry, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var results []LogEntry
	for _, log := range l.logs {
		if log.TraceID == traceID {
			results = append(results, log)
		}
	}

	return results, nil
}

func (l *LogService) GetRecentLogs(ctx context.Context, minutes int) ([]LogEntry, error) {
	query := LogQuery{
		StartTime: time.Now().Add(-time.Duration(minutes) * time.Minute),
		EndTime:   time.Now(),
		Limit:     100,
	}

	return l.QueryLogs(ctx, query)
}

func (l *LogService) ExportLogs(ctx context.Context, format string) (string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	output := ""
	for _, log := range l.logs {
		switch format {
		case "json":
			output += fmt.Sprintf("{\"timestamp\":\"%s\",\"level\":\"%s\",\"service\":\"%s\",\"message\":\"%s\"}\n",
				log.Timestamp.Format(time.RFC3339), log.Level, log.Service, log.Message)
		case "text":
			output += fmt.Sprintf("[%s] %s %s: %s\n",
				log.Timestamp.Format(time.RFC3339), log.Level, log.Service, log.Message)
		default:
			output += fmt.Sprintf("%s|%s|%s|%s\n",
				log.Timestamp.Format(time.RFC3339), log.Level, log.Service, log.Message)
		}
	}

	logger.Info(fmt.Sprintf("Exported %d logs in format %s", len(l.logs), format))
	return output, nil
}

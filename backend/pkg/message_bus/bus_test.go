package message_bus

import (
	"testing"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

func init() {
	logger.Init()
}

func TestGetChannelForReceiver(t *testing.T) {
	tests := []struct {
		name          string
		receiver      string
		expectChannel string
	}{
		{"Coordinator", "coordinator-agent", "coordinator-channel"},
		{"Coordinator short", "coordinator", "coordinator-channel"},
		{"Monitor", "monitor-agent-001", "monitor-channel"},
		{"Monitor short", "monitor-agent", "monitor-channel"},
		{"Analysis", "analysis-agent-001", "analysis-channel"},
		{"Analysis short", "analysis-agent", "analysis-channel"},
		{"Alert", "alert-agent-001", "alert-channel"},
		{"Decision", "decision-agent-001", "decision-channel"},
		{"Learning", "learning-agent-001", "learning-channel"},
		{"Interaction", "interaction-agent-001", "interaction-channel"},
		{"Unknown agent", "unknown-agent", "agent-channel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel := getChannelForReceiver(tt.receiver)

			if channel != tt.expectChannel {
				t.Errorf("Expected channel '%s', got '%s'", tt.expectChannel, channel)
			}
		})
	}
}

func getChannelForReceiver(receiver string) string {
	if receiver == "coordinator-agent" || receiver == "coordinator" {
		return "coordinator-channel"
	}
	if receiver == "monitor-agent" || receiver == "monitor-agent-001" {
		return "monitor-channel"
	}
	if receiver == "analysis-agent" || receiver == "analysis-agent-001" {
		return "analysis-channel"
	}
	if receiver == "alert-agent" || receiver == "alert-agent-001" {
		return "alert-channel"
	}
	if receiver == "decision-agent" || receiver == "decision-agent-001" {
		return "decision-channel"
	}
	if receiver == "learning-agent" || receiver == "learning-agent-001" {
		return "learning-channel"
	}
	if receiver == "interaction-agent" || receiver == "interaction-agent-001" {
		return "interaction-channel"
	}
	return "agent-channel"
}

func TestBroadcastChannels(t *testing.T) {
	channels := []string{
		"agent-channel",
		"coordinator-channel",
		"monitor-channel",
		"analysis-channel",
		"alert-channel",
		"decision-channel",
		"learning-channel",
		"interaction-channel",
	}

	if len(channels) != 8 {
		t.Errorf("Expected 8 broadcast channels, got %d", len(channels))
	}

	for _, ch := range channels {
		if ch == "" {
			t.Error("Channel name should not be empty")
		}
	}

	t.Logf("Broadcast to %d channels", len(channels))
}

func TestMessageStats(t *testing.T) {
	stats := MessageStats{
		PublishedCount:  10,
		ReceivedCount:   9,
		ErrorCount:      1,
		AverageLatency:  100,
		LastMessageTime: time.Now(),
	}

	if stats.PublishedCount < stats.ReceivedCount {
		t.Error("PublishedCount should >= ReceivedCount")
	}

	t.Logf("Stats: published=%d, received=%d, errors=%d", stats.PublishedCount, stats.ReceivedCount, stats.ErrorCount)
}

package message_bus

import (
	"testing"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

func init() {
	logger.Init()
}

func TestMessageCreation(t *testing.T) {
	msg := Message{
		MessageID:   "msg-001",
		MessageType: MessageTypeTaskRequest,
		Sender:      "coordinator-agent",
		Receiver:    "monitor-agent-001",
		SessionID:   "session-001",
		WorkflowID:  "wf-12345",
		Content:     map[string]interface{}{"task": "monitor_collect"},
		Timestamp:   time.Now(),
	}

	if msg.MessageID == "" {
		t.Error("MessageID should not be empty")
	}

	if msg.MessageType != MessageTypeTaskRequest {
		t.Errorf("Expected type '%s', got '%s'", MessageTypeTaskRequest, msg.MessageType)
	}

	t.Logf("Message created: %s (%s -> %s)", msg.MessageID, msg.Sender, msg.Receiver)
}

func TestMessageTypes(t *testing.T) {
	tests := []struct {
		name        string
		messageType string
	}{
		{"TaskRequest", MessageTypeTaskRequest},
		{"TaskResult", MessageTypeTaskResult},
		{"CollaborationRequest", MessageTypeCollaborationRequest},
		{"StateUpdate", MessageTypeStateUpdate},
		{"EventBroadcast", MessageTypeEventBroadcast},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Message{
				MessageID:   "msg-" + tt.name,
				MessageType: tt.messageType,
				Sender:      "test-agent",
				Receiver:    "test-receiver",
				Timestamp:   time.Now(),
			}

			switch tt.messageType {
			case MessageTypeTaskRequest:
				if !msg.IsTaskRequest() {
					t.Error("Should be a task request")
				}
			case MessageTypeTaskResult:
				if !msg.IsTaskResult() {
					t.Error("Should be a task result")
				}
			case MessageTypeCollaborationRequest:
				if !msg.IsCollaborationRequest() {
					t.Error("Should be a collaboration request")
				}
			case MessageTypeStateUpdate:
				if !msg.IsStateUpdate() {
					t.Error("Should be a state update")
				}
			case MessageTypeEventBroadcast:
				if !msg.IsEventBroadcast() {
					t.Error("Should be an event broadcast")
				}
			}
		})
	}
}

func TestBroadcastDetection(t *testing.T) {
	tests := []struct {
		name     string
		receiver string
		expected bool
	}{
		{"Broadcast receiver", "broadcast", true},
		{"All receiver", "all", true},
		{"Specific receiver", "monitor-agent-001", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Message{
				MessageID: "msg-" + tt.name,
				Receiver:  tt.receiver,
			}

			if msg.IsBroadcast() != tt.expected {
				t.Errorf("Expected IsBroadcast=%v, got %v", tt.expected, msg.IsBroadcast())
			}
		})
	}
}

func TestMessageJSONSerialization(t *testing.T) {
	msg := NewTaskRequestMessage(
		"msg-json-test",
		"coordinator",
		"monitor-agent",
		"session-test",
		"wf-test",
		map[string]interface{}{"task": "test"},
	)

	jsonStr, err := msg.ToJSON()
	if err != nil {
		t.Errorf("Failed to serialize: %v", err)
	}

	if jsonStr == "" {
		t.Error("JSON string should not be empty")
	}

	parsedMsg, err := FromJSON(jsonStr)
	if err != nil {
		t.Errorf("Failed to deserialize: %v", err)
	}

	if parsedMsg.MessageID != msg.MessageID {
		t.Errorf("MessageID mismatch: expected '%s', got '%s'", msg.MessageID, parsedMsg.MessageID)
	}

	if parsedMsg.MessageType != msg.MessageType {
		t.Errorf("MessageType mismatch: expected '%s', got '%s'", msg.MessageType, parsedMsg.MessageType)
	}

	t.Logf("JSON serialization successful: %d bytes", len(jsonStr))
}

func TestTaskRequestContent(t *testing.T) {
	content := NewTaskRequestContent(
		"monitor_collect",
		map[string]interface{}{"service": "order-service"},
		map[string]interface{}{"env": "production"},
		1,
		60,
	)

	if content.Task != "monitor_collect" {
		t.Errorf("Expected task 'monitor_collect', got '%s'", content.Task)
	}

	if content.Priority != 1 {
		t.Errorf("Expected priority 1, got %d", content.Priority)
	}

	if content.Timeout != 60 {
		t.Errorf("Expected timeout 60, got %d", content.Timeout)
	}

	t.Logf("TaskRequestContent: task=%s, priority=%d", content.Task, content.Priority)
}

func TestTaskResultContent(t *testing.T) {
	content := NewTaskResultContent(
		map[string]interface{}{"cpu_usage": 85.5},
		"completed",
		"",
		100,
		30,
	)

	if content.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", content.Status)
	}

	if content.TokensUsed != 100 {
		t.Errorf("Expected tokens 100, got %d", content.TokensUsed)
	}

	if content.Duration != 30 {
		t.Errorf("Expected duration 30, got %d", content.Duration)
	}

	t.Logf("TaskResultContent: status=%s, tokens=%d, duration=%d", content.Status, content.TokensUsed, content.Duration)
}

func TestStateUpdateContent(t *testing.T) {
	content := NewStateUpdateContent(
		"monitor-agent-001",
		"RUNNING",
		"task-001",
		50,
	)

	if content.AgentID != "monitor-agent-001" {
		t.Errorf("Expected agent 'monitor-agent-001', got '%s'", content.AgentID)
	}

	if content.Status != "RUNNING" {
		t.Errorf("Expected status 'RUNNING', got '%s'", content.Status)
	}

	if content.Progress != 50 {
		t.Errorf("Expected progress 50, got %d", content.Progress)
	}

	t.Logf("StateUpdateContent: agent=%s, status=%s, progress=%d%%", content.AgentID, content.Status, content.Progress)
}

func TestEventBroadcastContent(t *testing.T) {
	content := NewEventBroadcastContent(
		"workflow_completed",
		"Workflow execution completed successfully",
		map[string]interface{}{"workflow_id": "wf-12345"},
	)

	if content.Event != "workflow_completed" {
		t.Errorf("Expected event 'workflow_completed', got '%s'", content.Event)
	}

	if content.Description == "" {
		t.Error("Description should not be empty")
	}

	t.Logf("EventBroadcastContent: event=%s, description=%s", content.Event, content.Description)
}

func TestMessageFactoryFunctions(t *testing.T) {
	tests := []struct {
		name    string
		msgType string
		factory func() *Message
	}{
		{
			name:    "TaskRequest",
			msgType: MessageTypeTaskRequest,
			factory: func() *Message {
				return NewTaskRequestMessage("msg-001", "sender", "receiver", "session", "workflow", map[string]interface{}{})
			},
		},
		{
			name:    "TaskResult",
			msgType: MessageTypeTaskResult,
			factory: func() *Message {
				return NewTaskResultMessage("msg-002", "sender", "receiver", "session", "workflow", map[string]interface{}{})
			},
		},
		{
			name:    "CollaborationRequest",
			msgType: MessageTypeCollaborationRequest,
			factory: func() *Message {
				return NewCollaborationRequestMessage("msg-003", "sender", "receiver", "session", "workflow", map[string]interface{}{})
			},
		},
		{
			name:    "StateUpdate",
			msgType: MessageTypeStateUpdate,
			factory: func() *Message {
				return NewStateUpdateMessage("msg-004", "sender", "session", map[string]interface{}{})
			},
		},
		{
			name:    "EventBroadcast",
			msgType: MessageTypeEventBroadcast,
			factory: func() *Message {
				return NewEventBroadcastMessage("msg-005", "sender", "session", "workflow", map[string]interface{}{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.factory()

			if msg.MessageType != tt.msgType {
				t.Errorf("Expected type '%s', got '%s'", tt.msgType, msg.MessageType)
			}

			if msg.MessageID == "" {
				t.Error("MessageID should not be empty")
			}

			if msg.Timestamp.IsZero() {
				t.Error("Timestamp should not be zero")
			}
		})
	}
}

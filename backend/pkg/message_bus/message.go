package message_bus

import (
	"encoding/json"
	"time"
)

type Message struct {
	MessageID   string                 `json:"message_id"`
	MessageType string                 `json:"message_type"`
	Sender      string                 `json:"sender"`
	Receiver    string                 `json:"receiver"`
	SessionID   string                 `json:"session_id"`
	WorkflowID  string                 `json:"workflow_id"`
	Content     map[string]interface{} `json:"content"`
	Timestamp   time.Time              `json:"timestamp"`
}

const (
	MessageTypeTaskRequest          = "task_request"
	MessageTypeTaskResult           = "task_result"
	MessageTypeCollaborationRequest = "collaboration_request"
	MessageTypeStateUpdate          = "state_update"
	MessageTypeEventBroadcast       = "event_broadcast"
)

func NewTaskRequestMessage(messageID, sender, receiver, sessionID, workflowID string, content map[string]interface{}) *Message {
	return &Message{
		MessageID:   messageID,
		MessageType: MessageTypeTaskRequest,
		Sender:      sender,
		Receiver:    receiver,
		SessionID:   sessionID,
		WorkflowID:  workflowID,
		Content:     content,
		Timestamp:   time.Now(),
	}
}

func NewTaskResultMessage(messageID, sender, receiver, sessionID, workflowID string, content map[string]interface{}) *Message {
	return &Message{
		MessageID:   messageID,
		MessageType: MessageTypeTaskResult,
		Sender:      sender,
		Receiver:    receiver,
		SessionID:   sessionID,
		WorkflowID:  workflowID,
		Content:     content,
		Timestamp:   time.Now(),
	}
}

func NewCollaborationRequestMessage(messageID, sender, receiver, sessionID, workflowID string, content map[string]interface{}) *Message {
	return &Message{
		MessageID:   messageID,
		MessageType: MessageTypeCollaborationRequest,
		Sender:      sender,
		Receiver:    receiver,
		SessionID:   sessionID,
		WorkflowID:  workflowID,
		Content:     content,
		Timestamp:   time.Now(),
	}
}

func NewStateUpdateMessage(messageID, sender, sessionID string, content map[string]interface{}) *Message {
	return &Message{
		MessageID:   messageID,
		MessageType: MessageTypeStateUpdate,
		Sender:      sender,
		Receiver:    "broadcast",
		SessionID:   sessionID,
		WorkflowID:  "",
		Content:     content,
		Timestamp:   time.Now(),
	}
}

func NewEventBroadcastMessage(messageID, sender, sessionID, workflowID string, content map[string]interface{}) *Message {
	return &Message{
		MessageID:   messageID,
		MessageType: MessageTypeEventBroadcast,
		Sender:      sender,
		Receiver:    "all",
		SessionID:   sessionID,
		WorkflowID:  workflowID,
		Content:     content,
		Timestamp:   time.Now(),
	}
}

func (m *Message) ToJSON() (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func FromJSON(data string) (*Message, error) {
	var message Message
	err := json.Unmarshal([]byte(data), &message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (m *Message) IsTaskRequest() bool {
	return m.MessageType == MessageTypeTaskRequest
}

func (m *Message) IsTaskResult() bool {
	return m.MessageType == MessageTypeTaskResult
}

func (m *Message) IsCollaborationRequest() bool {
	return m.MessageType == MessageTypeCollaborationRequest
}

func (m *Message) IsStateUpdate() bool {
	return m.MessageType == MessageTypeStateUpdate
}

func (m *Message) IsEventBroadcast() bool {
	return m.MessageType == MessageTypeEventBroadcast
}

func (m *Message) IsBroadcast() bool {
	return m.Receiver == "broadcast" || m.Receiver == "all"
}

type TaskRequestContent struct {
	Task       string                 `json:"task"`
	Parameters map[string]interface{} `json:"parameters"`
	Context    map[string]interface{} `json:"context"`
	Priority   int                    `json:"priority"`
	Timeout    int                    `json:"timeout"` // seconds
}

type TaskResultContent struct {
	Result     map[string]interface{} `json:"result"`
	Status     string                 `json:"status"`
	Error      string                 `json:"error"`
	TokensUsed int                    `json:"tokens_used"`
	Duration   int                    `json:"duration"` // seconds
}

type CollaborationRequestContent struct {
	Request    string                 `json:"request"`
	Parameters map[string]interface{} `json:"parameters"`
	Priority   int                    `json:"priority"`
	Timeout    int                    `json:"timeout"`
}

type StateUpdateContent struct {
	AgentID     string `json:"agent_id"`
	Status      string `json:"status"`
	Progress    int    `json:"progress"`
	CurrentTask string `json:"current_task"`
}

type EventBroadcastContent struct {
	Event       string                 `json:"event"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
}

func NewTaskRequestContent(task string, parameters, context map[string]interface{}, priority, timeout int) *TaskRequestContent {
	return &TaskRequestContent{
		Task:       task,
		Parameters: parameters,
		Context:    context,
		Priority:   priority,
		Timeout:    timeout,
	}
}

func NewTaskResultContent(result map[string]interface{}, status, error string, tokensUsed, duration int) *TaskResultContent {
	return &TaskResultContent{
		Result:     result,
		Status:     status,
		Error:      error,
		TokensUsed: tokensUsed,
		Duration:   duration,
	}
}

func NewCollaborationRequestContent(request string, parameters map[string]interface{}, priority, timeout int) *CollaborationRequestContent {
	return &CollaborationRequestContent{
		Request:    request,
		Parameters: parameters,
		Priority:   priority,
		Timeout:    timeout,
	}
}

func NewStateUpdateContent(agentID, status, currentTask string, progress int) *StateUpdateContent {
	return &StateUpdateContent{
		AgentID:     agentID,
		Status:      status,
		Progress:    progress,
		CurrentTask: currentTask,
	}
}

func NewEventBroadcastContent(event, description string, data map[string]interface{}) *EventBroadcastContent {
	return &EventBroadcastContent{
		Event:       event,
		Description: description,
		Data:        data,
	}
}

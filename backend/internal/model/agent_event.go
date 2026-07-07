package model

import "time"

type AgentEventType string

const (
	EventThinking      AgentEventType = "thinking"
	EventToolCall      AgentEventType = "tool_call"
	EventToolResult    AgentEventType = "tool_result"
	EventContentChunk  AgentEventType = "content_chunk"
	EventAgentTransfer AgentEventType = "agent_transfer"
	EventError         AgentEventType = "error"
	EventDone          AgentEventType = "done"
	EventRagReferences AgentEventType = "rag_references"
	EventUserMessage   AgentEventType = "user_message"
	EventAIMessage     AgentEventType = "ai_message"
)

type AgentEvent struct {
	Type      AgentEventType `json:"type"`
	AgentName string         `json:"agent_name"`
	RunPath   []AgentRunStep `json:"run_path"`
	Data      interface{}    `json:"data"`
	Timestamp int64          `json:"timestamp"`
}

type AgentRunStep struct {
	AgentID   string `json:"agent_id"`
	AgentName string `json:"agent_name"`
	Action    string `json:"action"`
}

type ThinkingEventData struct {
	Content string `json:"content"`
}

type ToolCallEventData struct {
	ToolID       string `json:"tool_id"`
	ToolName     string `json:"tool_name"`
	ArgsRaw      string `json:"args_raw"`      // 流式输出的原始参数分片（不解析）
	ArgsComplete string `json:"args_complete"` // 完整的参数JSON（流完成后）
}

type ToolResultEventData struct {
	ToolID   string `json:"tool_id"`
	ToolName string `json:"tool_name"`
	Result   string `json:"result"`
	Success  bool   `json:"success"`
}

type ContentChunkEventData struct {
	Content string `json:"content"`
}

type AgentTransferEventData struct {
	FromAgent string `json:"from_agent"`
	ToAgent   string `json:"to_agent"`
	Reason    string `json:"reason"`
}

type ErrorEventData struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type RagReferencesEventData struct {
	References []map[string]interface{} `json:"references"`
}

type UserMessageEventData struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AIMessageEventData struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewAgentEvent(eventType AgentEventType, agentName string, data interface{}) *AgentEvent {
	return &AgentEvent{
		Type:      eventType,
		AgentName: agentName,
		Data:      data,
		Timestamp: time.Now().Unix(),
		RunPath:   []AgentRunStep{},
	}
}

func NewAgentEventWithPath(eventType AgentEventType, agentName string, data interface{}, runPath []AgentRunStep) *AgentEvent {
	return &AgentEvent{
		Type:      eventType,
		AgentName: agentName,
		Data:      data,
		Timestamp: time.Now().Unix(),
		RunPath:   runPath,
	}
}

func NewThinkingEvent(agentName, content string) *AgentEvent {
	return NewAgentEvent(EventThinking, agentName, ThinkingEventData{Content: content})
}

func NewToolCallEvent(agentName, toolID, toolName string, argsRaw string) *AgentEvent {
	return NewAgentEvent(EventToolCall, agentName, ToolCallEventData{
		ToolID:   toolID,
		ToolName: toolName,
		ArgsRaw:  argsRaw,
	})
}

// NewToolCallCompleteEvent 用于发送完整的工具调用参数（流完成后）
func NewToolCallCompleteEvent(agentName, toolID, toolName string, argsComplete string) *AgentEvent {
	return NewAgentEvent(EventToolCall, agentName, ToolCallEventData{
		ToolID:       toolID,
		ToolName:     toolName,
		ArgsComplete: argsComplete,
	})
}

func NewToolResultEvent(agentName, toolID, toolName, result string, success bool) *AgentEvent {
	return NewAgentEvent(EventToolResult, agentName, ToolResultEventData{
		ToolID:   toolID,
		ToolName: toolName,
		Result:   result,
		Success:  success,
	})
}

func NewContentChunkEvent(agentName, content string) *AgentEvent {
	return NewAgentEvent(EventContentChunk, agentName, ContentChunkEventData{Content: content})
}

func NewAgentTransferEvent(fromAgent, toAgent, reason string) *AgentEvent {
	return NewAgentEvent(EventAgentTransfer, toAgent, AgentTransferEventData{
		FromAgent: fromAgent,
		ToAgent:   toAgent,
		Reason:    reason,
	})
}

func NewErrorEvent(agentName, message string, code int) *AgentEvent {
	return NewAgentEvent(EventError, agentName, ErrorEventData{
		Message: message,
		Code:    code,
	})
}

func NewDoneEvent(agentName string, runPath []AgentRunStep) *AgentEvent {
	return NewAgentEventWithPath(EventDone, agentName, nil, runPath)
}

func NewRagReferencesEvent(references []map[string]interface{}) *AgentEvent {
	return NewAgentEvent(EventRagReferences, "", RagReferencesEventData{References: references})
}

func NewUserMessageEvent(id, role, content string) *AgentEvent {
	return NewAgentEvent(EventUserMessage, "", UserMessageEventData{
		ID:      id,
		Role:    role,
		Content: content,
	})
}

func NewAIMessageEvent(id, role, content string) *AgentEvent {
	return NewAgentEvent(EventAIMessage, "", AIMessageEventData{
		ID:      id,
		Role:    role,
		Content: content,
	})
}

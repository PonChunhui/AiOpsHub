package message_bus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/aiops/AiOpsHub/backend/pkg/redis"
)

type MessageBus struct {
	RedisClient *redis.RedisClient
	Channels    map[string]context.CancelFunc
	Mutex       sync.RWMutex
}

func NewMessageBus(redisClient *redis.RedisClient) *MessageBus {
	bus := &MessageBus{
		RedisClient: redisClient,
		Channels:    make(map[string]context.CancelFunc),
	}

	logger.Info("Created Message Bus")
	return bus
}

func (mb *MessageBus) Publish(channel string, message *Message) error {
	messageJSON, err := message.ToJSON()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to marshal message: %v", err))
		return err
	}

	err = mb.RedisClient.Publish(context.Background(), channel, messageJSON)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to publish message to channel %s: %v", channel, err))
		return err
	}

	logger.Info(fmt.Sprintf("Published message %s to channel %s (type: %s, sender: %s, receiver: %s)",
		message.MessageID, channel, message.MessageType, message.Sender, message.Receiver))

	return nil
}

func (mb *MessageBus) Subscribe(channel string, handler MessageHandler) error {
	mb.Mutex.Lock()
	defer mb.Mutex.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		pubsub := mb.RedisClient.Subscribe(ctx, channel)
		ch := pubsub.Channel()

		for {
			select {
			case <-ctx.Done():
				logger.Info(fmt.Sprintf("Stopped subscribing to channel %s", channel))
				return
			case msg := <-ch:
				message, err := FromJSON(msg.Payload)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to unmarshal message from channel %s: %v", channel, err))
					continue
				}

				logger.Info(fmt.Sprintf("Received message %s from channel %s (type: %s, sender: %s)",
					message.MessageID, channel, message.MessageType, message.Sender))

				if handler != nil {
					handler.Handle(message)
				}
			}
		}
	}()

	mb.Channels[channel] = cancel
	logger.Info(fmt.Sprintf("Started subscribing to channel %s", channel))

	return nil
}

func (mb *MessageBus) Unsubscribe(channel string) error {
	mb.Mutex.Lock()
	defer mb.Mutex.Unlock()

	cancel, exists := mb.Channels[channel]
	if !exists {
		logger.Info(fmt.Sprintf("Channel %s not subscribed", channel))
		return nil
	}

	cancel()
	delete(mb.Channels, channel)

	logger.Info(fmt.Sprintf("Unsubscribed from channel %s", channel))
	return nil
}

func (mb *MessageBus) Route(message *Message) error {
	channel := mb.GetChannelForReceiver(message.Receiver)

	if message.IsBroadcast() {
		return mb.broadcastEventInternal(message)
	}

	return mb.Publish(channel, message)
}

func (mb *MessageBus) GetChannelForReceiver(receiver string) string {
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

func (mb *MessageBus) broadcastEventInternal(message *Message) error {
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

	for _, channel := range channels {
		err := mb.Publish(channel, message)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to broadcast to channel %s: %v", channel, err))
			continue
		}
	}

	logger.Info(fmt.Sprintf("Broadcasted message %s to all channels", message.MessageID))
	return nil
}

func (mb *MessageBus) SendTaskRequest(taskID, sender, receiver, sessionID, workflowID string, content *TaskRequestContent) error {
	messageContent := map[string]interface{}{
		"task":       content.Task,
		"parameters": content.Parameters,
		"context":    content.Context,
		"priority":   content.Priority,
		"timeout":    content.Timeout,
	}

	message := NewTaskRequestMessage(taskID, sender, receiver, sessionID, workflowID, messageContent)

	return mb.Route(message)
}

func (mb *MessageBus) SendTaskResult(resultID, sender, receiver, sessionID, workflowID string, content *TaskResultContent) error {
	messageContent := map[string]interface{}{
		"result":      content.Result,
		"status":      content.Status,
		"error":       content.Error,
		"tokens_used": content.TokensUsed,
		"duration":    content.Duration,
	}

	message := NewTaskResultMessage(resultID, sender, receiver, sessionID, workflowID, messageContent)

	return mb.Route(message)
}

func (mb *MessageBus) SendCollaborationRequest(requestID, sender, receiver, sessionID, workflowID string, content *CollaborationRequestContent) error {
	messageContent := map[string]interface{}{
		"request":    content.Request,
		"parameters": content.Parameters,
		"priority":   content.Priority,
		"timeout":    content.Timeout,
	}

	message := NewCollaborationRequestMessage(requestID, sender, receiver, sessionID, workflowID, messageContent)

	return mb.Route(message)
}

func (mb *MessageBus) SendStateUpdate(updateID, sender, sessionID string, content *StateUpdateContent) error {
	messageContent := map[string]interface{}{
		"agent_id":     content.AgentID,
		"status":       content.Status,
		"progress":     content.Progress,
		"current_task": content.CurrentTask,
	}

	message := NewStateUpdateMessage(updateID, sender, sessionID, messageContent)

	return mb.Publish("state-channel", message)
}

func (mb *MessageBus) BroadcastEvent(eventID, sender, sessionID, workflowID string, content *EventBroadcastContent) error {
	messageContent := map[string]interface{}{
		"event":       content.Event,
		"description": content.Description,
		"data":        content.Data,
	}

	message := NewEventBroadcastMessage(eventID, sender, sessionID, workflowID, messageContent)

	return mb.broadcastEventInternal(message)
}

func (mb *MessageBus) Close() {
	mb.Mutex.Lock()
	defer mb.Mutex.Unlock()

	for channel, cancel := range mb.Channels {
		cancel()
		logger.Info(fmt.Sprintf("Closed subscription to channel %s", channel))
	}

	mb.Channels = make(map[string]context.CancelFunc)
	logger.Info("Message Bus closed")
}

type MessageHandler interface {
	Handle(message *Message)
}

type DefaultMessageHandler struct {
	HandlerFunc func(message *Message)
}

func (h *DefaultMessageHandler) Handle(message *Message) {
	if h.HandlerFunc != nil {
		h.HandlerFunc(message)
	}
}

func NewDefaultMessageHandler(handlerFunc func(message *Message)) *DefaultMessageHandler {
	return &DefaultMessageHandler{
		HandlerFunc: handlerFunc,
	}
}

type MessageStats struct {
	PublishedCount  int
	ReceivedCount   int
	ErrorCount      int
	AverageLatency  time.Duration
	LastMessageTime time.Time
}

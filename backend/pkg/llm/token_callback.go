package llm

import (
	"context"
	"fmt"
	"sync"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type TokenRecorder struct {
	sessionID string
	agentID   string
	modelName string
	tokenSvc  TokenServiceInterface
	mu        sync.Mutex
}

type TokenServiceInterface interface {
	RecordUsageFromData(ctx context.Context, usage TokenUsageData) error
}

type TokenUsageData struct {
	SessionID    string
	AgentID      string
	Model        string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

func NewTokenRecorder(sessionID, agentID, modelName string, tokenSvc TokenServiceInterface) *TokenRecorder {
	return &TokenRecorder{
		sessionID: sessionID,
		agentID:   agentID,
		modelName: modelName,
		tokenSvc:  tokenSvc,
	}
}

func (r *TokenRecorder) CreateCallbackHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			mo := model.ConvCallbackOutput(output)
			if mo == nil || mo.Message == nil || mo.Message.ResponseMeta == nil || mo.Message.ResponseMeta.Usage == nil {
				logger.Info("No token usage information available in callback")
				return ctx
			}

			usage := mo.Message.ResponseMeta.Usage
			r.mu.Lock()
			defer r.mu.Unlock()

			tokenData := TokenUsageData{
				SessionID:    r.sessionID,
				AgentID:      r.agentID,
				Model:        r.modelName,
				InputTokens:  usage.PromptTokens,
				OutputTokens: usage.CompletionTokens,
				TotalTokens:  usage.TotalTokens,
			}

			if r.tokenSvc != nil {
				if err := r.tokenSvc.RecordUsageFromData(ctx, tokenData); err != nil {
					logger.Error(fmt.Sprintf("Failed to record token usage via callback: %v", err))
				} else {
					logger.Info(fmt.Sprintf("✅ Token usage recorded via callback: session=%s, input=%d, output=%d, total=%d",
						r.sessionID, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens))
				}
			}

			return ctx
		}).
		Build()
}

func ExtractTokenUsage(msg *schema.Message) *TokenUsageData {
	if msg == nil || msg.ResponseMeta == nil || msg.ResponseMeta.Usage == nil {
		return nil
	}

	usage := msg.ResponseMeta.Usage
	return &TokenUsageData{
		InputTokens:  usage.PromptTokens,
		OutputTokens: usage.CompletionTokens,
		TotalTokens:  usage.TotalTokens,
	}
}

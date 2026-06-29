package llm

import (
	"context"
	"fmt"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type EinoLLMConfig struct {
	Model       string
	Temperature float64
	MaxTokens   int
	Provider    string
	APIKey      string
	BaseURL     string
}

type EinoLLM struct {
	chatModel model.ChatModel
	config    EinoLLMConfig
}

func NewEinoLLM(config EinoLLMConfig) (*EinoLLM, error) {
	var chatModel model.ChatModel
	var err error

	switch config.Provider {
	case "openai":
		chatModel, err = createOpenAIChatModel(config)
	case "aliyun_bailian", "qwen":
		chatModel, err = createAliyunBailianChatModel(config)
	default:
		chatModel, err = createOpenAIChatModel(config)
	}

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create EinoLLM: %v", err))
		return nil, err
	}

	llm := &EinoLLM{
		chatModel: chatModel,
		config:    config,
	}

	logger.Info(fmt.Sprintf("Created EinoLLM with provider: %s, model: %s", config.Provider, config.Model))
	return llm, nil
}

func createOpenAIChatModel(config EinoLLMConfig) (model.ChatModel, error) {
	modelName := config.Model
	if modelName == "" {
		modelName = "gpt-3.5-turbo"
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	ctx := context.Background()
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   modelName,
		APIKey:  config.APIKey,
		BaseURL: config.BaseURL,
	})
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("Created OpenAI ChatModel with model: %s", modelName))
	return chatModel, nil
}

func createAliyunBailianChatModel(config EinoLLMConfig) (model.ChatModel, error) {
	modelName := config.Model
	if modelName == "" {
		modelName = "qwen3.7-max"
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("Aliyun Bailian API key not configured")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}

	ctx := context.Background()
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   modelName,
		APIKey:  config.APIKey,
		BaseURL: baseURL,
	})
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("Created Aliyun Bailian ChatModel with model: %s at %s", modelName, baseURL))
	return chatModel, nil
}

func (e *EinoLLM) Generate(ctx context.Context, prompt string) (string, error) {
	logger.Info(fmt.Sprintf("EinoLLM generating response for prompt: %s", prompt[:min(50, len(prompt))]))

	messages := []*schema.Message{
		schema.UserMessage(prompt),
	}

	result, err := e.chatModel.Generate(ctx, messages)
	if err != nil {
		logger.Error(fmt.Sprintf("EinoLLM generation failed: %v", err))
		return "", err
	}

	response := result.Content
	logger.Info(fmt.Sprintf("EinoLLM generated response length: %d", len(response)))

	return response, nil
}

func (e *EinoLLM) GenerateWithCallback(ctx context.Context, prompt string, handler callbacks.Handler) (string, *schema.Message, error) {
	logger.Info(fmt.Sprintf("EinoLLM generating with callback for prompt: %s", prompt[:min(50, len(prompt))]))

	ctx = callbacks.InitCallbacks(ctx, &callbacks.RunInfo{
		Name:      "EinoLLM",
		Type:      e.config.Provider,
		Component: "ChatModel",
	}, handler)

	messages := []*schema.Message{
		schema.UserMessage(prompt),
	}

	result, err := e.chatModel.Generate(ctx, messages)
	if err != nil {
		logger.Error(fmt.Sprintf("EinoLLM generation with callback failed: %v", err))
		return "", nil, err
	}

	response := result.Content
	logger.Info(fmt.Sprintf("EinoLLM generated response with callback, length: %d", len(response)))

	return response, result, nil
}

func (e *EinoLLM) GenerateWithSystemPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	logger.Info(fmt.Sprintf("EinoLLM generating with system prompt"))

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	}

	result, err := e.chatModel.Generate(ctx, messages)
	if err != nil {
		logger.Error(fmt.Sprintf("EinoLLM generation failed: %v", err))
		return "", err
	}

	return result.Content, nil
}

func (e *EinoLLM) StreamGenerate(ctx context.Context, prompt string) (<-chan string, error) {
	logger.Info(fmt.Sprintf("EinoLLM streaming generation for prompt: %s", prompt[:min(50, len(prompt))]))

	messages := []*schema.Message{
		schema.UserMessage(prompt),
	}

	streamReader, err := e.chatModel.Stream(ctx, messages)
	if err != nil {
		logger.Error(fmt.Sprintf("EinoLLM streaming failed: %v", err))
		return nil, err
	}

	outputChan := make(chan string, 100)

	go func() {
		defer close(outputChan)
		defer streamReader.Close()

		for {
			chunk, err := streamReader.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					logger.Info("Stream completed")
					return
				}
				logger.Error(fmt.Sprintf("Stream receive error: %v", err))
				outputChan <- fmt.Sprintf("错误: %v", err)
				return
			}
			if chunk.Content != "" {
				outputChan <- chunk.Content
			}
		}
	}()

	return outputChan, nil
}

func (e *EinoLLM) StreamGenerateWithCallback(ctx context.Context, prompt string, handler callbacks.Handler) (<-chan string, *schema.Message, error) {
	logger.Info(fmt.Sprintf("EinoLLM streaming with callback for prompt: %s", prompt[:min(50, len(prompt))]))

	ctx = callbacks.InitCallbacks(ctx, &callbacks.RunInfo{
		Name:      "EinoLLM",
		Type:      e.config.Provider,
		Component: "ChatModel",
	}, handler)

	messages := []*schema.Message{
		schema.UserMessage(prompt),
	}

	streamReader, err := e.chatModel.Stream(ctx, messages)
	if err != nil {
		logger.Error(fmt.Sprintf("EinoLLM streaming with callback failed: %v", err))
		return nil, nil, err
	}

	outputChan := make(chan string, 100)
	var finalMsg *schema.Message

	go func() {
		defer close(outputChan)
		defer streamReader.Close()

		for {
			chunk, err := streamReader.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					logger.Info("Stream with callback completed")
					return
				}
				logger.Error(fmt.Sprintf("Stream receive error: %v", err))
				return
			}
			if chunk.Content != "" {
				outputChan <- chunk.Content
			}
			if chunk.ResponseMeta != nil {
				finalMsg = chunk
			}
		}
	}()

	return outputChan, finalMsg, nil
}

func (e *EinoLLM) GetChatModel() model.ChatModel {
	return e.chatModel
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

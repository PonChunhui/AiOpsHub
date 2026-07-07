package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// OpenAI ChatCompletionChunk 格式
// 用于兼容 Tiny-Robot-kit

type OpenAIChatCompletionChunk struct {
	ID      string              `json:"id"`
	Object  string              `json:"object"`
	Created int64               `json:"created"`
	Model   string              `json:"model"`
	Choices []OpenAIChunkChoice `json:"choices"`
}

type OpenAIChunkChoice struct {
	Index        int              `json:"index"`
	Delta        OpenAIChunkDelta `json:"delta"`
	FinishReason string           `json:"finish_reason"`
}

type OpenAIChunkDelta struct {
	Content          string                `json:"content,omitempty"`
	ReasoningContent string                `json:"reasoning_content,omitempty"`
	ToolCalls        []OpenAIToolCallDelta `json:"tool_calls,omitempty"`
}

type OpenAIToolCallDelta struct {
	Index    int                 `json:"index"`
	ID       string              `json:"id"`
	Type     string              `json:"type"`
	Function OpenAIFunctionDelta `json:"function"`
}

type OpenAIFunctionDelta struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments"` // 🔧 移除omitempty，确保arguments总是被发送
}

// ToolCallsBuffer 用于合并流式工具调用分片（改进版）
// 修复问题：确保同一个工具调用的所有分片使用相同的index和ID
type ToolCallsBuffer struct {
	toolCalls    map[string]*ToolCallBuilder
	indexCounter int // 用于跟踪当前已使用的最大index
}

type ToolCallBuilder struct {
	id         string
	name       string
	argsBuffer string
	index      int
}

func NewToolCallsBuffer() *ToolCallsBuffer {
	return &ToolCallsBuffer{
		toolCalls:    make(map[string]*ToolCallBuilder),
		indexCounter: 0,
	}
}

// ConvertAgentEventToOpenAIChunk 将 AgentEvent 转换为 OpenAI ChatCompletionChunk
func ConvertAgentEventToOpenAIChunk(event *AgentEvent, buffer *ToolCallsBuffer) (*OpenAIChatCompletionChunk, error) {
	chunkID := fmt.Sprintf("chatcmpl-%d", event.Timestamp)
	if chunkID == "chatcmpl-0" {
		chunkID = fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	}

	delta := OpenAIChunkDelta{}
	finishReason := ""

	switch event.Type {
	case EventContentChunk:
		// 内容分片 - Data 是 ContentChunkEventData
		if contentData, ok := event.Data.(ContentChunkEventData); ok {
			delta.Content = contentData.Content
		}

	case EventThinking:
		// 思考内容 - Data 是 ThinkingEventData
		if thinkingData, ok := event.Data.(ThinkingEventData); ok {
			delta.ReasoningContent = thinkingData.Content
		} else {
			// 调试：打印实际类型
			fmt.Printf("DEBUG: EventThinking Data type: %T, value: %v\n", event.Data, event.Data)
		}

	case EventToolCall:
		// 工具调用 - Data 是 ToolCallEventData
		// 🔧 修复方案：不发送中间分片事件，只在buffer中合并
		//    在Done事件时发送完整的tool_call，避免前端处理中间状态
		if toolCallData, ok := event.Data.(ToolCallEventData); ok {
			toolID := toolCallData.ToolID
			toolName := toolCallData.ToolName
			argsRaw := toolCallData.ArgsRaw
			argsComplete := toolCallData.ArgsComplete

			var matchedBuilder *ToolCallBuilder
			var matchedKey string

			// 🔧 最终修复：正确处理工具调用分片
			// 问题：LLM可能为同一调用的不同分片使用不同toolID（第一个有ID，后续没有）
			// 解决：优先用ID匹配，无ID时找"参数不完整"的builder

			// 1. 有ID -> 尝试通过ID匹配
			if toolID != "" {
				if builder, exists := buffer.toolCalls[toolID]; exists {
					matchedBuilder = builder
					matchedKey = toolID
				}
			}

			// 2. 无ID或ID匹配失败 -> 找参数不完整的builder
			if matchedBuilder == nil {
				// 遍历所有builder，找到参数不完整的（JSON未闭合）
				for existingID, builder := range buffer.toolCalls {
					// 检查参数是否完整
					isComplete := false
					if len(builder.argsBuffer) > 0 && builder.argsBuffer[len(builder.argsBuffer)-1] == '}' {
						// 最后字符是'}'，尝试parse验证完整性
						var testMap map[string]interface{}
						if err := json.Unmarshal([]byte(builder.argsBuffer), &testMap); err == nil {
							isComplete = true // JSON完整，parse成功
						}
					}

					// 找到第一个参数不完整的builder
					if !isComplete {
						matchedBuilder = builder
						matchedKey = existingID
						break // 找到就停止
					}
				}
			}

			// 3. 如果还是没有匹配到，创建新的builder
			if matchedBuilder == nil {
				if toolID != "" {
					matchedKey = toolID
				} else {
					// 生成临时key（这种情况不应该发生）
					matchedKey = fmt.Sprintf("tc_%d", buffer.indexCounter)
				}
				buffer.toolCalls[matchedKey] = &ToolCallBuilder{
					id:    toolID,
					name:  toolName,
					index: buffer.indexCounter,
				}
				buffer.indexCounter++
				matchedBuilder = buffer.toolCalls[matchedKey]
			}

			// 更新builder数据
			if toolName != "" && matchedBuilder.name == "" {
				matchedBuilder.name = toolName
			}
			if toolID != "" && matchedBuilder.id == "" {
				matchedBuilder.id = toolID
			}

			// 合并参数：优先使用argsComplete，否则累加argsRaw
			if argsComplete != "" {
				matchedBuilder.argsBuffer = argsComplete
			} else if argsRaw != "" {
				matchedBuilder.argsBuffer += argsRaw
			}
		}

	case EventDone:
		// 完成 - 发送完整的工具调用（如果有）
		if len(buffer.toolCalls) > 0 {
			// 发送合并后的完整工具调用
			var completeToolCalls []OpenAIToolCallDelta
			for _, builder := range buffer.toolCalls {
				completeToolCalls = append(completeToolCalls, OpenAIToolCallDelta{
					Index: builder.index,
					ID:    builder.id,
					Type:  "function",
					Function: OpenAIFunctionDelta{
						Name:      builder.name,
						Arguments: builder.argsBuffer,
					},
				})
			}

			return &OpenAIChatCompletionChunk{
				ID:      chunkID,
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   "agent",
				Choices: []OpenAIChunkChoice{
					{
						Index: 0,
						Delta: OpenAIChunkDelta{
							ToolCalls: completeToolCalls,
						},
						FinishReason: "tool_calls",
					},
				},
			}, nil
		}

		finishReason = "stop"

	case EventError:
		// 错误 - Data 是 ErrorEventData
		if errorData, ok := event.Data.(ErrorEventData); ok {
			delta.Content = fmt.Sprintf("\n\n**错误**: %s", errorData.Message)
		}

	default:
		// 其他事件不转换为 OpenAI 格式
		return nil, nil
	}

	return &OpenAIChatCompletionChunk{
		ID:      chunkID,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   "agent",
		Choices: []OpenAIChunkChoice{
			{
				Index:        0,
				Delta:        delta,
				FinishReason: finishReason,
			},
		},
	}, nil
}

// ToSSE 将 OpenAIChatCompletionChunk 转换为 SSE 格式字符串
func (chunk *OpenAIChatCompletionChunk) ToSSE() string {
	data, err := json.Marshal(chunk)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("data: %s\n\n", string(data))
}

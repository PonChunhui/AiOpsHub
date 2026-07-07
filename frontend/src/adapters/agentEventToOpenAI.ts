import type { ChatCompletionStreamResponse, ChatCompletionStreamResponseDelta } from '@opentiny/tiny-robot-kit'
import type { AgentEvent } from '@/types/agentEvent'

/**
 * 将 AgentEvent SSE 流转换为 OpenAI ChatCompletionStreamResponse 格式
 * 用于兼容 Tiny-Robot-kit 的 useMessage 和 toolPlugin
 */

export class AgentEventToOpenAIAdapter {
  private toolCallsBuffer: Map<string, { id: string; name: string; arguments: string }> = new Map()
  private currentContent: string = ''
  private currentReasoning: string = ''

  /**
   * 处理单个 AgentEvent，转换为 ChatCompletionStreamResponse
   */
  processEvent(event: AgentEvent): ChatCompletionStreamResponse | null {
    const delta: ChatCompletionStreamResponseDelta = {}
    
    switch (event.type) {
      case 'content_chunk':
        // 内容分片
        const content = event.data?.content || ''
        this.currentContent += content
        delta.content = content
        break
      
      case 'thinking':
        // 思考内容
        const reasoning = event.data?.content || ''
        this.currentReasoning += reasoning
        delta.reasoning_content = reasoning
        break
      
      case 'tool_call':
        // 工具调用 - 合并分片
        const toolId = event.data?.tool_id || ''
        const toolName = event.data?.tool_name || ''
        const argsRaw = event.data?.args_raw || event.data?.args_complete || ''
        
        if (toolId) {
          if (!this.toolCallsBuffer.has(toolId)) {
            this.toolCallsBuffer.set(toolId, {
              id: toolId,
              name: toolName,
              arguments: argsRaw
            })
          } else {
            const existing = this.toolCallsBuffer.get(toolId)!
            if (toolName) existing.name = toolName
            existing.arguments += argsRaw
          }
          
          // 发送增量更新
          delta.tool_calls = [{
            index: this.toolCallsBuffer.size - 1,
            id: toolId,
            type: 'function',
            function: {
              name: toolName,
              arguments: argsRaw
            }
          }]
        }
        break
      
      case 'tool_result':
        // 工具结果 - 不需要转换，由 toolPlugin 处理
        break
      
      case 'done':
        // 完成 - 发送最终状态
        if (this.toolCallsBuffer.size > 0) {
          // 发送合并后的完整工具调用
          const completeToolCalls = Array.from(this.toolCallsBuffer.values()).map((tc, index) => ({
            index,
            id: tc.id,
            type: 'function',
            function: {
              name: tc.name,
              arguments: tc.arguments
            }
          }))
          
          return {
            id: `completion-${Date.now()}`,
            object: 'chat.completion.chunk',
            created: Math.floor(Date.now() / 1000),
            model: 'agent',
            choices: [{
              index: 0,
              delta: {
                tool_calls: completeToolCalls
              },
              finish_reason: 'tool_calls'
            }]
          }
        }
        
        return {
          id: `completion-${Date.now()}`,
          object: 'chat.completion.chunk',
          created: Math.floor(Date.now() / 1000),
          model: 'agent',
          choices: [{
            index: 0,
            delta: {},
            finish_reason: 'stop'
          }]
        }
      
      case 'error':
        // 错误
        delta.content = `\n\n**错误**: ${event.data?.message || '未知错误'}`
        break
      
      default:
        return null
    }
    
    // 返回增量响应
    return {
      id: `completion-${event.timestamp || Date.now()}`,
      object: 'chat.completion.chunk',
      created: Math.floor((event.timestamp || Date.now()) / 1000),
      model: 'agent',
      choices: [{
        index: 0,
        delta,
        finish_reason: null
      }]
    }
  }

  /**
   * 获取当前累积的内容
   */
  getCurrentContent(): string {
    return this.currentContent
  }

  /**
   * 获取当前累积的思考内容
   */
  getCurrentReasoning(): string {
    return this.currentReasoning
  }

  /**
   * 获取合并后的工具调用列表
   */
  getCompleteToolCalls(): Array<{ id: string; type: string; function: { name: string; arguments: string } }> {
    return Array.from(this.toolCallsBuffer.values()).map(tc => ({
      id: tc.id,
      type: 'function',
      function: {
        name: tc.name,
        arguments: tc.arguments
      }
    }))
  }

  /**
   * 重置缓冲区
   */
  reset(): void {
    this.toolCallsBuffer.clear()
    this.currentContent = ''
    this.currentReasoning = ''
  }
}

/**
 * 将 AgentEvent SSE 流转换为 ChatCompletionStreamResponse AsyncGenerator
 */
export async function* agentEventStreamToOpenAIGenerator(
  eventStream: AsyncGenerator<AgentEvent>
): AsyncGenerator<ChatCompletionStreamResponse> {
  const adapter = new AgentEventToOpenAIAdapter()
  
  for await (const event of eventStream) {
    const openAIResponse = adapter.processEvent(event)
    if (openAIResponse) {
      yield openAIResponse
    }
  }
}
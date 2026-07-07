import type { AgentEvent } from '@/types/agentEvent'

export interface TinyRobotBubbleMessage {
  id?: string
  role?: 'user' | 'assistant' | 'tool' | string
  content?: string | Array<{ type: string; text?: string; [key: string]: any }>
  reasoning_content?: string
  tool_calls?: TinyRobotToolCall[]
  tool_call_id?: string
  name?: string
  loading?: boolean
  state?: {
    agentVisualization?: AgentVisualizationData
    success?: boolean
    [key: string]: any
  }
}

export interface TinyRobotToolCall {
  id: string
  type: 'function'
  function: {
    name: string
    arguments: string
  }
}

export interface AgentVisualizationData {
  agentPath: Array<{
    agent_id?: string
    agent_name: string
    action: string
    timestamp?: number
  }>
  currentAgent?: string
  events: AgentEvent[]
  toolCallsBuffer?: Map<string, { id: string; name: string; args: string }>
}

export function convertAgentEventsToTinyRobotMessages(
  events: AgentEvent[],
  userMessage?: string
): TinyRobotBubbleMessage[] {
  const messages: TinyRobotBubbleMessage[] = []
  
  if (userMessage) {
    messages.push({
      role: 'user',
      content: userMessage
    })
  }
  
  const assistantMessage: TinyRobotBubbleMessage = {
    role: 'assistant',
    content: '',
    state: {
      agentVisualization: {
        agentPath: [],
        events: []
      }
    }
  }
  
  // 使用 Map 合并流式工具调用分片
  const toolCallsMap = new Map<string, { id: string; name: string; argsBuffer: string }>()
  
  const toolMessages: TinyRobotBubbleMessage[] = []
  
  for (const event of events) {
    switch (event.type) {
      case 'thinking':
        assistantMessage.reasoning_content = (assistantMessage.reasoning_content || '') + (event.data?.content || '')
        if (assistantMessage.state && assistantMessage.state.agentVisualization) {
          assistantMessage.state.agentVisualization.events.push(event)
        }
        break
      
      case 'tool_call':
        const toolId = event.data?.tool_id || ''
        const toolName = event.data?.tool_name || ''
        // 直接使用原始参数字符串（args_raw）或完整参数（args_complete）
        const argsStr = event.data?.args_complete || event.data?.args_raw || ''
        
        if (toolId) {
          // 合并工具调用分片
          if (!toolCallsMap.has(toolId)) {
            toolCallsMap.set(toolId, { id: toolId, name: toolName, argsBuffer: argsStr })
          } else {
            const builder = toolCallsMap.get(toolId)!
            if (toolName) builder.name = toolName
            builder.argsBuffer += argsStr
          }
        }
        
        if (assistantMessage.state && assistantMessage.state.agentVisualization) {
          assistantMessage.state.agentVisualization.events.push(event)
        }
        break
      
      case 'tool_result':
        toolMessages.push({
          role: 'tool',
          tool_call_id: event.data?.tool_id || '',
          name: event.data?.tool_name || '',
          content: event.data?.result || '',
          state: {
            success: event.data?.success || false
          }
        })
        if (assistantMessage.state && assistantMessage.state.agentVisualization) {
          assistantMessage.state.agentVisualization.events.push(event)
        }
        break
      
      case 'content_chunk':
        assistantMessage.content = (assistantMessage.content || '') + (event.data?.content || '')
        break
      
      case 'agent_transfer':
        if (assistantMessage.state && assistantMessage.state.agentVisualization) {
          assistantMessage.state.agentVisualization.agentPath.push({
            agent_name: event.data?.from_agent || '',
            action: 'transfer_start'
          })
          assistantMessage.state.agentVisualization.agentPath.push({
            agent_name: event.data?.to_agent || '',
            action: 'transfer_complete'
          })
          assistantMessage.state.agentVisualization.currentAgent = event.data?.to_agent
          assistantMessage.state.agentVisualization.events.push(event)
        }
        break
      
      case 'done':
        if (event.run_path && event.run_path.length > 0 && assistantMessage.state && assistantMessage.state.agentVisualization) {
          assistantMessage.state.agentVisualization.agentPath = event.run_path
        }
        break
      
      case 'error':
        assistantMessage.content = (assistantMessage.content || '') + '\n\n**错误**: ' + (event.data?.message || '未知错误')
        if (assistantMessage.state && assistantMessage.state.agentVisualization) {
          assistantMessage.state.agentVisualization.events.push(event)
        }
        break
      
      default:
        break
    }
  }
  
  // 将合并后的工具调用转换为数组
  if (toolCallsMap.size > 0) {
    assistantMessage.tool_calls = []
    toolCallsMap.forEach((builder) => {
      if (builder.name && builder.argsBuffer) {
        assistantMessage.tool_calls!.push({
          id: builder.id,
          type: 'function',
          function: {
            name: builder.name,
            arguments: builder.argsBuffer
          }
        })
      }
    })
  }
  
  messages.push(assistantMessage)
  messages.push(...toolMessages)
  
  return messages
}

export function appendAgentEventToMessage(
  existingMessage: TinyRobotBubbleMessage,
  event: AgentEvent
): TinyRobotBubbleMessage {
  const updatedMessage: TinyRobotBubbleMessage = {
    ...existingMessage,
    content: existingMessage.content 
      ? (typeof existingMessage.content === 'string' 
          ? existingMessage.content 
          : [...existingMessage.content])
      : existingMessage.content,
    tool_calls: existingMessage.tool_calls ? [...existingMessage.tool_calls] : undefined,
    state: {
      ...existingMessage.state,
      agentVisualization: existingMessage.state?.agentVisualization ? {
        ...existingMessage.state.agentVisualization,
        agentPath: [...(existingMessage.state.agentVisualization.agentPath || [])],
        events: [...(existingMessage.state.agentVisualization.events || [])],
        toolCallsBuffer: existingMessage.state.agentVisualization.toolCallsBuffer 
          ? new Map(existingMessage.state.agentVisualization.toolCallsBuffer) 
          : undefined
      } : {
        agentPath: [],
        events: [],
        toolCallsBuffer: undefined
      }
    }
  }
  
  if (!updatedMessage.state) {
    updatedMessage.state = {}
  }
  if (!updatedMessage.state.agentVisualization) {
    updatedMessage.state.agentVisualization = {
      agentPath: [],
      events: []
    }
  }
  
  switch (event.type) {
    case 'thinking':
      updatedMessage.reasoning_content = (updatedMessage.reasoning_content || '') + (event.data?.content || '')
      updatedMessage.state.agentVisualization.events.push(event)
      break
    
    case 'tool_call':
      const tcToolId = event.data?.tool_id || ''
      const tcToolName = event.data?.tool_name || ''
      const tcArgsRaw = event.data?.args_raw || ''
      const tcArgsComplete = event.data?.args_complete || ''
      
      console.log('[ToolCall] Received:', {
        toolId: tcToolId,
        toolName: tcToolName,
        hasArgsRaw: Boolean(tcArgsRaw),
        hasArgsComplete: Boolean(tcArgsComplete)
      })
      
      if (!updatedMessage.state.agentVisualization.toolCallsBuffer) {
        updatedMessage.state.agentVisualization.toolCallsBuffer = new Map()
      }
      
      const buffer = updatedMessage.state.agentVisualization.toolCallsBuffer
      
      if (tcToolId) {
        const existing = buffer.get(tcToolId)
        const newName = tcToolName || existing?.name || ''
        const newArgs = tcArgsComplete || (tcArgsRaw ? (existing?.args || '') + tcArgsRaw : (existing?.args || ''))
        
        buffer.set(tcToolId, {
          id: tcToolId,
          name: newName,
          args: newArgs
        })
        
        console.log('[ToolCall] Buffer updated:', tcToolId, 'name:', newName, 'argsLen:', newArgs.length)
        
        updatedMessage.tool_calls = []
        buffer.forEach((tc) => {
          if (tc.name && tc.args) {
            updatedMessage.tool_calls!.push({
              id: tc.id,
              type: 'function',
              function: {
                name: tc.name,
                arguments: tc.args
              }
            })
          } else {
            console.log('[ToolCall] Buffer incomplete:', tc.id, 'waiting for', tc.name ? 'args' : 'name')
          }
        })
        
        console.log('[ToolCall] Display count:', updatedMessage.tool_calls.length, 'Buffer count:', buffer.size)
      }
      
      updatedMessage.state.agentVisualization.events.push(event)
      break
    
    case 'content_chunk':
      if (event.data?.content) {
        if (typeof updatedMessage.content === 'string') {
          updatedMessage.content = updatedMessage.content + event.data.content
        } else {
          updatedMessage.content = event.data.content
        }
        console.log('[ContentChunk] Total length:', 
          typeof updatedMessage.content === 'string' ? updatedMessage.content.length : 'array')
      }
      break
    
    case 'agent_transfer':
      if (updatedMessage.state.agentVisualization) {
        updatedMessage.state.agentVisualization.agentPath.push({
          agent_name: event.data?.from_agent || '',
          action: 'transfer_start'
        })
        updatedMessage.state.agentVisualization.agentPath.push({
          agent_name: event.data?.to_agent || '',
          action: 'transfer_complete'
        })
        updatedMessage.state.agentVisualization.currentAgent = event.data?.to_agent
        updatedMessage.state.agentVisualization.events.push(event)
      }
      break
    
    case 'done':
      if (event.run_path && event.run_path.length > 0) {
        if (updatedMessage.state.agentVisualization) {
          updatedMessage.state.agentVisualization.agentPath = event.run_path
        }
      }
      
      if (updatedMessage.state.agentVisualization.toolCallsBuffer) {
        updatedMessage.state.agentVisualization.toolCallsBuffer.clear()
        console.log('[ToolCall] Buffer cleared')
      }
      
      if (updatedMessage.tool_calls && updatedMessage.tool_calls.length > 0) {
        updatedMessage.tool_calls = updatedMessage.tool_calls.filter(tc => {
          const hasName = tc.function?.name && tc.function.name.trim() !== ''
          const hasArgs = tc.function?.arguments && tc.function.arguments.trim() !== ''
          const isValid = hasName && hasArgs
          
          if (!isValid) {
            console.warn('[ToolCall] Filtered invalid:', tc.id, 'name:', tc.function?.name)
          }
          
          return isValid
        })
        
        console.log('[ToolCall] Final valid count:', updatedMessage.tool_calls.length)
      }
      
      console.log('[Done] Content finalized, length:', 
        typeof updatedMessage.content === 'string' ? updatedMessage.content.length : 'array')
      
      updatedMessage.loading = false
      break
    
    case 'error':
      const errorMsg = event.data?.message || event.data?.error || '未知错误'
      if (typeof updatedMessage.content === 'string') {
        updatedMessage.content = updatedMessage.content + '\n\n**错误**: ' + errorMsg
      }
      updatedMessage.state.agentVisualization.events.push(event)
      updatedMessage.loading = false
      break
    
    default:
      break
  }
  
  return updatedMessage
}

export function createInitialAssistantMessage(): TinyRobotBubbleMessage {
  return {
    role: 'assistant',
    content: '',
    loading: true,
    state: {
      agentVisualization: {
        agentPath: [],
        events: []
      }
    }
  }
}
import type { AgentEvent } from '@/types/agentEvent'

export interface SchemaJsonNode {
  componentName: string
  props?: Record<string, any>
  children?: SchemaJsonNode[]
  text?: string
}

export function convertAgentEventToSchemaJson(event: AgentEvent): SchemaJsonNode | null {
  switch (event.type) {
    case 'thinking':
      return {
        componentName: 'ThinkingBlock',
        props: {
          agentName: event.agent_name,
          content: event.data?.content || '',
          timestamp: event.timestamp
        }
      }
    
    case 'tool_call':
      return {
        componentName: 'ToolCallCard',
        props: {
          toolName: event.data?.tool_name || '',
          toolId: event.data?.tool_id || '',
          args: event.data?.args || {}
        }
      }
    
    case 'tool_result':
      return {
        componentName: 'ToolResultCard',
        props: {
          toolName: event.data?.tool_name || '',
          result: event.data?.result || '',
          success: event.data?.success || false
        }
      }
    
    case 'content_chunk':
      return {
        componentName: 'TextContent',
        props: {
          text: event.data?.content || ''
        }
      }
    
    case 'agent_transfer':
      return {
        componentName: 'AgentTransferBlock',
        props: {
          fromAgent: event.data?.from_agent || '',
          toAgent: event.data?.to_agent || '',
          reason: event.data?.reason || ''
        }
      }
    
    case 'error':
      return {
        componentName: 'ErrorBlock',
        props: {
          message: event.data?.message || '',
          code: event.data?.code || 500
        }
      }
    
    case 'done':
      return null
    
    default:
      return null
  }
}

export function convertEventsToSchemaJson(events: AgentEvent[]): SchemaJsonNode {
  const children = events
    .map(convertAgentEventToSchemaJson)
    .filter(node => node !== null) as SchemaJsonNode[]
  
  return {
    componentName: 'Page',
    children
  }
}

export function appendEventToSchemaJson(
  existingSchema: SchemaJsonNode,
  event: AgentEvent
): SchemaJsonNode {
  const newNode = convertAgentEventToSchemaJson(event)
  if (!newNode) return existingSchema
  
  return {
    ...existingSchema,
    children: [...(existingSchema.children || []), newNode]
  }
}
export enum AgentEventType {
  THINKING = 'thinking',
  TOOL_CALL = 'tool_call',
  TOOL_RESULT = 'tool_result',
  CONTENT_CHUNK = 'content_chunk',
  AGENT_TRANSFER = 'agent_transfer',
  ERROR = 'error',
  DONE = 'done',
  RAG_REFERENCES = 'rag_references',
  USER_MESSAGE = 'user_message',
  AI_MESSAGE = 'ai_message'
}

export interface AgentEvent {
  type: AgentEventType
  agent_name: string
  run_path: AgentRunStep[]
  data: any
  timestamp: number
}

export interface AgentRunStep {
  agent_id: string
  agent_name: string
  action: string
}

export interface ThinkingEventData {
  content: string
}

export interface ToolCallEventData {
  tool_id: string
  tool_name: string
  args: Record<string, any>
}

export interface ToolResultEventData {
  tool_id: string
  tool_name: string
  result: string
  success: boolean
}

export interface ContentChunkEventData {
  content: string
}

export interface AgentTransferEventData {
  from_agent: string
  to_agent: string
  reason: string
}

export interface ErrorEventData {
  message: string
  code: number
}

export interface RagReferencesEventData {
  references: Array<{
    id: string
    title: string
    score: number
  }>
}

export interface UserMessageEventData {
  id: string
  role: string
  content: string
}

export interface AIMessageEventData {
  id: string
  role: string
  content: string
}

export type UIComponentType = 
  | 'ThinkingCard'
  | 'ToolCallCard'
  | 'ToolResultCard'
  | 'ContentChunk'
  | 'AgentTransferCard'
  | 'ErrorCard'
  | 'AgentPathVisual'

export const eventToUIComponent: Record<AgentEventType, UIComponentType | null> = {
  [AgentEventType.THINKING]: 'ThinkingCard',
  [AgentEventType.TOOL_CALL]: 'ToolCallCard',
  [AgentEventType.TOOL_RESULT]: 'ToolResultCard',
  [AgentEventType.CONTENT_CHUNK]: 'ContentChunk',
  [AgentEventType.AGENT_TRANSFER]: 'AgentTransferCard',
  [AgentEventType.ERROR]: 'ErrorCard',
  [AgentEventType.DONE]: 'AgentPathVisual',
  [AgentEventType.RAG_REFERENCES]: null,
  [AgentEventType.USER_MESSAGE]: null,
  [AgentEventType.AI_MESSAGE]: null
}

export function getEventComponent(event: AgentEvent): UIComponentType | null {
  return eventToUIComponent[event.type]
}

export function isThinkingEvent(event: AgentEvent): boolean {
  return event.type === AgentEventType.THINKING
}

export function isToolCallEvent(event: AgentEvent): boolean {
  return event.type === AgentEventType.TOOL_CALL
}

export function isToolResultEvent(event: AgentEvent): boolean {
  return event.type === AgentEventType.TOOL_RESULT
}

export function isContentChunkEvent(event: AgentEvent): boolean {
  return event.type === AgentEventType.CONTENT_CHUNK
}

export function isAgentTransferEvent(event: AgentEvent): boolean {
  return event.type === AgentEventType.AGENT_TRANSFER
}

export function isErrorEvent(event: AgentEvent): boolean {
  return event.type === AgentEventType.ERROR
}

export function isDoneEvent(event: AgentEvent): boolean {
  return event.type === AgentEventType.DONE
}
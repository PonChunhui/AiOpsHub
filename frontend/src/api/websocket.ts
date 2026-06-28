export interface WebSocketMessage {
  type: string
  timestamp: number
  data: any
}

export interface SubscriptionRequest {
  action: 'subscribe' | 'unsubscribe'
  workflow_id?: string
  session_id?: string
}

export class WebSocketClient {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private messageHandlers: Map<string, (message: WebSocketMessage) => void> = new Map()

  connect(url: string = 'ws://localhost:8080/api/v1/ws') {
    this.ws = new WebSocket(url)

    this.ws.onopen = () => {
      console.log('WebSocket connected')
      this.reconnectAttempts = 0
    }

    this.ws.onmessage = (event) => {
      const message: WebSocketMessage = JSON.parse(event.data)
      console.log('WebSocket message received:', message)
      
      const handler = this.messageHandlers.get(message.type)
      if (handler) {
        handler(message)
      }

      const allHandler = this.messageHandlers.get('*')
      if (allHandler) {
        allHandler(message)
      }
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }

    this.ws.onclose = () => {
      console.log('WebSocket closed')
      this.handleReconnect(url)
    }
  }

  private handleReconnect(url: string) {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      console.log(`Attempting reconnect ${this.reconnectAttempts}/${this.maxReconnectAttempts}`)
      
      setTimeout(() => {
        this.connect(url)
      }, this.reconnectDelay * this.reconnectAttempts)
    }
  }

  subscribe(workflowId?: string, sessionId?: string) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('WebSocket not connected')
      return
    }

    const request: SubscriptionRequest = {
      action: 'subscribe',
      workflow_id: workflowId,
      session_id: sessionId
    }

    this.ws.send(JSON.stringify(request))
    console.log('Subscribed:', request)
  }

  unsubscribe(workflowId?: string, sessionId?: string) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('WebSocket not connected')
      return
    }

    const request: SubscriptionRequest = {
      action: 'unsubscribe',
      workflow_id: workflowId,
      session_id: sessionId
    }

    this.ws.send(JSON.stringify(request))
    console.log('Unsubscribed:', request)
  }

  onMessage(type: string, handler: (message: WebSocketMessage) => void) {
    this.messageHandlers.set(type, handler)
  }

  onAllMessages(handler: (message: WebSocketMessage) => void) {
    this.messageHandlers.set('*', handler)
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.messageHandlers.clear()
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }
}

export const wsClient = new WebSocketClient()
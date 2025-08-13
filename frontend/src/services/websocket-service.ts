export type WebSocketMessage = {
  type: string
  data: any
  timestamp?: string
}

export type NotificationUpdate = {
  type: 'notification'
  id: string
  title: string
  message: string
  priority: string
  entity_type?: string
  entity_id?: string
  created_at: string
}

export type UnreadCountUpdate = {
  type: 'unread_count'
  count: number
}

export type AlertNotification = {
  type: 'alert'
  alert_type: string
  camera_name: string
  message: string
  timestamp: string
  data: any
}

export type WebSocketEventHandler = (data: any) => void

class WebSocketService {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private isConnecting = false
  private eventHandlers: Map<string, WebSocketEventHandler[]> = new Map()
  private userID: string | null = null

  private getWebSocketUrl(): string {
    const baseUrl =
      import.meta.env.VITE_API_BASE_URL || 'http://localhost:3002/api'
    const wsUrl = baseUrl.replace(/^http/, 'ws')
    return `${wsUrl}/ws`
  }

  async connect(): Promise<void> {
    if (
      this.isConnecting ||
      (this.ws && this.ws.readyState === WebSocket.OPEN)
    ) {
      return
    }

    this.isConnecting = true

    try {
      // For this project, we'll use a simple user ID or generate one
      // In a real app, you would get this from your auth system
      this.userID = 'user_' + Math.random().toString(36).substr(2, 9)

      const wsUrl = this.getWebSocketUrl()
      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        this.reconnectAttempts = 0
        this.isConnecting = false
        this.emit('connected', { connected: true })

        // Subscribe to channels
        this.send({
          type: 'subscribe',
          data: {
            channels: [
              'alerts',
              'people_count',
              'vehicle_count',
              'face_recognition',
            ],
          },
        })
      }

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          if (typeof window !== 'undefined') {
            ;(window as any).lastWSMessage = message
          }
          this.handleMessage(message)
        } catch (_error) {
          // Message parsing failed - continue silently
        }
      }

      this.ws.onclose = (event) => {
        this.isConnecting = false
        this.emit('disconnected', { connected: false })

        // Attempt to reconnect if not a clean close
        if (
          event.code !== 1000 &&
          this.reconnectAttempts < this.maxReconnectAttempts
        ) {
          this.scheduleReconnect()
        }
      }

      this.ws.onerror = (error) => {
        this.isConnecting = false
        this.emit('error', { error })
      }
    } catch (_error) {
      this.isConnecting = false
      this.scheduleReconnect()
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      return
    }

    this.reconnectAttempts++
    const delay = 1000 * Math.pow(2, this.reconnectAttempts - 1)

    setTimeout(() => {
      this.connect()
    }, delay)
  }

  private handleMessage(message: WebSocketMessage): void {
    // Temporary debug logging
    if (typeof window !== 'undefined') {
      ;(window as any).wsMessageCount =
        ((window as any).wsMessageCount || 0) + 1
      ;(window as any).wsMessages = [
        ...((window as any).wsMessages || []).slice(-9),
        message,
      ]
    }

    // Handle ping/pong
    if (message.type === 'ping') {
      this.send({
        type: 'pong',
        data: { timestamp: new Date().toISOString() },
      })
      return
    }

    // Emit to specific event handlers
    // Pass the data field if it exists, otherwise pass the whole message
    const eventData = message.data !== undefined ? message.data : message
    this.emit(message.type, eventData)

    // Also emit to general message handlers
    this.emit('message', message)
  }

  private emit(eventType: string, data: any): void {
    const handlers = this.eventHandlers.get(eventType) || []

    handlers.forEach((handler) => {
      try {
        handler({ type: eventType, data })
      } catch (_error) {
        // Handler error - continue silently
      }
    })
  }

  on(eventType: string, handler: WebSocketEventHandler): void {
    if (!this.eventHandlers.has(eventType)) {
      this.eventHandlers.set(eventType, [])
    }
    this.eventHandlers.get(eventType)!.push(handler)
  }

  off(eventType: string, handler: WebSocketEventHandler): void {
    const handlers = this.eventHandlers.get(eventType)
    if (handlers) {
      const index = handlers.indexOf(handler)
      if (index > -1) {
        handlers.splice(index, 1)
      }
    }
  }

  send(data: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    }
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close(1000, 'Client disconnecting')
      this.ws = null
    }
    this.userID = null
    this.reconnectAttempts = 0
    this.eventHandlers.clear()
  }

  isConnected(): boolean {
    return this.ws ? this.ws.readyState === WebSocket.OPEN : false
  }

  getConnectionInfo(): any {
    return {
      connected: this.isConnected(),
      readyState: this.ws?.readyState,
      url: this.ws?.url,
      userID: this.userID,
      handlerCounts: Object.fromEntries(
        Array.from(this.eventHandlers.entries()).map(([key, handlers]) => [
          key,
          handlers.length,
        ])
      ),
    }
  }
}

// Create a singleton instance
export const webSocketService = new WebSocketService()

// Auto-connect when the service is imported
if (typeof window !== 'undefined') {
  webSocketService.connect()
}

export default webSocketService

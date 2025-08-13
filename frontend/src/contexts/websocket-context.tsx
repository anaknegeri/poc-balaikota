import { createContext, ReactNode, useContext } from 'react'
import { useWebSocketConnection } from '@/hooks/use-websocket'

interface WebSocketContextType {
  isConnected: boolean
  connect: () => Promise<void>
  disconnect: () => void
  send: (message: any) => void
  on: (eventType: string, handler: (event: any) => void) => void
  off: (eventType: string, handler: (event: any) => void) => void
  recentAlerts: any[]
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(
  undefined
)

interface WebSocketProviderProps {
  children: ReactNode
}

export function WebSocketProvider({ children }: WebSocketProviderProps) {
  const { isConnected, connect, disconnect, send, on, off, recentAlerts } =
    useWebSocketConnection()

  const contextValue: WebSocketContextType = {
    isConnected,
    connect,
    disconnect,
    send,
    on,
    off,
    recentAlerts: recentAlerts || [],
  }

  return (
    <WebSocketContext.Provider value={contextValue}>
      {children}
    </WebSocketContext.Provider>
  )
}

export function useWebSocketContext() {
  const context = useContext(WebSocketContext)
  if (context === undefined) {
    throw new Error(
      'useWebSocketContext must be used within a WebSocketProvider'
    )
  }
  return context
}

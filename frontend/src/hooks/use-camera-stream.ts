import { useEffect, useRef, useState } from 'react'

interface UseCameraStreamProps {
  wsUrl?: string
  fallbackImageUrl?: string
  cameraId?: number
  enabled?: boolean
}

interface CameraStreamState {
  imageSrc: string | null
  isConnected: boolean
  isConnecting: boolean
  error: string | null
  usingWebSocket: boolean
}

export function useCameraStream({
  wsUrl,
  fallbackImageUrl,
  cameraId,
  enabled = true,
}: UseCameraStreamProps): CameraStreamState {
  const [state, setState] = useState<CameraStreamState>({
    imageSrc: fallbackImageUrl || null, // Start with fallback image immediately
    isConnected: false,
    isConnecting: false,
    error: null,
    usingWebSocket: false,
  })

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const reconnectAttempts = useRef(0)

  const connectWebSocket = () => {
    if (!wsUrl || !enabled) {
      // Use fallback image if no WebSocket URL or disabled
      setState((prev) => ({
        ...prev,
        imageSrc: fallbackImageUrl || null,
        isConnected: false,
        isConnecting: false,
        error: null,
        usingWebSocket: false,
      }))
      return
    }

    setState((prev) => ({
      ...prev,
      isConnecting: true,
      error: null,
    }))

    try {
      const ws = new WebSocket(wsUrl)
      wsRef.current = ws

      ws.onopen = () => {
        setState((prev) => ({
          ...prev,
          isConnected: true,
          isConnecting: false,
          error: null,
          usingWebSocket: true,
        }))
        reconnectAttempts.current = 0
      }

      ws.onmessage = (event) => {
        try {
          // Parse JSON response: {"image": "jpg_as_text", "camera_id": 1}
          const data = JSON.parse(event.data)

          // if (
          //   cameraId !== undefined &&
          //   data.camera_id !== undefined &&
          //   data.camera_id !== cameraId - 1
          // ) {
          //   return
          // }

          if (data.image === undefined) {
            console.warn('Received WebSocket message without image field')
            return
          }

          const imageSrc = `data:image/jpeg;base64,${data.image}`
          setState((prev) => ({
            ...prev,
            imageSrc,
            usingWebSocket: true,
          }))
        } catch (error) {
          console.error('Error parsing WebSocket message:', error)
          // Try using event.data directly as fallback
          const imageSrc = `data:image/jpeg;base64,${event.data}`
          setState((prev) => ({
            ...prev,
            imageSrc,
            usingWebSocket: true,
          }))
        }
      }

      ws.onclose = () => {
        setState((prev) => ({
          ...prev,
          isConnected: false,
          usingWebSocket: false,
          imageSrc: fallbackImageUrl || '',
          error: 'Connection closed',
        }))
      }

      ws.onerror = (error) => {
        console.error('Camera WebSocket error:', error)
        setState((prev) => ({
          ...prev,
          error: 'WebSocket connection error',
          imageSrc: fallbackImageUrl || '',
        }))
      }
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error)
      setState((prev) => ({
        ...prev,
        isConnecting: false,
        error: 'Failed to connect to WebSocket',
        imageSrc: fallbackImageUrl || '',
      }))
    }
  }

  const cleanup = () => {
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }
    reconnectAttempts.current = 0
  }

  useEffect(() => {
    if (enabled) {
      connectWebSocket()
    } else {
      cleanup()
      setState((prev) => ({
        ...prev,
        imageSrc: fallbackImageUrl || null,
        isConnected: false,
        isConnecting: false,
        error: null,
        usingWebSocket: false,
      }))
    }

    return cleanup
  }, [wsUrl, fallbackImageUrl, cameraId, enabled])

  return state
}

import { useCallback, useEffect, useRef, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { NotificationService, type Notification } from '@/services/notification'
import { webSocketService } from '@/services/websocket-service'

export function useWebSocketConnection() {
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout>>(null)
  const [recentAlerts, setRecentAlerts] = useState<any[]>([])

  const connect = useCallback(async () => {
    await webSocketService.connect()
  }, [])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    webSocketService.disconnect()
  }, [])

  useEffect(() => {
    connect()

    // Listen for alert events
    const handleAlert = (alertData: any) => {
      setRecentAlerts(prev => {
        const newAlert = {
          ...alertData,
          timestamp: alertData.timestamp || new Date().toISOString(),
          id: alertData.id || Date.now().toString()
        }
        // Keep only last 10 alerts
        return [newAlert, ...prev.slice(0, 9)]
      })
    }

    webSocketService.on('alert', handleAlert)

    return () => {
      webSocketService.off('alert', handleAlert)
      disconnect()
    }
  }, [connect, disconnect])

  return {
    connect,
    disconnect,
    isConnected: webSocketService.isConnected(),
    send: webSocketService.send.bind(webSocketService),
    on: webSocketService.on.bind(webSocketService),
    off: webSocketService.off.bind(webSocketService),
    recentAlerts,
  }
}

export function useWebSocketEvent<T = any>(
  eventType: string,
  handler: (data: T) => void
) {
  const handlerRef = useRef(handler)
  handlerRef.current = handler

  useEffect(() => {
    const wrappedHandler = (data: T) => {
      handlerRef.current(data)
    }

    webSocketService.on(eventType, wrappedHandler)

    return () => {
      webSocketService.off(eventType, wrappedHandler)
    }
  }, [eventType])
}

// Job progress tracking
export function useJobProgress(jobId?: string) {
  const [progress, setProgress] = useState<{
    progress: number
    message: string
    status?: 'running' | 'completed' | 'failed'
  }>({ progress: 0, message: '' })

  useWebSocketEvent(
    'job_progress',
    useCallback(
      (data: any) => {
        if (!jobId || data.job_id === jobId) {
          setProgress({
            progress: data.progress || 0,
            message: data.message || '',
            status: 'running',
          })
        }
      },
      [jobId]
    )
  )

  useWebSocketEvent(
    'job_completed',
    useCallback(
      (data: any) => {
        if (!jobId || data.job_id === jobId) {
          setProgress((prev) => ({
            ...prev,
            progress: 100,
            message: data.message || 'Job completed',
            status: data.status || 'completed',
          }))
        }
      },
      [jobId]
    )
  )

  return progress
}

// Notification handling
export function useNotifications() {
  const queryClient = useQueryClient()
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [unreadCount, setUnreadCount] = useState(0)

  // Fetch initial notifications from API
  const {
    data: notificationData,
    isSuccess: notificationSuccess,
    isLoading: notificationLoading,
    error: _notificationError,
  } = useQuery({
    queryKey: ['notifications'],
    queryFn: () => NotificationService.getNotifications({ per_page: 50 }),
    staleTime: 1000 * 60 * 5, // 5 minutes
    refetchOnWindowFocus: false,
  })

  // Fetch unread count
  const {
    data: unreadCountData,
    isSuccess: unreadCountSuccess,
    isLoading: unreadCountLoading,
    error: _unreadCountError,
  } = useQuery({
    queryKey: ['notifications', 'unread-count'],
    queryFn: () => NotificationService.getUnreadCount(),
    staleTime: 1000 * 60, // 1 minute
    refetchOnWindowFocus: false,
  })

  // Set initial data from API
  useEffect(() => {
    if (notificationSuccess && notificationData) {
      // Handle different response formats from backend
      if (notificationData.data && Array.isArray(notificationData.data)) {
        setNotifications(notificationData.data)
      } else if (Array.isArray(notificationData)) {
        setNotifications(notificationData)
      } else {
        setNotifications([])
      }
    }
  }, [notificationData, notificationSuccess])

  useEffect(() => {
    if (unreadCountSuccess && unreadCountData) {
      // Handle different response formats
      if (typeof unreadCountData === 'number') {
        setUnreadCount(unreadCountData)
      } else if (typeof (unreadCountData as any).count === 'number') {
        setUnreadCount((unreadCountData as any).count)
      } else if (typeof (unreadCountData as any).data?.count === 'number') {
        setUnreadCount((unreadCountData as any).data.count)
      } else {
        setUnreadCount(0)
      }
    }
  }, [unreadCountData, unreadCountSuccess])

  // Listen for real-time notifications via WebSocket
  useWebSocketEvent(
    'notification',
    useCallback(
      (data: any) => {
        const newNotification: Notification = {
          id: data.id || Date.now().toString(),
          type: data.type || 'system_alert',
          title: data.title || 'New Notification',
          message: data.message || '',
          data: data.data || data,
          entity_type: data.entity_type || null,
          entity_id: data.entity_id || null,
          channels: data.channels || ['web'],
          priority: data.priority || 'normal',
          status: 'unread',
          read_at: null,
          user_id: 'anonymous',
          created_at: data.timestamp || new Date().toISOString(),
        }

        // Add to local state (avoid duplicates)
        setNotifications((prev) => {
          const exists = prev.some((n) => n.id === newNotification.id)
          if (exists) return prev
          return [newNotification, ...prev]
        })

        // Update unread count
        setUnreadCount((prev) => prev + 1)

        // Invalidate queries to refresh data
        queryClient.invalidateQueries({ queryKey: ['notifications'] })
        queryClient.invalidateQueries({
          queryKey: ['notifications', 'unread-count'],
        })
      },
      [queryClient]
    )
  )

  const markAsRead = useCallback(
    async (notificationId: string) => {
      try {
        await NotificationService.markAsRead(notificationId)

        // Update local state
        setNotifications((prev) =>
          prev.map((n) =>
            n.id === notificationId
              ? { ...n, status: 'read', read_at: new Date().toISOString() }
              : n
          )
        )

        // Update unread count
        setUnreadCount((prev) => Math.max(0, prev - 1))

        // Invalidate queries
        queryClient.invalidateQueries({
          queryKey: ['notifications', 'unread-count'],
        })
      } catch (error) {
        console.error('Failed to mark notification as read:', error)
      }
    },
    [queryClient]
  )

  const markAllAsRead = useCallback(async () => {
    try {
      await NotificationService.markAllAsRead()

      // Update local state
      setNotifications((prev) =>
        prev.map((n) => ({
          ...n,
          status: 'read',
          read_at: n.read_at || new Date().toISOString(),
        }))
      )

      // Reset unread count
      setUnreadCount(0)

      // Invalidate queries
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({
        queryKey: ['notifications', 'unread-count'],
      })
    } catch (error) {
      console.error('Failed to mark all notifications as read:', error)
    }
  }, [queryClient])

  return {
    notifications,
    unreadCount,
    isLoading: notificationLoading || unreadCountLoading,
    markAsRead,
    markAllAsRead,
    refetch: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({
        queryKey: ['notifications', 'unread-count'],
      })
    },
  }
}

export default useWebSocketConnection

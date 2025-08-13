export interface Notification {
  id: string
  type: string
  title: string
  message: string
  data?: any
  entity_type?: string | null
  entity_id?: string | null
  channels: string[]
  priority: 'low' | 'normal' | 'high' | 'urgent'
  status: 'unread' | 'read'
  read_at?: string | null
  user_id: string
  created_at: string
}

export interface NotificationFilters {
  per_page?: number
  page?: number
  type?: string
  status?: 'unread' | 'read'
  priority?: 'low' | 'normal' | 'high' | 'urgent'
}

export class NotificationService {
  private static baseUrl = '/api/notifications'

  static async getNotifications(filters: NotificationFilters = {}): Promise<{
    data: Notification[]
    total: number
    pages: number
  }> {
    const params = new URLSearchParams()
    
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined) {
        params.append(key, value.toString())
      }
    })

    const response = await fetch(`${this.baseUrl}?${params}`)
    if (!response.ok) {
      throw new Error(`Failed to fetch notifications: ${response.statusText}`)
    }

    return response.json()
  }

  static async getUnreadCount(): Promise<{ count: number }> {
    const response = await fetch(`${this.baseUrl}/unread-count`)
    if (!response.ok) {
      throw new Error(`Failed to fetch unread count: ${response.statusText}`)
    }

    return response.json()
  }

  static async markAsRead(notificationId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/${notificationId}/read`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      throw new Error(`Failed to mark notification as read: ${response.statusText}`)
    }
  }

  static async markAllAsRead(): Promise<void> {
    const response = await fetch(`${this.baseUrl}/mark-all-read`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      throw new Error(`Failed to mark all notifications as read: ${response.statusText}`)
    }
  }

  static async deleteNotification(notificationId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/${notificationId}`, {
      method: 'DELETE',
    })

    if (!response.ok) {
      throw new Error(`Failed to delete notification: ${response.statusText}`)
    }
  }
}
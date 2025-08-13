import { config } from '@/config/config'
import type {
  Alert,
  DateRange,
  FaceRecognitionData,
  PeakHoursAnalysis,
  PeopleCountSummary,
  PeopleCountTrend,
  VehicleCountSummary,
  VehicleCountTrend,
} from '@/types/dashboard'
import type { ApiResponse } from '@/types/response'

// Helper function to build query parameters with date range
const buildQueryParams = (
  params: Record<string, any>,
  dateRange?: DateRange
): string => {
  const queryParams = new URLSearchParams()

  // Add standard parameters
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null) {
      queryParams.append(key, value.toString())
    }
  })

  // Add date range parameters if provided
  if (dateRange) {
    if (dateRange.from) {
      queryParams.append('from', dateRange.from)
    }
    if (dateRange.to) {
      queryParams.append('to', dateRange.to)
    }
  }

  const queryString = queryParams.toString()
  return queryString ? `?${queryString}` : ''
}

// API Functions
export const dashboardService = {
  // People Count APIs
  async getPeopleCountSummary(
    dateRange?: DateRange
  ): Promise<PeopleCountSummary> {
    const queryParams = buildQueryParams({}, dateRange)
    const response = await fetch(
      `${config.API_BASE_URL}/counts/summary${queryParams}`
    )
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    const result: ApiResponse<PeopleCountSummary> = await response.json()
    return result.data
  },

  async getPeopleCountTrends(
    interval: 'hour' | 'day' | 'week' | 'month' = 'hour',
    limit: number = 24,
    dateRange?: DateRange
  ): Promise<PeopleCountTrend[]> {
    const queryParams = buildQueryParams({ interval, limit }, dateRange)
    const response = await fetch(
      `${config.API_BASE_URL}/counts/trends${queryParams}`
    )
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    const result: ApiResponse<{ interval: string; data: PeopleCountTrend[] }> =
      await response.json()
    return result.data.data
  },

  // Peak Hours Analysis API with date range support
  async getPeakHoursAnalysis(
    cameraId?: string,
    timeWindow: '24h' | '7d' | '30d' | 'all' = '24h',
    dateRange?: DateRange
  ): Promise<PeakHoursAnalysis> {
    const params: Record<string, any> = { window: timeWindow }
    if (cameraId) {
      params.camera_id = cameraId
    }

    const queryParams = buildQueryParams(params, dateRange)
    const response = await fetch(
      `${config.API_BASE_URL}/counts/peak-hours${queryParams}`
    )
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    const result: ApiResponse<PeakHoursAnalysis> = await response.json()
    return result.data
  },

  // Vehicle Count APIs
  async getVehicleCountSummary(
    dateRange?: DateRange
  ): Promise<VehicleCountSummary> {
    const queryParams = buildQueryParams({}, dateRange)
    const response = await fetch(
      `${config.API_BASE_URL}/vehicles/summary${queryParams}`
    )
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    const result: ApiResponse<VehicleCountSummary> = await response.json()
    return result.data
  },

  async getVehicleCountTrends(
    interval: 'hour' | 'day' | 'week' | 'month' = 'hour',
    limit: number = 24,
    dateRange?: DateRange
  ): Promise<VehicleCountTrend[]> {
    const queryParams = buildQueryParams({ interval, limit }, dateRange)
    const response = await fetch(
      `${config.API_BASE_URL}/vehicles/trends${queryParams}`
    )
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    const result: ApiResponse<{ interval: string; data: VehicleCountTrend[] }> =
      await response.json()
    return result.data.data
  },

  // Alert APIs
  async getActiveAlerts(dateRange?: DateRange): Promise<{
    count: number
    data: Alert[]
    total: number
    pages: number
  }> {
    const queryParams = buildQueryParams({ include_relations: true }, dateRange)
    const response = await fetch(
      `${config.API_BASE_URL}/alerts/active${queryParams}`
    )
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    const result: ApiResponse<Alert[]> = await response.json()
    return {
      count: result.count_all_active || 0,
      data: result.data,
      total: result.total || 0,
      pages: result.pages || 1,
    }
  },

  // Face Recognition APIs
  async getFaceRecognitions(
    page: number = 1,
    limit: number = 10,
    dateRange?: DateRange
  ): Promise<FaceRecognitionData> {
    const queryParams = buildQueryParams(
      {
        page,
        limit,
      },
      dateRange
    )
    const response = await fetch(
      `${config.API_BASE_URL}/recognitions${queryParams}`
    )
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    const result: FaceRecognitionData = await response.json()
    return result
  },

  async getAlerts(
    page: number = 1,
    limit: number = 10,
    includeRelations: boolean = true,
    dateRange?: DateRange,
    filters?: {
      search?: string
      alert_type_id?: number
      severity?: string
      status?: string
      camera_id?: number
    }
  ): Promise<{ alerts: Alert[]; total: number; pages: number }> {
    const params: Record<string, any> = {
      page,
      limit,
      include_relations: includeRelations,
    }

    // Add filter parameters if provided
    if (filters) {
      if (filters.search) params.search = filters.search
      if (filters.alert_type_id) params.alert_type_id = filters.alert_type_id
      if (filters.severity) params.severity = filters.severity
      if (filters.status) params.status = filters.status
      if (filters.camera_id) params.camera_id = filters.camera_id
    }

    const queryParams = buildQueryParams(params, dateRange)
    const response = await fetch(`${config.API_BASE_URL}/alerts${queryParams}`)
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    const result: ApiResponse<Alert[]> = await response.json()
    return {
      alerts: result.data,
      total: result.total || 0,
      pages: result.pages || 1,
    }
  },

  // Alert Types API
  async getAlertTypes(): Promise<
    {
      id: number
      name: string
      display_name: string
      icon: string
      color: string
      description: string
    }[]
  > {
    const response = await fetch(`${config.API_BASE_URL}/alerts/types`)
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    const result: ApiResponse<
      {
        id: number
        name: string
        display_name: string
        icon: string
        color: string
        description: string
      }[]
    > = await response.json()
    return result.data
  },

  // Health Check
  async healthCheck(): Promise<{ status: string; time: string }> {
    const response = await fetch(`${config.API_BASE_URL}/health`)
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    return response.json()
  },
}

// Utility functions
export const formatDateTime = (dateString: string): string => {
  return new Date(dateString).toLocaleString('id-ID', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export const getTimeAgo = (dateString: string): string => {
  try {
    // Pastikan kita mendapatkan waktu WIB saat ini
    const now = new Date()

    // Parse date dari string yang sudah pasti WIB format
    const date = new Date(dateString)

    // Validasi parsing
    if (isNaN(date.getTime())) {
      return 'Waktu tidak valid'
    }

    // Hitung selisih dalam milidetik, lalu convert ke detik
    const diffInMs = now.getTime() - date.getTime()
    const diffInSeconds = Math.floor(diffInMs / 1000)

    // Handle future dates (negatif) - treat as "baru saja"
    if (diffInSeconds < 0) {
      return 'Baru saja'
    }

    // Handle waktu yang sangat baru
    if (diffInSeconds < 10) {
      return 'Baru saja'
    }

    if (diffInSeconds < 60) {
      return `${diffInSeconds} detik yang lalu`
    } else if (diffInSeconds < 3600) {
      const minutes = Math.floor(diffInSeconds / 60)
      return `${minutes} menit yang lalu`
    } else if (diffInSeconds < 86400) {
      const hours = Math.floor(diffInSeconds / 3600)
      return `${hours} jam yang lalu`
    } else {
      const days = Math.floor(diffInSeconds / 86400)
      return `${days} hari yang lalu`
    }
  } catch (_error) {
    return 'Waktu tidak valid'
  }
}

export const getSeverityColor = (severity: string): string => {
  switch (severity.toLowerCase()) {
    case 'critical':
      return 'bg-red-500'
    case 'high':
      return 'bg-orange-500'
    case 'medium':
      return 'bg-yellow-500'
    case 'low':
      return 'bg-blue-500'
    default:
      return 'bg-gray-500'
  }
}

export const getSeverityTextColor = (severity: string): string => {
  switch (severity.toLowerCase()) {
    case 'critical':
      return 'text-red-600 dark:text-red-400'
    case 'high':
      return 'text-orange-600 dark:text-orange-400'
    case 'medium':
      return 'text-yellow-600 dark:text-yellow-400'
    case 'low':
      return 'text-blue-600 dark:text-blue-400'
    default:
      return 'text-gray-600 dark:text-gray-400'
  }
}

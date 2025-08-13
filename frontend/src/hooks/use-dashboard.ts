// hooks/use-dashboard-data.ts
import { useCallback, useEffect, useRef, useState } from 'react'
import { config } from '@/config/config'
import { dashboardService } from '@/services/dashboardService'
import type {
  Alert,
  FaceRecognitionData,
  PeakHoursAnalysis,
  PeopleCountSummary,
  PeopleCountTrend,
  VehicleCountSummary,
  VehicleCountTrend,
} from '@/types/dashboard'

// Date range interface
export interface DateRange {
  from: string
  to: string
}

export interface DashboardData {
  peopleCountSummary: PeopleCountSummary | null
  vehicleCountSummary: VehicleCountSummary | null
  faceRecognitions: FaceRecognitionData | null
  peopleCountTrends: PeopleCountTrend[]
  vehicleCountTrends: VehicleCountTrend[]
  peakHoursAnalysis: PeakHoursAnalysis | null
  activeAlert: {
    count: number
    data: Alert[]
    total: number
    pages: number
  } | null
}

export interface DashboardState {
  data: DashboardData
  loading: boolean
  error: string | null
  lastUpdated: Date | null
}

export const useDashboardData = (
  refreshInterval?: number,
  dateRange?: DateRange
) => {
  const actualRefreshInterval =
    refreshInterval ?? config.REFRESH_INTERVAL.DASHBOARD
  const isInitialLoadRef = useRef(true)

  const [state, setState] = useState<DashboardState>({
    data: {
      peopleCountSummary: null,
      vehicleCountSummary: null,
      faceRecognitions: null,
      peopleCountTrends: [],
      vehicleCountTrends: [],
      peakHoursAnalysis: null,
      activeAlert: null,
    },
    loading: true,
    error: null,
    lastUpdated: null,
  })

  const fetchDashboardData = useCallback(
    async (showLoading = false) => {
      try {
        // Show loading only on initial load or manual refresh
        if (showLoading || isInitialLoadRef.current) {
          setState((prev) => ({ ...prev, loading: true, error: null }))
        }

        // Fetch all data in parallel - including peak hours with date range
        const [
          peopleCountSummary,
          vehicleCountSummary,
          faceRecognitions,
          peopleCountTrends,
          vehicleCountTrends,
          peakHoursAnalysis,
          activeAlert,
        ] = await Promise.all([
          dashboardService.getPeopleCountSummary(dateRange),
          dashboardService.getVehicleCountSummary(dateRange),
          dashboardService.getFaceRecognitions(
            1,
            config.CHART_LIMITS.ALERTS,
            dateRange
          ),
          dashboardService.getPeopleCountTrends(
            'hour',
            config.CHART_LIMITS.TRENDS,
            dateRange
          ),
          dashboardService.getVehicleCountTrends(
            'hour',
            config.CHART_LIMITS.TRENDS,
            dateRange
          ),
          dashboardService.getPeakHoursAnalysis(undefined, '24h', dateRange),
          dashboardService.getActiveAlerts(dateRange),
        ])

        setState({
          data: {
            peopleCountSummary,
            vehicleCountSummary,
            faceRecognitions,
            peopleCountTrends,
            vehicleCountTrends,
            peakHoursAnalysis,
            activeAlert,
          },
          loading: false,
          error: null,
          lastUpdated: new Date(),
        })

        // Mark initial load as complete
        if (isInitialLoadRef.current) {
          isInitialLoadRef.current = false
        }
      } catch (error) {
        console.error('Error fetching dashboard data:', error)
        setState((prev) => ({
          ...prev,
          loading: false,
          error:
            error instanceof Error ? error.message : 'Unknown error occurred',
        }))

        // Mark initial load as complete even on error
        if (isInitialLoadRef.current) {
          isInitialLoadRef.current = false
        }
      }
    },
    [dateRange]
  )

  // Manual refresh function that shows loading
  const manualRefetch = useCallback(() => {
    fetchDashboardData(true)
  }, [fetchDashboardData])

  // Initial fetch and refetch when date range changes
  useEffect(() => {
    // Reset initial load flag when date range changes
    isInitialLoadRef.current = true
    fetchDashboardData(false)
  }, [fetchDashboardData])

  // Auto refresh (silent, no loading indicator)
  useEffect(() => {
    if (actualRefreshInterval > 0) {
      const interval = setInterval(() => {
        // Only do silent refresh if not initial load
        if (!isInitialLoadRef.current) {
          fetchDashboardData(false)
        }
      }, actualRefreshInterval)

      return () => clearInterval(interval)
    }
  }, [fetchDashboardData, actualRefreshInterval])

  return {
    ...state,
    refetch: manualRefetch,
    isInitialLoad: isInitialLoadRef.current,
  }
}

// Hook specifically for Peak Hours Analysis with date range support
export const usePeakHoursAnalysis = (
  cameraId?: string,
  timeWindow: '24h' | '7d' | '30d' | 'all' = '24h',
  refreshInterval?: number,
  dateRange?: DateRange
) => {
  const actualRefreshInterval =
    refreshInterval ?? config.REFRESH_INTERVAL.CHARTS
  const isInitialLoadRef = useRef(true)

  const [peakHours, setPeakHours] = useState<PeakHoursAnalysis | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchPeakHours = useCallback(
    async (showLoading = false) => {
      try {
        if (showLoading || isInitialLoadRef.current) {
          setLoading(true)
        }
        setError(null)

        const data = await dashboardService.getPeakHoursAnalysis(
          cameraId,
          timeWindow,
          dateRange
        )
        setPeakHours(data)

        if (isInitialLoadRef.current) {
          isInitialLoadRef.current = false
        }
      } catch (error) {
        setError(
          error instanceof Error ? error.message : 'Unknown error occurred'
        )
        if (isInitialLoadRef.current) {
          isInitialLoadRef.current = false
        }
      } finally {
        if (showLoading || isInitialLoadRef.current) {
          setLoading(false)
        }
      }
    },
    [cameraId, timeWindow, dateRange]
  )

  // Initial fetch and refetch when parameters change
  useEffect(() => {
    isInitialLoadRef.current = true
    fetchPeakHours(false)
  }, [fetchPeakHours])

  // Auto refresh for peak hours (silent)
  useEffect(() => {
    if (actualRefreshInterval > 0) {
      const intervalId = setInterval(() => {
        if (!isInitialLoadRef.current) {
          fetchPeakHours(false)
        }
      }, actualRefreshInterval)
      return () => clearInterval(intervalId)
    }
  }, [fetchPeakHours, actualRefreshInterval])

  return {
    peakHours,
    loading: loading && isInitialLoadRef.current,
    error,
    refetch: () => fetchPeakHours(true),
  }
}

// Hook for people count trends with date range support
export const usePeopleCountTrends = (
  interval: 'hour' | 'day' | 'week' | 'month' = 'hour',
  dateRange?: DateRange
) => {
  const [trends, setTrends] = useState<PeopleCountTrend[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const isInitialLoadRef = useRef(true)

  const fetchTrends = useCallback(
    async (showLoading = false) => {
      try {
        if (showLoading || isInitialLoadRef.current) {
          setLoading(true)
        }
        setError(null)

        const data = await dashboardService.getPeopleCountTrends(
          interval,
          config.CHART_LIMITS.TRENDS,
          dateRange
        )
        setTrends(data)

        if (isInitialLoadRef.current) {
          isInitialLoadRef.current = false
        }
      } catch (error) {
        setError(
          error instanceof Error ? error.message : 'Unknown error occurred'
        )
        if (isInitialLoadRef.current) {
          isInitialLoadRef.current = false
        }
      } finally {
        if (showLoading || isInitialLoadRef.current) {
          setLoading(false)
        }
      }
    },
    [interval, dateRange]
  )

  // Initial fetch and refetch when parameters change
  useEffect(() => {
    isInitialLoadRef.current = true
    fetchTrends(false)
  }, [fetchTrends])

  // Auto refresh for trends (silent)
  useEffect(() => {
    const interval_ms = config.REFRESH_INTERVAL.CHARTS
    if (interval_ms > 0) {
      const intervalId = setInterval(() => {
        if (!isInitialLoadRef.current) {
          fetchTrends(false)
        }
      }, interval_ms)
      return () => clearInterval(intervalId)
    }
  }, [fetchTrends])

  return {
    trends,
    loading: loading && isInitialLoadRef.current,
    error,
    refetch: () => fetchTrends(true),
  }
}

// Hook for vehicle count trends with date range support
export const useVehicleCountTrends = (
  interval: 'hour' | 'day' | 'week' | 'month' = 'hour',
  dateRange?: DateRange
) => {
  const [trends, setTrends] = useState<VehicleCountTrend[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const isInitialLoadRef = useRef(true)

  const fetchTrends = useCallback(
    async (showLoading = false) => {
      try {
        if (showLoading || isInitialLoadRef.current) {
          setLoading(true)
        }
        setError(null)

        const data = await dashboardService.getVehicleCountTrends(
          interval,
          config.CHART_LIMITS.TRENDS,
          dateRange
        )
        setTrends(data)

        if (isInitialLoadRef.current) {
          isInitialLoadRef.current = false
        }
      } catch (error) {
        setError(
          error instanceof Error ? error.message : 'Unknown error occurred'
        )
        if (isInitialLoadRef.current) {
          isInitialLoadRef.current = false
        }
      } finally {
        if (showLoading || isInitialLoadRef.current) {
          setLoading(false)
        }
      }
    },
    [interval, dateRange]
  )

  // Initial fetch and refetch when parameters change
  useEffect(() => {
    isInitialLoadRef.current = true
    fetchTrends(false)
  }, [fetchTrends])

  // Auto refresh for trends (silent)
  useEffect(() => {
    const interval_ms = config.REFRESH_INTERVAL.CHARTS
    if (interval_ms > 0) {
      const intervalId = setInterval(() => {
        if (!isInitialLoadRef.current) {
          fetchTrends(false)
        }
      }, interval_ms)
      return () => clearInterval(intervalId)
    }
  }, [fetchTrends])

  return {
    trends,
    loading: loading && isInitialLoadRef.current,
    error,
    refetch: () => fetchTrends(true),
  }
}

// Hook for active alerts with date range support
export const useActiveAlerts = (
  refreshInterval?: number,
  dateRange?: DateRange
) => {
  const actualRefreshInterval =
    refreshInterval ?? config.REFRESH_INTERVAL.ALERTS
  const isInitialLoadRef = useRef(true)

  const [alerts, setAlerts] = useState<Alert[] | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchAlerts = useCallback(
    async (showLoading = false) => {
      try {
        if (showLoading || isInitialLoadRef.current) {
          setLoading(true)
        }
        setError(null)

        const data = await dashboardService.getActiveAlerts(dateRange)
        setAlerts(data.data)

        if (isInitialLoadRef.current) {
          isInitialLoadRef.current = false
        }
      } catch (error) {
        setError(
          error instanceof Error ? error.message : 'Unknown error occurred'
        )
        if (isInitialLoadRef.current) {
          isInitialLoadRef.current = false
        }
      } finally {
        if (showLoading || isInitialLoadRef.current) {
          setLoading(false)
        }
      }
    },
    [dateRange]
  )

  // Initial fetch and refetch when date range changes
  useEffect(() => {
    isInitialLoadRef.current = true
    fetchAlerts(false)
  }, [fetchAlerts])

  // Auto refresh (silent)
  useEffect(() => {
    if (actualRefreshInterval > 0) {
      const interval = setInterval(() => {
        if (!isInitialLoadRef.current) {
          fetchAlerts(false)
        }
      }, actualRefreshInterval)
      return () => clearInterval(interval)
    }
  }, [fetchAlerts, actualRefreshInterval])

  return {
    alerts,
    loading: loading && isInitialLoadRef.current,
    error,
    refetch: () => fetchAlerts(true),
  }
}

// Hook for alerts with filtering support
export const useAlerts = (
  page: number = 1,
  limit: number = 10,
  filters?: {
    search?: string
    alert_type_id?: number
    severity?: string
    status?: string
    camera_id?: number
  },
  refreshInterval?: number,
  dateRange?: DateRange
) => {
  const actualRefreshInterval =
    refreshInterval ?? config.REFRESH_INTERVAL.ALERTS
  const isInitialLoadRef = useRef(true)

  const [alerts, setAlerts] = useState<Alert[]>([])
  const [total, setTotal] = useState(0)
  const [pages, setPages] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchAlerts = useCallback(
    async (showLoading = false) => {
      try {
        if (showLoading || isInitialLoadRef.current) {
          setLoading(true)
        }
        setError(null)

        const data = await dashboardService.getAlerts(
          page,
          limit,
          true,
          dateRange,
          filters
        )

        setAlerts(data.alerts ?? [])
        setTotal(data.total)
        setPages(data.pages)

        if (isInitialLoadRef.current) {
          isInitialLoadRef.current = false
        }
      } catch (error) {
        setError(
          error instanceof Error ? error.message : 'Unknown error occurred'
        )
        if (isInitialLoadRef.current) {
          isInitialLoadRef.current = false
        }
      } finally {
        if (showLoading || isInitialLoadRef.current) {
          setLoading(false)
        }
      }
    },
    [page, limit, filters, dateRange]
  )

  // Initial fetch and refetch when parameters change
  useEffect(() => {
    isInitialLoadRef.current = true
    fetchAlerts(false)
  }, [fetchAlerts])

  // Auto refresh (silent)
  useEffect(() => {
    if (actualRefreshInterval > 0) {
      const interval = setInterval(() => {
        if (!isInitialLoadRef.current) {
          fetchAlerts(false)
        }
      }, actualRefreshInterval)
      return () => clearInterval(interval)
    }
  }, [fetchAlerts, actualRefreshInterval])

  return {
    alerts,
    total,
    pages,
    loading: loading && isInitialLoadRef.current,
    error,
    refetch: () => fetchAlerts(true),
  }
}

// Hook for alert types
export const useAlertTypes = () => {
  const [alertTypes, setAlertTypes] = useState<
    {
      id: number
      name: string
      display_name: string
      icon: string
      color: string
      description: string
    }[]
  >([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchAlertTypes = async () => {
      try {
        setLoading(true)
        setError(null)
        const data = await dashboardService.getAlertTypes()
        setAlertTypes(data)
      } catch (error) {
        setError(
          error instanceof Error ? error.message : 'Unknown error occurred'
        )
      } finally {
        setLoading(false)
      }
    }

    fetchAlertTypes()
  }, [])

  return {
    alertTypes,
    loading,
    error,
  }
}

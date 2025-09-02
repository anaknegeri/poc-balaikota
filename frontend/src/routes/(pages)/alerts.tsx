import { useMemo, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useWebSocketContext } from '@/contexts/websocket-context'
import type { Alert } from '@/types/dashboard'
import {
  AlertTriangle,
  Camera,
  CheckCircle,
  Clock,
  Eye,
  RefreshCw,
  Shield,
  TrendingUp,
} from 'lucide-react'
import { useActiveAlerts, useAlerts } from '@/hooks/use-dashboard'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ErrorMessage } from '@/components/ErrorMessage'
import { LoadingSpinner } from '@/components/LoadingSpinner'

export const Route = createFileRoute('/(pages)/alerts')({
  component: AlertsPage,
})

// SummaryCard component for alert statistics
function SummaryCard({
  title,
  value,
  icon: Icon,
  trend,
  trendValue,
  description,
  className = '',
}: {
  title: string
  value: string
  icon: any
  trend?: 'up' | 'down'
  trendValue?: string
  description?: string
  className?: string
}) {
  return (
    <Card
      className={`rounded-lg border border-gray-200 bg-white/80 p-6 shadow-xs backdrop-blur-sm dark:border-gray-700/50 dark:bg-gray-800/80 ${className}`}
    >
      <CardContent className='relative px-0'>
        <div className='flex items-start justify-between'>
          <div className='flex-1 space-y-3'>
            <div className='flex items-center gap-3'>
              <div
                className={`rounded-xl p-3 transition-colors duration-300 group-hover:bg-blue-100 dark:group-hover:bg-blue-900/30 ${
                  title.includes('Alert') || title.includes('Critical')
                    ? 'bg-red-50 dark:bg-red-900/20'
                    : title.includes('Resolved')
                      ? 'bg-green-50 dark:bg-green-900/20'
                      : 'bg-blue-50 dark:bg-blue-900/20'
                }`}
              >
                <Icon
                  className={`h-5 w-5 transition-colors duration-300 ${
                    title.includes('Alert') || title.includes('Critical')
                      ? 'text-red-600 dark:text-red-400'
                      : title.includes('Resolved')
                        ? 'text-green-600 dark:text-green-400'
                        : 'text-blue-600 dark:text-blue-400'
                  }`}
                />
              </div>
              <div>
                <p className='text-sm font-medium tracking-wide text-gray-600 uppercase dark:text-gray-400'>
                  {title}
                </p>
                {description && (
                  <p className='mt-1 text-xs text-gray-500 dark:text-gray-500'>
                    {description}
                  </p>
                )}
              </div>
            </div>
            <div className='flex items-end gap-3'>
              <p className='text-3xl font-bold tracking-tight text-gray-900 dark:text-white'>
                {value}
              </p>
              {trend && trendValue && (
                <div className='mb-1 flex items-center gap-1'>
                  <TrendingUp
                    className={`h-4 w-4 transition-transform duration-300 ${
                      trend === 'up'
                        ? 'text-emerald-600 dark:text-emerald-400'
                        : 'rotate-180 text-red-600 dark:text-red-400'
                    }`}
                  />
                  <span
                    className={`text-sm font-semibold ${
                      trend === 'up'
                        ? 'text-emerald-600 dark:text-emerald-400'
                        : 'text-red-600 dark:text-red-400'
                    }`}
                  >
                    {trendValue}
                  </span>
                </div>
              )}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function AlertsPage() {
  const [selectedAlert, setSelectedAlert] = useState<Alert | null>(null)
  const [currentPage, setCurrentPage] = useState(1)
  const [filters, _setFilters] = useState<{
    search?: string
    severity?: string
    status?: string
  }>({})

  const { recentAlerts } = useWebSocketContext()
  const { alerts, total, pages, loading, error, refetch } = useAlerts(
    currentPage,
    20,
    filters
  )

  const {
    count: activeAlertsCount,
    loading: _loadingActive,
    refetch: refetchActive,
  } = useActiveAlerts()

  // Alert statistics
  const alertStats = useMemo(() => {
    const totalAlerts = total
    // Use the count from active alerts response instead of array length
    const activeCount = activeAlertsCount || 0
    const criticalCount = alerts.filter((a) => a.severity === 'high').length
    const resolvedCount = alerts.filter((a) => a.resolved_at !== null).length

    return { totalAlerts, activeCount, criticalCount, resolvedCount }
  }, [alerts, activeAlertsCount, total])

  // Get severity badge variant
  const getSeverityBadge = (severity: string) => {
    switch (severity) {
      case 'high':
        return 'destructive'
      case 'medium':
        return 'default'
      case 'low':
        return 'secondary'
      default:
        return 'default'
    }
  }

  // Get severity color
  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'high':
        return 'bg-red-500'
      case 'medium':
        return 'bg-amber-500'
      case 'low':
        return 'bg-emerald-500'
      default:
        return 'bg-gray-500'
    }
  }

  // Loading state
  if (loading && alerts.length === 0) {
    return <LoadingSpinner message='Loading alerts...' />
  }

  // Error state
  if (error && alerts.length === 0) {
    return <ErrorMessage message={error} onRetry={refetch} />
  }

  // Single Alert Detail Mode
  if (selectedAlert) {
    return (
      <div className='mx-auto w-full space-y-4 p-4'>
        {/* Back Button */}
        <div className='flex flex-shrink-0 items-center justify-between'>
          <Button
            onClick={() => setSelectedAlert(null)}
            variant='outline'
            size='sm'
          >
            ← Back to Alerts
          </Button>
        </div>

        {/* Alert Detail */}
        <div className='grid flex-1 grid-cols-1 gap-6 lg:grid-cols-3'>
          {/* Alert Image - 2 columns */}
          <div className='flex min-h-0 flex-col lg:col-span-2'>
            <Card className='overflow-hidden'>
              <div className='relative aspect-video w-full'>
                {selectedAlert.image_url ? (
                  <img
                    src={selectedAlert.image_url}
                    alt={`Alert ${selectedAlert.id}`}
                    className='h-full w-full bg-gray-900 object-contain'
                  />
                ) : (
                  <div className='flex h-full w-full items-center justify-center bg-gray-900'>
                    <AlertTriangle className='h-16 w-16 text-gray-400' />
                    <span className='ml-2 text-gray-400'>
                      No image available
                    </span>
                  </div>
                )}

                {/* Alert Info Overlay */}
                <div className='absolute top-0 right-0 left-0 bg-gradient-to-b from-black/60 via-black/20 to-transparent p-4'>
                  <div className='flex items-start justify-between'>
                    <div>
                      <h3 className='mb-1 text-lg font-semibold text-white drop-shadow-lg'>
                        {selectedAlert.alert_type?.display_name ||
                          'Security Alert'}
                      </h3>
                      <p className='text-sm text-gray-200 drop-shadow'>
                        {selectedAlert.camera?.name ||
                          `Camera ${selectedAlert.camera_id}`}
                      </p>
                    </div>
                    <Badge
                      variant={getSeverityBadge(selectedAlert.severity) as any}
                      className='text-xs'
                    >
                      {selectedAlert.severity.toUpperCase()}
                    </Badge>
                  </div>
                </div>
              </div>
            </Card>
          </div>

          {/* Alert Details Sidebar - 1 column */}
          <Card className='flex min-h-0 flex-col'>
            <CardHeader className='flex-shrink-0 border-b'>
              <CardTitle className='text-base'>Alert Details</CardTitle>
            </CardHeader>
            <CardContent className='min-h-0 flex-1 space-y-4 p-4'>
              <div>
                <label className='text-sm font-medium text-gray-600 dark:text-gray-400'>
                  Message
                </label>
                <p className='mt-1 text-sm text-gray-900 dark:text-white'>
                  {selectedAlert.message}
                </p>
              </div>

              <div>
                <label className='text-sm font-medium text-gray-600 dark:text-gray-400'>
                  Location
                </label>
                <p className='mt-1 text-sm text-gray-900 dark:text-white'>
                  {selectedAlert.camera?.location || 'Unknown location'}
                </p>
              </div>

              <div>
                <label className='text-sm font-medium text-gray-600 dark:text-gray-400'>
                  Detected At
                </label>
                <p className='mt-1 text-sm text-gray-900 dark:text-white'>
                  {new Date(selectedAlert.detected_at).toLocaleString()}
                </p>
              </div>

              <div>
                <label className='text-sm font-medium text-gray-600 dark:text-gray-400'>
                  Status
                </label>
                <div className='mt-1 flex items-center gap-2'>
                  <div
                    className={`h-2 w-2 rounded-full ${getSeverityColor(selectedAlert.severity)}`}
                  />
                  <span className='text-sm text-gray-900 dark:text-white'>
                    {selectedAlert.is_active ? 'Active' : 'Resolved'}
                  </span>
                </div>
              </div>

              {selectedAlert.resolved_at && (
                <div>
                  <label className='text-sm font-medium text-gray-600 dark:text-gray-400'>
                    Resolved At
                  </label>
                  <p className='mt-1 text-sm text-gray-900 dark:text-white'>
                    {new Date(selectedAlert.resolved_at).toLocaleString()}
                  </p>
                </div>
              )}

              {selectedAlert.resolution_note && (
                <div>
                  <label className='text-sm font-medium text-gray-600 dark:text-gray-400'>
                    Resolution Note
                  </label>
                  <p className='mt-1 text-sm text-gray-900 dark:text-white'>
                    {selectedAlert.resolution_note}
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  // Grid View Mode
  return (
    <div className='mx-auto w-full space-y-4 p-4'>
      {/* Header */}
      <div className='flex items-center justify-between'>
        <div>
          <h1 className='text-3xl font-bold tracking-tight'>Security Alerts</h1>
          <p className='text-muted-foreground'>
            Monitor and manage security alerts from surveillance system
          </p>
        </div>
        <Button
          className='hover:cursor-pointer'
          size='sm'
          onClick={() => {
            refetch()
            refetchActive()
          }}
          disabled={loading}
        >
          <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          <span>Refresh</span>
        </Button>
      </div>

      {/* Recent Real-time Alerts */}
      {recentAlerts.length > 0 && (
        <Card className='border-blue-200 bg-blue-50/50 dark:border-blue-800 dark:bg-blue-900/20'>
          <CardHeader>
            <CardTitle className='flex items-center gap-2 text-blue-800 dark:text-blue-200'>
              <AlertTriangle className='h-5 w-5' />
              Recent Real-time Alerts
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className='space-y-3'>
              {recentAlerts.slice(0, 3).map((alert, index) => (
                <div
                  key={index}
                  className='flex items-center justify-between rounded-lg border border-blue-200 bg-white p-3 dark:border-blue-700 dark:bg-blue-900/30'
                >
                  <div className='flex items-center gap-3'>
                    <AlertTriangle className='h-4 w-4 text-blue-600' />
                    <div>
                      <p className='font-medium text-blue-900 dark:text-blue-100'>
                        {alert.data?.alert_type || alert.alert_type} -{' '}
                        {alert.data?.camera_name || alert.camera_name}
                      </p>
                      <p className='text-sm text-blue-700 dark:text-blue-300'>
                        {alert.data?.message || alert.message}
                      </p>
                    </div>
                  </div>
                  <div className='text-right'>
                    <p className='text-xs text-blue-600 dark:text-blue-400'>
                      {new Date(alert.timestamp).toLocaleTimeString()}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Summary Cards */}
      <div className='grid gap-4 md:grid-cols-2 lg:grid-cols-4'>
        <SummaryCard
          title='Total Alerts'
          value={alertStats.totalAlerts.toString()}
          icon={AlertTriangle}
          description='All time alerts'
        />
        <SummaryCard
          title='Active Alerts'
          value={alertStats.activeCount.toString()}
          icon={Shield}
          description='Requires attention'
        />
        <SummaryCard
          title='Critical Alerts'
          value={alertStats.criticalCount.toString()}
          icon={AlertTriangle}
          description='High severity'
        />
        <SummaryCard
          title='Resolved Today'
          value={alertStats.resolvedCount.toString()}
          icon={CheckCircle}
          description='Successfully handled'
        />
      </div>

      {/* Error Message for Partial Errors */}
      {error && alerts.length > 0 && (
        <ErrorMessage message={error} onRetry={refetch} compact />
      )}

      {/* Alerts List */}
      <div className='space-y-4'>
        {loading && alerts.length === 0 ? (
          <div className='space-y-4'>
            {Array.from({ length: 5 }).map((_, index) => (
              <Card key={index} className='overflow-hidden'>
                <CardContent className='p-4'>
                  <div className='flex items-center gap-4'>
                    <div className='h-16 w-24 animate-pulse rounded bg-gray-300 dark:bg-gray-600' />
                    <div className='flex-1 space-y-2'>
                      <div className='h-4 w-32 animate-pulse rounded bg-gray-300 dark:bg-gray-600' />
                      <div className='h-3 w-48 animate-pulse rounded bg-gray-300 dark:bg-gray-600' />
                      <div className='h-3 w-24 animate-pulse rounded bg-gray-300 dark:bg-gray-600' />
                    </div>
                    <div className='h-6 w-16 animate-pulse rounded bg-gray-300 dark:bg-gray-600' />
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        ) : alerts.length === 0 ? (
          <div className='py-12 text-center'>
            <Shield className='mx-auto mb-4 h-16 w-16 text-gray-400' />
            <h3 className='mb-2 text-lg font-semibold text-gray-900 dark:text-white'>
              No Alerts Found
            </h3>
            <p className='text-gray-600 dark:text-gray-400'>
              No security alerts have been detected recently.
            </p>
          </div>
        ) : (
          <div className='space-y-3'>
            {alerts.map((alert) => (
              <Card
                key={alert.id}
                onClick={() => setSelectedAlert(alert)}
                className='cursor-pointer transition-all hover:border-red-300 hover:shadow-md dark:hover:border-red-600'
              >
                <CardContent className='p-4'>
                  <div className='flex items-center gap-4'>
                    {/* Alert Thumbnail */}
                    <div className='h-16 w-24 flex-shrink-0 overflow-hidden rounded-md bg-gray-100 dark:bg-gray-800'>
                      {alert.image_url ? (
                        <img
                          src={alert.image_url}
                          alt={`Alert ${alert.id}`}
                          className='h-full w-full object-cover'
                          onError={(e) => {
                            const target = e.target as HTMLImageElement
                            target.style.display = 'none'
                            const parent = target.parentElement
                            if (parent) {
                              parent.innerHTML = `
                                <div class="flex h-full w-full items-center justify-center bg-gray-200 dark:bg-gray-700">
                                  <svg class="h-6 w-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
                                  </svg>
                                </div>
                              `
                            }
                          }}
                        />
                      ) : (
                        <div className='flex h-full w-full items-center justify-center bg-gray-200 dark:bg-gray-700'>
                          <AlertTriangle className='h-6 w-6 text-gray-400' />
                        </div>
                      )}
                    </div>

                    {/* Alert Info */}
                    <div className='min-w-0 flex-1'>
                      <div className='flex items-start justify-between'>
                        <div className='min-w-0 flex-1'>
                          <h4 className='truncate text-sm font-semibold text-gray-900 dark:text-white'>
                            {alert.alert_type?.display_name || 'Security Alert'}
                          </h4>
                          <p className='mt-1 truncate text-sm text-gray-600 dark:text-gray-400'>
                            {alert.message}
                          </p>
                          <div className='mt-2 flex items-center gap-4 text-xs text-gray-500 dark:text-gray-500'>
                            <span className='flex items-center gap-1'>
                              <Camera className='h-3 w-3' />
                              {alert.camera?.name ||
                                `Camera ${alert.camera_id}`}
                            </span>
                            <span className='flex items-center gap-1'>
                              <Clock className='h-3 w-3' />
                              {new Date(alert.detected_at).toLocaleString()}
                            </span>
                          </div>
                        </div>

                        {/* Status and Severity */}
                        <div className='ml-4 flex flex-col items-end gap-2'>
                          <Badge
                            variant={getSeverityBadge(alert.severity) as any}
                            className='text-xs'
                          >
                            {alert.severity.toUpperCase()}
                          </Badge>
                          {alert.is_active ? (
                            <span className='inline-flex items-center rounded-full bg-red-100 px-2 py-1 text-xs font-medium text-red-800 dark:bg-red-900 dark:text-red-200'>
                              Active
                            </span>
                          ) : (
                            <span className='inline-flex items-center rounded-full bg-green-100 px-2 py-1 text-xs font-medium text-green-800 dark:bg-green-900 dark:text-green-200'>
                              Resolved
                            </span>
                          )}
                        </div>
                      </div>
                    </div>

                    {/* View Icon */}
                    <div className='flex-shrink-0'>
                      <Eye className='h-5 w-5 text-gray-400' />
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* Pagination */}
      {pages > 1 && (
        <div className='flex items-center justify-center gap-2'>
          <Button
            variant='outline'
            size='sm'
            onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
            disabled={currentPage === 1 || loading}
          >
            Previous
          </Button>
          <span className='text-sm text-gray-600 dark:text-gray-400'>
            Page {currentPage} of {pages}
          </span>
          <Button
            variant='outline'
            size='sm'
            onClick={() => setCurrentPage(Math.min(pages, currentPage + 1))}
            disabled={currentPage === pages || loading}
          >
            Next
          </Button>
        </div>
      )}

      {/* Results Info */}
      <div className='text-center text-sm text-gray-600 dark:text-gray-400'>
        Showing {alerts.length} of {total} alerts
      </div>
    </div>
  )
}

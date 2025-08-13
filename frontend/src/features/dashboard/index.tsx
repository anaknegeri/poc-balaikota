import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import type { Alert } from '@/types/dashboard'
import {
  Activity,
  AlertTriangle,
  Baby,
  Calendar,
  Camera,
  Clock,
  Eye,
  Loader2,
  RefreshCw,
  Shield,
  User,
  UserCheck,
  Users,
} from 'lucide-react'
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  XAxis,
  YAxis,
} from 'recharts'
import { useCameras } from '@/hooks/use-cameras'
import { useDashboardData, type DateRange } from '@/hooks/use-dashboard'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart'

// Chart configurations

// Fallback data for when backend is not available

const peakHoursConfig = {
  visitors: {
    label: 'Current',
    color: 'hsl(217, 91%, 60%)',
  },
  previous: {
    label: 'Previous Day',
    color: 'hsl(217, 91%, 80%)',
  },
} satisfies ChartConfig

const demographicsConfig = {
  male: {
    label: 'Male',
    color: 'hsl(217, 91%, 60%)',
  },
  female: {
    label: 'Female',
    color: 'hsl(142, 76%, 36%)',
  },
  children: {
    label: 'Children',
    color: 'hsl(47, 96%, 53%)',
  },
} satisfies ChartConfig

// Loading component
function LoadingCard({ className = '' }: { className?: string }) {
  return (
    <Card className={className}>
      <CardContent className='flex items-center justify-center p-6'>
        <Loader2 className='h-6 w-6 animate-spin text-gray-400' />
      </CardContent>
    </Card>
  )
}

function SummaryCard({
  title,
  value,
  icon: Icon,
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
                  title.includes('Alert')
                    ? 'bg-red-50 dark:bg-red-900/20'
                    : title.includes('Camera')
                      ? 'bg-blue-50 dark:bg-blue-900/20'
                      : 'bg-green-50 dark:bg-green-900/20'
                }`}
              >
                <Icon
                  className={`h-5 w-5 transition-colors duration-300 ${
                    title.includes('Alert')
                      ? 'text-red-600 dark:text-red-400'
                      : title.includes('Camera')
                        ? 'text-blue-600 dark:text-blue-400'
                        : 'text-green-600 dark:text-green-400'
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
              {/* {trend && trendValue && (
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
              )} */}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function HourlyDemographicsChart({
  data,
  loading,
}: {
  data: any
  loading: boolean
}) {
  if (loading) {
    return <LoadingCard />
  }

  // Create hourly demographics data from people count trends
  const trendsData = data?.peopleCountTrends || []
  const hourlyDemographicsData = trendsData
    .map((item: any) => ({
      hour: new Date(item.time_period).getHours(),
      time_label: new Date(item.time_period).toLocaleTimeString('en-US', {
        hour: 'numeric',
        hour12: true,
      }),
      male: item.male_count || 0,
      female: item.female_count || 0,
      children: item.child_count || 0,
      adults: item.adult_count || 0,
      elderly: item.elderly_count || 0,
      total: item.total_count || 0,
    }))
    .slice(0, 24) // Show last 24 hours

  return (
    <Card>
      <CardHeader className='pb-3'>
        <div className='flex items-center gap-3'>
          <div className='rounded-lg bg-purple-50 p-2 dark:bg-purple-900/20'>
            <Users className='h-5 w-5 text-purple-600 dark:text-purple-400' />
          </div>
          <CardTitle className='text-lg font-semibold text-gray-900 dark:text-white'>
            Hourly Demographics Breakdown
          </CardTitle>
        </div>
      </CardHeader>
      <CardContent className='pt-0'>
        <ChartContainer
          config={demographicsConfig}
          className='aspect-auto h-80 w-full'
        >
          <AreaChart
            accessibilityLayer
            data={hourlyDemographicsData || []}
            margin={{ left: 12, right: 12, top: 12, bottom: 12 }}
          >
            <defs>
              <linearGradient id='fillMale' x1='0' y1='0' x2='0' y2='1'>
                <stop
                  offset='5%'
                  stopColor='hsl(217, 91%, 60%)'
                  stopOpacity={0.8}
                />
                <stop
                  offset='95%'
                  stopColor='hsl(217, 91%, 60%)'
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient id='fillFemale' x1='0' y1='0' x2='0' y2='1'>
                <stop
                  offset='5%'
                  stopColor='hsl(142, 76%, 36%)'
                  stopOpacity={0.8}
                />
                <stop
                  offset='95%'
                  stopColor='hsl(142, 76%, 36%)'
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient id='fillChildren' x1='0' y1='0' x2='0' y2='1'>
                <stop
                  offset='5%'
                  stopColor='hsl(47, 96%, 53%)'
                  stopOpacity={0.8}
                />
                <stop
                  offset='95%'
                  stopColor='hsl(47, 96%, 53%)'
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray='3 3' stroke='hsl(217, 91%, 95%)' />
            <XAxis
              dataKey='time_label'
              tickLine={false}
              axisLine={false}
              tickMargin={16}
              className='text-xs'
            />
            <YAxis
              axisLine={false}
              tickLine={false}
              tickMargin={16}
              className='text-xs'
            />
            <ChartTooltip
              cursor={false}
              content={<ChartTooltipContent className='w-40' />}
            />
            <Area
              dataKey='male'
              type='monotone'
              stackId='1'
              fill='url(#fillMale)'
              stroke='hsl(217, 91%, 60%)'
              strokeWidth={1}
            />
            <Area
              dataKey='female'
              type='monotone'
              stackId='1'
              fill='url(#fillFemale)'
              stroke='hsl(142, 76%, 36%)'
              strokeWidth={1}
            />
            <Area
              dataKey='children'
              type='monotone'
              stackId='1'
              fill='url(#fillChildren)'
              stroke='hsl(47, 96%, 53%)'
              strokeWidth={1}
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  )
}

function DemographicsChart({ data, loading }: { data: any; loading: boolean }) {
  if (loading) {
    return <LoadingCard />
  }

  const totals = data?.peopleCountSummary?.totals
  const demographicsData = totals
    ? [
        { name: 'Male', value: totals.male, fill: 'var(--color-male)' },
        { name: 'Female', value: totals.female, fill: 'var(--color-female)' },
        {
          name: 'Children',
          value: totals.child,
          fill: 'var(--color-children)',
        },
      ]
    : []

  return (
    <Card>
      <CardHeader className='pb-3'>
        <div className='flex items-center gap-3'>
          <div className='rounded-lg bg-purple-50 p-2 dark:bg-purple-900/20'>
            <Users className='h-5 w-5 text-purple-600 dark:text-purple-400' />
          </div>
          <CardTitle className='text-lg font-semibold text-gray-900 dark:text-white'>
            Demographics Distribution
          </CardTitle>
        </div>
      </CardHeader>
      <CardContent className='pt-0'>
        <ChartContainer
          config={demographicsConfig}
          className='aspect-auto h-80 w-full'
        >
          <BarChart
            accessibilityLayer
            data={demographicsData || []}
            margin={{ left: 12, right: 12, top: 12, bottom: 12 }}
          >
            <CartesianGrid strokeDasharray='3 3' stroke='hsl(217, 91%, 95%)' />
            <XAxis
              dataKey='name'
              tickLine={false}
              axisLine={false}
              tickMargin={16}
              className='text-xs'
            />
            <YAxis
              axisLine={false}
              tickLine={false}
              tickMargin={16}
              className='text-xs'
            />
            <ChartTooltip content={<ChartTooltipContent className='w-40' />} />
            <Bar dataKey='value' radius={[4, 4, 0, 0]} />
          </BarChart>
        </ChartContainer>
      </CardContent>
    </Card>
  )
}

function VisitorDistributionCard({
  data,
  loading,
}: {
  data: any
  loading: boolean
}) {
  if (loading) {
    return <LoadingCard />
  }

  // Menggunakan struktur data yang benar dari backend
  const visitorStats =
    data?.peopleCountSummary?.data?.totals || data?.peopleCountSummary?.totals

  return (
    <Card className='gap-4'>
      <CardHeader>
        <div className='flex items-center gap-3'>
          <div className='rounded-lg bg-indigo-50 p-2 dark:bg-indigo-900/20'>
            <Users className='h-5 w-5 text-indigo-600 dark:text-indigo-400' />
          </div>
          <CardTitle className='text-lg font-semibold text-gray-900 dark:text-white'>
            Visitor Distribution
          </CardTitle>
        </div>
      </CardHeader>
      <CardContent className='space-y-4 pt-0'>
        {/* Gender in single row */}
        <div className='grid grid-cols-2 gap-3'>
          <div className='flex items-center justify-between rounded-lg border border-blue-200 bg-blue-50/50 p-3 dark:border-blue-800/30 dark:bg-blue-900/20'>
            <div className='flex items-center gap-2'>
              <User className='h-4 w-4 text-blue-600 dark:text-blue-400' />
              <span className='text-xs font-medium text-gray-700 dark:text-gray-300'>
                Male
              </span>
            </div>
            <div className='text-right'>
              <div className='text-sm font-bold text-gray-900 dark:text-white'>
                {visitorStats?.male || 0}
              </div>
              <div className='text-xs text-gray-500 dark:text-gray-400'>
                Real-time
              </div>
            </div>
          </div>

          <div className='flex items-center justify-between rounded-lg border border-pink-200 bg-pink-50/50 p-3 dark:border-pink-800/30 dark:bg-pink-900/20'>
            <div className='flex items-center gap-2'>
              <Users className='h-4 w-4 text-pink-600 dark:text-pink-400' />
              <span className='text-xs font-medium text-gray-700 dark:text-gray-300'>
                Female
              </span>
            </div>
            <div className='text-right'>
              <div className='text-sm font-bold text-gray-900 dark:text-white'>
                {visitorStats?.female || 0}
              </div>
              <div className='text-xs text-gray-500 dark:text-gray-400'>
                Real-time
              </div>
            </div>
          </div>
        </div>

        {/* Age Groups */}
        <div className='grid grid-cols-3 gap-3'>
          <div className='flex items-center justify-between rounded-lg border border-green-200 bg-green-50/50 p-3 dark:border-green-800/30 dark:bg-green-900/20'>
            <div className='flex items-center gap-2'>
              <Baby className='h-4 w-4 text-green-600 dark:text-green-400' />
              <span className='text-xs font-medium text-gray-700 dark:text-gray-300'>
                Children
              </span>
            </div>
            <div className='text-right'>
              <div className='text-sm font-bold text-gray-900 dark:text-white'>
                {visitorStats?.child || 0}
              </div>
              <div className='text-xs text-gray-500 dark:text-gray-400'>
                Real-time
              </div>
            </div>
          </div>

          <div className='flex items-center justify-between rounded-lg border border-purple-200 bg-purple-50/50 p-3 dark:border-purple-800/30 dark:bg-purple-900/20'>
            <div className='flex items-center gap-2'>
              <UserCheck className='h-4 w-4 text-purple-600 dark:text-purple-400' />
              <span className='text-xs font-medium text-gray-700 dark:text-gray-300'>
                Adults
              </span>
            </div>
            <div className='text-right'>
              <div className='text-sm font-bold text-gray-900 dark:text-white'>
                {visitorStats?.adult || 0}
              </div>
              <div className='text-xs text-gray-500 dark:text-gray-400'>
                Real-time
              </div>
            </div>
          </div>

          <div className='flex items-center justify-between rounded-lg border border-orange-200 bg-orange-50/50 p-3 dark:border-orange-800/30 dark:bg-orange-900/20'>
            <div className='flex items-center gap-2'>
              <Users className='h-4 w-4 text-orange-600 dark:text-orange-400' />
              <span className='text-xs font-medium text-gray-700 dark:text-gray-300'>
                Elderly
              </span>
            </div>
            <div className='text-right'>
              <div className='text-sm font-bold text-gray-900 dark:text-white'>
                {visitorStats?.elderly || 0}
              </div>
              <div className='text-xs text-gray-500 dark:text-gray-400'>
                Real-time
              </div>
            </div>
          </div>
        </div>

        {/* Total Summary */}
        <div className='border-t border-gray-200 pt-3 dark:border-gray-700'>
          <div className='flex items-center justify-between rounded-lg border border-gray-200 bg-gray-50/50 p-3 dark:border-gray-700 dark:bg-gray-800/50'>
            <div className='flex items-center gap-2'>
              <Activity className='h-4 w-4 text-gray-600 dark:text-gray-400' />
              <span className='text-xs font-medium text-gray-800 dark:text-gray-200'>
                Total Today
              </span>
            </div>
            <div className='text-right'>
              <div className='text-lg font-bold text-gray-900 dark:text-white'>
                {visitorStats?.total || 0}
              </div>
              <div className='text-xs text-gray-500 dark:text-gray-400'>
                Real-time data
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function PeakHoursAnalysisChart({
  data,
  loading,
}: {
  data: any
  loading: boolean
}) {
  if (loading) {
    return <LoadingCard />
  }

  const peakHoursData = data?.peakHoursAnalysis?.data || []
  const summary = data?.peakHoursAnalysis?.summary

  return (
    <Card>
      <CardHeader className='pb-3'>
        <div className='flex items-center gap-3'>
          <div className='rounded-lg bg-blue-50 p-2 dark:bg-blue-900/20'>
            <Clock className='h-5 w-5 text-blue-600 dark:text-blue-400' />
          </div>
          <CardTitle className='text-lg font-semibold text-gray-900 dark:text-white'>
            Peak Hours Analysis
          </CardTitle>
        </div>
      </CardHeader>
      <CardContent className='pt-0'>
        <div className='space-y-4'>
          {/* Summary Stats */}
          {summary && (
            <div className='mb-4 grid grid-cols-2 gap-4'>
              <div className='rounded-lg bg-blue-50 p-3 text-center dark:bg-blue-900/20'>
                <div className='text-2xl font-bold text-blue-600 dark:text-blue-400'>
                  {summary.peak_hour}:00
                </div>
                <div className='text-sm text-gray-600 dark:text-gray-400'>
                  Peak Hour ({summary.peak_count} visitors)
                </div>
              </div>
              <div className='rounded-lg bg-green-50 p-3 text-center dark:bg-green-900/20'>
                <div className='text-2xl font-bold text-green-600 dark:text-green-400'>
                  {summary.average_per_hour?.toFixed(1)}
                </div>
                <div className='text-sm text-gray-600 dark:text-gray-400'>
                  Avg per Hour
                </div>
              </div>
            </div>
          )}

          {/* Chart */}
          <ChartContainer
            config={peakHoursConfig}
            className='aspect-auto h-64 w-full'
          >
            <AreaChart
              accessibilityLayer
              data={peakHoursData || []}
              margin={{ left: 12, right: 12, top: 12, bottom: 12 }}
            >
              <defs>
                <linearGradient id='fillVisitors2' x1='0' y1='0' x2='0' y2='1'>
                  <stop
                    offset='5%'
                    stopColor='hsl(217, 91%, 60%)'
                    stopOpacity={0.3}
                  />
                  <stop
                    offset='95%'
                    stopColor='hsl(217, 91%, 60%)'
                    stopOpacity={0.1}
                  />
                </linearGradient>
              </defs>
              <CartesianGrid
                strokeDasharray='3 3'
                stroke='hsl(217, 91%, 95%)'
              />
              <XAxis
                dataKey='time_label'
                tickLine={false}
                axisLine={false}
                tickMargin={16}
                className='text-xs'
              />
              <YAxis
                axisLine={false}
                tickLine={false}
                tickMargin={16}
                className='text-xs'
              />
              <ChartTooltip
                cursor={false}
                content={<ChartTooltipContent className='w-40' />}
              />
              <Area
                dataKey='visitors'
                type='monotone'
                fill='url(#fillVisitors2)'
                fillOpacity={0.4}
                stroke='hsl(217, 91%, 60%)'
                strokeWidth={2}
              />
            </AreaChart>
          </ChartContainer>
        </div>
      </CardContent>
    </Card>
  )
}

function RecentAlertsCard({
  data,
  loading,
  onImageClick,
}: {
  data: any
  loading: boolean
  onImageClick: (image: string) => void
}) {
  if (loading) {
    return <LoadingCard />
  }

  const alertsData = data?.activeAlert?.data || []

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

  return (
    <Card>
      <CardHeader className='pb-3'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-3'>
            <div className='rounded-lg bg-red-50 p-2 dark:bg-red-900/20'>
              <AlertTriangle className='h-5 w-5 text-red-600 dark:text-red-400' />
            </div>
            <div>
              <CardTitle className='text-lg font-semibold text-gray-900 dark:text-white'>
                Security Alerts
              </CardTitle>
            </div>
          </div>
          <Link to='/alerts'>
            <Button variant='outline' size='sm'>
              <Eye className='mr-2 h-4 w-4' />
              View All
            </Button>
          </Link>
        </div>
      </CardHeader>
      <CardContent className='pt-0'>
        <div className='max-h-80 space-y-3 overflow-y-auto pr-2'>
          {(alertsData || []).map((alert: Alert) => (
            <div
              key={alert.id}
              className='rounded-lg border border-gray-200 p-3 dark:border-gray-700'
            >
              <div className='flex items-start gap-3'>
                {/* Alert Thumbnail */}
                <div className='h-12 w-16 flex-shrink-0 overflow-hidden rounded-md bg-gray-100 dark:bg-gray-800'>
                  {alert.image_url ? (
                    <img
                      src={alert.image_url}
                      alt={`Alert ${alert.id}`}
                      className='h-full w-full cursor-pointer object-cover transition-transform hover:scale-105'
                      onClick={() => onImageClick(alert.image_url)}
                      onError={(e) => {
                        const target = e.target as HTMLImageElement
                        target.style.display = 'none'
                        const parent = target.parentElement
                        if (parent) {
                          parent.innerHTML = `
                            <div class="flex h-full w-full items-center justify-center bg-gray-200 dark:bg-gray-700">
                              <svg class="h-4 w-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
                              </svg>
                            </div>
                          `
                        }
                      }}
                    />
                  ) : (
                    <div className='flex h-full w-full items-center justify-center bg-gray-200 dark:bg-gray-700'>
                      <Camera className='h-4 w-4 text-gray-400' />
                    </div>
                  )}
                </div>

                <div className='min-w-0 flex-1'>
                  <div className='flex items-start justify-between gap-3'>
                    <div className='flex-1'>
                      <h4 className='text-sm font-medium text-gray-900 dark:text-white'>
                        {alert.alert_type?.display_name}
                      </h4>
                      <p className='mt-1 text-xs text-gray-600 dark:text-gray-400'>
                        {alert.message}
                      </p>
                    </div>
                    <Badge
                      variant={getSeverityBadge(alert.severity) as any}
                      className='text-xs'
                    >
                      {alert.severity}
                    </Badge>
                  </div>
                  <div className='mt-2 flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400'>
                    <span>
                      {alert.camera?.name || `Camera ${alert.camera_id}`}
                    </span>
                    <span>{alert.camera?.location}</span>
                    <span>{new Date(alert.detected_at).toLocaleString()}</span>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

export default function Dashboard() {
  const [selectedImage, setSelectedImage] = useState<string | null>(null)

  // Initialize date range with today's date
  const today = new Date()
  const [dateRange, _setDateRange] = useState<DateRange>({
    from: today.toISOString().split('T')[0],
    to: today.toISOString().split('T')[0],
  })

  const { data, loading, error, refetch } = useDashboardData(
    undefined,
    dateRange
  )

  const { cameras, loading: loadingCamera } = useCameras('active')

  const activeAlertsCount =
    data.activeAlert?.data?.filter((alert) => alert.is_active).length || 0

  // Image viewer modal
  const ImageModal = ({
    image,
    onClose,
  }: {
    image: string
    onClose: () => void
  }) => (
    <div
      className='fixed inset-0 z-50 flex items-center justify-center bg-black/80'
      onClick={onClose}
    >
      <div className='relative max-h-[90vh] max-w-4xl p-4'>
        <button
          onClick={onClose}
          className='absolute top-2 right-2 flex h-8 w-8 cursor-pointer items-center justify-center rounded-full bg-white/20 text-xl font-bold text-white transition-colors hover:bg-white/30'
        >
          ×
        </button>
        <img
          src={image}
          alt='Detection snapshot'
          className='max-h-full max-w-full rounded-lg'
        />
      </div>
    </div>
  )

  // Error state only if no data at all
  if (error && !data.peopleCountSummary) {
    return (
      <div className='flex min-h-[400px] items-center justify-center'>
        <div className='text-center'>
          <p className='mb-4 text-red-500'>{error}</p>
          <Button onClick={refetch}>Retry</Button>
        </div>
      </div>
    )
  }

  return (
    <>
      {/* Image Modal */}
      {selectedImage && (
        <ImageModal
          image={selectedImage}
          onClose={() => setSelectedImage(null)}
        />
      )}

      <div className='mx-auto w-full space-y-4 p-4'>
        {/* Header */}
        <div className='flex items-center justify-between'>
          <div>
            <h1 className='text-3xl font-bold tracking-tight'>
              Security Dashboard
            </h1>
            <p className='text-muted-foreground'>
              Real-time monitoring and analytics
            </p>
          </div>
          <div className='flex items-center space-x-2'>
            <div className='flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400'>
              <Calendar className='h-4 w-4' />
              <span>
                {dateRange.from === dateRange.to
                  ? `Date: ${new Date(dateRange.from).toLocaleDateString()}`
                  : `Range: ${new Date(dateRange.from).toLocaleDateString()} - ${new Date(dateRange.to).toLocaleDateString()}`}
              </span>
            </div>

            <Button
              className='hover:cursor-pointer'
              size='sm'
              onClick={refetch}
              disabled={loading}
            >
              <RefreshCw
                className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`}
              />
              <span>Refresh</span>
            </Button>
          </div>
        </div>

        {/* Summary Cards */}
        <div className='grid gap-4 md:grid-cols-2 lg:grid-cols-4'>
          <SummaryCard
            title='Total Visitors Today'
            value={
              loading
                ? '...'
                : data.peopleCountSummary?.totals?.total?.toString() || '0'
            }
            icon={Users}
            trend='up'
            trendValue='+12.4%'
            description='Compared to yesterday'
          />
          <SummaryCard
            title='Active Security Alerts'
            value={activeAlertsCount.toString()}
            icon={AlertTriangle}
            trend={activeAlertsCount > 0 ? 'up' : 'down'}
            trendValue={activeAlertsCount > 0 ? `+${activeAlertsCount}` : '0'}
            description='Recent incidents'
          />
          <SummaryCard
            title='Active Cameras'
            value={loadingCamera ? '...' : cameras?.length?.toString() || '0'}
            icon={Camera}
            description='System coverage'
          />
          <SummaryCard
            title='System Status'
            value='Online'
            icon={Shield}
            trend='up'
            description='All systems operational'
          />
        </div>

        {/* Charts Grid */}
        <div className='grid grid-cols-2 gap-4'>
          <VisitorDistributionCard data={data} loading={loading} />
          <RecentAlertsCard
            data={data}
            loading={loading}
            onImageClick={setSelectedImage}
          />
        </div>

        <div className='grid gap-4 md:grid-cols-2'>
          <DemographicsChart data={data} loading={loading} />
          <PeakHoursAnalysisChart data={data} loading={loading} />
        </div>

        {/* Hourly Demographics */}
        <HourlyDemographicsChart data={data} loading={loading} />
      </div>
    </>
  )
}

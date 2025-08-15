import { useMemo, useState, useEffect } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import type { FaceRecognition } from '@/types/dashboard'
import {
  Camera,
  Clock,
  Eye,
  RefreshCw,
  TrendingUp,
  User,
  UserCheck,
  Users,
} from 'lucide-react'
import { dashboardService } from '@/services/dashboardService'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ErrorMessage } from '@/components/ErrorMessage'
import { LoadingSpinner } from '@/components/LoadingSpinner'

export const Route = createFileRoute('/(pages)/face-recognition')({
  component: FaceRecognitionPage,
})

// Custom hook for face recognition data
function useFaceRecognitions(page: number = 1, limit: number = 20) {
  const [data, setData] = useState<{
    recognitions: FaceRecognition[]
    total: number
    pages: number
    count: number
  }>({
    recognitions: [],
    total: 0,
    pages: 1,
    count: 0,
  })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchData = async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await dashboardService.getFaceRecognitions(page, limit)
      setData({
        recognitions: result.data || [],
        total: result.total || 0,
        pages: result.pages || 1,
        count: result.count || 0,
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error occurred')
    } finally {
      setLoading(false)
    }
  }

  // Fetch data on mount and when page changes
  useEffect(() => {
    fetchData()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, limit])

  return {
    ...data,
    loading,
    error,
    refetch: fetchData,
  }
}

// SummaryCard component for face recognition statistics
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
              <div className='rounded-xl bg-blue-50 p-3 transition-colors duration-300 group-hover:bg-blue-100 dark:bg-blue-900/20 dark:group-hover:bg-blue-900/30'>
                <Icon className='h-5 w-5 text-blue-600 transition-colors duration-300 dark:text-blue-400' />
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

function FaceRecognitionPage() {
  const [selectedRecognition, setSelectedRecognition] = useState<FaceRecognition | null>(null)
  const [currentPage, setCurrentPage] = useState(1)

  const { recognitions, total, pages, count, loading, error, refetch } = useFaceRecognitions(
    currentPage,
    20
  )

  // Face recognition statistics
  const faceRecognitionStats = useMemo(() => {
    const totalRecognitions = total
    const todayCount = count
    const uniquePersons = new Set(recognitions.map((r) => r.object_name)).size
    const detectedToday = recognitions.filter((r) => {
      const today = new Date()
      const detectedDate = new Date(r.detected_at)
      return detectedDate.toDateString() === today.toDateString()
    }).length

    return { totalRecognitions, todayCount, uniquePersons, detectedToday }
  }, [recognitions, total, count])

  // Loading state
  if (loading && recognitions.length === 0) {
    return <LoadingSpinner message='Loading face recognitions...' />
  }

  // Error state
  if (error && recognitions.length === 0) {
    return <ErrorMessage message={error} onRetry={refetch} />
  }

  // Single Face Recognition Detail Mode
  if (selectedRecognition) {
    return (
      <div className='mx-auto w-full space-y-4 p-4'>
        {/* Back Button */}
        <div className='flex flex-shrink-0 items-center justify-between'>
          <Button
            onClick={() => setSelectedRecognition(null)}
            variant='outline'
            size='sm'
          >
            ← Back to Face Recognition
          </Button>
        </div>

        {/* Recognition Detail */}
        <div className='grid flex-1 grid-cols-1 gap-6 lg:grid-cols-3'>
          {/* Recognition Image - 2 columns */}
          <div className='flex min-h-0 flex-col lg:col-span-2'>
            <Card className='overflow-hidden'>
              <div className='relative aspect-video w-full'>
                {selectedRecognition.image_url ? (
                  <img
                    src={selectedRecognition.image_url}
                    alt={`Face Recognition ${selectedRecognition.object_name}`}
                    className='h-full w-full bg-gray-900 object-contain'
                  />
                ) : (
                  <div className='flex h-full w-full items-center justify-center bg-gray-900'>
                    <UserCheck className='h-16 w-16 text-gray-400' />
                    <span className='ml-2 text-gray-400'>
                      No image available
                    </span>
                  </div>
                )}

                {/* Recognition Info Overlay */}
                <div className='absolute top-0 right-0 left-0 bg-gradient-to-b from-black/60 via-black/20 to-transparent p-4'>
                  <div className='flex items-start justify-between'>
                    <div>
                      <h3 className='mb-1 text-lg font-semibold text-white drop-shadow-lg'>
                        {selectedRecognition.object_name || 'Unknown Person'}
                      </h3>
                      <p className='text-sm text-gray-200 drop-shadow'>
                        Camera {selectedRecognition.camera_id}
                      </p>
                    </div>
                    <div className='rounded-full bg-green-100 px-3 py-1 dark:bg-green-900/20'>
                      <span className='text-xs font-medium text-green-600 dark:text-green-400'>
                        Detected
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </Card>
          </div>

          {/* Recognition Details Sidebar - 1 column */}
          <Card className='flex min-h-0 flex-col'>
            <CardHeader className='flex-shrink-0 border-b'>
              <CardTitle className='text-base'>Recognition Details</CardTitle>
            </CardHeader>
            <CardContent className='min-h-0 flex-1 space-y-4 p-4'>
              <div>
                <label className='text-sm font-medium text-gray-600 dark:text-gray-400'>
                  Person Name
                </label>
                <p className='mt-1 text-sm text-gray-900 dark:text-white'>
                  {selectedRecognition.object_name || 'Unknown Person'}
                </p>
              </div>

              <div>
                <label className='text-sm font-medium text-gray-600 dark:text-gray-400'>
                  Camera ID
                </label>
                <p className='mt-1 text-sm text-gray-900 dark:text-white'>
                  Camera {selectedRecognition.camera_id}
                </p>
              </div>

              <div>
                <label className='text-sm font-medium text-gray-600 dark:text-gray-400'>
                  Detected At
                </label>
                <p className='mt-1 text-sm text-gray-900 dark:text-white'>
                  {new Date(selectedRecognition.detected_at).toLocaleString()}
                </p>
              </div>

              <div>
                <label className='text-sm font-medium text-gray-600 dark:text-gray-400'>
                  Recognition ID
                </label>
                <p className='mt-1 text-sm font-mono text-gray-900 dark:text-white'>
                  {selectedRecognition.ID}
                </p>
              </div>

              <div>
                <label className='text-sm font-medium text-gray-600 dark:text-gray-400'>
                  Created At
                </label>
                <p className='mt-1 text-sm text-gray-900 dark:text-white'>
                  {new Date(selectedRecognition.created_at).toLocaleString()}
                </p>
              </div>

              <div>
                <label className='text-sm font-medium text-gray-600 dark:text-gray-400'>
                  Last Updated
                </label>
                <p className='mt-1 text-sm text-gray-900 dark:text-white'>
                  {new Date(selectedRecognition.updated_at).toLocaleString()}
                </p>
              </div>
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
          <h1 className='text-3xl font-bold tracking-tight'>Face Recognition</h1>
          <p className='text-muted-foreground'>
            Monitor and manage face recognition from surveillance system
          </p>
        </div>
        <Button
          className='hover:cursor-pointer'
          size='sm'
          onClick={refetch}
          disabled={loading}
        >
          <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          <span>Refresh</span>
        </Button>
      </div>

      {/* Summary Cards */}
      <div className='grid gap-4 md:grid-cols-2 lg:grid-cols-4'>
        <SummaryCard
          title='Total Recognitions'
          value={faceRecognitionStats.totalRecognitions.toString()}
          icon={UserCheck}
          description='All time recognitions'
        />
        <SummaryCard
          title='Today Count'
          value={faceRecognitionStats.todayCount.toString()}
          icon={Users}
          description='Detected today'
        />
        <SummaryCard
          title='Unique Persons'
          value={faceRecognitionStats.uniquePersons.toString()}
          icon={User}
          description='Different individuals'
        />
        <SummaryCard
          title='Recent Detections'
          value={faceRecognitionStats.detectedToday.toString()}
          icon={Eye}
          description='Today only'
        />
      </div>

      {/* Error Message for Partial Errors */}
      {error && recognitions.length > 0 && (
        <ErrorMessage message={error} onRetry={refetch} compact />
      )}

      {/* Face Recognition List */}
      <div className='space-y-4'>
        {loading && recognitions.length === 0 ? (
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
        ) : recognitions.length === 0 ? (
          <div className='py-12 text-center'>
            <UserCheck className='mx-auto mb-4 h-16 w-16 text-gray-400' />
            <h3 className='mb-2 text-lg font-semibold text-gray-900 dark:text-white'>
              No Face Recognitions Found
            </h3>
            <p className='text-gray-600 dark:text-gray-400'>
              No face recognitions have been detected recently.
            </p>
          </div>
        ) : (
          <div className='space-y-3'>
            {recognitions.map((recognition) => (
              <Card
                key={recognition.ID}
                onClick={() => setSelectedRecognition(recognition)}
                className='cursor-pointer transition-all hover:border-blue-300 hover:shadow-md dark:hover:border-blue-600'
              >
                <CardContent className='p-4'>
                  <div className='flex items-center gap-4'>
                    {/* Recognition Thumbnail */}
                    <div className='h-16 w-24 flex-shrink-0 overflow-hidden rounded-md bg-gray-100 dark:bg-gray-800'>
                      {recognition.image_url ? (
                        <img
                          src={recognition.image_url}
                          alt={`Face Recognition ${recognition.object_name}`}
                          className='h-full w-full object-cover'
                          onError={(e) => {
                            const target = e.target as HTMLImageElement
                            target.style.display = 'none'
                            const parent = target.parentElement
                            if (parent) {
                              parent.innerHTML = `
                                <div class="flex h-full w-full items-center justify-center bg-gray-200 dark:bg-gray-700">
                                  <svg class="h-6 w-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                                  </svg>
                                </div>
                              `
                            }
                          }}
                        />
                      ) : (
                        <div className='flex h-full w-full items-center justify-center bg-gray-200 dark:bg-gray-700'>
                          <UserCheck className='h-6 w-6 text-gray-400' />
                        </div>
                      )}
                    </div>

                    {/* Recognition Info */}
                    <div className='min-w-0 flex-1'>
                      <div className='flex items-start justify-between'>
                        <div className='min-w-0 flex-1'>
                          <h4 className='truncate text-sm font-semibold text-gray-900 dark:text-white'>
                            {recognition.object_name || 'Unknown Person'}
                          </h4>
                          <p className='mt-1 truncate text-sm text-gray-600 dark:text-gray-400'>
                            Face detected and recognized
                          </p>
                          <div className='mt-2 flex items-center gap-4 text-xs text-gray-500 dark:text-gray-500'>
                            <span className='flex items-center gap-1'>
                              <Camera className='h-3 w-3' />
                              Camera {recognition.camera_id}
                            </span>
                            <span className='flex items-center gap-1'>
                              <Clock className='h-3 w-3' />
                              {new Date(recognition.detected_at).toLocaleString()}
                            </span>
                          </div>
                        </div>

                        {/* Status */}
                        <div className='ml-4 flex flex-col items-end gap-2'>
                          <div className='rounded-full bg-green-100 px-2 py-1 dark:bg-green-900/20'>
                            <span className='text-xs font-medium text-green-600 dark:text-green-400'>
                              Detected
                            </span>
                          </div>
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
        Showing {recognitions.length} of {total} face recognitions
      </div>
    </div>
  )
}

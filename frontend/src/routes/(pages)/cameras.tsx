import { useMemo, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import type { Camera as CameraType } from '@/types/camera'
import {
  AlertTriangle,
  Camera,
  Maximize,
  RefreshCw,
} from 'lucide-react'
import { useCameras } from '@/hooks/use-cameras'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ErrorMessage } from '@/components/ErrorMessage'
import { LoadingSpinner } from '@/components/LoadingSpinner'
import { CameraStream } from '@/features/camera/CameraStream'

export const Route = createFileRoute('/(pages)/cameras')({
  component: CamerasPage,
})

// Convert API camera to display format
const convertApiCameraToDisplay = (apiCamera: CameraType) => ({
  id: apiCamera.id,
  name: apiCamera.name,
  location: apiCamera.location,
  status:
    apiCamera.status === 'active'
      ? 'online'
      : apiCamera.status === 'inactive'
        ? 'offline'
        : apiCamera.status,
  recording: apiCamera.status === 'active',
  audio: true, // Default assumption
  alerts: 0, // Could be enhanced with alert count from API
  ip_address: apiCamera.ip_address,
  ws_url: apiCamera.ws_url,
  stream_url: apiCamera.stream_url,
  image_url: apiCamera.image_url,
  apiCamera: apiCamera, // Keep reference to original API data
})

type DisplayCameraType = {
  id: number
  name: string
  location: string
  status: 'online' | 'offline' | 'maintenance'
  recording: boolean
  audio: boolean
  alerts: number
  ip_address: string
  stream_url: string
  image_url: string
  apiCamera: CameraType
}

function CamerasPage() {
  const [selectedCamera, setSelectedCamera] = useState<DisplayCameraType | any>(
    null
  )

  const {
    cameras: apiCameras,
    loading,
    error,
    refetch,
  } = useCameras()

  // Convert API cameras to display format
  const cameras = useMemo(() => {
    return apiCameras.map(convertApiCameraToDisplay)
  }, [apiCameras])

  // Camera statistics
  const cameraStats = useMemo(() => {
    const total = cameras.length
    const online = cameras.filter((c) => c.status === 'online').length
    const offline = cameras.filter((c) => c.status === 'offline').length
    const maintenance = cameras.filter((c) => c.status === 'maintenance').length

    return { total, online, offline, maintenance }
  }, [cameras])

  // Loading state
  if (loading && cameras.length === 0) {
    return <LoadingSpinner message='Loading cameras...' />
  }

  // Error state
  if (error && cameras.length === 0) {
    return <ErrorMessage message={error} onRetry={refetch} />
  }

  // Single Camera Preview Mode
  if (selectedCamera) {
    return (
      <div className='mx-auto w-full space-y-4 p-4'>
        {/* Back Button */}
        <div className='flex flex-shrink-0 items-center justify-between'>
          <Button
            onClick={() => setSelectedCamera(null)}
            variant='outline'
            size='sm'
          >
            ← Back to Grid
          </Button>
        </div>

        {/* Main Content */}
        <div className='grid flex-1 grid-cols-1 gap-6 lg:grid-cols-6'>
          {/* Video Player - 4 columns */}
          <div className='flex min-h-0 flex-col lg:col-span-4'>
            <Card className='overflow-hidden'>
              {/* 16:9 Video Container */}
              <div
                className='relative w-full'
                style={{ paddingBottom: '56.25%' }}
              >
                <div className='absolute inset-0 flex items-center justify-center bg-gray-900'>
                  {/* WebSocket camera stream with fallback */}
                  <CameraStream
                    wsUrl={selectedCamera.apiCamera.ws_url}
                    fallbackImageUrl={selectedCamera.apiCamera.stream_url}
                    alt={`Camera ${selectedCamera.name}`}
                    className='absolute inset-0 h-full w-full object-contain'
                    enabled={true}
                    showStatus={true}
                    size='large'
                    cameraId={selectedCamera.id}
                  />

                  {/* Header Overlay */}
                  <div className='absolute top-0 right-0 left-0 bg-gradient-to-b from-black/60 via-black/20 to-transparent p-4'>
                    <div className='flex items-start justify-between'>
                      <div>
                        <h3 className='mb-1 text-lg font-semibold text-white drop-shadow-lg'>
                          {selectedCamera.name}
                        </h3>
                        <p className='text-sm text-gray-200 drop-shadow'>
                          {selectedCamera.location}
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* Controls Footer */}
                  <div className='absolute right-0 bottom-0 left-0 bg-gradient-to-t from-black/60 via-black/20 to-transparent p-4'>
                    <div className='flex items-center justify-between'>
                      {selectedCamera.alerts > 0 && (
                        <div className='flex items-center gap-2'>
                          <AlertTriangle className='h-4 w-4 text-red-400' />
                          <span className='text-sm font-medium text-white'>
                            {selectedCamera.alerts} Alert
                            {selectedCamera.alerts > 1 ? 's' : ''}
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </Card>
          </div>

          {/* Camera Sidebar - 2 columns */}
          <Card className='flex min-h-0 flex-col lg:col-span-2'>
            {/* Sidebar Header */}
            <CardHeader className='flex-shrink-0 border-b'>
              <CardTitle className='text-base'>All Cameras</CardTitle>
              <p className='mt-1 text-xs text-gray-600 dark:text-gray-400'>
                {cameraStats.online} of {cameraStats.total} cameras online
              </p>
            </CardHeader>

            {/* Camera List */}
            <CardContent className='min-h-0 flex-1 space-y-3 overflow-y-auto p-3'>
              {cameras.map((camera) => (
                <div
                  key={camera.id}
                  onClick={() => setSelectedCamera(camera)}
                  className={`cursor-pointer rounded-sm border p-3 transition-all hover:shadow-sm ${
                    camera.id === selectedCamera.id
                      ? 'border-blue-200 bg-blue-50 ring-1 ring-blue-500 dark:border-blue-700 dark:bg-blue-900/20'
                      : 'border-gray-200 hover:bg-gray-50 dark:border-gray-600 dark:hover:bg-gray-700'
                  }`}
                >
                  <div className='flex gap-3'>
                    {/* Thumbnail */}
                    <div className='relative flex h-18 w-24 flex-shrink-0 items-center justify-center overflow-hidden rounded bg-gray-900'>
                      <CameraStream
                        wsUrl={camera.apiCamera.ws_url}
                        fallbackImageUrl={camera.apiCamera.image_url}
                        alt={`Camera ${camera.name}`}
                        className='h-full w-full object-cover'
                        enabled={true}
                        showStatus={false}
                        size='small'
                        cameraId={camera.id}
                      />

                      {/* Camera ID */}
                      <div className='absolute right-0 bottom-0 left-0 bg-black/80 px-1 py-0.5 text-center text-[10px] font-medium text-white'>
                        CAM-{camera.id.toString().padStart(2, '0')}
                      </div>
                    </div>

                    {/* Camera Info */}
                    <div className='min-w-0 flex-1'>
                      <h4 className='mb-1 truncate text-sm font-medium dark:text-white'>
                        {camera.name}
                      </h4>
                      <p className='mb-2 truncate text-xs text-gray-600 dark:text-gray-400'>
                        {camera.location}
                      </p>
                      <div className='flex gap-1'>
                        <span
                          className={`inline-flex w-fit items-center rounded-full px-2 py-0.5 text-[10px] font-medium ${
                            camera.status === 'online'
                              ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                              : camera.status === 'maintenance'
                                ? 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400'
                                : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                          }`}
                        >
                          {camera.status.toUpperCase()}
                        </span>
                        {camera.alerts > 0 && (
                          <span className='inline-flex w-fit items-center rounded-full bg-red-100 px-2 py-0.5 text-[10px] font-medium text-red-800 dark:bg-red-900/30 dark:text-red-400'>
                            <AlertTriangle className='mr-1 h-2.5 w-2.5' />
                            {camera.alerts} alert{camera.alerts > 1 ? 's' : ''}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
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
          <h1 className='text-3xl font-bold tracking-tight'>
            Camera Management
          </h1>
          <p className='text-muted-foreground'>
            Monitor and manage surveillance cameras
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

      {/* Error Message for Partial Errors */}
      {error && cameras.length > 0 && (
        <ErrorMessage message={error} onRetry={refetch} compact />
      )}

      {/* Camera Grid */}
      <div className='space-y-4'>
        {loading && cameras.length === 0 ? (
          <div className='grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'>
            {Array.from({ length: 8 }).map((_, index) => (
              <Card key={index} className='overflow-hidden'>
                <div
                  className='relative w-full animate-pulse bg-gray-300 dark:bg-gray-600'
                  style={{ paddingBottom: '56.25%' }}
                />
                <CardContent className='p-3'>
                  <div className='flex items-center justify-between'>
                    <div className='h-4 w-20 animate-pulse rounded bg-gray-300 dark:bg-gray-600' />
                    <div className='h-3 w-16 animate-pulse rounded bg-gray-300 dark:bg-gray-600' />
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        ) : cameras.length === 0 ? (
          <div className='py-12 text-center'>
            <Camera className='mx-auto mb-4 h-16 w-16 text-gray-400' />
            <h3 className='mb-2 text-lg font-semibold text-gray-900 dark:text-white'>
              No Cameras Available
            </h3>
            <p className='text-gray-600 dark:text-gray-400'>
              Get started by adding your first camera to the system.
            </p>
          </div>
        ) : (
          <div className='grid grid-cols-1 content-start gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'>
            {cameras.map((camera) => (
              <Card
                key={camera.id}
                onClick={() => setSelectedCamera(camera)}
                className='cursor-pointer overflow-hidden p-0 transition-all hover:border-cyan-300 hover:shadow-md dark:hover:border-cyan-600'
              >
                {/* 16:9 Video Container */}
                <div
                  className='relative w-full'
                  style={{ paddingBottom: '56.25%' }}
                >
                  <div className='absolute inset-0 flex items-center justify-center bg-gray-900'>
                    {/* WebSocket camera stream with fallback */}
                    <CameraStream
                      wsUrl={camera.apiCamera.ws_url}
                      fallbackImageUrl={camera.apiCamera.image_url}
                      alt={`Camera ${camera.name}`}
                      className='absolute inset-0 h-full w-full object-contain'
                      enabled={true}
                      showStatus={false}
                      size='large'
                      cameraId={camera.id}
                    />

                    {/* Title Overlay */}
                    <div className='absolute top-3 right-3 left-3'>
                      <h4 className='truncate text-sm font-medium text-white drop-shadow-lg'>
                        {camera.name}
                      </h4>
                      <p className='mt-0.5 truncate text-xs text-gray-200 drop-shadow'>
                        {camera.location}
                      </p>
                    </div>

                    {/* Top Right Info */}
                    <div className='absolute top-3 right-3 flex flex-col items-end gap-1'>
                      <div className='flex items-center gap-1'>
                        <span className='rounded bg-black/70 px-2 py-0.5 text-xs font-medium text-white'>
                          CAM-{camera.id.toString().padStart(2, '0')}
                        </span>
                      </div>
                    </div>

                    {/* Bottom Right Alerts */}
                    {camera.alerts > 0 && (
                      <div className='absolute right-3 bottom-3'>
                        <span className='inline-flex items-center rounded-full bg-red-600 px-2 py-0.5 text-xs font-medium text-white'>
                          {camera.alerts} Alert{camera.alerts > 1 ? 's' : ''}
                        </span>
                      </div>
                    )}

                    {/* Hover Overlay */}
                    <div className='absolute inset-0 flex items-center justify-center bg-black/20 opacity-0 transition-opacity hover:opacity-100'>
                      <Maximize className='h-8 w-8 text-white drop-shadow-lg' />
                    </div>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* Results Info */}

    </div>
  )
}

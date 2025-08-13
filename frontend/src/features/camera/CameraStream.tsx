import { useState } from 'react'
import { Camera, Wifi, WifiOff } from 'lucide-react'
import { useCameraStream } from '@/hooks/use-camera-stream'

interface CameraStreamProps {
  wsUrl?: string
  fallbackImageUrl?: string
  alt: string
  className?: string
  enabled?: boolean
  showStatus?: boolean
  size?: 'small' | 'large' // Add size prop for different layouts
  cameraId?: number // Add camera ID for filtering WebSocket messages
}

export function CameraStream({
  wsUrl,
  fallbackImageUrl,
  alt,
  className = 'absolute inset-0 w-full h-full object-cover',
  enabled = true,
  showStatus = false,
  size = 'large',
  cameraId,
}: CameraStreamProps) {
  const { imageSrc, isConnected, isConnecting, error, usingWebSocket } =
    useCameraStream({
      wsUrl,
      fallbackImageUrl,
      enabled,
      cameraId,
    })

  const [imageError, setImageError] = useState(false)

  const handleImageError = (e: React.SyntheticEvent<HTMLImageElement>) => {
    // Hide the image if it fails to load
    e.currentTarget.style.display = 'none'
    setImageError(true)
  }

  return (
    <>
      {/* Main image - Always show image if available, fallback if not */}
      {imageSrc && !imageError ? (
        <img
          src={imageSrc}
          alt={alt}
          className={className}
          onError={handleImageError}
        />
      ) : (
        // Placeholder when no image is available or image failed to load
        <div className='absolute inset-0 flex flex-col items-center justify-center bg-gray-800 text-gray-400'>
          {size === 'large' ? (
            <>
              <Camera className='mb-2 h-16 w-16' />
              <span className='text-sm font-medium'>
                {imageError
                  ? 'Unavailable'
                  : isConnecting
                    ? 'Connecting...'
                    : 'No signal'}
              </span>
            </>
          ) : (
            // Small size placeholder for thumbnails
            <Camera className='h-8 w-8 text-gray-500' />
          )}
        </div>
      )}

      {/* Connection status indicator */}
      {showStatus && wsUrl && (
        <div className='absolute top-2 right-2 flex items-center gap-1'>
          {isConnecting ? (
            <div className='inline-flex items-center rounded-full bg-yellow-600 p-2 text-white'>
              <div className='h-2 w-2 animate-pulse rounded-full bg-white' />
            </div>
          ) : isConnected ? (
            <div className='inline-flex items-center rounded-full bg-green-600 p-1.5 text-white'>
              <Wifi className='h-3 w-3' />
            </div>
          ) : usingWebSocket === false && wsUrl ? (
            <div className='inline-flex items-center rounded-full bg-red-600 p-1.5 text-white'>
              <WifiOff className='h-3 w-3' />
            </div>
          ) : null}
        </div>
      )}

      {/* Error indicator */}
      {error && showStatus && (
        <div className='absolute bottom-2 left-2 rounded bg-black/60 px-2 py-1 text-xs text-red-400'>
          Connection error
        </div>
      )}
    </>
  )
}

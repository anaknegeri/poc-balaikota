import { useCallback, useEffect, useState } from 'react'
import { config } from '@/config/config'
import { cameraApiService } from '@/services/cameraService'
import type {
  Camera,
  CreateCameraRequest,
  UpdateCameraRequest,
  UpdateCameraStatusRequest,
} from '@/types/camera'

export interface CamerasState {
  cameras: Camera[]
  loading: boolean
  error: string | null
  lastUpdated: Date | null
}

export const useCameras = (statusFilter?: string, refreshInterval?: number) => {
  const actualRefreshInterval =
    refreshInterval ?? config.REFRESH_INTERVAL.DASHBOARD

  const [state, setState] = useState<CamerasState>({
    cameras: [],
    loading: true,
    error: null,
    lastUpdated: null,
  })

  const fetchCameras = useCallback(async () => {
    try {
      setState((prev) => ({
        ...prev,
        loading: prev.cameras.length === 0,
        error: null,
      }))

      const cameras = await cameraApiService.getCameras(statusFilter)

      setState({
        cameras,
        loading: false,
        error: null,
        lastUpdated: new Date(),
      })
    } catch (error) {
      setState((prev) => ({
        ...prev,
        loading: false,
        error:
          error instanceof Error ? error.message : 'Unknown error occurred',
      }))
    }
  }, [statusFilter])

  // Initial fetch
  useEffect(() => {
    fetchCameras()
  }, [fetchCameras])

  // Auto refresh
  useEffect(() => {
    if (actualRefreshInterval > 0) {
      const interval = setInterval(fetchCameras, actualRefreshInterval)
      return () => clearInterval(interval)
    }
  }, [fetchCameras, actualRefreshInterval])

  return {
    ...state,
    refetch: fetchCameras,
  }
}

// Hook for single camera management
export const useCamera = (id: number) => {
  const [camera, setCamera] = useState<Camera | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchCamera = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const cameraData = await cameraApiService.getCameraById(id)
      setCamera(cameraData)
    } catch (error) {
      setError(
        error instanceof Error ? error.message : 'Unknown error occurred'
      )
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    if (id) {
      fetchCamera()
    }
  }, [fetchCamera, id])

  return { camera, loading, error, refetch: fetchCamera }
}

// Hook for camera operations (CRUD)
export const useCameraOperations = () => {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const createCamera = useCallback(
    async (cameraData: CreateCameraRequest): Promise<Camera> => {
      try {
        setLoading(true)
        setError(null)
        const newCamera = await cameraApiService.createCamera(cameraData)
        return newCamera
      } catch (error) {
        const errorMessage =
          error instanceof Error ? error.message : 'Failed to create camera'
        setError(errorMessage)
        throw new Error(errorMessage)
      } finally {
        setLoading(false)
      }
    },
    []
  )

  const updateCamera = useCallback(
    async (id: number, cameraData: UpdateCameraRequest): Promise<Camera> => {
      try {
        setLoading(true)
        setError(null)
        const updatedCamera = await cameraApiService.updateCamera(
          id,
          cameraData
        )
        return updatedCamera
      } catch (error) {
        const errorMessage =
          error instanceof Error ? error.message : 'Failed to update camera'
        setError(errorMessage)
        throw new Error(errorMessage)
      } finally {
        setLoading(false)
      }
    },
    []
  )

  const updateCameraStatus = useCallback(
    async (id: number, status: UpdateCameraStatusRequest): Promise<void> => {
      try {
        setLoading(true)
        setError(null)
        await cameraApiService.updateCameraStatus(id, status)
      } catch (error) {
        const errorMessage =
          error instanceof Error
            ? error.message
            : 'Failed to update camera status'
        setError(errorMessage)
        throw new Error(errorMessage)
      } finally {
        setLoading(false)
      }
    },
    []
  )

  const deleteCamera = useCallback(async (id: number): Promise<void> => {
    try {
      setLoading(true)
      setError(null)
      await cameraApiService.deleteCamera(id)
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to delete camera'
      setError(errorMessage)
      throw new Error(errorMessage)
    } finally {
      setLoading(false)
    }
  }, [])

  const clearError = useCallback(() => {
    setError(null)
  }, [])

  return {
    createCamera,
    updateCamera,
    updateCameraStatus,
    deleteCamera,
    loading,
    error,
    clearError,
  }
}

// Hook for camera stream status
export const useCameraStream = (cameraId: number) => {
  const [streamStatus, setStreamStatus] = useState<
    'loading' | 'connected' | 'error'
  >('loading')
  const [imageError, setImageError] = useState(false)

  const checkStreamStatus = useCallback(() => {
    const img = new Image()
    img.onload = () => setStreamStatus('connected')
    img.onerror = () => setStreamStatus('error')
    img.src = cameraApiService.getCameraImageUrl(cameraId)
  }, [cameraId])

  useEffect(() => {
    checkStreamStatus()

    // Check stream status every 30 seconds
    const interval = setInterval(checkStreamStatus, 30000)
    return () => clearInterval(interval)
  }, [checkStreamStatus])

  const handleImageError = useCallback(() => {
    setImageError(true)
    setStreamStatus('error')
  }, [])

  const handleImageLoad = useCallback(() => {
    setImageError(false)
    setStreamStatus('connected')
  }, [])

  return {
    streamStatus,
    imageError,
    handleImageError,
    handleImageLoad,
    retryStream: checkStreamStatus,
  }
}

import { config } from '@/config/config'
import type {
  Camera,
  CameraApiResponse,
  CreateCameraRequest,
  UpdateCameraRequest,
  UpdateCameraStatusRequest,
} from '@/types/camera'

export const cameraApiService = {
  // Get all cameras
  async getCameras(status?: string): Promise<Camera[]> {
    const url = new URL(`${config.API_BASE_URL}/cameras`)
    if (status) {
      url.searchParams.append('status', status)
    }

    const response = await fetch(url.toString())
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const result: CameraApiResponse = await response.json()
    if (result.error) {
      throw new Error(result.msg || 'Failed to fetch cameras')
    }

    return result.data
  },

  // Get camera by ID
  async getCameraById(id: number): Promise<Camera> {
    const response = await fetch(`${config.API_BASE_URL}/cameras/${id}`)
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const result = await response.json()
    if (result.error) {
      throw new Error(result.msg || 'Failed to fetch camera')
    }

    return result.data
  },

  // Create new camera
  async createCamera(camera: CreateCameraRequest): Promise<Camera> {
    const response = await fetch(`${config.API_BASE_URL}/cameras`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(camera),
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const result = await response.json()
    if (result.error) {
      throw new Error(result.msg || 'Failed to create camera')
    }

    return result.data
  },

  // Update camera
  async updateCamera(id: number, camera: UpdateCameraRequest): Promise<Camera> {
    const response = await fetch(`${config.API_BASE_URL}/cameras/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(camera),
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const result = await response.json()
    if (result.error) {
      throw new Error(result.msg || 'Failed to update camera')
    }

    return result.data
  },

  // Update camera status only
  async updateCameraStatus(
    id: number,
    statusUpdate: UpdateCameraStatusRequest
  ): Promise<void> {
    const response = await fetch(
      `${config.API_BASE_URL}/cameras/${id}/status`,
      {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(statusUpdate),
      }
    )

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const result = await response.json()
    if (result.error) {
      throw new Error(result.msg || 'Failed to update camera status')
    }
  },

  // Delete camera
  async deleteCamera(id: number): Promise<void> {
    const response = await fetch(`${config.API_BASE_URL}/cameras/${id}`, {
      method: 'DELETE',
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const result = await response.json()
    if (result.error) {
      throw new Error(result.msg || 'Failed to delete camera')
    }
  },

  // Get camera stream URL
  getCameraStreamUrl(id: number): string {
    return `${config.API_BASE_URL}/cameras/${id}/stream`
  },

  // Get camera image URL
  getCameraImageUrl(id: number): string {
    return `${config.API_BASE_URL}/cameras/${id}/image`
  },
}

// Utility functions
export const getCameraStatusColor = (status: string): string => {
  switch (status) {
    case 'active':
      return 'bg-emerald-500'
    case 'maintenance':
      return 'bg-amber-500'
    case 'inactive':
    case 'issue':
      return 'bg-red-500'
    default:
      return 'bg-gray-500'
  }
}

export const getCameraStatusTextColor = (status: string): string => {
  switch (status) {
    case 'active':
      return 'text-emerald-600 dark:text-emerald-400'
    case 'maintenance':
      return 'text-amber-600 dark:text-amber-400'
    case 'inactive':
    case 'issue':
      return 'text-red-600 dark:text-red-400'
    default:
      return 'text-gray-600 dark:text-gray-400'
  }
}

export const getCameraStatusIcon = (status: string): string => {
  switch (status) {
    case 'active':
      return '●'
    case 'maintenance':
      return '⚠'
    case 'inactive':
    case 'issue':
      return '●'
    default:
      return '●'
  }
}

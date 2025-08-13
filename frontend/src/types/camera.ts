export interface Camera {
  id: number
  name: string
  ip_address: string
  location: string
  status: 'active' | 'inactive' | 'maintenance' | 'issue'
  stream_url: string
  ws_url: string
  image_url: string
  created_at: string
  updated_at: string
}

export interface CameraApiResponse {
  error: boolean
  count: number
  data: Camera[]
  msg?: string
}

export interface CreateCameraRequest {
  name: string
  ip_address: string
  location: string
  status: 'active' | 'inactive' | 'maintenance' | 'issue'
}

export interface UpdateCameraRequest {
  name?: string
  ip_address?: string
  location?: string
  status?: 'active' | 'inactive' | 'maintenance' | 'issue'
}

export interface UpdateCameraStatusRequest {
  status: 'active' | 'inactive' | 'maintenance' | 'issue'
}

export interface UseCameraStreamProps {
  wsUrl?: string
  fallbackImageUrl?: string
  cameraId?: number
  enabled?: boolean
}

export interface CameraStreamState {
  imageSrc: string | null
  isConnected: boolean
  isConnecting: boolean
  error: string | null
  usingWebSocket: boolean
}

export interface PeopleCountSummary {
  cameras: {
    camera_id: number
    camera_name: string
    male_count: number
    female_count: number
    child_count: number
    adult_count: number
    elderly_count: number
    total_count: number
    last_updated: string
  }[]
  totals: {
    male: number
    female: number
    child: number
    adult: number
    elderly: number
    total: number
  }
}

export interface VehicleCountSummary {
  cctvs: {
    cctv_id: number
    cctv_name: string
    in_count_car: number
    in_count_truck: number
    in_count_people: number
    out_count: number
    total_in_count: number
    total_vehicle_in_count: number
    net_count: number
    last_updated: string
  }[]
  totals: {
    in_car: number
    in_truck: number
    in_people: number
    out: number
    total_in: number
    total_vehicle_in: number
    net_count: number
  }
}

export interface AlertData {
  count: number
  data: Alert[]
}
export interface Alert {
  id: string
  alert_type_id: number
  camera_id: number
  message: string
  severity: string
  is_active: boolean
  detected_at: string
  resolved_at: string | null
  resolved_by: string
  resolution_note: string
  image_path: string
  image_url: string
  created_at: string
  updated_at: string
  alert_type?: {
    id: number
    name: string
    display_name: string
    icon: string
    color: string
    description: string
  }
  camera?: {
    id: number
    name: string
    location: string
    status: string
  }
}

export interface FaceRecognitionData {
  count: number
  data: FaceRecognition[]
  error: boolean
  page: number
  pages: number
  total: number
}

export interface FaceRecognition {
  ID: string
  camera_id: number
  image_path: string
  image_url: string
  object_name: string
  detected_at: string
  created_at: string
  updated_at: string
}

export interface PeopleCountTrend {
  time_period: string
  male_count: number
  female_count: number
  total_count: number
  child_count: number
  adult_count: number
  elderly_count: number
}

export interface VehicleCountTrend {
  time_period: string
  in_count_car: number
  in_count_truck: number
  in_count_people: number
  out_count: number
  total_in_count: number
  total_vehicle_in_count: number
}

// Peak Hours Analysis Interfaces
export interface PeakHourPoint {
  hour: number
  time_label: string
  visitors: number
  period: string // "peak" | "high" | "moderate" | "low"
  percentage: number
}

export interface PeakHoursSummary {
  peak_hour: number
  peak_count: number
  low_hour: number
  low_count: number
  average_per_hour: number
  total_visitors: number
  peak_vs_low_ratio: number
}

export interface PeakHoursInsights {
  busiest_period: string
  quietest_period: string
  recommended_staffing_hours: string[]
  traffic_pattern: string
}

export interface PeakHoursAnalysis {
  data: PeakHourPoint[]
  summary: PeakHoursSummary
  insights: PeakHoursInsights
}

// Date range interface
export interface DateRange {
  from: string
  to: string
}

export interface ApiResponse<T> {
  error: boolean
  count?: number
  count_all_active?: number
  total?: number
  page?: number
  pages?: number
  data: T
  msg?: string
}

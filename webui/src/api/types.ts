export interface ApiError {
  code?: string
  detail?: unknown
  i18n_key?: string
  type?: string
}

export interface ApiResponse<T> {
  success: boolean
  code: number
  message: string
  data?: T
  error?: ApiError
  timestamp: number
}

export interface LoginPayload {
  user_code: string
  password: string
}

export interface LoginData {
  user_id: number
  module_code: string
  roles: string[]
  access_token: string
  refresh_token: string
}

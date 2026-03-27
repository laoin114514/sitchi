import axios from 'axios'

import type { ApiResponse } from '@/api/types'

const mode = import.meta.env.MODE
const devBaseURL = import.meta.env.VITE_API_BASE_URL_DEV || 'http://127.0.0.1:5090/api/webui'
const prodBaseURL = import.meta.env.VITE_API_BASE_URL_PROD || '/api/webui'

const http = axios.create({
  baseURL: mode === 'prod' ? prodBaseURL : devBaseURL,
  timeout: 10000,
})

http.interceptors.response.use(
  (response) => response,
  (error) => Promise.reject(error),
)

export async function postJson<T, P extends object>(
  url: string,
  payload: P,
): Promise<ApiResponse<T>> {
  const response = await http.post<ApiResponse<T>>(url, payload)
  if (!response.data.success) {
    throw new Error(response.data.message || 'request failed')
  }

  return response.data
}

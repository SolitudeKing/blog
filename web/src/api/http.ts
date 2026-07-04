import axios from 'axios'
import type { ApiListResponse, ApiResponse } from './types'
import { ApiError } from './types'

const API_VERSION = 'v1'

export const http = axios.create({
  baseURL: '/',
  timeout: 15000,
  headers: {
    'X-API-Version': API_VERSION,
  },
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  config.headers['X-API-Version'] = API_VERSION
  return config
})

export async function request<T>(config: Parameters<typeof http.request>[0]): Promise<T> {
  const response = await http.request<ApiResponse<T>>(config)
  const body = response.data
  if (body.code !== 0) {
    throw new ApiError(body.code, body.message)
  }
  return body.data
}

export async function requestList<T>(config: Parameters<typeof http.request>[0]): Promise<ApiListResponse<T>> {
  const response = await http.request<ApiListResponse<T>>(config)
  const body = response.data
  if (body.code !== 0) {
    throw new ApiError(body.code, body.message)
  }
  return body
}

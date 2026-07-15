import axios from 'axios'
import type { AxiosError, InternalAxiosRequestConfig } from 'axios'
import type { ApiListResponse, ApiResponse } from './types'
import { ApiError } from './types'
import { clearStoredSession, emitUnauthorized, getStoredSession, storeTokenPair } from './session'
import type { TokenPair } from '@/types/auth'

const API_VERSION = 'v1'

interface RetriableRequestConfig extends InternalAxiosRequestConfig {
  _authRetry?: boolean
}

export const http = axios.create({
  baseURL: '/',
  timeout: 15000,
  headers: {
    'X-API-Version': API_VERSION,
  },
})

const refreshHttp = axios.create({
  baseURL: '/',
  timeout: 15000,
  headers: {
    'X-API-Version': API_VERSION,
  },
})

let refreshPromise: Promise<string> | null = null

function isAuthEndpoint(url?: string) {
  if (!url) {
    return false
  }

  try {
    const pathname = new URL(url, window.location.origin).pathname.replace(/\/$/, '')
    return pathname === '/auth/login' || pathname === '/auth/refresh'
  } catch {
    return false
  }
}

function refreshAccessToken() {
  if (refreshPromise) {
    return refreshPromise
  }

  const { refreshToken } = getStoredSession()
  if (!refreshToken) {
    invalidateSession()
    return Promise.reject(new ApiError(10000, 'unauthorized'))
  }

  refreshPromise = refreshHttp
    .post<ApiResponse<TokenPair>>('auth/refresh', { refresh_token: refreshToken })
    .then(({ data }) => {
      if (data.code !== 0) {
        throw new ApiError(data.code, data.message)
      }
      storeTokenPair(data.data)
      return data.data.access_token
    })
    .catch((error: unknown) => {
      invalidateSession()
      throw error
    })
    .finally(() => {
      refreshPromise = null
    })

  return refreshPromise
}

function invalidateSession() {
  const session = getStoredSession()
  const hadSession = Boolean(session.accessToken || session.refreshToken)
  clearStoredSession()
  if (hadSession) {
    emitUnauthorized()
  }
}

http.interceptors.request.use((config) => {
  const { accessToken } = getStoredSession()
  if (accessToken && !isAuthEndpoint(config.url)) {
    config.headers.Authorization = `Bearer ${accessToken}`
  }
  config.headers['X-API-Version'] = API_VERSION
  return config
})

http.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiResponse<unknown>>) => {
    const requestConfig = error.config as RetriableRequestConfig | undefined
    const isUnauthorized = error.response?.status === 401

    if (!isUnauthorized || !requestConfig || isAuthEndpoint(requestConfig.url)) {
      return Promise.reject(error)
    }

    if (requestConfig._authRetry) {
      invalidateSession()
      return Promise.reject(error)
    }

    requestConfig._authRetry = true

    try {
      const accessToken = await refreshAccessToken()
      requestConfig.headers.Authorization = `Bearer ${accessToken}`
      return await http.request(requestConfig)
    } catch (refreshError) {
      return Promise.reject(refreshError)
    }
  },
)

function normalizeApiError(error: unknown) {
  if (axios.isAxiosError<ApiResponse<unknown>>(error)) {
    const body = error.response?.data
    if (body && typeof body.code === 'number' && typeof body.message === 'string') {
      return new ApiError(body.code, body.message)
    }
  }
  return error
}

export async function request<T>(config: Parameters<typeof http.request>[0]): Promise<T> {
  let response
  try {
    response = await http.request<ApiResponse<T>>(config)
  } catch (error) {
    throw normalizeApiError(error)
  }
  const body = response.data
  if (body.code !== 0) {
    throw new ApiError(body.code, body.message)
  }
  return body.data
}

export async function requestList<T>(config: Parameters<typeof http.request>[0]): Promise<ApiListResponse<T>> {
  let response
  try {
    response = await http.request<ApiListResponse<T>>(config)
  } catch (error) {
    throw normalizeApiError(error)
  }
  const body = response.data
  if (body.code !== 0) {
    throw new ApiError(body.code, body.message)
  }
  return body
}

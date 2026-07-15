import type { TokenPair } from '@/types/auth'

const ACCESS_TOKEN_KEY = 'access_token'
const REFRESH_TOKEN_KEY = 'refresh_token'
const ACCESS_TOKEN_EXPIRES_AT_KEY = 'access_token_expires_at'

export const AUTH_UNAUTHORIZED_EVENT = 'solitude:auth-unauthorized'

export interface SessionSnapshot {
  accessToken: string
  refreshToken: string
}

type SessionListener = (session: SessionSnapshot) => void

const listeners = new Set<SessionListener>()

function readStorage(key: string) {
  try {
    return window.localStorage.getItem(key) ?? ''
  } catch {
    return ''
  }
}

function notifySessionChanged() {
  const session = getStoredSession()
  listeners.forEach((listener) => listener(session))
}

export function getStoredSession(): SessionSnapshot {
  return {
    accessToken: readStorage(ACCESS_TOKEN_KEY),
    refreshToken: readStorage(REFRESH_TOKEN_KEY),
  }
}

export function storeTokenPair(tokens: TokenPair) {
  try {
    window.localStorage.setItem(ACCESS_TOKEN_KEY, tokens.access_token)
    window.localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refresh_token)
    window.localStorage.setItem(ACCESS_TOKEN_EXPIRES_AT_KEY, tokens.expires_at)
  } finally {
    notifySessionChanged()
  }
}

export function clearStoredSession() {
  try {
    window.localStorage.removeItem(ACCESS_TOKEN_KEY)
    window.localStorage.removeItem(REFRESH_TOKEN_KEY)
    window.localStorage.removeItem(ACCESS_TOKEN_EXPIRES_AT_KEY)
  } catch {
    // Storage can be unavailable in hardened browser contexts.
  } finally {
    notifySessionChanged()
  }
}

export function subscribeToSession(listener: SessionListener) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

export function emitUnauthorized() {
  window.dispatchEvent(new Event(AUTH_UNAUTHORIZED_EVENT))
}

import { computed, onScopeDispose, ref } from 'vue'
import { defineStore } from 'pinia'
import { getUserInfo, login as loginRequest } from '@/api/modules/auth'
import {
  clearStoredSession,
  getStoredSession,
  storeTokenPair,
  subscribeToSession,
} from '@/api/session'
import type { LoginPayload, UserInfo } from '@/types/auth'

export const useAuthStore = defineStore('auth', () => {
  const storedSession = getStoredSession()
  const accessToken = ref(storedSession.accessToken)
  const refreshToken = ref(storedSession.refreshToken)
  const user = ref<UserInfo | null>(null)
  const hasSession = computed(() => Boolean(accessToken.value || refreshToken.value))

  let sessionCheck: Promise<boolean> | null = null

  function syncFromStorage() {
    const session = getStoredSession()
    accessToken.value = session.accessToken
    refreshToken.value = session.refreshToken
    if (!session.accessToken && !session.refreshToken) {
      user.value = null
    }
  }

  const unsubscribe = subscribeToSession(syncFromStorage)
  onScopeDispose(unsubscribe)

  function clearSession() {
    user.value = null
    clearStoredSession()
  }

  async function login(payload: LoginPayload) {
    const tokens = await loginRequest(payload)
    storeTokenPair(tokens)
    user.value = await getUserInfo()
  }

  async function restoreSession(force = false) {
    syncFromStorage()
    if (!hasSession.value) {
      return false
    }
    if (user.value && !force) {
      return true
    }
    if (sessionCheck) {
      return sessionCheck
    }

    sessionCheck = getUserInfo()
      .then((currentUser) => {
        user.value = currentUser
        syncFromStorage()
        return true
      })
      .catch((error: unknown) => {
        syncFromStorage()
        if (!hasSession.value) {
          user.value = null
          return false
        }
        throw error
      })
      .finally(() => {
        sessionCheck = null
      })

    return sessionCheck
  }

  function logout() {
    clearSession()
  }

  return {
    accessToken,
    refreshToken,
    user,
    hasSession,
    login,
    restoreSession,
    clearSession,
    logout,
  }
})

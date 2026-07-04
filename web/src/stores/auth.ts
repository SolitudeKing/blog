import { defineStore } from 'pinia'
import { login as loginRequest } from '@/api/modules/auth'
import type { LoginPayload, UserInfo } from '@/types/auth'

interface AuthState {
  accessToken: string
  refreshToken: string
  user: UserInfo | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    accessToken: localStorage.getItem('access_token') ?? '',
    refreshToken: localStorage.getItem('refresh_token') ?? '',
    user: null,
  }),
  actions: {
    async login(payload: LoginPayload) {
      const tokens = await loginRequest(payload)
      this.accessToken = tokens.access_token
      this.refreshToken = tokens.refresh_token
      localStorage.setItem('access_token', tokens.access_token)
      localStorage.setItem('refresh_token', tokens.refresh_token)
    },
    logout() {
      this.accessToken = ''
      this.refreshToken = ''
      this.user = null
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
    },
  },
})


import { request } from '@/api/http'
import type { LoginPayload, TokenPair, UserInfo } from '@/types/auth'

export function login(payload: LoginPayload) {
  return request<TokenPair>({
    method: 'POST',
    url: 'auth/login',
    data: payload,
  })
}

export function getUserInfo() {
  return request<UserInfo>({
    method: 'GET',
    url: 'user/info',
  })
}


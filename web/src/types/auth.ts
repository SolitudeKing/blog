export interface LoginPayload {
  username: string
  password: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_at: string
}

export interface UserInfo {
  id: number
  username: string
  nickname?: string
  role: 'owner' | 'editor' | 'viewer'
}

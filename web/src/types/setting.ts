export interface LobbySetting {
  site_name: string
  author: string
  essay: string
  theme: 'forest' | 'strawberry'
  mode: 'light' | 'dark'
  social_links: Record<string, string>
}

export type SettingPayload = LobbySetting

export type ThemeName = 'mist-sea-salt' | 'mist-forest'
export type ModeName = 'light' | 'dark'

export interface LobbySetting {
  site_name: string
  author: string
  essay: string
  theme: ThemeName
  mode: ModeName
  social_links: Record<string, string>
}

export type SettingPayload = LobbySetting

export type ThemeName = 'mist-sea-salt' | 'mist-forest'
export type ModeName = 'light' | 'dark'

export interface ThemeElements {
  home_latest_empty_description: string
  home_latest_end_text: string
}

export type ThemeElementMap = Record<ThemeName, ThemeElements>

export interface LobbySetting {
  site_name: string
  author: string
  author_avatar_url: string
  essay: string
  icp_number: string
  theme: ThemeName
  mode: ModeName
  theme_elements: ThemeElementMap
  social_links: Record<string, string>
}

export interface SettingPayload {
  site_name: string
  author: string
  author_avatar_url: string
  essay: string
  icp_number: string
  theme: ThemeName
  mode: ModeName
  theme_elements: ThemeElementMap
  social_links: Record<string, string>
}

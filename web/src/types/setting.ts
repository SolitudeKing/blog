export type ThemeName = 'mist-sea-salt' | 'mist-forest'
export type ModeName = 'light' | 'dark'

export interface ThemeElements {
  home_latest_empty_description: string
  home_latest_end_text: string
}

export type ThemeElementMap = Record<ThemeName, ThemeElements>

// 主题无关的主页文案聚合：与 theme_elements 的差异在于不按主题分组，整组替换。
export interface HomeContent {
  home_profile_kicker: string
  home_heading_prefix: string
  home_status_fallback: string
  home_intro_heading: string
  home_intro_paragraph: string
  home_action_view_recent_label: string
  home_action_view_archive_label: string
  home_latest_kicker: string
  home_latest_heading: string
  home_latest_view_all_label: string
  home_latest_empty_title: string
}

export interface LobbySetting {
  site_name: string
  author: string
  author_handle: string
  author_avatar_url: string
  essay: string
  icp_number: string
  theme: ThemeName
  mode: ModeName
  theme_elements: ThemeElementMap
  social_links: Record<string, string>
  home_content: HomeContent
}

export interface SettingPayload {
  site_name: string
  author: string
  author_handle: string
  author_avatar_url: string
  essay: string
  icp_number: string
  theme: ThemeName
  mode: ModeName
  theme_elements: ThemeElementMap
  social_links: Record<string, string>
  home_content: HomeContent
}

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
  home_topics_kicker: string
  home_topics_heading: string
  home_notice_kicker: string
  home_notice_action_label: string
}

// 主题无关的归档页文案聚合，结构语义与 HomeContent 一致。
export interface ArchiveContent {
  archive_kicker: string
  archive_heading: string
  archive_intro: string
  archive_empty_title: string
  archive_empty_description: string
}

// 主题无关的搜索页文案聚合；search_suggestion_fallbacks 是换行分隔的兜底航标词，
// 当每日航标接口不可用或为空时由前端拆分渲染。
export interface SearchContent {
  search_kicker: string
  search_heading: string
  search_intro: string
  search_placeholder: string
  search_suggestion_label: string
  search_suggestion_fallbacks: string
  search_empty_title: string
  search_empty_description: string
}

// 主题无关的关于页文案聚合；关于页 hero 使用独立文案，避免与首页 profile 重复。
export interface AboutContent {
  about_kicker: string
  about_heading: string
  about_lead: string
  about_signature: string
  about_contact_label: string
  about_reading_label: string
  about_principles_kicker: string
  about_principles_heading: string
  about_principles_intro: string
  principle_1_title: string
  principle_1_description: string
  principle_2_title: string
  principle_2_description: string
  principle_3_title: string
  principle_3_description: string
  about_contact_kicker: string
  about_contact_heading_with: string
  about_contact_heading_empty: string
  about_contact_intro_with: string
  about_contact_intro_empty: string
  about_contact_empty_cta: string
  about_portrait_line1: string
  about_portrait_line2: string
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
  archive_content: ArchiveContent
  search_content: SearchContent
  about_content: AboutContent
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
  archive_content: ArchiveContent
  search_content: SearchContent
  about_content: AboutContent
}

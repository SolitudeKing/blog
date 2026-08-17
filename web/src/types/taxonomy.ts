export interface TopicItem {
  id: number
  name: string
  label: string
  slug: string
  description: string
  cover_url: string
  sort_order: number
  article_count?: number
  created_at: string
  updated_at: string
}

export interface TopicPayload {
  name: string
  label: string
  slug: string
  description: string
  cover_url: string
  sort_order: number
}

export interface TagItem {
  id: number
  name: string
  slug: string
  description: string
  color: string
  created_at: string
  updated_at: string
}

export interface TagPayload {
  name: string
  slug: string
  description: string
  color: string
}

// 每日搜索航标：由服务端按 UTC 日期从专题与标签中确定性抽样。
export type SuggestionKind = 'topic' | 'tag'

export interface SuggestionItem {
  text: string
  kind: SuggestionKind
}

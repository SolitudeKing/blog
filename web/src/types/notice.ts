export interface NoticeItem {
  id: number
  title: string
  content: string
  enabled: boolean
  sort_order: number
  starts_at: string | null
  ends_at: string | null
  created_at: string
  updated_at: string
}

export interface NoticePayload {
  title: string
  content: string
  enabled: boolean
  sort_order: number
  starts_at: string | null
  ends_at: string | null
}

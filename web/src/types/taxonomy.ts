export interface CategoryItem {
  id: number
  name: string
  slug: string
  description: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface CategoryPayload {
  name: string
  slug: string
  description: string
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

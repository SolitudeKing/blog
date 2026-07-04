export interface ArticleListItem {
  id: number
  title: string
  slug: string
  summary: string
  status: 'draft' | 'published' | 'private' | 'archived'
  category: string
  tags: string[]
  view_count: number
  published_at: string
  created_at: string
  updated_at: string
}

export interface ArticleDetail extends ArticleListItem {
  content_md: string
}


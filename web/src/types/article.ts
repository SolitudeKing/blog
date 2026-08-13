export interface ArticleTopic {
  id: number
  name: string
  label: string
  slug: string
}

export interface ArticleListItem {
  id: number
  title: string
  slug: string
  summary: string
  cover_url: string
  status: 'draft' | 'published' | 'private' | 'archived'
  topic_id: number
  topic: ArticleTopic
  tags: string[]
  view_count: number
  published_at: string
  created_at: string
  updated_at: string
}

export interface ArticleDetail extends ArticleListItem {
  content_md: string
  tag_ids: number[]
}

export interface ArticleSearchItem extends ArticleListItem {
  snippet: string
  matched_fields: string[]
}

export interface ArticleVersionItem {
  id: number
  article_id: number
  title: string
  summary: string
  content_md: string
  status: ArticleListItem['status']
  created_at: string
}

export interface ArticleSavePayload {
  title: string
  slug: string
  summary: string
  cover_url: string
  content_md: string
  topic_id: number
  tag_ids: number[]
  status: ArticleListItem['status']
}

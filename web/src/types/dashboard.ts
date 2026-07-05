export interface DashboardArticleCounts {
  total: number
  published: number
  draft: number
  private: number
  archived: number
}

export interface DashboardTaxonomyCounts {
  categories: number
  tags: number
}

export interface DashboardNoticeCounts {
  total: number
  enabled: number
}

export interface DashboardArticleItem {
  id: number
  title: string
  slug: string
  status: 'draft' | 'published' | 'private' | 'archived'
  updated_at: string
}

export interface DashboardNoticeItem {
  id: number
  title: string
  content: string
  starts_at: string | null
  ends_at: string | null
  updated_at: string
}

export interface DashboardSummary {
  article_counts: DashboardArticleCounts
  taxonomy_counts: DashboardTaxonomyCounts
  notice_counts: DashboardNoticeCounts
  total_views: number
  recent_articles: DashboardArticleItem[]
  active_notice: DashboardNoticeItem | null
  generated_at: string
}

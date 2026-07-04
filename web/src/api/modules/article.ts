import { request, requestList } from '@/api/http'
import type { ArticleDetail, ArticleListItem } from '@/types/article'

export interface ArticleListParams {
  cursor?: string
  limit?: number
  keyword?: string
  category?: string
  tag?: string
}

export function getArticleList(params: ArticleListParams = {}) {
  return requestList<ArticleListItem>({
    method: 'GET',
    url: 'article/list',
    params: {
      limit: 20,
      ...params,
    },
  })
}

export function getArticleDetail(slug: string) {
  return request<ArticleDetail>({
    method: 'GET',
    url: `article/detail/${slug}`,
  })
}


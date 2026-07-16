import { request, requestList } from '@/api/http'
import type {
  ArticleDetail,
  ArticleListItem,
  ArticleSavePayload,
  ArticleSearchItem,
  ArticleVersionItem,
} from '@/types/article'

export interface ArticleListParams {
  cursor?: string
  limit?: number
  keyword?: string
  topic?: string
  tag?: string
  status?: string
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

export function searchArticles(params: Pick<ArticleListParams, 'cursor' | 'limit' | 'keyword'> = {}) {
  return requestList<ArticleSearchItem>({
    method: 'GET',
    url: 'search/article',
    params: {
      limit: 20,
      ...params,
    },
  })
}

export function getManagedArticleList(params: ArticleListParams = {}) {
  return requestList<ArticleListItem>({
    method: 'GET',
    url: 'article/manage-list',
    params: {
      limit: 20,
      ...params,
    },
  })
}

export function getManagedArticleInfo(id: number | string) {
  return request<ArticleDetail>({
    method: 'GET',
    url: `article/info/${id}`,
  })
}

export function getArticleVersions(id: number | string) {
  return request<ArticleVersionItem[]>({
    method: 'GET',
    url: `article/version-list/${id}`,
  })
}

export function createArticle(payload: ArticleSavePayload) {
  return request<ArticleListItem>({
    method: 'POST',
    url: 'article/create',
    data: payload,
  })
}

export function updateArticle(id: number | string, payload: ArticleSavePayload) {
  return request<ArticleDetail>({
    method: 'PUT',
    url: `article/update/${id}`,
    data: payload,
  })
}

export function deleteArticle(id: number | string) {
  return request<{ id: string; deleted: boolean }>({
    method: 'DELETE',
    url: `article/delete/${id}`,
  })
}

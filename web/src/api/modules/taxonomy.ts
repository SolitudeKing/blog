import { request } from '@/api/http'
import type { CategoryItem, CategoryPayload, TagItem, TagPayload } from '@/types/taxonomy'

export function getCategoryList() {
  return request<CategoryItem[]>({
    method: 'GET',
    url: 'category/list',
  })
}

export function createCategory(payload: CategoryPayload) {
  return request<CategoryItem>({
    method: 'POST',
    url: 'category/create',
    data: payload,
  })
}

export function updateCategory(id: number | string, payload: CategoryPayload) {
  return request<CategoryItem>({
    method: 'PUT',
    url: `category/update/${id}`,
    data: payload,
  })
}

export function deleteCategory(id: number | string) {
  return request<{ id: string; deleted: boolean }>({
    method: 'DELETE',
    url: `category/delete/${id}`,
  })
}

export function getTagList() {
  return request<TagItem[]>({
    method: 'GET',
    url: 'tag/list',
  })
}

export function createTag(payload: TagPayload) {
  return request<TagItem>({
    method: 'POST',
    url: 'tag/create',
    data: payload,
  })
}

export function updateTag(id: number | string, payload: TagPayload) {
  return request<TagItem>({
    method: 'PUT',
    url: `tag/update/${id}`,
    data: payload,
  })
}

export function deleteTag(id: number | string) {
  return request<{ id: string; deleted: boolean }>({
    method: 'DELETE',
    url: `tag/delete/${id}`,
  })
}

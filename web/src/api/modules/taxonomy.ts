import { request } from '@/api/http'
import type { TagItem, TagPayload, TopicItem, TopicPayload } from '@/types/taxonomy'

export function getTopicList() {
  return request<TopicItem[]>({
    method: 'GET',
    url: 'topic/list',
  })
}

export function createTopic(payload: TopicPayload) {
  return request<TopicItem>({
    method: 'POST',
    url: 'topic/create',
    data: payload,
  })
}

export function updateTopic(id: number | string, payload: TopicPayload) {
  return request<TopicItem>({
    method: 'PUT',
    url: `topic/update/${id}`,
    data: payload,
  })
}

export function deleteTopic(id: number | string) {
  return request<{ id: string; deleted: boolean }>({
    method: 'DELETE',
    url: `topic/delete/${id}`,
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

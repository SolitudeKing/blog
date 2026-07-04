import { request, requestList } from '@/api/http'
import type { NoticeItem, NoticePayload } from '@/types/notice'

export interface NoticeListParams {
  cursor?: string
  limit?: number
  keyword?: string
  enabled?: boolean
}

export function getActiveNotice() {
  return request<NoticeItem | null>({
    method: 'GET',
    url: 'notice/active',
  })
}

export function getManagedNoticeList(params: NoticeListParams = {}) {
  return requestList<NoticeItem>({
    method: 'GET',
    url: 'notice/manage-list',
    params: {
      limit: 20,
      ...params,
    },
  })
}

export function createNotice(payload: NoticePayload) {
  return request<NoticeItem>({
    method: 'POST',
    url: 'notice/create',
    data: payload,
  })
}

export function updateNotice(id: number | string, payload: NoticePayload) {
  return request<NoticeItem>({
    method: 'PUT',
    url: `notice/update/${id}`,
    data: payload,
  })
}

export function deleteNotice(id: number | string) {
  return request<{ id: string; deleted: boolean }>({
    method: 'DELETE',
    url: `notice/delete/${id}`,
  })
}

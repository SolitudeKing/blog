import { request, requestList } from '@/api/http'
import type { AssetItem, AssetReferenceItem, AssetUpdatePayload } from '@/types/asset'

export interface AssetListParams {
  page?: number
  page_size?: number
  keyword?: string
  mime?: string
}

export function getAssetList(params: AssetListParams = {}) {
  return requestList<AssetItem>({
    method: 'GET',
    url: 'asset/list',
    params: {
      page_size: 40,
      ...params,
    },
  })
}

export function uploadAsset(file: File, displayName = '') {
  const data = new FormData()
  data.append('file', file)
  if (displayName) {
    data.append('display_name', displayName)
  }
  return request<AssetItem>({
    method: 'POST',
    url: 'asset/upload',
    data,
  })
}

export function updateAsset(id: number | string, payload: AssetUpdatePayload) {
  return request<AssetItem>({
    method: 'PUT',
    url: `asset/update/${id}`,
    data: payload,
  })
}

export function deleteAsset(id: number | string) {
  return request<{ id: string; deleted: boolean }>({
    method: 'DELETE',
    url: `asset/delete/${id}`,
  })
}

export function getAssetReferenceList(id: number | string, page = 1) {
  return requestList<AssetReferenceItem>({
    method: 'GET',
    url: `asset/reference-list/${id}`,
    params: {
      page,
      page_size: 20,
    },
  })
}

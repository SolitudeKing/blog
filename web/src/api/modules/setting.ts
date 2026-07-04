import { request } from '@/api/http'
import type { LobbySetting, SettingPayload } from '@/types/setting'

export function getLobbySetting() {
  return request<LobbySetting>({
    method: 'GET',
    url: 'setting/lobby',
  })
}

export function getSettingDetail() {
  return request<LobbySetting>({
    method: 'GET',
    url: 'setting/detail',
  })
}

export function updateSetting(payload: SettingPayload) {
  return request<LobbySetting>({
    method: 'PUT',
    url: 'setting/update',
    data: payload,
  })
}

import { request } from '@/api/http'
import type { LobbySetting } from '@/types/setting'

export function getLobbySetting() {
  return request<LobbySetting>({
    method: 'GET',
    url: 'setting/lobby',
  })
}


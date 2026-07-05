import { request } from '@/api/http'
import type { DashboardSummary } from '@/types/dashboard'

export function getDashboardSummary() {
  return request<DashboardSummary>({
    method: 'GET',
    url: 'dashboard/summary',
  })
}

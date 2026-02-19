import { request } from './http'

import type {
  VisitorInsights,
  VisitorListParams,
  VisitorListResponse,
  VisitorProfileResponse,
} from '@/types/visitors'

export function listVisitors(params: VisitorListParams = {}) {
  const query = Object.fromEntries(
    Object.entries(params).filter(([, value]) => value !== undefined && value !== ''),
  )
  return request<VisitorListResponse>('/admin/visitors', {
    method: 'GET',
    query,
  })
}

export function getVisitorProfile(visitorId: string, recentLimit = 20) {
  return request<VisitorProfileResponse>(`/admin/visitors/${encodeURIComponent(visitorId)}`, {
    method: 'GET',
    query: { recentLimit },
  })
}

export function getVisitorInsights(days = 30) {
  return request<VisitorInsights>('/admin/visitors/insights', {
    method: 'GET',
    query: { days },
  })
}

import { request } from './http'

import type { DashboardStats, HitokotoResponse } from '@/types/stats'

export async function getDashboardStats() {
  return request<DashboardStats>('/admin/stats/dashboard')
}

export async function getHitokoto() {
  return request<HitokotoResponse>('/admin/hitokoto')
}

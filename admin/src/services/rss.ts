import { request } from './http'

import type { RssAccessStats } from '@/types/rss'

export function getRssAccessStats(days = 7, top = 12) {
  return request<RssAccessStats>('/admin/rss/access-stats', {
    method: 'GET',
    query: { days, top },
  })
}

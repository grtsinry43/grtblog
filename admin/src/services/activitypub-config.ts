import { request } from './http'

import type { SysConfigTreeResponse, SysConfigUpdateItem } from './sysconfig'

export function listActivityPubConfigs(keys?: string[]) {
  const query = keys && keys.length > 0 ? { keys: keys.join(',') } : undefined
  return request<SysConfigTreeResponse>('/admin/activitypub/config', {
    method: 'GET',
    query,
  })
}

export function updateActivityPubConfigs(items: SysConfigUpdateItem[]) {
  return request<SysConfigTreeResponse>('/admin/activitypub/config', {
    method: 'PUT',
    body: { items },
  })
}

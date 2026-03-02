import { request } from './http'

export type SysConfigValueType = 'string' | 'number' | 'bool' | 'enum' | 'json'

export type SysConfigEnumOption = string | { label: string; value: string }

export interface SysConfigItem {
  key: string
  groupPath?: string
  label?: string
  description?: string
  valueType: SysConfigValueType
  enumOptions: SysConfigEnumOption[]
  defaultValue?: unknown
  visibleWhen: unknown[]
  sort: number
  meta: Record<string, unknown>
  isSensitive: boolean
  value?: unknown
  createdAt: string
  updatedAt: string
}

export interface SysConfigGroup {
  key: string
  path: string
  label: string
  children?: SysConfigGroup[]
  items?: SysConfigItem[]
}

export interface SysConfigTreeResponse {
  groups: SysConfigGroup[]
  items?: SysConfigItem[]
}

export interface SysConfigUpdateItem {
  key: string
  value?: unknown
  isSensitive?: boolean
  groupPath?: string
  label?: string
  description?: string
  valueType?: SysConfigValueType
  enumOptions?: SysConfigEnumOption[]
  defaultValue?: unknown
  visibleWhen?: unknown[]
  sort?: number
  meta?: Record<string, unknown>
}

export function listSysConfigs(keys?: string[]) {
  const query = keys && keys.length > 0 ? { keys: keys.join(',') } : undefined
  return request<SysConfigTreeResponse>('/admin/sysconfig', {
    method: 'GET',
    query,
  })
}

export function updateSysConfigs(items: SysConfigUpdateItem[]) {
  return request<SysConfigTreeResponse>('/admin/sysconfig', {
    method: 'PUT',
    body: { items },
  })
}

export function listFederationConfigs(keys?: string[]) {
  const query = keys && keys.length > 0 ? { keys: keys.join(',') } : undefined
  return request<SysConfigTreeResponse>('/admin/federation/config', {
    method: 'GET',
    query,
  })
}

export function updateFederationConfigs(items: SysConfigUpdateItem[]) {
  return request<SysConfigTreeResponse>('/admin/federation/config', {
    method: 'PUT',
    body: { items },
  })
}

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

export interface ConfigExportData {
  version: number
  exportedAt: string
  configs: { key: string; value: unknown }[]
}

export function exportFederationConfigs() {
  return request<ConfigExportData>('/admin/federation/export', { method: 'GET' })
}

export function importFederationConfigs(data: ConfigExportData) {
  return request<void>('/admin/federation/import', { method: 'POST', body: data })
}

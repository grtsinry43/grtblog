import { request } from './http'

export interface WebsiteInfoItem {
  key: string
  name?: string | null
  value?: string | null
  infoJson?: unknown
  createdAt: string
  updatedAt: string
}

export interface UpdateWebsiteInfoPayload {
  name?: string | null
  value?: string | null
  infoJson?: unknown
}

export function listWebsiteInfo() {
  return request<WebsiteInfoItem[]>('/website-info', {
    method: 'GET',
  })
}

export function updateWebsiteInfo(key: string, payload: UpdateWebsiteInfoPayload) {
  return request<WebsiteInfoItem>(`/website-info/${key}`, {
    method: 'PUT',
    body: payload,
  })
}

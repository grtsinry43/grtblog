import { request } from '@/services/http'

export interface OwnerStatusMedia {
  title?: string
  artist?: string
  thumbnail?: string
}

export interface OwnerStatusPayload {
  ok: number
  process?: string
  extend?: string
  media?: OwnerStatusMedia | null
  timestamp?: number
  adminPanelOnline?: boolean
}

export interface UpdateOwnerStatusReq {
  ok?: 0 | 1
  process?: string
  extend?: string
  media?: OwnerStatusMedia | null
  timestamp?: number
}

export function getOwnerStatus() {
  return request<OwnerStatusPayload>('/onlineStatus', {
    method: 'GET',
  })
}

export function updateOwnerStatus(payload: UpdateOwnerStatusReq) {
  return request<OwnerStatusPayload>('/onlineStatus', {
    method: 'POST',
    body: payload,
  })
}

export function sendOwnerPanelHeartbeat() {
  return request<OwnerStatusPayload>('/admin/owner-status/panel-heartbeat', {
    method: 'POST',
  })
}

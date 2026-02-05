import { request } from './http'
import type {
  ObservabilityAlerts,
  ObservabilityControlPlane,
  ObservabilityFederation,
  ObservabilityOverview,
  ObservabilityRealtime,
  ObservabilityStorage,
  ObservabilityTimeline,
} from '@/types/observability'

export function getObservabilityOverview() {
  return request<ObservabilityOverview>('/admin/observability/overview')
}

export function getObservabilityControlPlane(window = '5m') {
  return request<ObservabilityControlPlane>('/admin/observability/control-plane', {
    method: 'GET',
    query: { window },
  })
}

export function getObservabilityRenderPlane() {
  return request<any>('/admin/observability/render-plane')
}

export function getObservabilityRealtime() {
  return request<ObservabilityRealtime>('/admin/observability/realtime')
}

export function getObservabilityFederation(window = '24h') {
  return request<ObservabilityFederation>('/admin/observability/federation', {
    method: 'GET',
    query: { window },
  })
}

export function getObservabilityStorage() {
  return request<ObservabilityStorage>('/admin/observability/storage')
}

export function getObservabilityTimeline(query: { since?: string; until?: string; group_by?: 'minute' | 'hour' | 'day' }) {
  return request<ObservabilityTimeline>('/admin/observability/timeline', {
    method: 'GET',
    query: query as any,
  })
}

export function getObservabilityAlerts(limit = 50) {
  return request<ObservabilityAlerts>('/admin/observability/alerts', {
    method: 'GET',
    query: { limit },
  })
}

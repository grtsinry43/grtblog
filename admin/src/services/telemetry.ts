import { request } from './http'

export interface TelemetrySnapshot {
  generatedAt: string
  instance: {
    instanceId: string
    version: string
    goVersion: string
    os: string
    arch: string
    uptimeSeconds: number
    deployMode: string
    features: {
      federationEnabled: boolean
      activityPubEnabled: boolean
      commentsDisabled: boolean
      emailEnabled: boolean
      turnstileEnabled: boolean
    }
  }
  metrics: {
    content: { articlesTotal: number; momentsTotal: number; commentsTotal: number; friendLinksTotal: number }
    traffic: { window: string; requestTotal: number; errorRate5xx: number; p95LatencyMs: number }
    isr: { renderTotal: number; renderSuccess: number; renderFailed: number; avgRenderMs: number; p95RenderMs: number }
    federation: { outboundTotal: number; outboundFailures: number; activeInstances: number }
    realtime: { wsConnectionsCurrent: number; wsRooms: number; broadcastTotal: number }
  }
  errors: TelemetryErrorDigest[]
  panics: TelemetryErrorDigest[]
  summary: { uniqueErrors: number; totalErrors: number; uniquePanics: number; totalPanics: number }
}

export interface TelemetryErrorDigest {
  fingerprint: string
  kind: string
  bizCode?: string
  location: string
  sampleMessage: string
  count: number
  firstSeen: string
  lastSeen: string
}

export interface TelemetryStats {
  uniqueErrors: number
  totalCount: number
}

export interface TelemetryReportRecord {
  timestamp: string
  status: 'success' | 'failed' | 'skipped'
  statusCode?: number
  message?: string
  durationMs?: number
}

export function getTelemetrySnapshot() {
  return request<TelemetrySnapshot>('/admin/telemetry/snapshot')
}

export function getTelemetryStats() {
  return request<TelemetryStats>('/admin/telemetry/stats')
}

export function resetTelemetryErrors() {
  return request<null>('/admin/telemetry/reset', { method: 'POST' })
}

export function getTelemetryReportHistory() {
  return request<{ history: TelemetryReportRecord[] }>('/admin/telemetry/report-history')
}

export function triggerTelemetryReport() {
  return request<TelemetryReportRecord>('/admin/telemetry/report-now', { method: 'POST' })
}

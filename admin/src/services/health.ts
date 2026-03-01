export type SystemMode = 'healthy' | 'maintenance' | 'degraded' | 'critical' | 'outage'

export interface HealthReadinessResponse {
  status: string
  maintenance: boolean
  healthBits: number
  healthMode: string
  isDev: boolean
  components: Array<{
    name: string
    status: string
    healthy: boolean
    version?: string
  }>
}

export interface HealthWSPayload {
  type: 'system.health.state'
  healthBits: number
  maintenance: boolean
  mode: SystemMode
  components: Record<string, boolean>
  isDev: boolean
  timestamp: string
}

const PROBE_TIMEOUT = 2000

/**
 * Derive mode from a 6-bit state value.
 *
 *   bit 0 = 0         → maintenance  (admin toggled maintenance on)
 *   63     (111111)   → healthy
 *   top-5 >= 11100    → degraded     (Redis / Renderer down)
 *   top-5 >= 11000    → critical     (Database down)
 *   else              → outage       (Backend or Nginx down)
 */
export function deriveMode(value: number): SystemMode {
  if ((value & 1) === 0) return 'maintenance'
  if (value === 63) return 'healthy'
  const top5 = value >> 1
  if (top5 >= 0b11100) return 'degraded'
  if (top5 >= 0b11000) return 'critical'
  return 'outage'
}

export async function probeNginx(): Promise<boolean> {
  try {
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), PROBE_TIMEOUT)
    const resp = await fetch(window.location.origin, {
      method: 'HEAD',
      signal: controller.signal,
    })
    clearTimeout(timer)
    return resp.ok
  } catch {
    return false
  }
}

export async function probeRenderer(): Promise<boolean> {
  const base = import.meta.env.VITE_RENDERER_BASE_URL
  if (!base) return true // not configured → skip probe, assume healthy
  try {
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), PROBE_TIMEOUT)
    const resp = await fetch(base, {
      method: 'HEAD',
      mode: 'no-cors',
      signal: controller.signal,
    })
    clearTimeout(timer)
    // no-cors responses have type "opaque" with status 0, which is still success
    return resp.ok || resp.type === 'opaque'
  } catch {
    return false
  }
}

export async function fetchReadiness(): Promise<HealthReadinessResponse | null> {
  try {
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), PROBE_TIMEOUT)
    const resp = await fetch('/health/readiness', { signal: controller.signal })
    clearTimeout(timer)
    if (!resp.ok) return null
    const envelope = await resp.json()
    return envelope?.data ?? null
  } catch {
    return null
  }
}

import { request } from './http'

export interface SystemUpdateReleaseInfo {
  name: string
  prerelease: boolean
  publishedAt: string
  tag: string
  url: string
}

export interface SystemUpdateInfo {
  channel: string
  checkedAt: string
  comparison: 'older' | 'equal' | 'newer' | 'unknown' | string
  currentPrerelease: boolean
  currentVersion: string
  enabled: boolean
  hasUpdate: boolean
  latestRelease?: SystemUpdateReleaseInfo
  latestStableRelease?: SystemUpdateReleaseInfo
  message?: string
  repo: string
  status: 'ok' | 'disabled' | 'error' | string
  targetRelease?: SystemUpdateReleaseInfo
  upgradeUrl?: string
}

export interface SystemStatus {
  app: {
    goVersion: string
    startTime: string
    uptime: string
    version: string
  }
  components: Array<{
    checkedAt: string
    healthy: boolean
    name: string
    status: string
    version?: string
  }>
  cpu: {
    cores: number
  }
  database: {
    driver: string
    poolStats: {
      idle: number
      inUse: number
      maxIdleClosed: number
      maxIdleTimeClosed: number
      maxLifetimeClosed: number
      maxOpenConnections: number
      openConnections: number
      waitCount: number
    }
    status: string
    version?: string
  }
  disk: {
    all: number
    free: number
    path: string
    used: number
  }
  memory: {
    alloc: number
    numGC: number
    sys: number
    totalAlloc: number
  }
  platform: {
    arch: string
    os: string
  }
  redis: {
    status: string
    usedMemory?: string
    version?: string
  }
  storage: {
    path: string
    size: number
  }
  update: SystemUpdateInfo
}

export function getSystemStatus() {
  return request<SystemStatus>('/admin/system/status')
}

export type SystemLogs = string[]

export function getSystemLogs() {
  return request<SystemLogs>('/admin/logs')
}

export function getSystemUpdateCheck(force = false) {
  return request<SystemUpdateInfo>('/admin/system/update-check', {
    method: 'GET',
    query: { force },
  })
}

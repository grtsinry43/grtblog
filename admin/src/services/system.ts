import { request } from './http'

export interface SystemStatus {
  app: {
    goVersion: string
    startTime: string
    uptime: string
    version: string
  }
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
    usedMemory: string
  }
  storage: {
    path: string
    size: number
  }
}

export function getSystemStatus() {
  return request<SystemStatus>('/admin/system/status')
}

export type SystemLogs = string[]

export function getSystemLogs() {
  return request<SystemLogs>('/admin/logs')
}
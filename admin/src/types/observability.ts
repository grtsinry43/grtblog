export interface ObservabilityOverview {
  generatedAt: string
  uptimeSec: number
  api: {
    window: string
    requests: number
    errorRate: number
    p95LatencyMs: number
  }
  realtime: {
    currentOnline: number
    wsRooms: number
    fanoutP95Ms: number
  }
  federation: {
    window: string
    deliveryTotal: number
    deliverySuccessRate: number
    verifyFailedTotal: number
    rateLimitedTotal: number
  }
  render: {
    successJobs: number
    failedJobs: number
    lastDurationMs: number
    p95DurationMs: number
    lastRenderedFiles: number
  }
}

export interface ObservabilityControlPlane {
  generatedAt: string
  api: {
    requests: number
    errors: number
    errorRate: number
    p95LatencyMs: number
    rps: number
    status2xx: number
    status4xx: number
    status5xx: number
  }
  database: {
    status: string
    maxOpenConnections: number
    openConnections: number
    inUse: number
    idle: number
    waitCount: number
  }
  goRuntime: {
    numGoroutine: number
    goVersion: string
  }
}

export interface ObservabilityRealtime {
  generatedAt: string
  snapshot: {
    currentOnline: number
    rooms: number
    joinTotal: number
    leaveTotal: number
    broadcastTotal: number
    broadcastErrors: number
    broadcastFanout: number
    broadcastP95Ms: number
    avgRecipients: number
    broadcastErrorRate: number
    byRoomType: Record<string, number>
  }
}

export interface ObservabilityFederation {
  generatedAt: string
  window: string
  outboundByStatus: Record<string, number>
  outboundTotal: number
  successRate: number
  retryReadyCount: number
  deadLetterCount: number
  pendingCitations: number
  pendingMentions: number
  instancesActive: number
  instancesBlocked: number
  verifyFailedTotal: number
  rateLimitedTotal: number
  inboundEventTotals: Record<string, number>
}

export interface ObservabilityStorage {
  generatedAt: string
  storageHtml: {
    path: string
    size: number
    files: number
  }
  storageLogs: {
    path: string
    size: number
    files: number
  }
  redis: {
    status: string
    usedMemory?: string
    connectedClients?: number
    analyticsQueueDepth?: number
  }
}

export interface ObservabilityTimeline {
  generatedAt: string
  groupBy: string
  since: string
  until: string
  series: Array<{
    metric: string
    timestamp: string
    value: number
    tags?: Record<string, string>
  }>
}

export interface ObservabilityAlerts {
  generatedAt: string
  items: Array<{
    id: number
    type: string
    title: string
    content: string
    isRead: boolean
    createdAt: string
  }>
}

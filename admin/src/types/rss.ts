export interface RssAccessBucket {
  name: string
  count: number
}

export interface RssAccessTrendPoint {
  hour: string
  requests: number
  uniqueIp: number
}

export interface RssAccessStats {
  days: number
  generatedAt: string
  total: number
  uniqueIp: number
  trend: RssAccessTrendPoint[]
  topClients: RssAccessBucket[]
  topIps: RssAccessBucket[]
  topPlatforms: RssAccessBucket[]
  topBrowsers: RssAccessBucket[]
  topLocations: RssAccessBucket[]
  topUserAgents: RssAccessBucket[]
  topHints: RssAccessBucket[]
}

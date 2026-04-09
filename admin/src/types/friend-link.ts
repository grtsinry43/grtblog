export interface FriendLink {
  id: number
  name: string
  url: string
  logo?: string
  description?: string
  rssUrl?: string
  type: 'federation' | 'rss' | 'norss'
  instanceId?: number
  lastSyncAt?: string
  lastSyncStatus?: string
  syncInterval?: number
  totalPostsCached: number
  userId?: number
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface FriendLinkApplication {
  id: number
  name?: string
  url: string
  logo?: string
  description?: string
  applyChannel: 'user' | 'federation' | 'admin'
  requestedSyncMode: string
  rssUrl?: string
  instanceUrl?: string
  signatureVerified: boolean
  userId?: number
  message?: string
  status: 'pending' | 'approved' | 'rejected' | 'blocked'
  createdAt: string
  updatedAt: string
}

export interface FriendLinkSyncJob {
  id: number
  targetType: 'friend_link'
  syncMethod: 'timeline' | 'rss' | 'rss_fallback'
  friendLinkId?: number
  instanceId?: number
  targetUrl: string
  feedUrl?: string
  status: 'queued' | 'running' | 'success' | 'failed'
  attemptCount: number
  maxAttempts: number
  nextRetryAt?: string
  startedAt?: string
  finishedAt?: string
  durationMs?: number
  pulledCount: number
  errorMessage?: string
  triggerSource: string
  createdAt: string
  updatedAt: string
}

export interface FriendLinkCreateReq {
  name: string
  url: string
  logo?: string
  description?: string
  rssUrl?: string
  type?: 'federation' | 'rss' | 'norss'
  instanceId?: number
  syncInterval?: number
  isActive: boolean
}

export interface FriendLinkUpdateReq extends FriendLinkCreateReq {}

export interface FriendLinkListAppsParams {
  page?: number
  pageSize?: number
  status?: string
  channel?: string
  keyword?: string
}

export interface FriendLinkListSyncJobsParams {
  page?: number
  pageSize?: number
  status?: string
  targetType?: string
  syncMethod?: string
  friendLinkId?: number
  instanceId?: number
  keyword?: string
}

export interface FriendLinkListParams {
  page?: number
  pageSize?: number
  active?: boolean
  type?: 'federation' | 'rss' | 'norss'
  keyword?: string
}

export interface FriendLinkFederationRequestReq {
  target_url: string
  message?: string
  rss_url?: string
}

export interface FriendLinkFederationRequestResp {
  request_id?: string
  delivery_id?: number
  status_code: number
  body: string
}

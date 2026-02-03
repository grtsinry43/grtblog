export interface FriendLink {
    id: number
    name: string
    url: string
    logo?: string
    description?: string
    rssUrl?: string
    kind: 'manual' | 'federation'
    syncMode: 'none' | 'rss' | 'federation'
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
    applyChannel: 'user' | 'federation'
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

export interface FriendLinkCreateReq {
    name: string
    url: string
    logo?: string
    description?: string
    rssUrl?: string
    kind?: string
    syncMode?: string
    instanceId?: number
    syncInterval?: number
    isActive: boolean
}

export interface FriendLinkUpdateReq extends FriendLinkCreateReq { }

export interface FriendLinkListAppsParams {
    page?: number
    pageSize?: number
    status?: string
    channel?: string
    keyword?: string
}

export interface FriendLinkListParams {
    page?: number
    pageSize?: number
    active?: boolean
    kind?: string
    syncMode?: string
    keyword?: string
}

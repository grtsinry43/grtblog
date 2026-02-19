export interface VisitorProfile {
  visitorId: string
  nickName?: string
  email?: string
  website?: string
  ip?: string
  location?: string
  platform?: string
  browser?: string
  totalComments: number
  approvedComments: number
  pendingComments: number
  rejectedComments: number
  blockedComments: number
  deletedComments: number
  topComments: number
  activeDays: number
  totalLikes: number
  uniqueLikedItems: number
  totalViews: number
  uniqueViewItems: number
  firstSeenAt: string
  lastSeenAt: string
  lastLikedAt?: string
  lastViewedAt?: string
}

export interface VisitorRecentComment {
  id: number
  areaId: number
  content: string
  status: string
  createdAt: string
  isDeleted: boolean
}

export interface VisitorListResponse {
  items: VisitorProfile[]
  total: number
  page: number
  size: number
}

export interface VisitorProfileResponse {
  profile: VisitorProfile
  recentComments: VisitorRecentComment[]
}

export interface VisitorListParams {
  keyword?: string
  page?: number
  pageSize?: number
}

export interface VisitorDistributionItem {
  name: string
  count: number
}

export interface VisitorTrendPoint {
  date: string
  activeVisitors: number
  newVisitors: number
  returningVisitors: number
  views: number
  likes: number
  comments: number
}

export interface VisitorInsights {
  days: number
  generatedAt: string
  dataSource: string
  platformTop: VisitorDistributionItem[]
  browserTop: VisitorDistributionItem[]
  locationTop: VisitorDistributionItem[]
  trend: VisitorTrendPoint[]
  funnel: {
    viewVisitors: number
    likeVisitors: number
    commentVisitors: number
    likeRate: number
    commentRateByView: number
    commentRateByLike: number
  }
  segments: {
    active1d: number
    active3d: number
    active7d: number
    active30d: number
    highlyEngaged: number
  }
}

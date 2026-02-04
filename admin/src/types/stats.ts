export interface OverviewStats {
    users: number
    articlesTotal: number
    articlesPublished: number
    articlesDraft: number
    momentsTotal: number
    momentsPublished: number
    momentsDraft: number
    pagesTotal: number
    pagesEnabled: number
    thinkingsTotal: number
    categoriesTotal: number
    columnsTotal: number
    tagsTotal: number
}

export interface InteractionStats {
    viewsTotal: number
    likesTotal: number
    commentsTotal: number

    articleViews: number
    articleLikes: number
    articleComments: number

    momentViews: number
    momentLikes: number
    momentComments: number

    pageViews: number
    pageLikes: number
    pageComments: number

    thinkingViews: number
    thinkingLikes: number
    thinkingComments: number
}

export interface WordCountStats {
    total: number
    articles: number
    moments: number
    pages: number
    thinkings: number
}

export interface PendingStats {
    unviewedComments: number
    friendLinkApplications: number
}

export interface PublishTrendPoint {
    date: string
    articles: number
    moments: number
    pages: number
    thinkings: number
}

export interface DayCountPoint {
    date: string
    count: number
}

export interface OnlineTrendPoint {
    hour: string
    peak: number
    avg: number
}

export interface DistributionItem {
    name: string
    count: number
}

export interface TopArticleItem {
    id: number
    title: string
    shortUrl: string
    views: number
    likes: number
    comments: number
    score: number
    createdAt: string
}

export interface TopMomentItem {
    id: number
    title: string
    shortUrl: string
    views: number
    likes: number
    comments: number
    score: number
    createdAt: string
}

export interface DashboardStats {
    generatedAt: string
    cached: boolean
    overview: OverviewStats
    interaction: InteractionStats
    words: WordCountStats
    pending: PendingStats
    trend: PublishTrendPoint[]
    viewTrend: DayCountPoint[]
    commentTrend: DayCountPoint[]
    online24h: OnlineTrendPoint[]
    currentOnline: number
    todayPeakOnline: number
    categories: DistributionItem[]
    columns: DistributionItem[]
    tagTop: DistributionItem[]
    platformTop: DistributionItem[]
    browserTop: DistributionItem[]
    locationTop: DistributionItem[]
    topArticles: TopArticleItem[]
    topMoments: TopMomentItem[]
    topPages: TopMomentItem[] // Using TopMomentItem as it shares the same structure (HotContentItem)
    topThinkings: TopMomentItem[]
}

export interface Hitokoto {
    id: number
    uuid: string
    hitokoto: string
    from: string
    from_who: string | null
    creator: string
    creator_uid: number
    reviewer: number
    commit_from: string
    created_at: string
    length: number
}

export interface HitokotoResponse {
    sentence: Hitokoto
    cached: boolean
}

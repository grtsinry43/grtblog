export enum CommentStatus {
    Pending = 'pending',
    Approved = 'approved',
    Rejected = 'rejected',
    Blocked = 'blocked',
}

export interface Comment {
    id: string
    areaId: number
    areaType?: string
    areaRefId?: number
    areaName?: string
    areaTitle?: string
    areaClosed?: boolean
    content?: string
    authorId?: number
    nickName?: string
    avatar?: string
    email?: string
    ip?: string
    location?: string
    platform?: string
    browser?: string
    website?: string
    isOwner: boolean
    isFriend: boolean
    isAuthor: boolean
    isViewed: boolean
    isTop: boolean
    status: CommentStatus
    parentId?: string
    createdAt: string
    updatedAt: string
    deletedAt?: string
    isDeleted: boolean
}

export interface CommentListResponse {
    items: Comment[]
    total: number
    page: number
    size: number
}

export interface ListCommentsParams {
    areaId?: number
    status?: string
    onlyUnviewed?: boolean
    page?: number
    pageSize?: number
}

export interface ReplyCommentPayload {
    content: string
}

export interface UpdateCommentStatusPayload {
    status: CommentStatus
}

export interface SetCommentAuthorPayload {
    isAuthor: boolean
}

export interface SetCommentTopPayload {
    isTop: boolean
}

export interface SetCommentAreaClosePayload {
    isClosed: boolean
}

export interface MarkCommentsViewedPayload {
    ids: string[]
    isViewed?: boolean
}

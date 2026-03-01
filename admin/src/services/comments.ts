import { request } from './http'
import type {
    CommentListResponse,
    ListCommentsParams,
    ReplyCommentPayload,
    UpdateCommentStatusPayload,
    SetCommentAuthorPayload,
    SetCommentTopPayload,
    SetCommentAreaClosePayload,
    MarkCommentsViewedPayload,
    Comment,
} from '@/types/comments'

export function listComments(params: ListCommentsParams) {
    // Filter out undefined values
    const query = Object.fromEntries(
        Object.entries(params).filter(([, v]) => v !== undefined && v !== ''),
    ) as unknown as Record<string, string | number | boolean>

    return request<CommentListResponse>('/admin/comments', {
        method: 'GET',
        query,
    })
}

export function replyComment(id: string, payload: ReplyCommentPayload) {
    return request<Comment>(`/admin/comments/${id}/reply`, {
        method: 'POST',
        body: payload,
    })
}

export function updateCommentStatus(id: string, payload: UpdateCommentStatusPayload) {
    return request<void>(`/admin/comments/${id}/status`, {
        method: 'PUT',
        body: payload,
    })
}

export function setCommentAuthor(id: string, payload: SetCommentAuthorPayload) {
    return request<void>(`/admin/comments/${id}/author`, {
        method: 'PUT',
        body: payload,
    })
}

export function setCommentTop(id: string, payload: SetCommentTopPayload) {
    return request<void>(`/admin/comments/${id}/top`, {
        method: 'PUT',
        body: payload,
    })
}

export function deleteComment(id: string) {
    return request<void>(`/admin/comments/${id}`, {
        method: 'DELETE',
    })
}

export function setCommentAreaClose(areaId: number, payload: SetCommentAreaClosePayload) {
    return request<void>(`/admin/comments/areas/${areaId}/close`, {
        method: 'PUT',
        body: payload,
    })
}

export function markCommentsViewed(payload: MarkCommentsViewedPayload) {
    return request<void>('/admin/comments/viewed', {
        method: 'PUT',
        body: payload,
    })
}

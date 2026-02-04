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

export function replyComment(id: number, payload: ReplyCommentPayload) {
    return request<Comment>(`/admin/comments/${id}/reply`, {
        method: 'POST',
        body: payload,
    })
}

export function updateCommentStatus(id: number, payload: UpdateCommentStatusPayload) {
    return request<void>(`/admin/comments/${id}/status`, {
        method: 'PUT',
        body: payload,
    })
}

export function setCommentAuthor(id: number, payload: SetCommentAuthorPayload) {
    return request<void>(`/admin/comments/${id}/author`, {
        method: 'PUT',
        body: payload,
    })
}

export function setCommentTop(id: number, payload: SetCommentTopPayload) {
    return request<void>(`/admin/comments/${id}/top`, {
        method: 'PUT',
        body: payload,
    })
}

export function deleteComment(id: number) {
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

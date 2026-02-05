import { request } from './http'

export interface GlobalNotificationItem {
    id: number
    content: string
    publishAt: string
    expireAt: string
    allowClose: boolean
    createdAt: string
    updatedAt: string
}

export interface GlobalNotificationListResponse {
    items: GlobalNotificationItem[]
    total: number
    page: number
    size: number
}

export interface CreateGlobalNotificationPayload {
    content: string
    publishAt: string
    expireAt: string
    allowClose?: boolean
}

export interface UpdateGlobalNotificationPayload {
    content: string
    publishAt: string
    expireAt: string
    allowClose?: boolean
}

export function listGlobalNotifications(params: { page?: number; pageSize?: number } = {}) {
    return request<GlobalNotificationListResponse>('/admin/global-notifications', {
        method: 'GET',
        query: params,
    })
}

export function getGlobalNotification(id: number) {
    return request<GlobalNotificationItem>(`/admin/global-notifications/${id}`, {
        method: 'GET',
    })
}

export function createGlobalNotification(payload: CreateGlobalNotificationPayload) {
    return request<GlobalNotificationItem>('/admin/global-notifications', {
        method: 'POST',
        body: payload,
    })
}

export function updateGlobalNotification(id: number, payload: UpdateGlobalNotificationPayload) {
    return request<GlobalNotificationItem>(`/admin/global-notifications/${id}`, {
        method: 'PUT',
        body: payload,
    })
}

export function deleteGlobalNotification(id: number) {
    return request<void>(`/admin/global-notifications/${id}`, {
        method: 'DELETE',
    })
}

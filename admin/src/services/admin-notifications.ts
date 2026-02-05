import { request } from './http'

export interface AdminNotificationResp {
    id: number
    type: string
    title: string
    content: string
    payload?: any
    is_read: boolean
    read_at?: string
    created_at: string
}

export interface AdminNotificationListResp {
    items: AdminNotificationResp[]
    total: number
    page: number
    size: number
}

export const adminNotificationService = {
    listMine: (unreadOnly = false, page = 1, size = 20) => {
        return request<AdminNotificationListResp>('/notifications', {
            method: 'GET',
            query: {
                unreadOnly,
                page,
                pageSize: size,
            },
        })
    },

    markRead: (id: number) => {
        return request<void>(`/notifications/${id}/read`, {
            method: 'POST',
        })
    },

    markAllRead: () => {
        return request<void>('/notifications/read-all', {
            method: 'POST',
        })
    },
}

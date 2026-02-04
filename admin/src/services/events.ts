import { request } from './http'

export interface AdminEventFieldResp {
    name: string
    type: string
    required: boolean
    description: string
}

export interface AdminEventDescriptorResp {
    name: string
    title: string
    category: string
    description: string
    publicEmail: boolean
    channels: string[]
    fields: AdminEventFieldResp[]
}

export interface AdminEventGroupResp {
    category: string
    events: string[]
}

export interface AdminEventListResp {
    groups: AdminEventGroupResp[]
}

export interface AdminEventCatalogResp {
    items: AdminEventDescriptorResp[]
}

/**
 * List event groups (categorized names only)
 * @param channel Filter by channel ('email' | 'webhook')
 */
export function listEvents(channel?: string) {
    return request<AdminEventListResp>('/admin/events', {
        method: 'GET',
        query: channel ? { channel } : undefined,
    })
}

/**
 * Get full catalog of events with details
 * @param channel Filter by channel ('email' | 'webhook')
 */
export function listEventCatalog(channel?: string) {
    return request<AdminEventCatalogResp>('/admin/events/catalog', {
        method: 'GET',
        query: channel ? { channel } : undefined,
    })
}

/**
 * Get details for a single event
 * @param name Event name (e.g. 'article.created')
 */
export function getEventCatalogItem(name: string) {
    return request<AdminEventDescriptorResp>(`/admin/events/catalog/${name}`, {
        method: 'GET',
    })
}

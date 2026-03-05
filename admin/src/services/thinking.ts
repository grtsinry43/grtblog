import { request } from './http'

export interface ThinkingListItem {
  id: number
  content: string
  authorName?: string
  authorAvatar?: string
  allowComment: boolean
  views: number
  likes: number
  comments: number
  createdAt: string
  updatedAt: string
}

export interface ThinkingListResponse {
  items: ThinkingListItem[]
  total: number
  page: number
  size: number
}

export interface ThinkingDetail extends ThinkingListItem {
  activityPubObjectId?: string | null
  activityPubLastPublishedAt?: string | null
}

export interface ListThinkingsParams {
  page?: number
  pageSize?: number
}

export interface CreateThinkingPayload {
  content: string
  allowComment?: boolean
  createdAt?: string | null
}

export interface UpdateThinkingPayload {
  content: string
  allowComment?: boolean
}

function stripEmpty<T extends object>(value: T): Record<string, unknown> {
  return Object.fromEntries(
    Object.entries(value).filter(
      ([, entry]) => entry !== undefined && entry !== null && entry !== '',
    ),
  )
}

export function listThinkings(params: ListThinkingsParams) {
  return request<ThinkingListResponse>('/thinkings', {
    method: 'GET',
    query: stripEmpty(params),
  })
}

export function getThinking(id: number) {
  return request<ThinkingDetail>(`/thinkings/${id}`, {
    method: 'GET',
  })
}

export function createThinking(payload: CreateThinkingPayload) {
  return request<ThinkingDetail>('/thinkings', {
    method: 'POST',
    body: payload,
  })
}

export function updateThinking(id: number, payload: UpdateThinkingPayload) {
  return request<ThinkingDetail>(`/thinkings/${id}`, {
    method: 'PUT',
    body: payload,
  })
}

export function deleteThinking(id: number) {
  return request<void>(`/thinkings/${id}`, {
    method: 'DELETE',
  })
}

export function batchDeleteThinkings(payload: { ids: number[] }) {
  return request<void>('/admin/thinkings/batch-delete', {
    method: 'POST',
    body: payload,
  })
}

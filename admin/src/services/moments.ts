import { request } from './http'

export interface MomentListItem {
  id: number
  title: string
  shortUrl: string
  authorName?: string
  summary: string
  avatar?: string
  image?: string[]
  views: number
  columnName?: string
  columnShortUrl?: string
  topics: string[]
  likes: number
  comments: number
  isTop: boolean
  isHot: boolean
  isOriginal: boolean
  createdAt: string
  updatedAt: string
}

export interface MomentListResponse {
  items: MomentListItem[]
  total: number
  page: number
  size: number
}

export interface MomentTopic {
  id: number
  name: string
}

export interface MomentDetail {
  id: number
  title: string
  summary: string
  aiSummary?: string | null
  content: string
  contentHash: string
  authorId: number
  image?: string[]
  columnId?: number | null
  shortUrl: string
  isPublished: boolean
  isTop: boolean
  isHot: boolean
  isOriginal: boolean
  topics?: MomentTopic[]
  createdAt: string
  updatedAt: string
}

export interface ListMomentsParams {
  page?: number
  pageSize?: number
  columnId?: number
  topicId?: number
  authorId?: number
  published?: boolean
  search?: string
}

export interface CreateMomentPayload {
  title: string
  summary: string
  content: string
  image?: string[]
  columnId?: number | null
  topicIds?: number[]
  shortUrl?: string | null
  isPublished: boolean
  isTop: boolean
  isHot: boolean
  isOriginal: boolean
  createdAt?: string | null
}

export interface UpdateMomentPayload {
  title: string
  summary: string
  content: string
  image?: string[]
  columnId?: number | null
  topicIds?: number[]
  shortUrl: string
  isPublished: boolean
  isTop: boolean
  isHot: boolean
  isOriginal: boolean
}

function stripEmpty<T extends Record<string, unknown>>(value: T) {
  return Object.fromEntries(
    Object.entries(value).filter(
      ([, entry]) => entry !== undefined && entry !== null && entry !== '',
    ),
  ) as T
}

export function listMoments(params: ListMomentsParams) {
  return request<MomentListResponse>('/admin/moments', {
    method: 'GET',
    query: stripEmpty(params),
  })
}

export function getMoment(id: number) {
  return request<MomentDetail>(`/moments/${id}`, {
    method: 'GET',
  })
}

export function createMoment(payload: CreateMomentPayload) {
  return request<MomentDetail>('/moments', {
    method: 'POST',
    body: payload,
  })
}

export function updateMoment(id: number, payload: UpdateMomentPayload) {
  return request<MomentDetail>(`/moments/${id}`, {
    method: 'PUT',
    body: payload,
  })
}

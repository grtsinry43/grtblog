import { request } from './http'
import type { ContentExtInfo, TOCNode } from '@/types/ext-info' // Assuming TOCNode and ContentExtInfo are defined here or similar

export interface PageMetrics {
  views: number
  likes: number
  comments: number
}

export interface PageListItem {
  id: number
  title: string
  description?: string
  shortUrl: string
  isEnabled: boolean
  isBuiltin: boolean
  metrics: PageMetrics
  createdAt: string
  updatedAt: string
}

export interface PageListResponse {
  items: PageListItem[]
  total: number
  page: number
  size: number
}

export interface PageDetail {
  id: number
  title: string
  description?: string
  aiSummary?: string
  toc?: TOCNode[]
  content: string
  contentHash: string
  commentId?: number
  shortUrl: string
  isEnabled: boolean
  isBuiltin: boolean
  allowComment: boolean
  extInfo?: ContentExtInfo
  metrics: PageMetrics
  createdAt: string
  updatedAt: string
}

export interface ListPagesParams {
  page?: number
  pageSize?: number
}

export interface CreatePagePayload {
  title: string
  description?: string
  content: string
  shortUrl: string
  isEnabled: boolean
  allowComment: boolean
  extInfo?: ContentExtInfo
  createdAt?: string // Optional for setting creation time
}

export interface UpdatePagePayload {
  title: string
  description?: string
  content: string
  shortUrl: string
  isEnabled: boolean
  allowComment: boolean
  extInfo?: ContentExtInfo
}

function stripEmpty<T extends Record<string, unknown>>(value: T) {
  return Object.fromEntries(
    Object.entries(value).filter(
      ([, entry]) => entry !== undefined && entry !== null && entry !== '',
    ),
  ) as T
}

export function listPages(params: ListPagesParams) {
  return request<PageListResponse>('/pages', {
    method: 'GET',
    query: stripEmpty(params),
  })
}

export function getPage(id: number) {
  return request<PageDetail>(`/pages/${id}`, {
    method: 'GET',
  })
}

export function getPageByShortUrl(shortUrl: string) {
  return request<PageDetail>(`/pages/short/${shortUrl}`, {
    method: 'GET',
  })
}

export function createPage(payload: CreatePagePayload) {
  return request<PageDetail>('/pages', {
    method: 'POST',
    body: payload,
  })
}

export function updatePage(id: number, payload: UpdatePagePayload) {
  return request<PageDetail>(`/pages/${id}`, {
    method: 'PUT',
    body: payload,
  })
}

export function deletePage(id: number) {
  return request<void>(`/pages/${id}`, {
    method: 'DELETE',
  })
}

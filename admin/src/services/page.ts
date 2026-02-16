import { request } from './http'
import type { ContentExtInfo } from '@/types/ext-info'

export interface TOCNode {
  id: string
  text: string
  level: number
  children?: TOCNode[]
}

export interface PageListItem {
  id: number
  title: string
  description?: string
  shortUrl: string
  isEnabled: boolean
  isBuiltin: boolean
  views: number
  likes: number
  comments: number
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
  views: number
  likes: number
  comments: number
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

function stripEmpty<T extends object>(value: T): Record<string, unknown> {
  return Object.fromEntries(
    Object.entries(value).filter(
      ([, entry]) => entry !== undefined && entry !== null && entry !== '',
    ),
  )
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

export function batchSetPageEnabled(payload: { ids: number[]; isEnabled: boolean }) {
  return request<void>('/admin/pages/enabled', {
    method: 'PUT',
    body: payload,
  })
}

export function batchDeletePages(payload: { ids: number[] }) {
  return request<void>('/admin/pages/batch-delete', {
    method: 'POST',
    body: payload,
  })
}

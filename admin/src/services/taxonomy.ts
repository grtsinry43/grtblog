import { request } from './http'

export interface CategoryItem {
  id: number
  name: string
  shortUrl: string
  createdAt: string
  updatedAt: string
}

export interface ColumnItem {
  id: number
  name: string
  shortUrl: string
  createdAt: string
  updatedAt: string
}

export interface TagItem {
  id: number
  name: string
  createdAt: string
  updatedAt: string
}

export interface TaxonomyNamePayload {
  name: string
}

export interface TaxonomySlugPayload {
  name: string
  shortUrl: string
}

export function listCategories() {
  return request<CategoryItem[]>('/categories', {
    method: 'GET',
  })
}

export function listColumns() {
  return request<ColumnItem[]>('/columns', {
    method: 'GET',
  })
}

export function listTags() {
  return request<TagItem[]>('/tags', {
    method: 'GET',
  })
}

export function createCategory(payload: { name: string; shortUrl: string }) {
  return request<CategoryItem>('/admin/categories', {
    method: 'POST',
    body: payload,
  })
}

export function updateCategory(id: number, payload: TaxonomySlugPayload) {
  return request<CategoryItem>(`/admin/categories/${id}`, {
    method: 'PUT',
    body: payload,
  })
}

export function deleteCategory(id: number) {
  return request<void>(`/admin/categories/${id}`, {
    method: 'DELETE',
  })
}

export function createColumn(payload: { name: string; shortUrl: string }) {
  return request<ColumnItem>('/admin/columns', {
    method: 'POST',
    body: payload,
  })
}

export function updateColumn(id: number, payload: TaxonomySlugPayload) {
  return request<ColumnItem>(`/admin/columns/${id}`, {
    method: 'PUT',
    body: payload,
  })
}

export function deleteColumn(id: number) {
  return request<void>(`/admin/columns/${id}`, {
    method: 'DELETE',
  })
}

export function createTag(name: string) {
  return request<TagItem>('/admin/tags', {
    method: 'POST',
    body: { name },
  })
}

export function updateTag(id: number, payload: TaxonomyNamePayload) {
  return request<TagItem>(`/admin/tags/${id}`, {
    method: 'PUT',
    body: payload,
  })
}

export function deleteTag(id: number) {
  return request<void>(`/admin/tags/${id}`, {
    method: 'DELETE',
  })
}

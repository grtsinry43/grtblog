import { request } from './http'

export interface AdminTokenItem {
  id: number
  userId: number
  username: string
  description: string
  tokenPreview: string
  expireAt: string
  createdAt: string
  updatedAt: string
  isExpired: boolean
}

export interface AdminTokenListResponse {
  items: AdminTokenItem[]
  total: number
  page: number
  size: number
}

export interface CreateAdminTokenPayload {
  description?: string
  expireAt: string
}

export interface CreateAdminTokenResponse extends AdminTokenItem {
  token: string
}

export function listAdminTokens(params: { page?: number; pageSize?: number } = {}) {
  return request<AdminTokenListResponse>('/admin/tokens', {
    method: 'GET',
    query: params,
  })
}

export function createAdminToken(payload: CreateAdminTokenPayload) {
  return request<CreateAdminTokenResponse>('/admin/tokens', {
    method: 'POST',
    body: payload,
  })
}

export function deleteAdminToken(id: number) {
  return request<void>(`/admin/tokens/${id}`, {
    method: 'DELETE',
  })
}

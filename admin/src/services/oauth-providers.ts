import { request } from './http'

export interface OAuthProviderPayload {
  key: string
  displayName: string
  clientId: string
  clientSecret?: string
  authorizationEndpoint: string
  tokenEndpoint: string
  userinfoEndpoint?: string
  redirectUriTemplate: string
  scopes?: string
  issuer?: string
  jwksUri?: string
  pkceRequired: boolean
  enabled: boolean
  extraParams?: Record<string, unknown>
}

export interface AdminOAuthProvider {
  key: string
  displayName: string
  clientId: string
  authorizationEndpoint: string
  tokenEndpoint: string
  userinfoEndpoint: string
  redirectUriTemplate: string
  scopes: string
  issuer: string
  jwksUri: string
  pkceRequired: boolean
  enabled: boolean
  extraParams: Record<string, unknown> | null
  createdAt: string
  updatedAt: string
}

export function listOAuthProviders() {
  return request<AdminOAuthProvider[]>('/admin/oauth-providers', {
    method: 'GET',
  })
}

export function createOAuthProvider(payload: OAuthProviderPayload) {
  return request<AdminOAuthProvider>('/admin/oauth-providers', {
    method: 'POST',
    body: payload,
  })
}

export function updateOAuthProvider(key: string, payload: OAuthProviderPayload) {
  return request<AdminOAuthProvider>(`/admin/oauth-providers/${key}`, {
    method: 'PUT',
    body: payload,
  })
}

export function deleteOAuthProvider(key: string) {
  return request<void>(`/admin/oauth-providers/${key}`, {
    method: 'DELETE',
  })
}

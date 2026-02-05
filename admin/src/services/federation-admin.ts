import { request } from './http'
import type {
  FederationAdminCitationReq,
  FederationAdminFriendLinkRequestReq,
  FederationAdminMentionReq,
  FederationAdminProxyResp,
  FederationAdminRemoteCheckResp,
  FederationInstanceDetailResp,
  FederationInstanceListReq,
  FederationInstanceListResp,
  FederationOutboundDeliveryListResp,
  FederationOutboundDeliveryResp,
  FederationOutboundListReq,
  FederationReviewDecisionReq,
  FederationReviewListResp,
} from '@/types/federation'

export function checkFederationRemote(targetUrl: string) {
  return request<FederationAdminRemoteCheckResp>('/admin/federation/remote/check', {
    method: 'GET',
    query: { target_url: targetUrl },
  })
}

export function requestFederationFriendlink(payload: FederationAdminFriendLinkRequestReq) {
  return request<FederationAdminProxyResp>('/admin/federation/friendlinks/request', {
    method: 'POST',
    body: payload,
  })
}

export function requestFederationCitation(payload: FederationAdminCitationReq) {
  return request<FederationAdminProxyResp>('/admin/federation/citations/request', {
    method: 'POST',
    body: payload,
  })
}

export function notifyFederationMention(payload: FederationAdminMentionReq) {
  return request<FederationAdminProxyResp>('/admin/federation/mentions/notify', {
    method: 'POST',
    body: payload,
  })
}

export function getFederationOutboundLog(query: FederationOutboundListReq) {
  return request<FederationOutboundDeliveryListResp>('/admin/federation/outbound', {
    method: 'GET',
    query: query as any,
  })
}

export function getFederationOutboundLogDetail(id: number | string) {
  return request<FederationOutboundDeliveryResp>(`/admin/federation/outbound/${id}`, {
    method: 'GET',
  })
}

export function retryFederationOutboundLog(id: number | string) {
  return request<void>(`/admin/federation/outbound/${id}/retry`, {
    method: 'POST',
  })
}

export function getFederationPendingReviews() {
  return request<FederationReviewListResp>('/admin/federation/reviews/pending', {
    method: 'GET',
  })
}

export function reviewFederationCitation(id: number | string, decision: FederationReviewDecisionReq) {
  return request<void>(`/admin/federation/citations/${id}/review`, {
    method: 'PUT',
    body: decision,
  })
}

export function reviewFederationMention(id: number | string, decision: FederationReviewDecisionReq) {
  return request<void>(`/admin/federation/mentions/${id}/review`, {
    method: 'PUT',
    body: decision,
  })
}

export function getFederationInstances(query?: FederationInstanceListReq) {
  return request<FederationInstanceListResp>('/admin/federation/instances', {
    method: 'GET',
    query: query as any,
  })
}

export function getFederationInstanceDetail(id: number | string) {
  return request<FederationInstanceDetailResp>(`/admin/federation/instances/${id}`, {
    method: 'GET',
  })
}

export function updateFederationInstanceStatus(id: number | string, status: string) {
  return request<void>(`/admin/federation/instances/${id}/status`, {
    method: 'PUT',
    body: { status },
  })
}

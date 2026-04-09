export interface FederationAdminCitationReq {
  target_instance_url: string
  target_post_id: string
  source_article_id?: number
  source_short_url?: string
  citation_context?: string
  citation_type?: string
}

export interface FederationAdminMentionReq {
  target_instance_url: string
  mentioned_user: string
  source_article_id?: number
  source_short_url?: string
  mention_context?: string
  mention_type?: string
}

export interface FederationAdminRemoteCheckReq {
  target_url: string
}

export interface FederationAdminProxyResp {
  request_id?: string
  delivery_id?: number
  status_code: number
  body: string
}

export interface FederationAdminRemoteCheckResp {
  manifest?: any
  public_key?: any
  endpoints?: any
}

export interface FederationOutboundListReq {
  request_id?: string
  type?: string
  status?: string
  target?: string
  page?: number
  pageSize?: number
}

export interface FederationOutboundDeliveryResp {
  id: number
  request_id: string
  type: string
  source_article_id?: number
  target_instance_url: string
  target_endpoint: string
  status: string
  attempt_count: number
  max_attempts: number
  next_retry_at?: string
  http_status?: number
  response_body?: string
  error_message?: string
  remote_ticket_id?: string
  trace_id?: string
  last_callback_at?: string
  created_at: string
  updated_at: string
}

export interface FederationOutboundDeliveryListResp {
  items: FederationOutboundDeliveryResp[]
  total: number
  page: number
  size: number
}

export interface FederationReviewDecisionReq {
  status: 'approved' | 'rejected'
  reason?: string
}

export interface FederationReviewItemResp {
  type: string
  id: number
  status: string
  source_instance_id: number
  source_request_id?: string
  summary: string
  requested_at: string
}

export interface FederationReviewListResp {
  items: FederationReviewItemResp[]
}

export interface FederationInstanceResp {
  id: number
  base_url: string
  name?: string
  description?: string
  protocol_version?: string
  key_id?: string
  status: string
  last_seen_at?: string
  created_at: string
  updated_at: string
}

export interface FederationInstanceListResp {
  items: FederationInstanceResp[]
  total: number
  page: number
  size: number
}

export interface FederationInstanceDetailResp extends FederationInstanceResp {
  public_key?: string
  features?: any
  policies?: any
  endpoints?: any
  manifest?: any
  public_key_doc?: any
  endpoints_doc?: any
  remote_error?: string
}

export interface FederationInstanceListReq {
  page?: number
  pageSize?: number
  keyword?: string
}

export interface FederationActivityPubPublishReq {
  source_type: 'article' | 'moment' | 'thinking'
  source_id: number
  summary?: string
}

export interface FederationActivityPubPublishResp {
  activity_id: string
  object_id: string
  source_type: string
  source_id: number
  deliveries: number
  success_count: number
  failure_count: number
  failed_target?: string[]
  published_at: string
}

export interface FederationActivityPubFollowerResp {
  id: number
  actor_id: string
  inbox_url: string
  shared_inbox_url?: string
  preferred_username?: string
  display_name?: string
  status: string
  followed_at: string
  last_seen_at?: string
  updated_at: string
}

export interface FederationActivityPubFollowerListResp {
  items: FederationActivityPubFollowerResp[]
  total: number
  page: number
  size: number
}

export interface FederationCachedPostResp {
  id: number
  remotePostId?: string
  instanceId: number
  url: string
  title: string
  summary: string
  coverImage?: string
  authorName?: string
  publishedAt: string
  allowCitation: boolean
}

export interface FederationCachedPostListResp {
  items: FederationCachedPostResp[]
}

export interface FederationRemotePostResp {
  id: string
  url: string
  title: string
  summary: string
  content_preview?: string
  author: { name: string; url?: string; avatar?: string }
  instance_name: string
  instance_url: string
  published_at: string
  updated_at?: string
  cover_image?: string
  language?: string
  allow_citation: boolean
  allow_comment: boolean
}

export interface FederationRemotePostListResp {
  items: FederationRemotePostResp[]
  total: number
  page: number
  size: number
}

export interface FederationAuthorResp {
  name: string
  instanceUrl: string
  instanceName: string
}

export interface FederationAuthorListResp {
  items: FederationAuthorResp[]
}

export interface FederationCitationInteractionResp {
  id: number
  source_instance_id: number
  source_post_url: string
  source_post_title?: string
  citation_type: string
  status: string
  requested_at: string
}

export interface FederationOutboundInteractionResp {
  id: number
  request_id: string
  type: string
  signal_key?: string
  target_instance_url: string
  status: string
  attempt_count: number
  http_status?: number
  error_message?: string
  remote_ticket_id?: string
  created_at: string
  updated_at: string
}

export interface FederationArticleInteractionsResp {
  article_id: number
  inbound_citations: FederationCitationInteractionResp[]
  outbound: FederationOutboundInteractionResp[]
}

export interface ActivityPubDeliveryDetailResp {
  inbox: string
  actor_id?: string
  status: 'success' | 'failed' | string
  http_status?: number
  error?: string
  delivered_at?: string
}

export interface ActivityPubOutboxItemResp {
  id: number
  activity_id: string
  object_id: string
  source_type: string
  source_id: number
  source_url: string
  summary: string
  activity?: string
  status: 'queued' | 'sending' | 'completed' | 'partial' | 'failed' | string
  trigger_source: 'auto' | 'manual' | string
  total_targets: number
  success_count: number
  failure_count: number
  deliveries?: ActivityPubDeliveryDetailResp[]
  started_at?: string
  finished_at?: string
  duration_ms?: number
  published_at: string
  created_at: string
  updated_at: string
}

export interface ActivityPubOutboxListResp {
  items: ActivityPubOutboxItemResp[]
  total: number
  page: number
  size: number
}

export interface ActivityPubOutboxListReq {
  page?: number
  pageSize?: number
  status?: string
  sourceType?: string
  search?: string
}

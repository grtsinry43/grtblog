export interface FederationAdminFriendLinkRequestReq {
    target_url: string
    message?: string
    rss_url?: string
}

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

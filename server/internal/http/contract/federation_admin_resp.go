package contract

// FederationAdminProxyResp 返回远端响应。
type FederationAdminProxyResp struct {
	RequestID  string `json:"request_id,omitempty"`
	DeliveryID int64  `json:"delivery_id,omitempty"`
	StatusCode int    `json:"status_code"`
	Body       string `json:"body"`
}

// FederationAdminRemoteCheckResp 返回远端 well-known 信息（仅用于文档与测试展示）。
type FederationAdminRemoteCheckResp struct {
	Manifest  any `json:"manifest,omitempty" swaggertype:"object"`
	PublicKey any `json:"public_key,omitempty" swaggertype:"object"`
	Endpoints any `json:"endpoints,omitempty" swaggertype:"object"`
}

type FederationOutboundDeliveryResp struct {
	ID                int64   `json:"id"`
	RequestID         string  `json:"request_id"`
	Type              string  `json:"type"`
	SourceArticleID   *int64  `json:"source_article_id,omitempty"`
	TargetInstanceURL string  `json:"target_instance_url"`
	TargetEndpoint    string  `json:"target_endpoint"`
	Status            string  `json:"status"`
	AttemptCount      int     `json:"attempt_count"`
	MaxAttempts       int     `json:"max_attempts"`
	NextRetryAt       *string `json:"next_retry_at,omitempty"`
	HTTPStatus        *int    `json:"http_status,omitempty"`
	ResponseBody      *string `json:"response_body,omitempty"`
	ErrorMessage      *string `json:"error_message,omitempty"`
	RemoteTicketID    *string `json:"remote_ticket_id,omitempty"`
	TraceID           *string `json:"trace_id,omitempty"`
	LastCallbackAt    *string `json:"last_callback_at,omitempty"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type FederationOutboundDeliveryListResp struct {
	Items []FederationOutboundDeliveryResp `json:"items"`
	Total int64                            `json:"total"`
	Page  int                              `json:"page"`
	Size  int                              `json:"size"`
}

type FederationReviewItemResp struct {
	Type             string  `json:"type"`
	ID               int64   `json:"id"`
	Status           string  `json:"status"`
	SourceInstanceID int64   `json:"source_instance_id"`
	SourceRequestID  *string `json:"source_request_id,omitempty"`
	Summary          string  `json:"summary"`
	RequestedAt      string  `json:"requested_at"`
}

type FederationReviewListResp struct {
	Items []FederationReviewItemResp `json:"items"`
}

type FederationInstanceResp struct {
	ID              int64   `json:"id"`
	BaseURL         string  `json:"base_url"`
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	ProtocolVersion *string `json:"protocol_version,omitempty"`
	KeyID           *string `json:"key_id,omitempty"`
	Status          string  `json:"status"`
	LastSeenAt      *string `json:"last_seen_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type FederationInstanceListResp struct {
	Items []FederationInstanceResp `json:"items"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Size  int                      `json:"size"`
}

type FederationInstanceDetailResp struct {
	ID              int64   `json:"id"`
	BaseURL         string  `json:"base_url"`
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	ProtocolVersion *string `json:"protocol_version,omitempty"`
	KeyID           *string `json:"key_id,omitempty"`
	PublicKey       *string `json:"public_key,omitempty"`
	Status          string  `json:"status"`
	Features        any     `json:"features,omitempty" swaggertype:"object"`
	Policies        any     `json:"policies,omitempty" swaggertype:"object"`
	Endpoints       any     `json:"endpoints,omitempty" swaggertype:"object"`
	Manifest        any     `json:"manifest,omitempty" swaggertype:"object"`
	PublicKeyDoc    any     `json:"public_key_doc,omitempty" swaggertype:"object"`
	EndpointsDoc    any     `json:"endpoints_doc,omitempty" swaggertype:"object"`
	RemoteError     *string `json:"remote_error,omitempty"`
	LastSeenAt      *string `json:"last_seen_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

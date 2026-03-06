package contract

import "time"

// FederationFriendLinkResponseResp 联合友链申请响应。
type FederationFriendLinkResponseResp struct {
	ApplicationID int64  `json:"applicationId"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

// FederationCitationResponseResp 引用请求响应。
type FederationCitationResponseResp struct {
	CitationID int64  `json:"citation_id"`
	Status     string `json:"status"`
}

// FederationCitationDecisionResp 引用审批响应。
type FederationCitationDecisionResp struct {
	CitationID int64  `json:"citation_id"`
	Status     string `json:"status"`
}

// FederationMentionNotifyResp 提及通知响应。
type FederationMentionNotifyResp struct {
	MentionID int64 `json:"mention_id"`
	Delivered bool  `json:"delivered"`
}

// FederationPostAuthorResp 联合时间线作者信息。
type FederationPostAuthorResp struct {
	Name   string  `json:"name"`
	URL    *string `json:"url,omitempty"`
	Avatar *string `json:"avatar,omitempty"`
}

// FederationPostResp 联合时间线文章条目。
type FederationPostResp struct {
	ID             string                   `json:"id"`
	URL            string                   `json:"url"`
	Title          string                   `json:"title"`
	Summary        string                   `json:"summary"`
	ContentPreview *string                  `json:"content_preview,omitempty"`
	Author         FederationPostAuthorResp `json:"author"`
	InstanceName   string                   `json:"instance_name"`
	InstanceURL    string                   `json:"instance_url"`
	PublishedAt    time.Time                `json:"published_at"`
	UpdatedAt      *time.Time               `json:"updated_at,omitempty"`
	CoverImage     *string                  `json:"cover_image,omitempty"`
	Language       *string                  `json:"language,omitempty"`
	AllowCitation  bool                     `json:"allow_citation"`
	AllowComment   bool                     `json:"allow_comment"`
}

// FederationTimelineResp 联合时间线响应。
type FederationTimelineResp struct {
	Items []FederationPostResp `json:"items"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
}

// FederationPostDetailResp 文章详情响应。
type FederationPostDetailResp struct {
	Post         FederationPostResp   `json:"post"`
	RelatedPosts []FederationPostResp `json:"related_posts,omitempty"`
}

type FederationOutboundResultResp struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

type FederationCitationInteractionResp struct {
	ID               int64   `json:"id"`
	SourceInstanceID int64   `json:"source_instance_id"`
	SourcePostURL    string  `json:"source_post_url"`
	SourcePostTitle  *string `json:"source_post_title,omitempty"`
	CitationType     string  `json:"citation_type"`
	Status           string  `json:"status"`
	RequestedAt      string  `json:"requested_at"`
}

type FederationOutboundInteractionResp struct {
	ID                int64   `json:"id"`
	RequestID         string  `json:"request_id"`
	Type              string  `json:"type"`
	SignalKey         *string `json:"signal_key,omitempty"`
	TargetInstanceURL string  `json:"target_instance_url"`
	Status            string  `json:"status"`
	AttemptCount      int     `json:"attempt_count"`
	HTTPStatus        *int    `json:"http_status,omitempty"`
	ErrorMessage      *string `json:"error_message,omitempty"`
	RemoteTicketID    *string `json:"remote_ticket_id,omitempty"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type FederationArticleInteractionsResp struct {
	ArticleID        int64                               `json:"article_id"`
	InboundCitations []FederationCitationInteractionResp `json:"inbound_citations"`
	Outbound         []FederationOutboundInteractionResp `json:"outbound"`
}

// FederationCachedPostResp 缓存文章搜索结果。
type FederationCachedPostResp struct {
	ID            int64   `json:"id"`
	RemotePostID  *string `json:"remotePostId,omitempty"`
	InstanceID    int64   `json:"instanceId"`
	URL           string  `json:"url"`
	Title         string  `json:"title"`
	Summary       string  `json:"summary"`
	CoverImage    *string `json:"coverImage,omitempty"`
	AuthorName    string  `json:"authorName,omitempty"`
	PublishedAt   string  `json:"publishedAt"`
	AllowCitation bool    `json:"allowCitation"`
}

// FederationCachedPostListResp 缓存文章搜索列表。
type FederationCachedPostListResp struct {
	Items []FederationCachedPostResp `json:"items"`
}

// FederationAuthorResp 作者搜索结果。
type FederationAuthorResp struct {
	Name         string `json:"name"`
	InstanceURL  string `json:"instanceUrl"`
	InstanceName string `json:"instanceName"`
}

// FederationAuthorListResp 作者搜索列表。
type FederationAuthorListResp struct {
	Items []FederationAuthorResp `json:"items"`
}

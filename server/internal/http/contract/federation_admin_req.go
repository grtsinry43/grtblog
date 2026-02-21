package contract

// FederationAdminFriendLinkRequestReq 管理后台发起友链申请。
type FederationAdminFriendLinkRequestReq struct {
	TargetURL string `json:"target_url"`
	Message   string `json:"message,omitempty"`
	RSSURL    string `json:"rss_url,omitempty"`
}

// FederationAdminCitationReq 管理后台发起引用请求。
type FederationAdminCitationReq struct {
	TargetInstanceURL string  `json:"target_instance_url"`
	TargetPostID      string  `json:"target_post_id"`
	SourceArticleID   *int64  `json:"source_article_id,omitempty"`
	SourceShortURL    *string `json:"source_short_url,omitempty"`
	CitationContext   string  `json:"citation_context,omitempty"`
	CitationType      string  `json:"citation_type,omitempty"`
}

// FederationAdminMentionReq 管理后台发起提及通知。
type FederationAdminMentionReq struct {
	TargetInstanceURL string  `json:"target_instance_url"`
	MentionedUser     string  `json:"mentioned_user"`
	SourceArticleID   *int64  `json:"source_article_id,omitempty"`
	SourceShortURL    *string `json:"source_short_url,omitempty"`
	MentionContext    string  `json:"mention_context,omitempty"`
	MentionType       string  `json:"mention_type,omitempty"`
}

// FederationAdminRemoteCheckReq 远端联通性检查请求。
type FederationAdminRemoteCheckReq struct {
	TargetURL string `json:"target_url"`
}

type FederationOutboundListReq struct {
	RequestID string `query:"request_id"`
	Type      string `query:"type"`
	Status    string `query:"status"`
	Target    string `query:"target"`
	Page      int    `query:"page"`
	PageSize  int    `query:"pageSize"`
}

type FederationReviewDecisionReq struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type FederationInstanceStatusUpdateReq struct {
	Status string `json:"status"`
}

type FederationActivityPubPublishReq struct {
	SourceType string `json:"source_type"`
	SourceID   int64  `json:"source_id"`
	Summary    string `json:"summary,omitempty"`
}

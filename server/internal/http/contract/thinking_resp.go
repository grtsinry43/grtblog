package contract

import "time"

type ThinkingResp struct {
	ID                         int64           `json:"id"`
	CommentID                  int64           `json:"commentId"`
	Content                    string          `json:"content"`
	AuthorID                   int64           `json:"authorId"`
	ActivityPubObjectID        *string         `json:"activityPubObjectId,omitempty"`
	ActivityPubLastPublishedAt *time.Time      `json:"activityPubLastPublishedAt,omitempty"`
	IsHot                      bool            `json:"isHot"`
	AllowComment               bool            `json:"allowComment"`
	AuthorName                 string          `json:"authorName,omitempty"`
	Avatar                     string          `json:"avatar,omitempty"`
	Metrics                    ThinkingMetrics `json:"metrics"`
	CreatedAt                  time.Time       `json:"createdAt"`
	UpdatedAt                  time.Time       `json:"updatedAt"`
}

type ThinkingMetrics struct {
	Views    int64 `json:"views"`
	Likes    int   `json:"likes"`
	Comments int   `json:"comments"`
}

type ListThinkingResp struct {
	Items []ThinkingListItemResp `json:"items"`
	Total int64                  `json:"total"`
}

// BatchThinkingMetricsReq 批量获取思考指标请求。
type BatchThinkingMetricsReq struct {
	IDs []int64 `json:"ids"`
}

// BatchThinkingMetricsResp 批量获取思考指标响应。
type BatchThinkingMetricsResp struct {
	Items []ThinkingMetricsItem `json:"items"`
}

// ThinkingMetricsItem 单条思考指标。
type ThinkingMetricsItem struct {
	ID       int64 `json:"id"`
	Views    int64 `json:"views"`
	Likes    int   `json:"likes"`
	Comments int   `json:"comments"`
}

type ThinkingListItemResp struct {
	ID                  int64     `json:"id"`
	CommentID           int64     `json:"commentId"`
	Content             string    `json:"content"`
	AuthorID            int64     `json:"authorId"`
	ActivityPubObjectID *string   `json:"activityPubObjectId,omitempty"`
	IsHot               bool      `json:"isHot"`
	AllowComment        bool      `json:"allowComment"`
	AuthorName          string    `json:"authorName,omitempty"`
	Avatar              string    `json:"avatar,omitempty"`
	Views               int64     `json:"views"`
	Likes               int       `json:"likes"`
	Comments            int       `json:"comments"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

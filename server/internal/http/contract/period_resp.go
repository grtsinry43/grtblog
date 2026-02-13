package contract

import "time"

// SamePeriodMomentItemResp 文章同一时期手记项。
type SamePeriodMomentItemResp struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	ShortURL  string    `json:"shortUrl"`
	Summary   string    `json:"summary"`
	Image     []string  `json:"image,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// SamePeriodMomentListResp 文章同一时期手记列表响应。
type SamePeriodMomentListResp struct {
	Items []SamePeriodMomentItemResp `json:"items"`
}

// SamePeriodArticleItemResp 手记同一时期文章项。
type SamePeriodArticleItemResp struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	ShortURL  string    `json:"shortUrl"`
	Summary   string    `json:"summary"`
	Cover     string    `json:"cover,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// SamePeriodArticleListResp 手记同一时期文章列表响应。
type SamePeriodArticleListResp struct {
	Items []SamePeriodArticleItemResp `json:"items"`
}

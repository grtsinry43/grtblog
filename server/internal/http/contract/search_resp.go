package contract

import "time"

type SiteSearchItemResp struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Snippet   string    `json:"snippet"`
	ShortURL  *string   `json:"shortUrl,omitempty"`
	Path      string    `json:"path"`
	Score     float64   `json:"score"`
	CreatedAt time.Time `json:"createdAt"`
}

type SiteSearchResp struct {
	Query     string               `json:"query"`
	Keywords  []string             `json:"keywords"`
	Cached    bool                 `json:"cached"`
	Articles  []SiteSearchItemResp `json:"articles"`
	Moments   []SiteSearchItemResp `json:"moments"`
	Pages     []SiteSearchItemResp `json:"pages"`
	Thinkings []SiteSearchItemResp `json:"thinkings"`
}

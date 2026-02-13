package contract

// TagRelatedContentsResp 按标签聚合的公开内容响应。
type TagRelatedContentsResp struct {
	Articles []ArticleListItemResp `json:"articles"`
	Moments  []MomentListItemResp  `json:"moments"`
}

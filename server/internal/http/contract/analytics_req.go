package contract

type TrackViewReq struct {
	ContentType string `json:"contentType"`
	ContentID   int64  `json:"contentId"`
	VisitorID   string `json:"visitorId"`
}

type TrackViewResp struct {
	VisitorID string `json:"visitorId"`
	Queued    bool   `json:"queued"`
}

type TrackLikeReq struct {
	ContentType string `json:"contentType"`
	ContentID   int64  `json:"contentId"`
	VisitorID   string `json:"visitorId"`
}

type TrackLikeResp struct {
	VisitorID string `json:"visitorId"`
	Affected  bool   `json:"affected"`
}

type ImportLikeBatchReq struct {
	ContentType string   `json:"contentType"`
	ContentID   int64    `json:"contentId"`
	VisitorIDs  []string `json:"visitorIds"`
}

type ImportLikeBatchResp struct {
	Inserted int64 `json:"inserted"`
}

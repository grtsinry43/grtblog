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

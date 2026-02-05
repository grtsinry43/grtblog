package contract

type FriendTimelineListResp struct {
	Items []FederationPostResp `json:"items"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
}

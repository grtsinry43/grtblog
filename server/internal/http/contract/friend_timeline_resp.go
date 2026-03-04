package contract

import "time"

type FriendTimelineAuthorResp struct {
	Name string `json:"name"`
}

type FriendTimelineItemResp struct {
	URL            string                   `json:"url"`
	Title          string                   `json:"title"`
	Summary        string                   `json:"summary"`
	ContentPreview *string                  `json:"content_preview,omitempty"`
	Author         FriendTimelineAuthorResp `json:"author"`
	PublishedAt    time.Time                `json:"published_at"`
	CoverImage     *string                  `json:"cover_image,omitempty"`
}

type FriendTimelineListResp struct {
	Items []FriendTimelineItemResp `json:"items"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Size  int                      `json:"size"`
}

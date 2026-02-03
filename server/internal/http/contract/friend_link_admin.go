package contract

import "time"

// FriendLinkResp 友链管理响应。
type FriendLinkResp struct {
	ID               int64      `json:"id"`
	Name             string     `json:"name"`
	URL              string     `json:"url"`
	Logo             *string    `json:"logo,omitempty"`
	Description      *string    `json:"description,omitempty"`
	RSSURL           *string    `json:"rssUrl,omitempty"`
	Kind             string     `json:"kind"`
	SyncMode         string     `json:"syncMode"`
	InstanceID       *int64     `json:"instanceId,omitempty"`
	LastSyncAt       *time.Time `json:"lastSyncAt,omitempty"`
	LastSyncStatus   *string    `json:"lastSyncStatus,omitempty"`
	SyncInterval     *int       `json:"syncInterval,omitempty"`
	TotalPostsCached int        `json:"totalPostsCached"`
	UserID           *int64     `json:"userId,omitempty"`
	IsActive         bool       `json:"isActive"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// FriendLinkListResp 友链列表响应。
type FriendLinkListResp struct {
	Items []FriendLinkResp `json:"items"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
}

// FriendLinkPublicResp 公开友链响应。
type FriendLinkPublicResp struct {
	Name        string  `json:"name"`
	URL         string  `json:"url"`
	Logo        *string `json:"logo,omitempty"`
	Description *string `json:"description,omitempty"`
	RSSURL      *string `json:"rssUrl,omitempty"`
	Kind        string  `json:"kind"`
	SyncMode    string  `json:"syncMode"`
}

// FriendLinkPublicListResp 公开友链列表响应。
type FriendLinkPublicListResp struct {
	Items []FriendLinkPublicResp `json:"items"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Size  int                    `json:"size"`
}

// FriendLinkApplicationListResp 友链申请列表响应。
type FriendLinkApplicationListResp struct {
	Items []FriendLinkApplicationResp `json:"items"`
	Total int64                       `json:"total"`
	Page  int                         `json:"page"`
	Size  int                         `json:"size"`
}

// FriendLinkApplicationStatusReq 友链申请状态变更请求。
type FriendLinkApplicationStatusReq struct {
	Status string `json:"status"`
}

// FriendLinkCreateReq 管理端创建友链请求。
type FriendLinkCreateReq struct {
	Name         string  `json:"name"`
	URL          string  `json:"url"`
	Logo         *string `json:"logo,omitempty"`
	Description  *string `json:"description,omitempty"`
	RSSURL       *string `json:"rssUrl,omitempty"`
	Kind         string  `json:"kind,omitempty"`
	SyncMode     string  `json:"syncMode,omitempty"`
	InstanceID   *int64  `json:"instanceId,omitempty"`
	SyncInterval *int    `json:"syncInterval,omitempty"`
	IsActive     bool    `json:"isActive"`
	UserID       *int64  `json:"userId,omitempty"`
}

// FriendLinkUpdateReq 管理端更新友链请求。
type FriendLinkUpdateReq struct {
	Name         string  `json:"name"`
	URL          string  `json:"url"`
	Logo         *string `json:"logo,omitempty"`
	Description  *string `json:"description,omitempty"`
	RSSURL       *string `json:"rssUrl,omitempty"`
	Kind         string  `json:"kind,omitempty"`
	SyncMode     string  `json:"syncMode,omitempty"`
	InstanceID   *int64  `json:"instanceId,omitempty"`
	SyncInterval *int    `json:"syncInterval,omitempty"`
	IsActive     bool    `json:"isActive"`
	UserID       *int64  `json:"userId,omitempty"`
}

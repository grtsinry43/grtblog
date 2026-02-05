package social

import (
	"encoding/json"
	"time"
)

type FriendLink struct {
	ID               int64
	Name             string
	URL              string
	Logo             *string
	Description      *string
	RSSURL           *string
	Kind             string
	SyncMode         string
	InstanceID       *int64
	LastSyncAt       *time.Time
	LastSyncStatus   *string
	SyncInterval     *int
	TotalPostsCached int
	UserID           *int64
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

const (
	FriendLinkAppStatusPending  = "pending"
	FriendLinkAppStatusApproved = "approved"
	FriendLinkAppStatusRejected = "rejected"
	FriendLinkAppStatusBlocked  = "blocked"
)

const (
	FriendLinkApplyChannelUser       = "user"
	FriendLinkApplyChannelFederation = "federation"
	FriendLinkApplyChannelAdmin      = "admin"
)

const (
	FriendLinkKindManual     = "manual"
	FriendLinkKindFederation = "federation"
)

const (
	FriendLinkSyncModeNone       = "none"
	FriendLinkSyncModeRSS        = "rss"
	FriendLinkSyncModeFederation = "federation"
)

type FriendLinkApplication struct {
	ID                int64
	Name              *string
	URL               string
	Logo              *string
	Description       *string
	ApplyChannel      string
	RequestedSyncMode string
	RSSURL            *string
	InstanceURL       *string
	Manifest          json.RawMessage
	SignatureKeyID    *string
	SignatureVerified bool
	UserID            *int64
	Message           *string
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type GlobalNotification struct {
	ID         int64
	Content    string
	PublishAt  time.Time
	ExpireAt   time.Time
	AllowClose bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type AdminNotification struct {
	ID        int64
	UserID    int64
	NotifType string
	Title     string
	Content   string
	Payload   json.RawMessage
	IsRead    bool
	ReadAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

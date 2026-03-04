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
	Type             string
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
	FriendLinkTypeFederation = "federation"
	FriendLinkTypeRSS        = "rss"
	FriendLinkTypeNoRSS      = "norss"
)

const (
	FriendLinkSyncJobTargetFriendLink = "friend_link"
)

const (
	FriendLinkSyncJobMethodTimeline = "timeline"
	FriendLinkSyncJobMethodRSS      = "rss"
)

const (
	FriendLinkSyncJobStatusQueued  = "queued"
	FriendLinkSyncJobStatusRunning = "running"
	FriendLinkSyncJobStatusSuccess = "success"
	FriendLinkSyncJobStatusFailed  = "failed"
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

type FriendLinkSyncJob struct {
	ID            int64
	TargetType    string
	SyncMethod    string
	FriendLinkID  *int64
	InstanceID    *int64
	TargetURL     string
	FeedURL       *string
	Status        string
	AttemptCount  int
	MaxAttempts   int
	NextRetryAt   *time.Time
	StartedAt     *time.Time
	FinishedAt    *time.Time
	DurationMS    *int64
	PulledCount   int
	ErrorMessage  *string
	TriggerSource string
	CreatedAt     time.Time
	UpdatedAt     time.Time
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

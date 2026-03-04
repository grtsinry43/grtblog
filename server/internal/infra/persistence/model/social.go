package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type FriendLink struct {
	ID               int64          `gorm:"column:id;primaryKey"`
	Name             string         `gorm:"column:name;size:255;not null"`
	URL              string         `gorm:"column:url;size:255;not null"`
	Logo             string         `gorm:"column:logo;size:255"`
	Description      string         `gorm:"column:description"`
	RSSURL           string         `gorm:"column:rss_url;size:255"`
	Type             string         `gorm:"column:type;size:20;not null"`
	InstanceID       *int64         `gorm:"column:instance_id"`
	LastSyncAt       *time.Time     `gorm:"column:last_sync_at"`
	LastSyncStatus   *string        `gorm:"column:last_sync_status;size:20"`
	SyncInterval     *int           `gorm:"column:sync_interval"`
	TotalPostsCached int            `gorm:"column:total_posts_cached;not null"`
	UserID           *int64         `gorm:"column:user_id"`
	IsActive         bool           `gorm:"column:is_active"`
	CreatedAt        time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (FriendLink) TableName() string { return "friend_link" }

type FriendLinkApplication struct {
	ID                int64          `gorm:"column:id;primaryKey"`
	Name              *string        `gorm:"column:name;size:255"`
	URL               string         `gorm:"column:url;size:255;not null"`
	Logo              *string        `gorm:"column:logo;size:255"`
	Description       *string        `gorm:"column:description"`
	ApplyChannel      string         `gorm:"column:apply_channel;size:20;not null"`
	RequestedSyncMode string         `gorm:"column:requested_sync_mode;size:20;not null"`
	RSSURL            *string        `gorm:"column:rss_url;size:255"`
	InstanceURL       *string        `gorm:"column:instance_url;size:255"`
	Manifest          datatypes.JSON `gorm:"column:manifest;type:jsonb"`
	SignatureKeyID    *string        `gorm:"column:signature_key_id;type:text"`
	SignatureVerified bool           `gorm:"column:signature_verified;not null"`
	UserID            *int64         `gorm:"column:user_id"`
	Message           *string        `gorm:"column:message"`
	Status            string         `gorm:"column:status;size:20"`
	CreatedAt         time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (FriendLinkApplication) TableName() string { return "friend_link_applications" }

type FriendLinkSyncJob struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	TargetType    string     `gorm:"column:target_type;size:30;not null"`
	SyncMethod    string     `gorm:"column:sync_method;size:20;not null"`
	FriendLinkID  *int64     `gorm:"column:friend_link_id"`
	InstanceID    *int64     `gorm:"column:instance_id"`
	TargetURL     string     `gorm:"column:target_url;size:500;not null"`
	FeedURL       *string    `gorm:"column:feed_url;size:500"`
	Status        string     `gorm:"column:status;size:20;not null"`
	AttemptCount  int        `gorm:"column:attempt_count;not null"`
	MaxAttempts   int        `gorm:"column:max_attempts;not null"`
	NextRetryAt   *time.Time `gorm:"column:next_retry_at"`
	StartedAt     *time.Time `gorm:"column:started_at"`
	FinishedAt    *time.Time `gorm:"column:finished_at"`
	DurationMS    *int64     `gorm:"column:duration_ms"`
	PulledCount   int        `gorm:"column:pulled_count;not null"`
	ErrorMessage  *string    `gorm:"column:error_message;type:text"`
	TriggerSource string     `gorm:"column:trigger_source;size:40;not null"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (FriendLinkSyncJob) TableName() string { return "friend_link_sync_job" }

type GlobalNotification struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	Content    string    `gorm:"column:content;type:text;not null"`
	PublishAt  time.Time `gorm:"column:publish_at;not null"`
	ExpireAt   time.Time `gorm:"column:expire_at;not null"`
	AllowClose bool      `gorm:"column:allow_close"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (GlobalNotification) TableName() string { return "global_notification" }

type AdminNotification struct {
	ID        int64          `gorm:"column:id;primaryKey"`
	UserID    int64          `gorm:"column:user_id;not null"`
	NotifType string         `gorm:"column:notif_type;size:50;not null"`
	Title     string         `gorm:"column:title;size:200;not null"`
	Content   string         `gorm:"column:content;type:text;not null"`
	Payload   datatypes.JSON `gorm:"column:payload;type:jsonb;not null"`
	IsRead    bool           `gorm:"column:is_read;not null"`
	ReadAt    *time.Time     `gorm:"column:read_at"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (AdminNotification) TableName() string { return "admin_notification" }

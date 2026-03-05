package model

import (
	"time"

	"gorm.io/datatypes"
)

type ActivityPubFollower struct {
	ID                int64      `gorm:"column:id;primaryKey"`
	ActorID           string     `gorm:"column:actor_id;size:500;not null"`
	InboxURL          string     `gorm:"column:inbox_url;size:500;not null"`
	SharedInboxURL    *string    `gorm:"column:shared_inbox_url;size:500"`
	PreferredUsername *string    `gorm:"column:preferred_username;size:255"`
	DisplayName       *string    `gorm:"column:display_name;size:255"`
	Status            string     `gorm:"column:status;size:20;not null"`
	FollowedAt        time.Time  `gorm:"column:followed_at;autoCreateTime"`
	LastSeenAt        *time.Time `gorm:"column:last_seen_at"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (ActivityPubFollower) TableName() string { return "activitypub_follower" }

type ActivityPubOutboxItem struct {
	ID            int64          `gorm:"column:id;primaryKey"`
	ActivityID    string         `gorm:"column:activity_id;size:500;not null"`
	ObjectID      string         `gorm:"column:object_id;size:500;not null"`
	SourceType    string         `gorm:"column:source_type;size:20;not null"`
	SourceID      int64          `gorm:"column:source_id;not null"`
	SourceURL     string         `gorm:"column:source_url;size:500;not null"`
	Summary       string         `gorm:"column:summary;type:text;not null"`
	Activity      datatypes.JSON `gorm:"column:activity;type:jsonb;not null"`
	Status        string         `gorm:"column:status;size:20;not null"`
	TriggerSource string         `gorm:"column:trigger_source;size:20;not null"`
	TotalTargets  int            `gorm:"column:total_targets;not null"`
	SuccessCount  int            `gorm:"column:success_count;not null"`
	FailureCount  int            `gorm:"column:failure_count;not null"`
	Deliveries    datatypes.JSON `gorm:"column:deliveries;type:jsonb;not null"`
	StartedAt     *time.Time     `gorm:"column:started_at"`
	FinishedAt    *time.Time     `gorm:"column:finished_at"`
	DurationMs    *int64         `gorm:"column:duration_ms"`
	PublishedAt   time.Time      `gorm:"column:published_at;autoCreateTime"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (ActivityPubOutboxItem) TableName() string { return "activitypub_outbox_item" }

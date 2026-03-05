package activitypub

import (
	"encoding/json"
	"time"
)

const (
	OutboxStatusQueued    = "queued"
	OutboxStatusSending   = "sending"
	OutboxStatusCompleted = "completed"
	OutboxStatusPartial   = "partial"
	OutboxStatusFailed    = "failed"
)

type Follower struct {
	ID                int64
	ActorID           string
	InboxURL          string
	SharedInboxURL    *string
	PreferredUsername *string
	DisplayName       *string
	Status            string
	FollowedAt        time.Time
	LastSeenAt        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type DeliveryDetail struct {
	Inbox       string     `json:"inbox"`
	ActorID     string     `json:"actor_id,omitempty"`
	Status      string     `json:"status"`
	HTTPStatus  *int       `json:"http_status,omitempty"`
	Error       string     `json:"error,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
}

type OutboxItem struct {
	ID            int64
	ActivityID    string
	ObjectID      string
	SourceType    string
	SourceID      int64
	SourceURL     string
	Summary       string
	Activity      json.RawMessage
	Status        string
	TriggerSource string
	TotalTargets  int
	SuccessCount  int
	FailureCount  int
	Deliveries    []DeliveryDetail
	StartedAt     *time.Time
	FinishedAt    *time.Time
	DurationMs    *int64
	PublishedAt   time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type OutboxListOptions struct {
	Page       int
	PageSize   int
	Status     string
	SourceType string
	Search     string
}

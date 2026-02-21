package activitypub

import (
	"encoding/json"
	"time"
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

type OutboxItem struct {
	ID          int64
	ActivityID  string
	ObjectID    string
	SourceType  string
	SourceID    int64
	SourceURL   string
	Summary     string
	Activity    json.RawMessage
	PublishedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

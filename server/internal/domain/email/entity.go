package email

import "time"

const (
	OutboxStatusPending = "pending"
	OutboxStatusSending = "sending"
	OutboxStatusSent    = "sent"
	OutboxStatusFailed  = "failed"

	SubscriptionStatusActive       = "active"
	SubscriptionStatusUnsubscribed = "unsubscribed"
	SubscriptionStatusBlocked      = "blocked"
)

type Template struct {
	ID              int64
	Code            string
	Name            string
	EventName       string
	SubjectTemplate string
	HTMLTemplate    string
	TextTemplate    string
	ToEmails        []string
	IsEnabled       bool
	IsInternal      bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type Outbox struct {
	ID           int64
	TemplateID   *int64
	TemplateCode string
	EventName    string
	ToEmails     []string
	Subject      string
	HTMLBody     string
	TextBody     string
	Status       string
	RetryCount   int
	NextRetryAt  time.Time
	LastError    string
	SentAt       *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Subscription struct {
	ID             int64
	Email          string
	EventName      string
	Status         string
	Token          string
	SourceIP       string
	UnsubscribedAt *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type SubscriptionListOptions struct {
	Page      int
	PageSize  int
	EventName *string
	Status    *string
	Search    *string
}

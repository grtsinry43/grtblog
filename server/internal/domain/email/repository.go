package email

import (
	"context"
	"time"
)

type Repository interface {
	CreateTemplate(ctx context.Context, tpl *Template) error
	UpdateTemplate(ctx context.Context, tpl *Template) error
	DeleteTemplateByCode(ctx context.Context, code string) error
	GetTemplateByCode(ctx context.Context, code string) (*Template, error)
	ListTemplates(ctx context.Context) ([]*Template, error)
	ListEnabledTemplatesByEvent(ctx context.Context, eventName string) ([]*Template, error)

	CreateOutbox(ctx context.Context, item *Outbox) error
	ClaimDueOutbox(ctx context.Context, limit int, dueAt time.Time, maxRetries int) ([]*Outbox, error)
	MarkOutboxSent(ctx context.Context, id int64, sentAt time.Time) error
	MarkOutboxFailed(ctx context.Context, id int64, retryCount int, nextRetryAt time.Time, lastError string) error

	CreateOrUpdateSubscription(ctx context.Context, sub *Subscription) error
	GetSubscriptionByEmailEvent(ctx context.Context, email string, eventName string) (*Subscription, error)
	UnsubscribeByToken(ctx context.Context, token string) error
	UnsubscribeByEmailEvent(ctx context.Context, email string, eventName string) error
	ListSubscriptions(ctx context.Context, options SubscriptionListOptions) ([]*Subscription, int64, error)
	BatchUpdateSubscriptionStatus(ctx context.Context, ids []int64, status string) error
	ListActiveSubscriberEmailsByEvent(ctx context.Context, eventName string) ([]string, error)
}

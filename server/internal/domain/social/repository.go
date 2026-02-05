package social

import (
	"context"
	"time"
)

type FriendLinkApplicationRepository interface {
	GetByID(ctx context.Context, id int64) (*FriendLinkApplication, error)
	FindByURL(ctx context.Context, url string) (*FriendLinkApplication, error)
	Create(ctx context.Context, app *FriendLinkApplication) error
	Update(ctx context.Context, app *FriendLinkApplication) error
	UpdateByID(ctx context.Context, app *FriendLinkApplication) error
	List(ctx context.Context, options FriendLinkApplicationListOptions) ([]FriendLinkApplication, int64, error)
}

type FriendLinkRepository interface {
	GetByID(ctx context.Context, id int64) (*FriendLink, error)
	FindByURL(ctx context.Context, url string) (*FriendLink, error)
	ExistsActiveByUserID(ctx context.Context, userID int64) (bool, error)
	Create(ctx context.Context, link *FriendLink) error
	Update(ctx context.Context, link *FriendLink) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, options FriendLinkListOptions) ([]FriendLink, int64, error)
}

type GlobalNotificationRepository interface {
	GetByID(ctx context.Context, id int64) (*GlobalNotification, error)
	Create(ctx context.Context, notification *GlobalNotification) error
	Update(ctx context.Context, notification *GlobalNotification) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, options GlobalNotificationListOptions) ([]GlobalNotification, int64, error)
	ListActive(ctx context.Context, at time.Time) ([]GlobalNotification, error)
}

type FriendLinkApplicationListOptions struct {
	Status       string
	ApplyChannel string
	Keyword      string
	Page         int
	PageSize     int
}

type FriendLinkListOptions struct {
	IsActive *bool
	Kind     string
	SyncMode string
	Keyword  string
	Page     int
	PageSize int
}

type GlobalNotificationListOptions struct {
	Status   string
	Page     int
	PageSize int
}

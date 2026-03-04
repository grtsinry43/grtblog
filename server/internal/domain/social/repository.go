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

type FriendLinkSyncJobRepository interface {
	Create(ctx context.Context, job *FriendLinkSyncJob) error
	Update(ctx context.Context, job *FriendLinkSyncJob) error
	List(ctx context.Context, options FriendLinkSyncJobListOptions) ([]FriendLinkSyncJob, int64, error)
	ListProcessable(ctx context.Context, now time.Time, limit int) ([]FriendLinkSyncJob, error)
}

type GlobalNotificationRepository interface {
	GetByID(ctx context.Context, id int64) (*GlobalNotification, error)
	Create(ctx context.Context, notification *GlobalNotification) error
	Update(ctx context.Context, notification *GlobalNotification) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, options GlobalNotificationListOptions) ([]GlobalNotification, int64, error)
	ListActive(ctx context.Context, at time.Time) ([]GlobalNotification, error)
}

type AdminNotificationRepository interface {
	Create(ctx context.Context, notification *AdminNotification) error
	GetByID(ctx context.Context, id int64) (*AdminNotification, error)
	ListByUser(ctx context.Context, userID int64, options AdminNotificationListOptions) ([]AdminNotification, int64, error)
	MarkRead(ctx context.Context, userID int64, id int64) error
	MarkAllRead(ctx context.Context, userID int64) error
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
	Type     string
	Keyword  string
	Page     int
	PageSize int
}

type FriendLinkSyncJobListOptions struct {
	Status       string
	TargetType   string
	SyncMethod   string
	FriendLinkID *int64
	InstanceID   *int64
	Keyword      string
	Page         int
	PageSize     int
}

type GlobalNotificationListOptions struct {
	Status   string
	Page     int
	PageSize int
}

type AdminNotificationListOptions struct {
	UnreadOnly bool
	Page       int
	PageSize   int
}

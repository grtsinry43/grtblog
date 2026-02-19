package comment

import "context"

type PublicListOptions struct {
	AreaID          int64
	ViewerAuthorID  *int64
	ViewerVisitorID string
}

type AdminListOptions struct {
	AreaID       *int64
	Status       string
	OnlyUnviewed bool
	Page         int
	PageSize     int
}

type AdminVisitorListOptions struct {
	Keyword  string
	Page     int
	PageSize int
}

type CommentRepository interface {
	GetAreaByID(ctx context.Context, id int64) (*CommentArea, error)
	SetAreaClosed(ctx context.Context, areaID int64, isClosed bool) error
	FindByID(ctx context.Context, id int64) (*Comment, error)
	ListPublicByAreaID(ctx context.Context, options PublicListOptions) ([]*Comment, error)
	ListForAdmin(ctx context.Context, options AdminListOptions) ([]*Comment, int64, error)
	Create(ctx context.Context, comment *Comment) error
	Update(ctx context.Context, comment *Comment) error
	Delete(ctx context.Context, id int64) error
	SetViewedStatus(ctx context.Context, ids []int64, isViewed bool) error
	SetAuthorStatus(ctx context.Context, id int64, isAuthor bool) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	SetTopStatus(ctx context.Context, id int64, isTop bool) error
	ExistsBlockedIdentity(ctx context.Context, authorID *int64, email *string) (bool, error)
	ListVisitors(ctx context.Context, options AdminVisitorListOptions) ([]VisitorProfile, int64, error)
	GetVisitorProfile(ctx context.Context, visitorID string, recentLimit int) (*VisitorProfile, []VisitorRecentComment, error)
	GetVisitorInsights(ctx context.Context, days int) (*VisitorInsights, error)
}

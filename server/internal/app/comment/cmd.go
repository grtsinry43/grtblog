package comment

import "time"

type CreateCommentLoginCmd struct {
	AreaID    int64
	Content   string
	ParentID  *int64
	VisitorID string
}

type CreateCommentVisitorCmd struct {
	AreaID    int64
	Content   string
	ParentID  *int64
	NickName  string
	Email     string
	Website   *string
	VisitorID string
}

type ListPublicCommentsCmd struct {
	AreaID          int64
	Page            int
	PageSize        int
	ViewerAuthorID  *int64
	ViewerVisitorID string
}

type ListAdminCommentsCmd struct {
	AreaID       *int64
	Status       string
	OnlyUnviewed bool
	Page         int
	PageSize     int
}

type ListAdminVisitorsCmd struct {
	Keyword  string
	Page     int
	PageSize int
}

type GetVisitorProfileCmd struct {
	VisitorID   string
	RecentLimit int
}

type GetVisitorInsightsCmd struct {
	Days int
}

type ReplyCommentCmd struct {
	ParentID int64
	Content  string
	AdminID  int64
}

type UpdateCommentStatusCmd struct {
	ID     int64
	Status string
}

type SetCommentAuthorCmd struct {
	ID       int64
	IsAuthor bool
}

type SetCommentTopCmd struct {
	ID    int64
	IsTop bool
}

type MarkCommentsViewedCmd struct {
	IDs      []int64
	IsViewed bool
}

type EditCommentCmd struct {
	ID              int64
	Content         string
	ViewerAuthorID  *int64
	ViewerVisitorID string
}

type DeleteOwnCommentCmd struct {
	ID              int64
	ViewerAuthorID  *int64
	ViewerVisitorID string
}

type ImportCommentCmd struct {
	ID                *int64
	AreaID            int64
	Content           string
	AuthorID          *int64
	VisitorID         *string
	NickName          *string
	IP                *string
	Location          *string
	Platform          *string
	Browser           *string
	Email             *string
	Website           *string
	IsOwner           *bool
	IsFriend          *bool
	IsAuthor          *bool
	IsViewed          *bool
	IsTop             *bool
	IsFederated       *bool
	FederatedProtocol *string
	FederatedActor    *string
	FederatedObjectID *string
	CanReply          *bool
	Status            *string
	ParentID          *int64
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
	DeletedAt         *time.Time
}

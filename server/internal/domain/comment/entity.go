package comment

import "time"

type CommentArea struct {
	ID        int64
	Name      string
	Type      string
	ContentID *int64
	IsClosed  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Comment struct {
	ID         int64
	AreaID     int64
	Content    string
	AuthorID   *int64
	VisitorID  *string
	NickName   *string
	Avatar     *string
	IP         *string
	Location   *string
	Platform   *string
	Browser    *string
	Email      *string
	Website    *string
	IsOwner    bool
	IsFriend   bool
	IsAuthor   bool
	IsViewed   bool
	IsTop      bool
	IsMy       bool
	Status     string
	AreaType   *string
	AreaName   *string
	AreaTitle  *string
	AreaRefID  *int64
	AreaClosed *bool
	ParentID   *int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

const (
	CommentStatusPending  = "pending"
	CommentStatusApproved = "approved"
	CommentStatusRejected = "rejected"
	CommentStatusBlocked  = "blocked"
)

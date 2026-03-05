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
	ID                int64
	AreaID            int64
	Content           string
	AuthorID          *int64
	VisitorID         *string
	NickName          *string
	Avatar            *string
	IP                *string
	Location          *string
	Platform          *string
	Browser           *string
	Email             *string
	Website           *string
	IsOwner           bool
	IsFriend          bool
	IsAuthor          bool
	IsViewed          bool
	IsTop             bool
	IsMy              bool
	IsFederated       bool
	FederatedProtocol *string
	FederatedActor    *string
	FederatedObjectID *string
	CanReply          bool
	Status            string
	IsEdited          bool
	AreaType          *string
	AreaName          *string
	AreaTitle         *string
	AreaRefID         *int64
	AreaClosed        *bool
	ParentID          *int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

type VisitorProfile struct {
	VisitorID        string
	NickName         *string
	Email            *string
	Website          *string
	IP               *string
	Location         *string
	Platform         *string
	Browser          *string
	TotalComments    int64
	ApprovedComments int64
	PendingComments  int64
	RejectedComments int64
	BlockedComments  int64
	DeletedComments  int64
	TopComments      int64
	ActiveDays       int64
	TotalLikes       int64
	UniqueLikedItems int64
	TotalViews       int64
	UniqueViewItems  int64
	FirstSeenAt      time.Time
	LastSeenAt       time.Time
	LastLikedAt      *time.Time
	LastViewedAt     *time.Time
}

type VisitorRecentComment struct {
	ID        int64
	AreaID    int64
	Content   string
	Status    string
	CreatedAt time.Time
	IsDeleted bool
}

type VisitorDistributionItem struct {
	Name  string
	Count int64
}

type VisitorTrendPoint struct {
	Date              string
	ActiveVisitors    int64
	NewVisitors       int64
	ReturningVisitors int64
	Views             int64
	Likes             int64
	Comments          int64
}

type VisitorFunnel struct {
	ViewVisitors      int64
	LikeVisitors      int64
	CommentVisitors   int64
	LikeRate          float64
	CommentRateByView float64
	CommentRateByLike float64
}

type VisitorSegments struct {
	Active1D      int64
	Active3D      int64
	Active7D      int64
	Active30D     int64
	HighlyEngaged int64
}

type VisitorInsights struct {
	Days        int
	GeneratedAt time.Time
	DataSource  string
	PlatformTop []VisitorDistributionItem
	BrowserTop  []VisitorDistributionItem
	LocationTop []VisitorDistributionItem
	Trend       []VisitorTrendPoint
	Funnel      VisitorFunnel
	Segments    VisitorSegments
}

const (
	CommentStatusPending  = "pending"
	CommentStatusApproved = "approved"
	CommentStatusRejected = "rejected"
	CommentStatusBlocked  = "blocked"
)

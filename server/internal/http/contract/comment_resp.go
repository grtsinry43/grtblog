package contract

import "time"

type CreateCommentResp struct {
	ID                int64      `json:"id"`
	AreaID            int64      `json:"areaId"`
	Content           string     `json:"content"`
	NickName          *string    `json:"nickName"`
	Avatar            *string    `json:"avatar"`
	Location          *string    `json:"location"`
	Platform          *string    `json:"platform"`
	Browser           *string    `json:"browser"`
	Website           *string    `json:"website"`
	IsOwner           bool       `json:"isOwner"`
	IsFriend          bool       `json:"isFriend"`
	IsAuthor          bool       `json:"isAuthor"`
	IsViewed          bool       `json:"isViewed"`
	IsTop             bool       `json:"isTop"`
	IsMy              bool       `json:"isMy"`
	IsFederated       bool       `json:"isFederated"`
	FederatedProtocol *string    `json:"federatedProtocol"`
	FederatedActor    *string    `json:"federatedActor"`
	CanReply          bool       `json:"canReply"`
	Status            string     `json:"status"`
	ParentID          *int64     `json:"parentId"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	DeletedAt         *time.Time `json:"deletedAt,omitempty"`
	IsDeleted         bool       `json:"isDeleted"`
}

type CommentNodeResp struct {
	ID                int64             `json:"id"`
	AreaID            int64             `json:"areaId"`
	Content           *string           `json:"content"`
	NickName          *string           `json:"nickName"`
	Avatar            *string           `json:"avatar"`
	Location          *string           `json:"location"`
	Platform          *string           `json:"platform"`
	Browser           *string           `json:"browser"`
	Website           *string           `json:"website"`
	IsOwner           bool              `json:"isOwner"`
	IsFriend          bool              `json:"isFriend"`
	IsAuthor          bool              `json:"isAuthor"`
	IsViewed          bool              `json:"isViewed"`
	IsTop             bool              `json:"isTop"`
	IsMy              bool              `json:"isMy"`
	IsFederated       bool              `json:"isFederated"`
	FederatedProtocol *string           `json:"federatedProtocol"`
	FederatedActor    *string           `json:"federatedActor"`
	CanReply          bool              `json:"canReply"`
	Status            string            `json:"status"`
	ParentID          *int64            `json:"parentId"`
	CreatedAt         time.Time         `json:"createdAt"`
	UpdatedAt         time.Time         `json:"updatedAt"`
	DeletedAt         *time.Time        `json:"deletedAt,omitempty"`
	IsDeleted         bool              `json:"isDeleted"`
	Children          []CommentNodeResp `json:"children,omitempty"`
}

type PublicCommentListResp struct {
	Items             []CommentNodeResp `json:"items"`
	Total             int64             `json:"total"`
	Page              int               `json:"page"`
	Size              int               `json:"size"`
	IsClosed          bool              `json:"isClosed"`
	RequireModeration bool              `json:"requireModeration"`
}

type AdminCommentResp struct {
	ID                string     `json:"id"`
	AreaID            int64      `json:"areaId"`
	AreaType          *string    `json:"areaType,omitempty"`
	AreaRefID         *int64     `json:"areaRefId,omitempty"`
	AreaName          *string    `json:"areaName,omitempty"`
	AreaTitle         *string    `json:"areaTitle,omitempty"`
	AreaClosed        *bool      `json:"areaClosed,omitempty"`
	Content           *string    `json:"content"`
	AuthorID          *int64     `json:"authorId,omitempty"`
	NickName          *string    `json:"nickName"`
	Avatar            *string    `json:"avatar"`
	Email             *string    `json:"email,omitempty"`
	IP                *string    `json:"ip,omitempty"`
	Location          *string    `json:"location"`
	Platform          *string    `json:"platform"`
	Browser           *string    `json:"browser"`
	Website           *string    `json:"website"`
	IsOwner           bool       `json:"isOwner"`
	IsFriend          bool       `json:"isFriend"`
	IsAuthor          bool       `json:"isAuthor"`
	IsViewed          bool       `json:"isViewed"`
	IsTop             bool       `json:"isTop"`
	IsFederated       bool       `json:"isFederated"`
	FederatedProtocol *string    `json:"federatedProtocol"`
	FederatedActor    *string    `json:"federatedActor"`
	CanReply          bool       `json:"canReply"`
	Status            string     `json:"status"`
	ParentID          *string    `json:"parentId"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	DeletedAt         *time.Time `json:"deletedAt,omitempty"`
	IsDeleted         bool       `json:"isDeleted"`
}

type AdminCommentListResp struct {
	Items []AdminCommentResp `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
}

type AdminVisitorResp struct {
	VisitorID        string     `json:"visitorId"`
	NickName         *string    `json:"nickName,omitempty"`
	Email            *string    `json:"email,omitempty"`
	Website          *string    `json:"website,omitempty"`
	IP               *string    `json:"ip,omitempty"`
	Location         *string    `json:"location,omitempty"`
	Platform         *string    `json:"platform,omitempty"`
	Browser          *string    `json:"browser,omitempty"`
	TotalComments    int64      `json:"totalComments"`
	ApprovedComments int64      `json:"approvedComments"`
	PendingComments  int64      `json:"pendingComments"`
	RejectedComments int64      `json:"rejectedComments"`
	BlockedComments  int64      `json:"blockedComments"`
	DeletedComments  int64      `json:"deletedComments"`
	TopComments      int64      `json:"topComments"`
	ActiveDays       int64      `json:"activeDays"`
	TotalLikes       int64      `json:"totalLikes"`
	UniqueLikedItems int64      `json:"uniqueLikedItems"`
	TotalViews       int64      `json:"totalViews"`
	UniqueViewItems  int64      `json:"uniqueViewItems"`
	FirstSeenAt      time.Time  `json:"firstSeenAt"`
	LastSeenAt       time.Time  `json:"lastSeenAt"`
	LastLikedAt      *time.Time `json:"lastLikedAt,omitempty"`
	LastViewedAt     *time.Time `json:"lastViewedAt,omitempty"`
}

type AdminVisitorListResp struct {
	Items []AdminVisitorResp `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
}

type AdminVisitorRecentCommentResp struct {
	ID        string    `json:"id"`
	AreaID    int64     `json:"areaId"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	IsDeleted bool      `json:"isDeleted"`
}

type AdminVisitorProfileResp struct {
	Profile        AdminVisitorResp                `json:"profile"`
	RecentComments []AdminVisitorRecentCommentResp `json:"recentComments"`
}

type AdminVisitorDistributionResp struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type AdminVisitorTrendResp struct {
	Date              string `json:"date"`
	ActiveVisitors    int64  `json:"activeVisitors"`
	NewVisitors       int64  `json:"newVisitors"`
	ReturningVisitors int64  `json:"returningVisitors"`
	Views             int64  `json:"views"`
	Likes             int64  `json:"likes"`
	Comments          int64  `json:"comments"`
}

type AdminVisitorFunnelResp struct {
	ViewVisitors      int64   `json:"viewVisitors"`
	LikeVisitors      int64   `json:"likeVisitors"`
	CommentVisitors   int64   `json:"commentVisitors"`
	LikeRate          float64 `json:"likeRate"`
	CommentRateByView float64 `json:"commentRateByView"`
	CommentRateByLike float64 `json:"commentRateByLike"`
}

type AdminVisitorSegmentsResp struct {
	Active1D      int64 `json:"active1d"`
	Active3D      int64 `json:"active3d"`
	Active7D      int64 `json:"active7d"`
	Active30D     int64 `json:"active30d"`
	HighlyEngaged int64 `json:"highlyEngaged"`
}

type AdminVisitorInsightsResp struct {
	Days        int                            `json:"days"`
	GeneratedAt time.Time                      `json:"generatedAt"`
	DataSource  string                         `json:"dataSource"`
	PlatformTop []AdminVisitorDistributionResp `json:"platformTop"`
	BrowserTop  []AdminVisitorDistributionResp `json:"browserTop"`
	LocationTop []AdminVisitorDistributionResp `json:"locationTop"`
	Trend       []AdminVisitorTrendResp        `json:"trend"`
	Funnel      AdminVisitorFunnelResp         `json:"funnel"`
	Segments    AdminVisitorSegmentsResp       `json:"segments"`
}

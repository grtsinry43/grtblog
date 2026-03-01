package contract

import "time"

type CreateCommentLoginReq struct {
	Content   string `json:"content" validate:"required"`
	ParentID  *int64 `json:"parentId"`
	VisitorID string `json:"visitorId"`
}

type CreateCommentVisitorReq struct {
	Content   string  `json:"content" validate:"required"`
	NickName  *string `json:"nickName" validate:"required,max=255"`
	Email     *string `json:"email" validate:"required,max=255"`
	Website   *string `json:"website" validate:"max=255"`
	ParentID  *int64  `json:"parentId"`
	VisitorID string  `json:"visitorId"`
}

type UpdateCommentReq struct {
	Content string `json:"content" validate:"required"`
}

type ListAdminCommentsReq struct {
	AreaID       *int64 `json:"areaId"`
	Status       string `json:"status"`
	OnlyUnviewed *bool  `json:"onlyUnviewed"`
	Page         int    `json:"page" validate:"min=1"`
	PageSize     int    `json:"pageSize" validate:"min=1,max=100"`
}

type ReplyCommentReq struct {
	Content string `json:"content" validate:"required"`
}

type UpdateCommentStatusReq struct {
	Status string `json:"status" validate:"required"`
}

type SetCommentAuthorReq struct {
	IsAuthor bool `json:"isAuthor"`
}

type SetCommentTopReq struct {
	IsTop bool `json:"isTop"`
}

type SetCommentAreaCloseReq struct {
	IsClosed bool `json:"isClosed"`
}

type MarkCommentsViewedReq struct {
	IDs      []string `json:"ids"`
	IsViewed *bool    `json:"isViewed,omitempty"`
}

type ImportCommentReq struct {
	ID                *int64     `json:"id,omitempty"`
	AreaID            int64      `json:"areaId" validate:"required"`
	Content           string     `json:"content" validate:"required"`
	AuthorID          *int64     `json:"authorId,omitempty"`
	VisitorID         *string    `json:"visitorId,omitempty"`
	NickName          *string    `json:"nickName,omitempty"`
	IP                *string    `json:"ip,omitempty"`
	Location          *string    `json:"location,omitempty"`
	Platform          *string    `json:"platform,omitempty"`
	Browser           *string    `json:"browser,omitempty"`
	Email             *string    `json:"email,omitempty"`
	Website           *string    `json:"website,omitempty"`
	IsOwner           *bool      `json:"isOwner,omitempty"`
	IsFriend          *bool      `json:"isFriend,omitempty"`
	IsAuthor          *bool      `json:"isAuthor,omitempty"`
	IsViewed          *bool      `json:"isViewed,omitempty"`
	IsTop             *bool      `json:"isTop,omitempty"`
	IsFederated       *bool      `json:"isFederated,omitempty"`
	FederatedProtocol *string    `json:"federatedProtocol,omitempty"`
	FederatedActor    *string    `json:"federatedActor,omitempty"`
	FederatedObjectID *string    `json:"federatedObjectId,omitempty"`
	CanReply          *bool      `json:"canReply,omitempty"`
	Status            *string    `json:"status,omitempty"`
	ParentID          *int64     `json:"parentId,omitempty"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	UpdatedAt         *time.Time `json:"updatedAt,omitempty"`
	DeletedAt         *time.Time `json:"deletedAt,omitempty"`
}

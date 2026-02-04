package contract

type CreateCommentLoginReq struct {
	Content  string `json:"content" validate:"required"`
	ParentID *int64 `json:"parentId"`
}

type CreateCommentVisitorReq struct {
	Content  string  `json:"content" validate:"required"`
	NickName *string `json:"nickName" validate:"required,max=255"`
	Email    *string `json:"email" validate:"required,max=255"`
	Website  *string `json:"website" validate:"max=255"`
	ParentID *int64  `json:"parentId"`
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
	IDs      []int64 `json:"ids"`
	IsViewed *bool   `json:"isViewed,omitempty"`
}

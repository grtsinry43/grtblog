package contract

import "time"

type CreateCommentResp struct {
	ID        int64      `json:"id"`
	AreaID    int64      `json:"areaId"`
	Content   string     `json:"content"`
	NickName  *string    `json:"nickName"`
	Location  *string    `json:"location"`
	Platform  *string    `json:"platform"`
	Browser   *string    `json:"browser"`
	Website   *string    `json:"website"`
	IsOwner   bool       `json:"isOwner"`
	IsFriend  bool       `json:"isFriend"`
	IsAuthor  bool       `json:"isAuthor"`
	IsViewed  bool       `json:"isViewed"`
	IsTop     bool       `json:"isTop"`
	Status    string     `json:"status"`
	ParentID  *int64     `json:"parentId"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
	IsDeleted bool       `json:"isDeleted"`
}

type CommentNodeResp struct {
	ID        int64             `json:"id"`
	AreaID    int64             `json:"areaId"`
	Content   *string           `json:"content"`
	NickName  *string           `json:"nickName"`
	Location  *string           `json:"location"`
	Platform  *string           `json:"platform"`
	Browser   *string           `json:"browser"`
	Website   *string           `json:"website"`
	IsOwner   bool              `json:"isOwner"`
	IsFriend  bool              `json:"isFriend"`
	IsAuthor  bool              `json:"isAuthor"`
	IsViewed  bool              `json:"isViewed"`
	IsTop     bool              `json:"isTop"`
	Status    string            `json:"status"`
	ParentID  *int64            `json:"parentId"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
	DeletedAt *time.Time        `json:"deletedAt,omitempty"`
	IsDeleted bool              `json:"isDeleted"`
	Children  []CommentNodeResp `json:"children,omitempty"`
}

type PublicCommentListResp struct {
	Items    []CommentNodeResp `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Size     int               `json:"size"`
	IsClosed bool              `json:"isClosed"`
}

type AdminCommentResp struct {
	ID         int64      `json:"id"`
	AreaID     int64      `json:"areaId"`
	AreaType   *string    `json:"areaType,omitempty"`
	AreaRefID  *int64     `json:"areaRefId,omitempty"`
	AreaName   *string    `json:"areaName,omitempty"`
	AreaTitle  *string    `json:"areaTitle,omitempty"`
	AreaClosed *bool      `json:"areaClosed,omitempty"`
	Content    *string    `json:"content"`
	AuthorID   *int64     `json:"authorId,omitempty"`
	NickName   *string    `json:"nickName"`
	Email      *string    `json:"email,omitempty"`
	IP         *string    `json:"ip,omitempty"`
	Location   *string    `json:"location"`
	Platform   *string    `json:"platform"`
	Browser    *string    `json:"browser"`
	Website    *string    `json:"website"`
	IsOwner    bool       `json:"isOwner"`
	IsFriend   bool       `json:"isFriend"`
	IsAuthor   bool       `json:"isAuthor"`
	IsViewed   bool       `json:"isViewed"`
	IsTop      bool       `json:"isTop"`
	Status     string     `json:"status"`
	ParentID   *int64     `json:"parentId"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt,omitempty"`
	IsDeleted  bool       `json:"isDeleted"`
}

type AdminCommentListResp struct {
	Items []AdminCommentResp `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
}

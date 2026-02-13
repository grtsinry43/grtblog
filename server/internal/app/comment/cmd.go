package comment

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

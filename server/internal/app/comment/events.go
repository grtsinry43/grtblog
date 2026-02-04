package comment

import "time"

type CommentCreated struct {
	ID       int64
	AreaID   int64
	ParentID *int64
	AuthorID *int64
	NickName string
	Email    string
	Content  string
	Status   string
	At       time.Time
}

func (e CommentCreated) Name() string { return "comment.created" }
func (e CommentCreated) OccurredAt() time.Time {
	return e.At
}

package like

import "time"

type TargetType string

const (
	TargetArticle  TargetType = "article"
	TargetMoment   TargetType = "moment"
	TargetPage     TargetType = "page"
	TargetThinking TargetType = "thinking"
)

type ContentLike struct {
	ID         int64
	TargetType TargetType
	TargetID   int64
	UserID     *int64
	VisitorID  *string
	CreatedAt  time.Time
}

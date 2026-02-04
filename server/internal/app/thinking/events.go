package thinking

import "time"

type ThinkingCreated struct {
	ID       int64
	AuthorID int64
	Content  string
	At       time.Time
}

func (e ThinkingCreated) Name() string { return "thinking.created" }
func (e ThinkingCreated) OccurredAt() time.Time {
	return e.At
}

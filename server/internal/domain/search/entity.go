package search

import "time"

type Kind string

const (
	KindArticle  Kind = "article"
	KindMoment   Kind = "moment"
	KindPage     Kind = "page"
	KindThinking Kind = "thinking"
)

type Hit struct {
	Kind      Kind
	ID        int64
	Title     string
	Summary   string
	Snippet   string
	ShortURL  *string
	CreatedAt time.Time
	Score     float64
}

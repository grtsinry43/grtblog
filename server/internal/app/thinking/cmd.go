package thinking

type CreateThinkingCmd struct {
	Content      string
	AuthorID     int64
	AllowComment *bool
}

type UpdateThinkingCmd struct {
	ID           int64
	Content      string
	AllowComment *bool
}

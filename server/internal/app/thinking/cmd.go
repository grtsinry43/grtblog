package thinking

import "time"

type CreateThinkingCmd struct {
	Content      string
	AuthorID     int64
	AllowComment *bool
	CreatedAt    *time.Time
}

type UpdateThinkingCmd struct {
	ID           int64
	Content      string
	AllowComment *bool
}

// BatchDeleteCmd 批量删除思考命令。
type BatchDeleteCmd struct {
	IDs []int64
}

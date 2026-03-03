package moment

import "time"

// CreateMomentCmd 创建手记命令。
type CreateMomentCmd struct {
	Title        string
	Summary      string
	AISummary    *string
	Content      string
	Image        *string
	ColumnID     *int64
	TopicIDs     []int64
	ShortURL     *string
	IsPublished  bool
	IsTop        bool
	AllowComment *bool
	IsOriginal   bool
	ExtInfo      []byte
	CreatedAt    *time.Time // 可选：因为可能会有自定义发布时间的需求
	Views        *int64     // 可选：迁移时可指定初始阅读量
}

// UpdateMomentCmd 更新手记命令。
type UpdateMomentCmd struct {
	ID           int64
	Title        string
	Summary      string
	AISummary    *string
	Content      string
	Image        *string
	ColumnID     *int64
	TopicIDs     []int64
	ShortURL     string
	IsPublished  bool
	IsTop        bool
	AllowComment *bool
	IsOriginal   bool
	ExtInfo      []byte
}

// BatchSetPublishedCmd 批量设置手记发布状态命令。
type BatchSetPublishedCmd struct {
	IDs         []int64
	IsPublished bool
}

// BatchSetTopCmd 批量设置手记置顶状态命令。
type BatchSetTopCmd struct {
	IDs   []int64
	IsTop bool
}

// BatchDeleteCmd 批量删除手记命令。
type BatchDeleteCmd struct {
	IDs []int64
}

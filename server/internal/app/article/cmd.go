package article

import "time"

// CreateArticleCmd 创建文章命令。
type CreateArticleCmd struct {
	Title        string
	Summary      string
	AISummary    *string
	LeadIn       *string
	Content      string
	Cover        *string
	CategoryID   *int64
	TagIDs       []int64
	ShortURL     *string
	IsPublished  bool
	IsTop        bool
	AllowComment *bool
	IsOriginal   bool
	ExtInfo      []byte
	CreatedAt    *time.Time // 可选：因为可能会有自定义发布时间的需求
	Views        *int64     // 可选：迁移时可指定初始阅读量
}

// UpdateArticleCmd 更新文章命令。
type UpdateArticleCmd struct {
	ID           int64
	Title        string
	Summary      string
	AISummary    *string
	LeadIn       *string
	Content      string
	Cover        *string
	CategoryID   *int64
	TagIDs       []int64
	ShortURL     string
	IsPublished  bool
	IsTop        bool
	AllowComment *bool
	IsOriginal   bool
	ExtInfo      []byte
}

// ResetFederationSignalsCmd 重置文章联合条目状态命令。
type ResetFederationSignalsCmd struct {
	ID        int64
	Mentions  []string
	Citations []string
	Retrigger bool
}

// BatchSetPublishedCmd 批量设置文章发布状态命令。
type BatchSetPublishedCmd struct {
	IDs         []int64
	IsPublished bool
}

// BatchSetTopCmd 批量设置文章置顶状态命令。
type BatchSetTopCmd struct {
	IDs   []int64
	IsTop bool
}

// BatchDeleteCmd 批量删除文章命令。
type BatchDeleteCmd struct {
	IDs []int64
}

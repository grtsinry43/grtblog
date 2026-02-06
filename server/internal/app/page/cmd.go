package page

import "time"

// CreatePageCmd 创建页面命令。
type CreatePageCmd struct {
	Title        string
	Description  *string
	Content      string
	ShortURL     *string
	AllowComment *bool
	IsEnabled    bool
	IsBuiltin    bool
	ExtInfo      []byte
	CreatedAt    *time.Time
}

// UpdatePageCmd 更新页面命令。
type UpdatePageCmd struct {
	ID           int64
	Title        string
	Description  *string
	Content      string
	ShortURL     string
	AllowComment *bool
	IsEnabled    bool
	IsBuiltin    bool
	ExtInfo      []byte
}

// BatchSetEnabledCmd 批量设置页面启用状态命令。
type BatchSetEnabledCmd struct {
	IDs       []int64
	IsEnabled bool
}

// BatchDeleteCmd 批量删除页面命令。
type BatchDeleteCmd struct {
	IDs []int64
}

package album

import "time"

// CreateAlbumCmd 创建相册命令。
type CreateAlbumCmd struct {
	Title        string
	Description  *string
	Cover        *string
	ShortURL     *string
	IsPublished  bool
	AllowComment *bool
	CreatedAt    *time.Time
}

// UpdateAlbumCmd 更新相册命令。
type UpdateAlbumCmd struct {
	ID           int64
	Title        string
	Description  *string
	Cover        *string
	ShortURL     string
	IsPublished  bool
	AllowComment *bool
}

// CreatePhotoCmd 添加照片命令。
type CreatePhotoCmd struct {
	URL         string
	Description *string
	Caption     *string
	Exif        []byte
	SortOrder   int
}

// UpdatePhotoCmd 更新照片命令。
type UpdatePhotoCmd struct {
	ID          int64
	URL         string
	Description *string
	Caption     *string
	Exif        []byte
	SortOrder   int
}

// BatchCreatePhotosCmd 批量添加照片命令。
type BatchCreatePhotosCmd struct {
	AlbumID int64
	Photos  []CreatePhotoCmd
}

// ReorderPhotosCmd 排序照片命令。
type ReorderPhotosCmd struct {
	AlbumID  int64
	PhotoIDs []int64
}

// BatchSetPublishedCmd 批量设置发布状态命令。
type BatchSetPublishedCmd struct {
	IDs         []int64
	IsPublished bool
}

// BatchDeleteCmd 批量删除相册命令。
type BatchDeleteCmd struct {
	IDs []int64
}

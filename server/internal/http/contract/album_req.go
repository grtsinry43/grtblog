package contract

import (
	"encoding/json"
	"strings"
	"time"
)

// CreateAlbumReq 创建相册请求。
type CreateAlbumReq struct {
	Title        string     `json:"title" validate:"required,max=255"`
	Description  *string    `json:"description,omitempty"`
	Cover        *string    `json:"cover,omitempty"`
	ShortURL     *string    `json:"shortUrl,omitempty"`
	IsPublished  bool       `json:"isPublished"`
	AllowComment *bool      `json:"allowComment,omitempty"`
	CreatedAt    *time.Time `json:"createdAt,omitempty"`
}

type createAlbumReqJSON struct {
	Title        string  `json:"title"`
	Description  *string `json:"description"`
	Cover        *string `json:"cover"`
	ShortURL     *string `json:"shortUrl"`
	IsPublished  bool    `json:"isPublished"`
	AllowComment *bool   `json:"allowComment"`
	CreatedAt    *string `json:"createdAt"`
}

func (r *CreateAlbumReq) UnmarshalJSON(data []byte) error {
	var aux createAlbumReqJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.Title = aux.Title
	r.Description = aux.Description
	r.Cover = aux.Cover
	r.ShortURL = aux.ShortURL
	r.IsPublished = aux.IsPublished
	r.AllowComment = aux.AllowComment

	if aux.CreatedAt == nil {
		r.CreatedAt = nil
		return nil
	}
	if strings.TrimSpace(*aux.CreatedAt) == "" {
		now := time.Now()
		r.CreatedAt = &now
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, *aux.CreatedAt)
	if err != nil {
		return err
	}
	r.CreatedAt = &parsed
	return nil
}

// UpdateAlbumReq 更新相册请求。
type UpdateAlbumReq struct {
	Title        string  `json:"title" validate:"required,max=255"`
	Description  *string `json:"description,omitempty"`
	Cover        *string `json:"cover,omitempty"`
	ShortURL     string  `json:"shortUrl" validate:"required"`
	IsPublished  bool    `json:"isPublished"`
	AllowComment *bool   `json:"allowComment,omitempty"`
}

// CreatePhotoReq 添加照片请求。
type CreatePhotoReq struct {
	URL         string   `json:"url" validate:"required"`
	Description *string  `json:"description,omitempty"`
	Caption     *string  `json:"caption,omitempty"`
	Exif        *JSONRaw `json:"exif,omitempty" swaggertype:"object"`
	SortOrder   int      `json:"sortOrder"`
}

// BatchCreatePhotosReq 批量添加照片请求。
type BatchCreatePhotosReq struct {
	Photos []CreatePhotoReq `json:"photos" validate:"required,min=1"`
}

// UpdatePhotoReq 更新照片请求。
type UpdatePhotoReq struct {
	URL         string   `json:"url" validate:"required"`
	Description *string  `json:"description,omitempty"`
	Caption     *string  `json:"caption,omitempty"`
	Exif        *JSONRaw `json:"exif,omitempty" swaggertype:"object"`
	SortOrder   int      `json:"sortOrder"`
}

// ReorderPhotosReq 排序照片请求。
type ReorderPhotosReq struct {
	PhotoIDs []int64 `json:"photoIds" validate:"required,min=1"`
}

// BatchSetAlbumPublishedReq 批量切换发布状态请求。
type BatchSetAlbumPublishedReq struct {
	IDs         []int64 `json:"ids"`
	IsPublished bool    `json:"isPublished"`
}

// BatchDeleteAlbumReq 批量删除相册请求。
type BatchDeleteAlbumReq struct {
	IDs []int64 `json:"ids"`
}

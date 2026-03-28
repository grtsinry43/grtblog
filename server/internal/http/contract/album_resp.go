package contract

import (
	"encoding/json"
	"time"
)

// AlbumResp 相册详情响应。
type AlbumResp struct {
	ID           int64        `json:"id"`
	Title        string       `json:"title"`
	Description  *string      `json:"description,omitempty"`
	Cover        *string      `json:"cover,omitempty"`
	ShortURL     string       `json:"shortUrl"`
	AuthorID     int64        `json:"authorId"`
	CommentID    *int64       `json:"commentAreaId,omitempty"`
	IsPublished  bool         `json:"isPublished"`
	AllowComment bool         `json:"allowComment"`
	PhotoCount   int64        `json:"photoCount"`
	Metrics      *MetricsResp `json:"metrics,omitempty"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
}

// AlbumListItemResp 相册列表项响应。
type AlbumListItemResp struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Cover       *string   `json:"cover,omitempty"`
	ShortURL    string    `json:"shortUrl"`
	IsPublished bool      `json:"isPublished"`
	PhotoCount  int64     `json:"photoCount"`
	Views       int64     `json:"views"`
	Likes       int       `json:"likes"`
	Comments    int       `json:"comments"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// AlbumListResp 相册列表响应。
type AlbumListResp struct {
	Items []AlbumListItemResp `json:"items"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Size  int                 `json:"size"`
}

// PhotoResp 照片响应。
type PhotoResp struct {
	ID           int64            `json:"id"`
	AlbumID      *int64           `json:"albumId,omitempty"`
	URL          string           `json:"url"`
	ThumbnailURL string           `json:"thumbnailUrl,omitempty"`
	Description  *string          `json:"description,omitempty"`
	Caption      *string          `json:"caption,omitempty"`
	Exif         *json.RawMessage `json:"exif,omitempty"`
	SortOrder    int              `json:"sortOrder"`
	CreatedAt    time.Time        `json:"createdAt"`
}

// AlbumDetailResp 相册详情响应（含照片列表）。
type AlbumDetailResp struct {
	AlbumResp
	Photos []PhotoResp `json:"photos"`
}

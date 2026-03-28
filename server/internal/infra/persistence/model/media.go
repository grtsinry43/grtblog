package model

import (
	"time"

	"gorm.io/gorm"
)

type UploadFile struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:name;size:255;not null"`
	Path      string    `gorm:"column:path;size:255;not null"`
	Type      string    `gorm:"column:type;size:45;not null"`
	Size      int64     `gorm:"column:size;not null"`
	Hash      string    `gorm:"column:hash;size:64"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (UploadFile) TableName() string { return "upload_file" }

type Photo struct {
	ID          int64          `gorm:"column:id;primaryKey"`
	AlbumID     *int64         `gorm:"column:album_id"`
	URL         string         `gorm:"column:url;size:255;not null"`
	Description *string        `gorm:"column:description"`
	Caption     *string        `gorm:"column:caption"`
	Exif        []byte         `gorm:"column:exif;type:jsonb"`
	SortOrder   int            `gorm:"column:sort_order;default:0"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Photo) TableName() string { return "photo" }

type Album struct {
	ID          int64          `gorm:"column:id;primaryKey"`
	Title       string         `gorm:"column:title;size:255;not null"`
	Description *string        `gorm:"column:description;type:text"`
	Cover       *string        `gorm:"column:cover;size:255"`
	ShortURL    string         `gorm:"column:short_url;size:255;not null"`
	AuthorID    int64          `gorm:"column:author_id;not null"`
	CommentID   *int64         `gorm:"column:comment_id"`
	IsPublished bool           `gorm:"column:is_published;default:false"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Album) TableName() string { return "album" }

type AlbumMetrics struct {
	AlbumID   int64     `gorm:"column:album_id;primaryKey"`
	Views     int64     `gorm:"column:views;default:0"`
	Likes     int       `gorm:"column:likes;default:0"`
	Comments  int       `gorm:"column:comments;default:0"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (AlbumMetrics) TableName() string { return "album_metrics" }

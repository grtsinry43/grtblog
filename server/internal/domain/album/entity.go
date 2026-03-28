package album

import "time"

type Album struct {
	ID          int64
	Title       string
	Description *string
	Cover       *string
	ShortURL    string
	AuthorID    int64
	CommentID   *int64
	IsPublished bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type AlbumMetrics struct {
	AlbumID   int64
	Views     int64
	Likes     int
	Comments  int
	UpdatedAt time.Time
}

type Photo struct {
	ID          int64
	AlbumID     *int64
	URL         string
	Description *string
	Caption     *string
	Exif        []byte // JSONB: 完整 EXIF 数据（含 GPS、设备、色调等）
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

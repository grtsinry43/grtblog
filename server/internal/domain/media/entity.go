package media

import "time"

type UploadFile struct {
	ID        int64
	Name      string
	Path      string
	Type      string
	Size      int64
	Hash      string
	CreatedAt time.Time
}

type Photo struct {
	ID          int64
	AlbumID     *int64
	URL         string
	Description *string
	Caption     *string
	Exif        []byte
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

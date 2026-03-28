package album

import "context"

// Repository 定义相册相关的持久化操作。
type Repository interface {
	// Album CRUD
	CreateAlbum(ctx context.Context, album *Album) error
	GetAlbumByID(ctx context.Context, id int64) (*Album, error)
	GetAlbumByShortURL(ctx context.Context, shortURL string) (*Album, error)
	UpdateAlbum(ctx context.Context, album *Album) error
	DeleteAlbum(ctx context.Context, id int64) error
	ListAlbums(ctx context.Context, opts AlbumListOptionsInternal) ([]*Album, int64, error)
	ListPublicAlbums(ctx context.Context, opts AlbumListOptions) ([]*Album, int64, error)
	// ListPublishedAlbumShortURLs 用于 ISR 路由发现。
	ListPublishedAlbumShortURLs(ctx context.Context) ([]string, error)

	// Album Metrics
	GetAlbumMetrics(ctx context.Context, albumID int64) (*AlbumMetrics, error)
	IncrementAlbumViews(ctx context.Context, albumID int64) error

	// Photo CRUD
	CreatePhoto(ctx context.Context, photo *Photo) error
	BatchCreatePhotos(ctx context.Context, photos []*Photo) error
	GetPhotoByID(ctx context.Context, id int64) (*Photo, error)
	UpdatePhoto(ctx context.Context, photo *Photo) error
	DeletePhoto(ctx context.Context, id int64) error
	ListPhotosByAlbumID(ctx context.Context, albumID int64) ([]*Photo, error)
	CountPhotosByAlbumID(ctx context.Context, albumID int64) (int64, error)
	ReorderPhotos(ctx context.Context, albumID int64, photoIDs []int64) error
}

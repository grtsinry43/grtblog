package album

import "time"

type AlbumCreated struct {
	ID        int64
	AuthorID  int64
	Title     string
	ShortURL  string
	Published bool
	At        time.Time
}

func (e AlbumCreated) Name() string        { return "album.created" }
func (e AlbumCreated) OccurredAt() time.Time { return e.At }

type AlbumUpdated struct {
	ID        int64
	AuthorID  int64
	Title     string
	ShortURL  string
	Published bool
	At        time.Time
}

func (e AlbumUpdated) Name() string        { return "album.updated" }
func (e AlbumUpdated) OccurredAt() time.Time { return e.At }

type AlbumPublished struct {
	ID       int64
	AuthorID int64
	Title    string
	ShortURL string
	At       time.Time
}

func (e AlbumPublished) Name() string        { return "album.published" }
func (e AlbumPublished) OccurredAt() time.Time { return e.At }

type AlbumUnpublished struct {
	ID       int64
	AuthorID int64
	Title    string
	ShortURL string
	At       time.Time
}

func (e AlbumUnpublished) Name() string        { return "album.unpublished" }
func (e AlbumUnpublished) OccurredAt() time.Time { return e.At }

type AlbumDeleted struct {
	ID       int64
	AuthorID int64
	Title    string
	ShortURL string
	At       time.Time
}

func (e AlbumDeleted) Name() string        { return "album.deleted" }
func (e AlbumDeleted) OccurredAt() time.Time { return e.At }

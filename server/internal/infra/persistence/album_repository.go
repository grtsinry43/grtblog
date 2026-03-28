package persistence

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/contentutil"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/album"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type AlbumRepository struct {
	db *gorm.DB
}

func NewAlbumRepository(db *gorm.DB) *AlbumRepository {
	return &AlbumRepository{db: db}
}

// --------------- Album CRUD ---------------

func (r *AlbumRepository) CreateAlbum(ctx context.Context, a *album.Album) error {
	m := mapAlbumToModel(a)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(m).Error; err != nil {
			if isAlbumShortURLConstraint(err) {
				return album.ErrAlbumShortURLExists
			}
			return err
		}

		areaID, err := createCommentArea(tx, contentutil.CommentAreaTypeAlbum, "相册", m.Title, m.ID)
		if err != nil {
			return err
		}
		if err := tx.Model(&model.Album{}).
			Where("id = ?", m.ID).
			Update("comment_id", areaID).Error; err != nil {
			return err
		}
		m.CommentID = &areaID

		metrics := model.AlbumMetrics{
			AlbumID:  m.ID,
			Views:    0,
			Likes:    0,
			Comments: 0,
		}
		if err := tx.Create(&metrics).Error; err != nil {
			return err
		}

		a.ID = m.ID
		a.CommentID = m.CommentID
		a.CreatedAt = m.CreatedAt
		a.UpdatedAt = m.UpdatedAt
		return nil
	})
}

func (r *AlbumRepository) GetAlbumByID(ctx context.Context, id int64) (*album.Album, error) {
	var m model.Album
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, album.ErrAlbumNotFound
		}
		return nil, err
	}
	entity := mapAlbumToDomain(m)
	return &entity, nil
}

func (r *AlbumRepository) GetAlbumByShortURL(ctx context.Context, shortURL string) (*album.Album, error) {
	var m model.Album
	if err := r.db.WithContext(ctx).Where("short_url = ?", shortURL).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, album.ErrAlbumNotFound
		}
		return nil, err
	}
	entity := mapAlbumToDomain(m)
	return &entity, nil
}

func (r *AlbumRepository) UpdateAlbum(ctx context.Context, a *album.Album) error {
	updates := map[string]any{
		"title":        a.Title,
		"description":  a.Description,
		"cover":        a.Cover,
		"short_url":    a.ShortURL,
		"is_published": a.IsPublished,
	}
	if err := r.db.WithContext(ctx).
		Model(&model.Album{}).
		Where("id = ?", a.ID).
		Updates(updates).Error; err != nil {
		if isAlbumShortURLConstraint(err) {
			return album.ErrAlbumShortURLExists
		}
		return err
	}
	return nil
}

func (r *AlbumRepository) DeleteAlbum(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m model.Album
		if err := tx.First(&m, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return album.ErrAlbumNotFound
			}
			return err
		}
		if m.CommentID != nil {
			if err := deleteCommentArea(tx, *m.CommentID); err != nil {
				return err
			}
		}
		// Soft-delete photos belonging to this album.
		if err := tx.Where("album_id = ?", id).Delete(&model.Photo{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Album{}, id).Error
	})
}

func (r *AlbumRepository) ListAlbums(ctx context.Context, opts album.AlbumListOptionsInternal) ([]*album.Album, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Album{})
	if opts.Published != nil {
		q = q.Where("is_published = ?", *opts.Published)
	}
	if opts.Search != nil && *opts.Search != "" {
		q = q.Where("title ILIKE ?", "%"+*opts.Search+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []model.Album
	offset := (opts.Page - 1) * opts.PageSize
	if err := q.Order("created_at DESC").Offset(offset).Limit(opts.PageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	albums := make([]*album.Album, len(records))
	for i, rec := range records {
		a := mapAlbumToDomain(rec)
		albums[i] = &a
	}
	return albums, total, nil
}

func (r *AlbumRepository) ListPublicAlbums(ctx context.Context, opts album.AlbumListOptions) ([]*album.Album, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Album{}).Where("is_published = ?", true)
	if opts.Search != nil && *opts.Search != "" {
		q = q.Where("title ILIKE ?", "%"+*opts.Search+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []model.Album
	offset := (opts.Page - 1) * opts.PageSize
	if err := q.Order("created_at DESC").Offset(offset).Limit(opts.PageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	albums := make([]*album.Album, len(records))
	for i, rec := range records {
		a := mapAlbumToDomain(rec)
		albums[i] = &a
	}
	return albums, total, nil
}

func (r *AlbumRepository) ListPublishedAlbumShortURLs(ctx context.Context) ([]string, error) {
	var urls []string
	if err := r.db.WithContext(ctx).
		Model(&model.Album{}).
		Where("is_published = ?", true).
		Pluck("short_url", &urls).Error; err != nil {
		return nil, err
	}
	return urls, nil
}

// --------------- Album Metrics ---------------

func (r *AlbumRepository) GetAlbumMetrics(ctx context.Context, albumID int64) (*album.AlbumMetrics, error) {
	var m model.AlbumMetrics
	if err := r.db.WithContext(ctx).Where("album_id = ?", albumID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &album.AlbumMetrics{AlbumID: albumID}, nil
		}
		return nil, err
	}
	return &album.AlbumMetrics{
		AlbumID:   m.AlbumID,
		Views:     m.Views,
		Likes:     m.Likes,
		Comments:  m.Comments,
		UpdatedAt: m.UpdatedAt,
	}, nil
}

func (r *AlbumRepository) IncrementAlbumViews(ctx context.Context, albumID int64) error {
	return r.db.WithContext(ctx).
		Model(&model.AlbumMetrics{}).
		Where("album_id = ?", albumID).
		UpdateColumn("views", gorm.Expr("views + 1")).Error
}

// --------------- Photo CRUD ---------------

func (r *AlbumRepository) CreatePhoto(ctx context.Context, p *album.Photo) error {
	m := mapPhotoToModel(p)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	p.ID = m.ID
	p.CreatedAt = m.CreatedAt
	p.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *AlbumRepository) BatchCreatePhotos(ctx context.Context, photos []*album.Photo) error {
	if len(photos) == 0 {
		return nil
	}
	models := make([]model.Photo, len(photos))
	for i, p := range photos {
		models[i] = *mapPhotoToModel(p)
	}
	if err := r.db.WithContext(ctx).Create(&models).Error; err != nil {
		return err
	}
	for i := range photos {
		photos[i].ID = models[i].ID
		photos[i].CreatedAt = models[i].CreatedAt
		photos[i].UpdatedAt = models[i].UpdatedAt
	}
	return nil
}

func (r *AlbumRepository) GetPhotoByID(ctx context.Context, id int64) (*album.Photo, error) {
	var m model.Photo
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, album.ErrPhotoNotFound
		}
		return nil, err
	}
	entity := mapPhotoToDomain(m)
	return &entity, nil
}

func (r *AlbumRepository) UpdatePhoto(ctx context.Context, p *album.Photo) error {
	updates := map[string]any{
		"url":         p.URL,
		"description": p.Description,
		"caption":     p.Caption,
		"exif":        p.Exif,
		"sort_order":  p.SortOrder,
	}
	return r.db.WithContext(ctx).
		Model(&model.Photo{}).
		Where("id = ?", p.ID).
		Updates(updates).Error
}

func (r *AlbumRepository) DeletePhoto(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Photo{}, id).Error
}

func (r *AlbumRepository) ListPhotosByAlbumID(ctx context.Context, albumID int64) ([]*album.Photo, error) {
	var records []model.Photo
	if err := r.db.WithContext(ctx).
		Where("album_id = ?", albumID).
		Order("sort_order ASC, id ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	photos := make([]*album.Photo, len(records))
	for i, rec := range records {
		p := mapPhotoToDomain(rec)
		photos[i] = &p
	}
	return photos, nil
}

func (r *AlbumRepository) CountPhotosByAlbumID(ctx context.Context, albumID int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Photo{}).
		Where("album_id = ?", albumID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *AlbumRepository) ReorderPhotos(ctx context.Context, albumID int64, photoIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, pid := range photoIDs {
			if err := tx.Model(&model.Photo{}).
				Where("id = ? AND album_id = ?", pid, albumID).
				Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// --------------- Mappers ---------------

func mapAlbumToModel(a *album.Album) *model.Album {
	return &model.Album{
		ID:          a.ID,
		Title:       a.Title,
		Description: a.Description,
		Cover:       a.Cover,
		ShortURL:    a.ShortURL,
		AuthorID:    a.AuthorID,
		CommentID:   a.CommentID,
		IsPublished: a.IsPublished,
	}
}

func mapAlbumToDomain(m model.Album) album.Album {
	a := album.Album{
		ID:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		Cover:       m.Cover,
		ShortURL:    m.ShortURL,
		AuthorID:    m.AuthorID,
		CommentID:   m.CommentID,
		IsPublished: m.IsPublished,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if m.DeletedAt.Valid {
		a.DeletedAt = &m.DeletedAt.Time
	}
	return a
}

func mapPhotoToModel(p *album.Photo) *model.Photo {
	return &model.Photo{
		ID:          p.ID,
		AlbumID:     p.AlbumID,
		URL:         p.URL,
		Description: p.Description,
		Caption:     p.Caption,
		Exif:        p.Exif,
		SortOrder:   p.SortOrder,
	}
}

func mapPhotoToDomain(m model.Photo) album.Photo {
	p := album.Photo{
		ID:          m.ID,
		AlbumID:     m.AlbumID,
		URL:         m.URL,
		Description: m.Description,
		Caption:     m.Caption,
		Exif:        m.Exif,
		SortOrder:   m.SortOrder,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if m.DeletedAt.Valid {
		p.DeletedAt = &m.DeletedAt.Time
	}
	return p
}

func isAlbumShortURLConstraint(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "idx_album_short_url")
}

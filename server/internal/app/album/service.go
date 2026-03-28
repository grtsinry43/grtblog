package album

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/contentutil"
	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	domainalbum "github.com/grtsinry43/grtblog-v2/server/internal/domain/album"
	domaincomment "github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
)

type Service struct {
	repo        domainalbum.Repository
	commentRepo domaincomment.CommentRepository
	events      appEvent.Bus
}

func NewService(repo domainalbum.Repository, commentRepo domaincomment.CommentRepository, events appEvent.Bus) *Service {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &Service{repo: repo, commentRepo: commentRepo, events: events}
}

// CreateAlbum 创建相册。
func (s *Service) CreateAlbum(ctx context.Context, authorID int64, cmd CreateAlbumCmd) (*domainalbum.Album, error) {
	shortURL := ""
	if cmd.ShortURL != nil {
		shortURL = strings.TrimSpace(*cmd.ShortURL)
	}
	if shortURL == "" {
		shortURL = contentutil.GenerateShortURLFromTitle(cmd.Title)
	}
	shortURL, err := s.ensureShortURLAvailable(ctx, shortURL)
	if err != nil {
		return nil, err
	}

	a := &domainalbum.Album{
		Title:       cmd.Title,
		Description: cmd.Description,
		Cover:       cmd.Cover,
		ShortURL:    shortURL,
		AuthorID:    authorID,
		IsPublished: cmd.IsPublished,
	}
	if cmd.CreatedAt != nil {
		a.CreatedAt = *cmd.CreatedAt
	}

	if err := s.repo.CreateAlbum(ctx, a); err != nil {
		return nil, err
	}
	if err := s.applyCommentAreaStatus(ctx, a.CommentID, cmd.AllowComment); err != nil {
		return nil, err
	}

	now := time.Now()
	_ = s.events.Publish(ctx, AlbumCreated{
		ID:        a.ID,
		AuthorID:  a.AuthorID,
		Title:     a.Title,
		ShortURL:  a.ShortURL,
		Published: a.IsPublished,
		At:        now,
	})
	if a.IsPublished {
		_ = s.events.Publish(ctx, AlbumPublished{
			ID:       a.ID,
			AuthorID: a.AuthorID,
			Title:    a.Title,
			ShortURL: a.ShortURL,
			At:       now,
		})
	}

	return a, nil
}

// UpdateAlbum 更新相册。
func (s *Service) UpdateAlbum(ctx context.Context, cmd UpdateAlbumCmd) (*domainalbum.Album, error) {
	existing, err := s.repo.GetAlbumByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	prevPublished := existing.IsPublished

	existing.Title = cmd.Title
	existing.Description = cmd.Description
	existing.Cover = cmd.Cover
	shortURL := strings.TrimSpace(cmd.ShortURL)
	if shortURL == "" {
		shortURL = existing.ShortURL
	}
	if shortURL != existing.ShortURL {
		shortURL, err = s.ensureShortURLAvailable(ctx, shortURL)
		if err != nil {
			return nil, err
		}
	}
	existing.ShortURL = shortURL
	existing.IsPublished = cmd.IsPublished

	if err := s.repo.UpdateAlbum(ctx, existing); err != nil {
		return nil, err
	}
	if err := s.applyCommentAreaStatus(ctx, existing.CommentID, cmd.AllowComment); err != nil {
		return nil, err
	}

	now := time.Now()
	_ = s.events.Publish(ctx, AlbumUpdated{
		ID:        existing.ID,
		AuthorID:  existing.AuthorID,
		Title:     existing.Title,
		ShortURL:  existing.ShortURL,
		Published: existing.IsPublished,
		At:        now,
	})
	if !prevPublished && existing.IsPublished {
		_ = s.events.Publish(ctx, AlbumPublished{
			ID:       existing.ID,
			AuthorID: existing.AuthorID,
			Title:    existing.Title,
			ShortURL: existing.ShortURL,
			At:       now,
		})
	}
	if prevPublished && !existing.IsPublished {
		_ = s.events.Publish(ctx, AlbumUnpublished{
			ID:       existing.ID,
			AuthorID: existing.AuthorID,
			Title:    existing.Title,
			ShortURL: existing.ShortURL,
			At:       now,
		})
	}

	return existing, nil
}

// GetAlbumByID 根据 ID 获取相册。
func (s *Service) GetAlbumByID(ctx context.Context, id int64) (*domainalbum.Album, error) {
	return s.repo.GetAlbumByID(ctx, id)
}

// GetAlbumByShortURL 根据短链接获取相册。
func (s *Service) GetAlbumByShortURL(ctx context.Context, shortURL string) (*domainalbum.Album, error) {
	return s.repo.GetAlbumByShortURL(ctx, shortURL)
}

// ListAlbums 获取相册列表（管理后台）。
func (s *Service) ListAlbums(ctx context.Context, opts domainalbum.AlbumListOptionsInternal) ([]*domainalbum.Album, int64, error) {
	return s.repo.ListAlbums(ctx, opts)
}

// ListPublicAlbums 获取已发布的相册列表。
func (s *Service) ListPublicAlbums(ctx context.Context, opts domainalbum.AlbumListOptions) ([]*domainalbum.Album, int64, error) {
	return s.repo.ListPublicAlbums(ctx, opts)
}

// BatchSetPublished 批量设置发布状态。
func (s *Service) BatchSetPublished(ctx context.Context, cmd BatchSetPublishedCmd) error {
	ids := normalizeIDs(cmd.IDs)
	for _, id := range ids {
		a, err := s.repo.GetAlbumByID(ctx, id)
		if err != nil {
			return err
		}
		if a.IsPublished == cmd.IsPublished {
			continue
		}
		a.IsPublished = cmd.IsPublished
		if err := s.repo.UpdateAlbum(ctx, a); err != nil {
			return err
		}

		now := time.Now()
		_ = s.events.Publish(ctx, AlbumUpdated{
			ID:        a.ID,
			AuthorID:  a.AuthorID,
			Title:     a.Title,
			ShortURL:  a.ShortURL,
			Published: a.IsPublished,
			At:        now,
		})
		if cmd.IsPublished {
			_ = s.events.Publish(ctx, AlbumPublished{
				ID: a.ID, AuthorID: a.AuthorID, Title: a.Title, ShortURL: a.ShortURL, At: now,
			})
		} else {
			_ = s.events.Publish(ctx, AlbumUnpublished{
				ID: a.ID, AuthorID: a.AuthorID, Title: a.Title, ShortURL: a.ShortURL, At: now,
			})
		}
	}
	return nil
}

// DeleteAlbum 删除相册。
func (s *Service) DeleteAlbum(ctx context.Context, id int64) error {
	a, err := s.repo.GetAlbumByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteAlbum(ctx, id); err != nil {
		return err
	}
	_ = s.events.Publish(ctx, AlbumDeleted{
		ID:       a.ID,
		AuthorID: a.AuthorID,
		Title:    a.Title,
		ShortURL: a.ShortURL,
		At:       time.Now(),
	})
	return nil
}

// BatchDelete 批量删除相册。
func (s *Service) BatchDelete(ctx context.Context, cmd BatchDeleteCmd) error {
	ids := normalizeIDs(cmd.IDs)
	for _, id := range ids {
		if err := s.DeleteAlbum(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// GetAlbumMetrics 获取相册指标。
func (s *Service) GetAlbumMetrics(ctx context.Context, albumID int64) (*domainalbum.AlbumMetrics, error) {
	return s.repo.GetAlbumMetrics(ctx, albumID)
}

// --------------- Photo 操作 ---------------

// AddPhoto 添加单张照片到相册。
func (s *Service) AddPhoto(ctx context.Context, albumID int64, cmd CreatePhotoCmd) (*domainalbum.Photo, error) {
	if _, err := s.repo.GetAlbumByID(ctx, albumID); err != nil {
		return nil, err
	}
	p := &domainalbum.Photo{
		AlbumID:     &albumID,
		URL:         cmd.URL,
		Description: cmd.Description,
		Caption:     cmd.Caption,
		Exif:        cmd.Exif,
		SortOrder:   cmd.SortOrder,
	}
	if err := s.repo.CreatePhoto(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// AddPhotos 批量添加照片到相册。
func (s *Service) AddPhotos(ctx context.Context, cmd BatchCreatePhotosCmd) ([]*domainalbum.Photo, error) {
	if _, err := s.repo.GetAlbumByID(ctx, cmd.AlbumID); err != nil {
		return nil, err
	}
	photos := make([]*domainalbum.Photo, len(cmd.Photos))
	for i, c := range cmd.Photos {
		albumID := cmd.AlbumID
		photos[i] = &domainalbum.Photo{
			AlbumID:     &albumID,
			URL:         c.URL,
			Description: c.Description,
			Caption:     c.Caption,
			Exif:        c.Exif,
			SortOrder:   c.SortOrder,
		}
	}
	if err := s.repo.BatchCreatePhotos(ctx, photos); err != nil {
		return nil, err
	}

	// Publish album updated event so ISR can regenerate.
	a, _ := s.repo.GetAlbumByID(ctx, cmd.AlbumID)
	if a != nil {
		_ = s.events.Publish(ctx, AlbumUpdated{
			ID: a.ID, AuthorID: a.AuthorID, Title: a.Title,
			ShortURL: a.ShortURL, Published: a.IsPublished, At: time.Now(),
		})
	}

	return photos, nil
}

// UpdatePhoto 更新照片信息。
func (s *Service) UpdatePhoto(ctx context.Context, cmd UpdatePhotoCmd) (*domainalbum.Photo, error) {
	existing, err := s.repo.GetPhotoByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	existing.URL = cmd.URL
	existing.Description = cmd.Description
	existing.Caption = cmd.Caption
	existing.Exif = cmd.Exif
	existing.SortOrder = cmd.SortOrder
	if err := s.repo.UpdatePhoto(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// DeletePhoto 删除照片。
func (s *Service) DeletePhoto(ctx context.Context, id int64) error {
	return s.repo.DeletePhoto(ctx, id)
}

// ListAlbumPhotos 列出相册中的照片。
func (s *Service) ListAlbumPhotos(ctx context.Context, albumID int64) ([]*domainalbum.Photo, error) {
	return s.repo.ListPhotosByAlbumID(ctx, albumID)
}

// ReorderPhotos 重新排序照片。
func (s *Service) ReorderPhotos(ctx context.Context, cmd ReorderPhotosCmd) error {
	if _, err := s.repo.GetAlbumByID(ctx, cmd.AlbumID); err != nil {
		return err
	}
	return s.repo.ReorderPhotos(ctx, cmd.AlbumID, cmd.PhotoIDs)
}

// CountAlbumPhotos 获取相册照片数量。
func (s *Service) CountAlbumPhotos(ctx context.Context, albumID int64) (int64, error) {
	return s.repo.CountPhotosByAlbumID(ctx, albumID)
}

// --------------- helpers ---------------

func (s *Service) ensureShortURLAvailable(ctx context.Context, shortURL string) (string, error) {
	shortURL = strings.TrimSpace(shortURL)
	if shortURL == "" {
		for i := 0; i < 5; i++ {
			candidate := contentutil.GenerateRandomShortURL()
			_, err := s.repo.GetAlbumByShortURL(ctx, candidate)
			if err != nil {
				if errors.Is(err, domainalbum.ErrAlbumNotFound) {
					return candidate, nil
				}
				return "", err
			}
		}
		return "", domainalbum.ErrAlbumShortURLExists
	}

	existing, err := s.repo.GetAlbumByShortURL(ctx, shortURL)
	if err != nil && !errors.Is(err, domainalbum.ErrAlbumNotFound) {
		return "", err
	}
	if err == nil && existing != nil {
		return "", domainalbum.ErrAlbumShortURLExists
	}
	return shortURL, nil
}

func (s *Service) applyCommentAreaStatus(ctx context.Context, areaID *int64, allowComment *bool) error {
	if s.commentRepo == nil || areaID == nil || *areaID <= 0 || allowComment == nil {
		return nil
	}
	return s.commentRepo.SetAreaClosed(ctx, *areaID, !*allowComment)
}

func normalizeIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

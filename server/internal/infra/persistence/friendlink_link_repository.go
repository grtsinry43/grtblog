package persistence

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type FriendLinkRepository struct {
	db   *gorm.DB
	repo *GormRepository[model.FriendLink]
}

func NewFriendLinkRepository(db *gorm.DB) *FriendLinkRepository {
	return &FriendLinkRepository{
		db:   db,
		repo: NewGormRepository[model.FriendLink](db),
	}
}

func (r *FriendLinkRepository) GetByID(ctx context.Context, id int64) (*social.FriendLink, error) {
	rec, err := r.repo.FirstByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, social.ErrFriendLinkNotFound
		}
		return nil, err
	}
	entity := mapFriendLinkToDomain(*rec)
	return &entity, nil
}

func (r *FriendLinkRepository) FindByURL(ctx context.Context, url string) (*social.FriendLink, error) {
	rec, err := r.repo.First(ctx, "url = ?", url)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, social.ErrFriendLinkNotFound
		}
		return nil, err
	}
	entity := mapFriendLinkToDomain(*rec)
	return &entity, nil
}

func (r *FriendLinkRepository) ExistsActiveByUserID(ctx context.Context, userID int64) (bool, error) {
	if userID <= 0 {
		return false, nil
	}
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.FriendLink{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *FriendLinkRepository) Create(ctx context.Context, link *social.FriendLink) error {
	rec := mapFriendLinkToModel(link)
	if err := r.repo.Create(ctx, &rec); err != nil {
		return err
	}
	link.ID = rec.ID
	link.CreatedAt = rec.CreatedAt
	link.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *FriendLinkRepository) Update(ctx context.Context, link *social.FriendLink) error {
	rec := mapFriendLinkToModel(link)
	return r.db.WithContext(ctx).Model(&model.FriendLink{}).
		Where("id = ?", link.ID).
		Updates(map[string]any{
			"name":               rec.Name,
			"url":                rec.URL,
			"logo":               rec.Logo,
			"description":        rec.Description,
			"rss_url":            rec.RSSURL,
			"kind":               rec.Kind,
			"sync_mode":          rec.SyncMode,
			"instance_id":        rec.InstanceID,
			"last_sync_at":       rec.LastSyncAt,
			"last_sync_status":   rec.LastSyncStatus,
			"sync_interval":      rec.SyncInterval,
			"total_posts_cached": rec.TotalPostsCached,
			"user_id":            rec.UserID,
			"is_active":          rec.IsActive,
			"updated_at":         time.Now().UTC(),
		}).Error
}

func (r *FriendLinkRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.FriendLink{}, id).Error
}

func (r *FriendLinkRepository) List(ctx context.Context, options social.FriendLinkListOptions) ([]social.FriendLink, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.FriendLink{})
	if options.IsActive != nil {
		query = query.Where("is_active = ?", *options.IsActive)
	}
	if strings.TrimSpace(options.Kind) != "" {
		query = query.Where("kind = ?", options.Kind)
	}
	if strings.TrimSpace(options.SyncMode) != "" {
		query = query.Where("sync_mode = ?", options.SyncMode)
	}
	if strings.TrimSpace(options.Keyword) != "" {
		search := "%" + strings.TrimSpace(options.Keyword) + "%"
		query = query.Where("url ILIKE ? OR name ILIKE ? OR description ILIKE ?", search, search, search)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (options.Page - 1) * options.PageSize
	var recs []model.FriendLink
	if err := query.
		Order("updated_at DESC").
		Limit(options.PageSize).
		Offset(offset).
		Find(&recs).Error; err != nil {
		return nil, 0, err
	}
	result := make([]social.FriendLink, len(recs))
	for i, rec := range recs {
		result[i] = mapFriendLinkToDomain(rec)
	}
	return result, total, nil
}

func mapFriendLinkToDomain(rec model.FriendLink) social.FriendLink {
	return social.FriendLink{
		ID:               rec.ID,
		Name:             rec.Name,
		URL:              rec.URL,
		Logo:             stringToPtr(rec.Logo),
		Description:      stringToPtr(rec.Description),
		RSSURL:           stringToPtr(rec.RSSURL),
		Kind:             rec.Kind,
		SyncMode:         rec.SyncMode,
		InstanceID:       rec.InstanceID,
		LastSyncAt:       rec.LastSyncAt,
		LastSyncStatus:   rec.LastSyncStatus,
		SyncInterval:     rec.SyncInterval,
		TotalPostsCached: rec.TotalPostsCached,
		UserID:           rec.UserID,
		IsActive:         rec.IsActive,
		CreatedAt:        rec.CreatedAt,
		UpdatedAt:        rec.UpdatedAt,
		DeletedAt:        deletedAtToPtr(rec.DeletedAt),
	}
}

func mapFriendLinkToModel(link *social.FriendLink) model.FriendLink {
	return model.FriendLink{
		ID:               link.ID,
		Name:             link.Name,
		URL:              link.URL,
		Logo:             optionalString(link.Logo),
		Description:      optionalString(link.Description),
		RSSURL:           optionalString(link.RSSURL),
		Kind:             link.Kind,
		SyncMode:         link.SyncMode,
		InstanceID:       link.InstanceID,
		LastSyncAt:       link.LastSyncAt,
		LastSyncStatus:   link.LastSyncStatus,
		SyncInterval:     link.SyncInterval,
		TotalPostsCached: link.TotalPostsCached,
		UserID:           link.UserID,
		IsActive:         link.IsActive,
	}
}

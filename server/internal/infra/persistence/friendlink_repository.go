package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type FriendLinkApplicationRepository struct {
	db   *gorm.DB
	repo *GormRepository[model.FriendLinkApplication]
}

func NewFriendLinkApplicationRepository(db *gorm.DB) *FriendLinkApplicationRepository {
	return &FriendLinkApplicationRepository{
		db:   db,
		repo: NewGormRepository[model.FriendLinkApplication](db),
	}
}

func (r *FriendLinkApplicationRepository) GetByID(ctx context.Context, id int64) (*social.FriendLinkApplication, error) {
	rec, err := r.repo.FirstByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, social.ErrFriendLinkApplicationNotFound
		}
		return nil, err
	}
	entity := mapFriendLinkApplicationToDomain(*rec)
	return &entity, nil
}

func (r *FriendLinkApplicationRepository) FindByURL(ctx context.Context, url string) (*social.FriendLinkApplication, error) {
	rec, err := r.repo.First(ctx, "url = ?", url)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, social.ErrFriendLinkApplicationNotFound
		}
		return nil, err
	}
	entity := mapFriendLinkApplicationToDomain(*rec)
	return &entity, nil
}

func (r *FriendLinkApplicationRepository) Create(ctx context.Context, app *social.FriendLinkApplication) error {
	rec := mapFriendLinkApplicationToModel(app)
	if err := r.repo.Create(ctx, &rec); err != nil {
		return err
	}
	app.ID = rec.ID
	app.CreatedAt = rec.CreatedAt
	app.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *FriendLinkApplicationRepository) Update(ctx context.Context, app *social.FriendLinkApplication) error {
	rec := mapFriendLinkApplicationToModel(app)
	return r.db.WithContext(ctx).Model(&model.FriendLinkApplication{}).
		Where("url = ?", app.URL).
		Updates(map[string]any{
			"name":                rec.Name,
			"logo":                rec.Logo,
			"description":         rec.Description,
			"apply_channel":       rec.ApplyChannel,
			"requested_sync_mode": rec.RequestedSyncMode,
			"rss_url":             rec.RSSURL,
			"instance_url":        rec.InstanceURL,
			"manifest":            rec.Manifest,
			"signature_key_id":    rec.SignatureKeyID,
			"signature_verified":  rec.SignatureVerified,
			"user_id":             rec.UserID,
			"message":             rec.Message,
			"status":              rec.Status,
		}).Error
}

func (r *FriendLinkApplicationRepository) UpdateByID(ctx context.Context, app *social.FriendLinkApplication) error {
	rec := mapFriendLinkApplicationToModel(app)
	return r.db.WithContext(ctx).Model(&model.FriendLinkApplication{}).
		Where("id = ?", app.ID).
		Updates(map[string]any{
			"name":                rec.Name,
			"url":                 rec.URL,
			"logo":                rec.Logo,
			"description":         rec.Description,
			"apply_channel":       rec.ApplyChannel,
			"requested_sync_mode": rec.RequestedSyncMode,
			"rss_url":             rec.RSSURL,
			"instance_url":        rec.InstanceURL,
			"manifest":            rec.Manifest,
			"signature_key_id":    rec.SignatureKeyID,
			"signature_verified":  rec.SignatureVerified,
			"user_id":             rec.UserID,
			"message":             rec.Message,
			"status":              rec.Status,
		}).Error
}

func (r *FriendLinkApplicationRepository) List(ctx context.Context, options social.FriendLinkApplicationListOptions) ([]social.FriendLinkApplication, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.FriendLinkApplication{})
	if options.Status != "" {
		query = query.Where("status = ?", options.Status)
	}
	if options.ApplyChannel != "" {
		query = query.Where("apply_channel = ?", options.ApplyChannel)
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
	var recs []model.FriendLinkApplication
	if err := query.
		Order("updated_at DESC").
		Limit(options.PageSize).
		Offset(offset).
		Find(&recs).Error; err != nil {
		return nil, 0, err
	}
	result := make([]social.FriendLinkApplication, len(recs))
	for i, rec := range recs {
		result[i] = mapFriendLinkApplicationToDomain(rec)
	}
	return result, total, nil
}

func mapFriendLinkApplicationToDomain(rec model.FriendLinkApplication) social.FriendLinkApplication {
	return social.FriendLinkApplication{
		ID:                rec.ID,
		Name:              rec.Name,
		URL:               rec.URL,
		Logo:              rec.Logo,
		Description:       rec.Description,
		ApplyChannel:      rec.ApplyChannel,
		RequestedSyncMode: rec.RequestedSyncMode,
		RSSURL:            rec.RSSURL,
		InstanceURL:       rec.InstanceURL,
		Manifest:          json.RawMessage(rec.Manifest),
		SignatureKeyID:    rec.SignatureKeyID,
		SignatureVerified: rec.SignatureVerified,
		UserID:            rec.UserID,
		Message:           rec.Message,
		Status:            rec.Status,
		CreatedAt:         rec.CreatedAt,
		UpdatedAt:         rec.UpdatedAt,
	}
}

func mapFriendLinkApplicationToModel(app *social.FriendLinkApplication) model.FriendLinkApplication {
	return model.FriendLinkApplication{
		ID:                app.ID,
		Name:              app.Name,
		URL:               app.URL,
		Logo:              app.Logo,
		Description:       app.Description,
		ApplyChannel:      app.ApplyChannel,
		RequestedSyncMode: app.RequestedSyncMode,
		RSSURL:            app.RSSURL,
		InstanceURL:       app.InstanceURL,
		Manifest:          datatypes.JSON(app.Manifest),
		SignatureKeyID:    app.SignatureKeyID,
		SignatureVerified: app.SignatureVerified,
		UserID:            app.UserID,
		Message:           app.Message,
		Status:            app.Status,
	}
}

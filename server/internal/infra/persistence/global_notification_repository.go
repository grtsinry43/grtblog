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

type GlobalNotificationRepository struct {
	db *gorm.DB
}

func NewGlobalNotificationRepository(db *gorm.DB) *GlobalNotificationRepository {
	return &GlobalNotificationRepository{db: db}
}

func (r *GlobalNotificationRepository) GetByID(ctx context.Context, id int64) (*social.GlobalNotification, error) {
	var rec model.GlobalNotification
	if err := r.db.WithContext(ctx).First(&rec, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, social.ErrGlobalNotificationNotFound
		}
		return nil, err
	}
	entity := mapGlobalNotificationToDomain(rec)
	return &entity, nil
}

func (r *GlobalNotificationRepository) Create(ctx context.Context, notification *social.GlobalNotification) error {
	rec := mapGlobalNotificationToModel(notification)
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	notification.ID = rec.ID
	notification.CreatedAt = rec.CreatedAt
	notification.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *GlobalNotificationRepository) Update(ctx context.Context, notification *social.GlobalNotification) error {
	rec := mapGlobalNotificationToModel(notification)
	result := r.db.WithContext(ctx).Model(&model.GlobalNotification{}).
		Where("id = ?", notification.ID).
		Updates(map[string]any{
			"content":     rec.Content,
			"publish_at":  rec.PublishAt,
			"expire_at":   rec.ExpireAt,
			"allow_close": rec.AllowClose,
			"updated_at":  time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return social.ErrGlobalNotificationNotFound
	}
	return nil
}

func (r *GlobalNotificationRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&model.GlobalNotification{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return social.ErrGlobalNotificationNotFound
	}
	return nil
}

func (r *GlobalNotificationRepository) List(ctx context.Context, options social.GlobalNotificationListOptions) ([]social.GlobalNotification, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.GlobalNotification{})
	now := time.Now().UTC()
	switch strings.TrimSpace(strings.ToLower(options.Status)) {
	case "active":
		query = query.Where("publish_at <= ? AND expire_at >= ?", now, now)
	case "upcoming":
		query = query.Where("publish_at > ?", now)
	case "expired":
		query = query.Where("expire_at < ?", now)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (options.Page - 1) * options.PageSize
	var recs []model.GlobalNotification
	if err := query.Order("publish_at DESC, id DESC").
		Limit(options.PageSize).
		Offset(offset).
		Find(&recs).Error; err != nil {
		return nil, 0, err
	}

	items := make([]social.GlobalNotification, len(recs))
	for i, rec := range recs {
		items[i] = mapGlobalNotificationToDomain(rec)
	}
	return items, total, nil
}

func (r *GlobalNotificationRepository) ListActive(ctx context.Context, at time.Time) ([]social.GlobalNotification, error) {
	if at.IsZero() {
		at = time.Now().UTC()
	}
	var recs []model.GlobalNotification
	if err := r.db.WithContext(ctx).
		Where("publish_at <= ? AND expire_at >= ?", at, at).
		Order("publish_at DESC, id DESC").
		Find(&recs).Error; err != nil {
		return nil, err
	}
	items := make([]social.GlobalNotification, len(recs))
	for i, rec := range recs {
		items[i] = mapGlobalNotificationToDomain(rec)
	}
	return items, nil
}

func mapGlobalNotificationToDomain(rec model.GlobalNotification) social.GlobalNotification {
	return social.GlobalNotification{
		ID:         rec.ID,
		Content:    rec.Content,
		PublishAt:  rec.PublishAt,
		ExpireAt:   rec.ExpireAt,
		AllowClose: rec.AllowClose,
		CreatedAt:  rec.CreatedAt,
		UpdatedAt:  rec.UpdatedAt,
	}
}

func mapGlobalNotificationToModel(notification *social.GlobalNotification) model.GlobalNotification {
	return model.GlobalNotification{
		ID:         notification.ID,
		Content:    notification.Content,
		PublishAt:  notification.PublishAt,
		ExpireAt:   notification.ExpireAt,
		AllowClose: notification.AllowClose,
	}
}

package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type AdminNotificationRepository struct {
	db *gorm.DB
}

func NewAdminNotificationRepository(db *gorm.DB) *AdminNotificationRepository {
	return &AdminNotificationRepository{db: db}
}

func (r *AdminNotificationRepository) Create(ctx context.Context, notification *social.AdminNotification) error {
	rec := mapAdminNotificationToModel(notification)
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	notification.ID = rec.ID
	notification.CreatedAt = rec.CreatedAt
	notification.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *AdminNotificationRepository) GetByID(ctx context.Context, id int64) (*social.AdminNotification, error) {
	var rec model.AdminNotification
	if err := r.db.WithContext(ctx).First(&rec, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, social.ErrAdminNotificationNotFound
		}
		return nil, err
	}
	item := mapAdminNotificationToDomain(rec)
	return &item, nil
}

func (r *AdminNotificationRepository) ListByUser(ctx context.Context, userID int64, options social.AdminNotificationListOptions) ([]social.AdminNotification, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.AdminNotification{}).Where("user_id = ?", userID)
	if options.UnreadOnly {
		query = query.Where("is_read = ?", false)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page := options.Page
	if page < 1 {
		page = 1
	}
	size := options.PageSize
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	var recs []model.AdminNotification
	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&recs).Error; err != nil {
		return nil, 0, err
	}
	items := make([]social.AdminNotification, len(recs))
	for i := range recs {
		items[i] = mapAdminNotificationToDomain(recs[i])
	}
	return items, total, nil
}

func (r *AdminNotificationRepository) MarkRead(ctx context.Context, userID int64, id int64) error {
	result := r.db.WithContext(ctx).Model(&model.AdminNotification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]any{
			"is_read":    true,
			"read_at":    time.Now().UTC(),
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return social.ErrAdminNotificationNotFound
	}
	return nil
}

func (r *AdminNotificationRepository) MarkAllRead(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Model(&model.AdminNotification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]any{
			"is_read":    true,
			"read_at":    time.Now().UTC(),
			"updated_at": time.Now().UTC(),
		}).Error
}

func mapAdminNotificationToDomain(rec model.AdminNotification) social.AdminNotification {
	return social.AdminNotification{
		ID:        rec.ID,
		UserID:    rec.UserID,
		NotifType: rec.NotifType,
		Title:     rec.Title,
		Content:   rec.Content,
		Payload:   json.RawMessage(rec.Payload),
		IsRead:    rec.IsRead,
		ReadAt:    rec.ReadAt,
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
	}
}

func mapAdminNotificationToModel(item *social.AdminNotification) model.AdminNotification {
	return model.AdminNotification{
		ID:        item.ID,
		UserID:    item.UserID,
		NotifType: item.NotifType,
		Title:     item.Title,
		Content:   item.Content,
		Payload:   datatypes.JSON(item.Payload),
		IsRead:    item.IsRead,
		ReadAt:    item.ReadAt,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

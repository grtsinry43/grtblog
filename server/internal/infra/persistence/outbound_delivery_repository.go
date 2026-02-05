package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type OutboundDeliveryRepository struct {
	db *gorm.DB
}

func NewOutboundDeliveryRepository(db *gorm.DB) *OutboundDeliveryRepository {
	return &OutboundDeliveryRepository{db: db}
}

func (r *OutboundDeliveryRepository) Create(ctx context.Context, delivery *federation.OutboundDelivery) error {
	rec := mapOutboundDeliveryToModel(delivery)
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	delivery.ID = rec.ID
	delivery.CreatedAt = rec.CreatedAt
	delivery.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *OutboundDeliveryRepository) GetByID(ctx context.Context, id int64) (*federation.OutboundDelivery, error) {
	var rec model.OutboundDelivery
	if err := r.db.WithContext(ctx).First(&rec, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, federation.ErrOutboundDeliveryNotFound
		}
		return nil, err
	}
	item := mapOutboundDeliveryToDomain(rec)
	return &item, nil
}

func (r *OutboundDeliveryRepository) GetByRequestID(ctx context.Context, requestID string) (*federation.OutboundDelivery, error) {
	var rec model.OutboundDelivery
	if err := r.db.WithContext(ctx).Where("request_id = ?", requestID).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, federation.ErrOutboundDeliveryNotFound
		}
		return nil, err
	}
	item := mapOutboundDeliveryToDomain(rec)
	return &item, nil
}

func (r *OutboundDeliveryRepository) Update(ctx context.Context, delivery *federation.OutboundDelivery) error {
	rec := mapOutboundDeliveryToModel(delivery)
	return r.db.WithContext(ctx).Model(&model.OutboundDelivery{}).
		Where("id = ?", delivery.ID).
		Updates(&rec).Error
}

func (r *OutboundDeliveryRepository) List(ctx context.Context, options federation.OutboundDeliveryListOptions) ([]federation.OutboundDelivery, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.OutboundDelivery{})
	if options.RequestID != "" {
		query = query.Where("request_id = ?", options.RequestID)
	}
	if options.Type != "" {
		query = query.Where("delivery_type = ?", options.Type)
	}
	if options.Status != "" {
		query = query.Where("status = ?", options.Status)
	}
	if target := strings.TrimSpace(options.Target); target != "" {
		like := "%" + target + "%"
		query = query.Where("target_instance_url ILIKE ?", like)
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
	offset := (page - 1) * size
	var recs []model.OutboundDelivery
	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&recs).Error; err != nil {
		return nil, 0, err
	}
	items := make([]federation.OutboundDelivery, len(recs))
	for i := range recs {
		items[i] = mapOutboundDeliveryToDomain(recs[i])
	}
	return items, total, nil
}

func (r *OutboundDeliveryRepository) ListRetryable(ctx context.Context, now time.Time, limit int) ([]federation.OutboundDelivery, error) {
	if limit <= 0 {
		limit = 100
	}
	var recs []model.OutboundDelivery
	if err := r.db.WithContext(ctx).
		Where("status IN ?", []string{federation.DeliveryStatusQueued, federation.DeliveryStatusFailed, federation.DeliveryStatusTimeout}).
		Where("(next_retry_at IS NULL OR next_retry_at <= ?)", now).
		Order("created_at ASC").
		Limit(limit).
		Find(&recs).Error; err != nil {
		return nil, err
	}
	items := make([]federation.OutboundDelivery, len(recs))
	for i := range recs {
		items[i] = mapOutboundDeliveryToDomain(recs[i])
	}
	return items, nil
}

func (r *OutboundDeliveryRepository) ListBySourceArticle(ctx context.Context, articleID int64, limit int) ([]federation.OutboundDelivery, error) {
	if limit <= 0 {
		limit = 100
	}
	var recs []model.OutboundDelivery
	if err := r.db.WithContext(ctx).
		Where("source_article_id = ?", articleID).
		Order("created_at DESC").
		Limit(limit).
		Find(&recs).Error; err != nil {
		return nil, err
	}
	items := make([]federation.OutboundDelivery, len(recs))
	for i := range recs {
		items[i] = mapOutboundDeliveryToDomain(recs[i])
	}
	return items, nil
}

func mapOutboundDeliveryToDomain(rec model.OutboundDelivery) federation.OutboundDelivery {
	return federation.OutboundDelivery{
		ID:                rec.ID,
		RequestID:         rec.RequestID,
		DeliveryType:      rec.DeliveryType,
		SourceArticleID:   rec.SourceArticleID,
		TargetInstanceURL: rec.TargetInstanceURL,
		TargetEndpoint:    rec.TargetEndpoint,
		Payload:           json.RawMessage(rec.Payload),
		Status:            rec.Status,
		AttemptCount:      rec.AttemptCount,
		MaxAttempts:       rec.MaxAttempts,
		NextRetryAt:       rec.NextRetryAt,
		HTTPStatus:        rec.HTTPStatus,
		ResponseBody:      rec.ResponseBody,
		ErrorMessage:      rec.ErrorMessage,
		RemoteTicketID:    rec.RemoteTicketID,
		TraceID:           rec.TraceID,
		LastCallbackAt:    rec.LastCallbackAt,
		CreatedAt:         rec.CreatedAt,
		UpdatedAt:         rec.UpdatedAt,
	}
}

func mapOutboundDeliveryToModel(item *federation.OutboundDelivery) model.OutboundDelivery {
	return model.OutboundDelivery{
		ID:                item.ID,
		RequestID:         item.RequestID,
		DeliveryType:      item.DeliveryType,
		SourceArticleID:   item.SourceArticleID,
		TargetInstanceURL: item.TargetInstanceURL,
		TargetEndpoint:    item.TargetEndpoint,
		Payload:           datatypes.JSON(item.Payload),
		Status:            item.Status,
		AttemptCount:      item.AttemptCount,
		MaxAttempts:       item.MaxAttempts,
		NextRetryAt:       item.NextRetryAt,
		HTTPStatus:        item.HTTPStatus,
		ResponseBody:      item.ResponseBody,
		ErrorMessage:      item.ErrorMessage,
		RemoteTicketID:    item.RemoteTicketID,
		TraceID:           item.TraceID,
		LastCallbackAt:    item.LastCallbackAt,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

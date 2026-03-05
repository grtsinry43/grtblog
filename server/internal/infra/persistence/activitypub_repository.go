package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	domainap "github.com/grtsinry43/grtblog-v2/server/internal/domain/activitypub"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type ActivityPubFollowerRepository struct {
	db *gorm.DB
}

func NewActivityPubFollowerRepository(db *gorm.DB) *ActivityPubFollowerRepository {
	return &ActivityPubFollowerRepository{db: db}
}

func (r *ActivityPubFollowerRepository) Upsert(ctx context.Context, follower *domainap.Follower) error {
	if follower == nil {
		return nil
	}
	rec := mapFollowerToModel(follower)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.ActivityPubFollower
		err := tx.Where("actor_id = ?", rec.ActorID).Take(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if rec.Status == "" {
					rec.Status = "active"
				}
				if err := tx.Create(&rec).Error; err != nil {
					return err
				}
				follower.ID = rec.ID
				follower.CreatedAt = rec.CreatedAt
				follower.UpdatedAt = rec.UpdatedAt
				return nil
			}
			return err
		}
		updates := map[string]any{
			"inbox_url":          rec.InboxURL,
			"shared_inbox_url":   rec.SharedInboxURL,
			"preferred_username": rec.PreferredUsername,
			"display_name":       rec.DisplayName,
			"status":             firstNonEmpty(rec.Status, existing.Status, "active"),
			"last_seen_at":       rec.LastSeenAt,
			"followed_at":        rec.FollowedAt,
			"updated_at":         rec.UpdatedAt,
		}
		if err := tx.Model(&model.ActivityPubFollower{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
			return err
		}
		follower.ID = existing.ID
		follower.CreatedAt = existing.CreatedAt
		follower.UpdatedAt = rec.UpdatedAt
		return nil
	})
}

func (r *ActivityPubFollowerRepository) GetByActorID(ctx context.Context, actorID string) (*domainap.Follower, error) {
	var rec model.ActivityPubFollower
	err := r.db.WithContext(ctx).Where("actor_id = ?", strings.TrimSpace(actorID)).Take(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainap.ErrFollowerNotFound
		}
		return nil, err
	}
	item := mapFollowerToDomain(rec)
	return &item, nil
}

func (r *ActivityPubFollowerRepository) List(ctx context.Context, status string, page, pageSize int) ([]domainap.Follower, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.ActivityPubFollower{})
	if trimmed := strings.TrimSpace(status); trimmed != "" {
		query = query.Where("status = ?", trimmed)
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var recs []model.ActivityPubFollower
	offset := (page - 1) * pageSize
	if err := query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&recs).Error; err != nil {
		return nil, 0, err
	}
	items := make([]domainap.Follower, len(recs))
	for i, rec := range recs {
		items[i] = mapFollowerToDomain(rec)
	}
	return items, total, nil
}

func (r *ActivityPubFollowerRepository) ListActive(ctx context.Context) ([]domainap.Follower, error) {
	var recs []model.ActivityPubFollower
	if err := r.db.WithContext(ctx).Where("status = ?", "active").Order("updated_at DESC").Find(&recs).Error; err != nil {
		return nil, err
	}
	items := make([]domainap.Follower, len(recs))
	for i, rec := range recs {
		items[i] = mapFollowerToDomain(rec)
	}
	return items, nil
}

type ActivityPubOutboxRepository struct {
	db *gorm.DB
}

func NewActivityPubOutboxRepository(db *gorm.DB) *ActivityPubOutboxRepository {
	return &ActivityPubOutboxRepository{db: db}
}

func (r *ActivityPubOutboxRepository) Create(ctx context.Context, item *domainap.OutboxItem) error {
	rec := mapActivityPubOutboxToModel(item)
	if rec.Summary == "" {
		rec.Summary = ""
	}
	if len(rec.Activity) == 0 {
		rec.Activity = datatypes.JSON([]byte("{}"))
	}
	if strings.TrimSpace(rec.Status) == "" {
		rec.Status = domainap.OutboxStatusQueued
	}
	if strings.TrimSpace(rec.TriggerSource) == "" {
		rec.TriggerSource = "auto"
	}
	if len(rec.Deliveries) == 0 {
		rec.Deliveries = datatypes.JSON([]byte("[]"))
	}
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	item.ID = rec.ID
	item.Status = rec.Status
	item.TriggerSource = rec.TriggerSource
	item.PublishedAt = rec.PublishedAt
	item.CreatedAt = rec.CreatedAt
	item.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *ActivityPubOutboxRepository) List(ctx context.Context, page, pageSize int) ([]domainap.OutboxItem, int64, error) {
	return r.ListWithOptions(ctx, domainap.OutboxListOptions{Page: page, PageSize: pageSize})
}

func (r *ActivityPubOutboxRepository) GetByID(ctx context.Context, id int64) (*domainap.OutboxItem, error) {
	var rec model.ActivityPubOutboxItem
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainap.ErrOutboxItemNotFound
		}
		return nil, err
	}
	item := mapActivityPubOutboxToDomain(rec)
	return &item, nil
}

func (r *ActivityPubOutboxRepository) ListWithOptions(ctx context.Context, opts domainap.OutboxListOptions) ([]domainap.OutboxItem, int64, error) {
	page := opts.Page
	if page <= 0 {
		page = 1
	}
	size := opts.PageSize
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	query := r.db.WithContext(ctx).Model(&model.ActivityPubOutboxItem{})
	if status := strings.TrimSpace(opts.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	if sourceType := strings.TrimSpace(opts.SourceType); sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}
	if keyword := strings.TrimSpace(opts.Search); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("summary ILIKE ? OR activity_id ILIKE ? OR object_id ILIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var recs []model.ActivityPubOutboxItem
	offset := (page - 1) * size
	if err := query.Order("published_at DESC").Offset(offset).Limit(size).Find(&recs).Error; err != nil {
		return nil, 0, err
	}
	items := make([]domainap.OutboxItem, len(recs))
	for i, rec := range recs {
		items[i] = mapActivityPubOutboxToDomain(rec)
	}
	return items, total, nil
}

func (r *ActivityPubOutboxRepository) UpdateDeliveryResult(ctx context.Context, item *domainap.OutboxItem) error {
	if item == nil || item.ID <= 0 {
		return nil
	}
	deliveries, err := json.Marshal(item.Deliveries)
	if err != nil {
		return err
	}
	updates := map[string]any{
		"status":         strings.TrimSpace(item.Status),
		"trigger_source": strings.TrimSpace(item.TriggerSource),
		"total_targets":  item.TotalTargets,
		"success_count":  item.SuccessCount,
		"failure_count":  item.FailureCount,
		"deliveries":     datatypes.JSON(deliveries),
		"started_at":     item.StartedAt,
		"finished_at":    item.FinishedAt,
		"duration_ms":    item.DurationMs,
	}
	return r.db.WithContext(ctx).Model(&model.ActivityPubOutboxItem{}).Where("id = ?", item.ID).Updates(updates).Error
}

func mapFollowerToModel(item *domainap.Follower) model.ActivityPubFollower {
	return model.ActivityPubFollower{
		ID:                item.ID,
		ActorID:           strings.TrimSpace(item.ActorID),
		InboxURL:          strings.TrimSpace(item.InboxURL),
		SharedInboxURL:    trimPtr(item.SharedInboxURL),
		PreferredUsername: trimPtr(item.PreferredUsername),
		DisplayName:       trimPtr(item.DisplayName),
		Status:            firstNonEmpty(strings.TrimSpace(item.Status), "active"),
		FollowedAt:        item.FollowedAt,
		LastSeenAt:        item.LastSeenAt,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

func mapFollowerToDomain(rec model.ActivityPubFollower) domainap.Follower {
	status := strings.TrimSpace(rec.Status)
	if status == "" {
		status = "active"
	}
	return domainap.Follower{
		ID:                rec.ID,
		ActorID:           rec.ActorID,
		InboxURL:          rec.InboxURL,
		SharedInboxURL:    trimPtr(rec.SharedInboxURL),
		PreferredUsername: trimPtr(rec.PreferredUsername),
		DisplayName:       trimPtr(rec.DisplayName),
		Status:            status,
		FollowedAt:        rec.FollowedAt,
		LastSeenAt:        rec.LastSeenAt,
		CreatedAt:         rec.CreatedAt,
		UpdatedAt:         rec.UpdatedAt,
	}
}

func mapActivityPubOutboxToModel(item *domainap.OutboxItem) model.ActivityPubOutboxItem {
	deliveries, _ := json.Marshal(item.Deliveries)
	if len(deliveries) == 0 {
		deliveries = []byte("[]")
	}
	status := strings.TrimSpace(item.Status)
	if status == "" {
		status = domainap.OutboxStatusQueued
	}
	triggerSource := strings.TrimSpace(item.TriggerSource)
	if triggerSource == "" {
		triggerSource = "auto"
	}
	return model.ActivityPubOutboxItem{
		ID:            item.ID,
		ActivityID:    strings.TrimSpace(item.ActivityID),
		ObjectID:      strings.TrimSpace(item.ObjectID),
		SourceType:    strings.TrimSpace(item.SourceType),
		SourceID:      item.SourceID,
		SourceURL:     strings.TrimSpace(item.SourceURL),
		Summary:       strings.TrimSpace(item.Summary),
		Activity:      datatypes.JSON(item.Activity),
		Status:        status,
		TriggerSource: triggerSource,
		TotalTargets:  item.TotalTargets,
		SuccessCount:  item.SuccessCount,
		FailureCount:  item.FailureCount,
		Deliveries:    datatypes.JSON(deliveries),
		StartedAt:     item.StartedAt,
		FinishedAt:    item.FinishedAt,
		DurationMs:    item.DurationMs,
		PublishedAt:   item.PublishedAt,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func mapActivityPubOutboxToDomain(rec model.ActivityPubOutboxItem) domainap.OutboxItem {
	item := domainap.OutboxItem{
		ID:            rec.ID,
		ActivityID:    rec.ActivityID,
		ObjectID:      rec.ObjectID,
		SourceType:    rec.SourceType,
		SourceID:      rec.SourceID,
		SourceURL:     rec.SourceURL,
		Summary:       rec.Summary,
		Activity:      json.RawMessage(rec.Activity),
		Status:        rec.Status,
		TriggerSource: rec.TriggerSource,
		TotalTargets:  rec.TotalTargets,
		SuccessCount:  rec.SuccessCount,
		FailureCount:  rec.FailureCount,
		StartedAt:     rec.StartedAt,
		FinishedAt:    rec.FinishedAt,
		DurationMs:    rec.DurationMs,
		PublishedAt:   rec.PublishedAt,
		CreatedAt:     rec.CreatedAt,
		UpdatedAt:     rec.UpdatedAt,
	}
	if len(rec.Deliveries) > 0 {
		_ = json.Unmarshal(rec.Deliveries, &item.Deliveries)
	}
	if item.Deliveries == nil {
		item.Deliveries = make([]domainap.DeliveryDetail, 0)
	}
	return item
}

func firstNonEmpty(values ...string) string {
	for _, val := range values {
		if strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
	}
	return ""
}

func trimPtr(raw *string) *string {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

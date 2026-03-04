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

type FriendLinkSyncJobRepository struct {
	db *gorm.DB
}

func NewFriendLinkSyncJobRepository(db *gorm.DB) *FriendLinkSyncJobRepository {
	return &FriendLinkSyncJobRepository{db: db}
}

func (r *FriendLinkSyncJobRepository) Create(ctx context.Context, job *social.FriendLinkSyncJob) error {
	rec := mapFriendLinkSyncJobToModel(job)
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	job.ID = rec.ID
	job.CreatedAt = rec.CreatedAt
	job.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *FriendLinkSyncJobRepository) Update(ctx context.Context, job *social.FriendLinkSyncJob) error {
	rec := mapFriendLinkSyncJobToModel(job)
	return r.db.WithContext(ctx).Model(&model.FriendLinkSyncJob{}).
		Where("id = ?", job.ID).
		Updates(map[string]any{
			"target_type":    rec.TargetType,
			"sync_method":    rec.SyncMethod,
			"friend_link_id": rec.FriendLinkID,
			"instance_id":    rec.InstanceID,
			"target_url":     rec.TargetURL,
			"feed_url":       rec.FeedURL,
			"status":         rec.Status,
			"attempt_count":  rec.AttemptCount,
			"max_attempts":   rec.MaxAttempts,
			"next_retry_at":  rec.NextRetryAt,
			"started_at":     rec.StartedAt,
			"finished_at":    rec.FinishedAt,
			"duration_ms":    rec.DurationMS,
			"pulled_count":   rec.PulledCount,
			"error_message":  rec.ErrorMessage,
			"trigger_source": rec.TriggerSource,
			"updated_at":     time.Now().UTC(),
		}).Error
}

func (r *FriendLinkSyncJobRepository) List(ctx context.Context, options social.FriendLinkSyncJobListOptions) ([]social.FriendLinkSyncJob, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.FriendLinkSyncJob{})

	if status := strings.TrimSpace(options.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	if targetType := strings.TrimSpace(options.TargetType); targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	if method := strings.TrimSpace(options.SyncMethod); method != "" {
		query = query.Where("sync_method = ?", method)
	}
	if options.FriendLinkID != nil && *options.FriendLinkID > 0 {
		query = query.Where("friend_link_id = ?", *options.FriendLinkID)
	}
	if options.InstanceID != nil && *options.InstanceID > 0 {
		query = query.Where("instance_id = ?", *options.InstanceID)
	}
	if kw := strings.TrimSpace(options.Keyword); kw != "" {
		like := "%" + kw + "%"
		query = query.Where("target_url ILIKE ? OR COALESCE(feed_url, '') ILIKE ? OR COALESCE(error_message, '') ILIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := options.Page
	if page < 1 {
		page = 1
	}
	pageSize := options.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	var recs []model.FriendLinkSyncJob
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&recs).Error; err != nil {
		return nil, 0, err
	}
	items := make([]social.FriendLinkSyncJob, len(recs))
	for i := range recs {
		items[i] = mapFriendLinkSyncJobToDomain(recs[i])
	}
	return items, total, nil
}

func (r *FriendLinkSyncJobRepository) ListProcessable(ctx context.Context, now time.Time, limit int) ([]social.FriendLinkSyncJob, error) {
	if limit <= 0 {
		limit = 100
	}
	var recs []model.FriendLinkSyncJob
	if err := r.db.WithContext(ctx).
		Where("status = ?", social.FriendLinkSyncJobStatusQueued).
		Where("(next_retry_at IS NULL OR next_retry_at <= ?)", now).
		Order("created_at ASC").
		Limit(limit).
		Find(&recs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	items := make([]social.FriendLinkSyncJob, len(recs))
	for i := range recs {
		items[i] = mapFriendLinkSyncJobToDomain(recs[i])
	}
	return items, nil
}

func mapFriendLinkSyncJobToDomain(rec model.FriendLinkSyncJob) social.FriendLinkSyncJob {
	return social.FriendLinkSyncJob{
		ID:            rec.ID,
		TargetType:    rec.TargetType,
		SyncMethod:    rec.SyncMethod,
		FriendLinkID:  rec.FriendLinkID,
		InstanceID:    rec.InstanceID,
		TargetURL:     rec.TargetURL,
		FeedURL:       rec.FeedURL,
		Status:        rec.Status,
		AttemptCount:  rec.AttemptCount,
		MaxAttempts:   rec.MaxAttempts,
		NextRetryAt:   rec.NextRetryAt,
		StartedAt:     rec.StartedAt,
		FinishedAt:    rec.FinishedAt,
		DurationMS:    rec.DurationMS,
		PulledCount:   rec.PulledCount,
		ErrorMessage:  rec.ErrorMessage,
		TriggerSource: rec.TriggerSource,
		CreatedAt:     rec.CreatedAt,
		UpdatedAt:     rec.UpdatedAt,
	}
}

func mapFriendLinkSyncJobToModel(job *social.FriendLinkSyncJob) model.FriendLinkSyncJob {
	return model.FriendLinkSyncJob{
		ID:            job.ID,
		TargetType:    strings.TrimSpace(job.TargetType),
		SyncMethod:    strings.TrimSpace(job.SyncMethod),
		FriendLinkID:  job.FriendLinkID,
		InstanceID:    job.InstanceID,
		TargetURL:     strings.TrimSpace(job.TargetURL),
		FeedURL:       job.FeedURL,
		Status:        strings.TrimSpace(job.Status),
		AttemptCount:  job.AttemptCount,
		MaxAttempts:   job.MaxAttempts,
		NextRetryAt:   job.NextRetryAt,
		StartedAt:     job.StartedAt,
		FinishedAt:    job.FinishedAt,
		DurationMS:    job.DurationMS,
		PulledCount:   job.PulledCount,
		ErrorMessage:  job.ErrorMessage,
		TriggerSource: strings.TrimSpace(job.TriggerSource),
		CreatedAt:     job.CreatedAt,
		UpdatedAt:     job.UpdatedAt,
	}
}

package persistence

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const cleanupBatchSize = 5000

// CleanupRepository implements cleanup.Repository using GORM.
type CleanupRepository struct {
	db *gorm.DB
}

func NewCleanupRepository(db *gorm.DB) *CleanupRepository {
	return &CleanupRepository{db: db}
}

func (r *CleanupRepository) PurgeContentHourlyStats(ctx context.Context, before time.Time) (int64, error) {
	return r.deleteWhere(ctx, "analytics_content_hourly", "hour_bucket < ?", before)
}

func (r *CleanupRepository) PurgeOnlineHourlyStats(ctx context.Context, before time.Time) (int64, error) {
	return r.deleteWhere(ctx, "analytics_online_hourly", "hour_bucket < ?", before)
}

func (r *CleanupRepository) PurgeRSSAccessHourlyStats(ctx context.Context, before time.Time) (int64, error) {
	return r.deleteWhere(ctx, "analytics_rss_access_hourly", "hour_bucket < ?", before)
}

func (r *CleanupRepository) PurgeStaleVisitorViews(ctx context.Context, lastViewBefore time.Time) (int64, error) {
	return r.deleteWhere(ctx, "analytics_visitor_view", "last_view_at < ?", lastViewBefore)
}

func (r *CleanupRepository) PurgeAITaskLogs(ctx context.Context, before time.Time) (int64, error) {
	return r.deleteWhere(ctx, "ai_task_log", "created_at < ?", before)
}

func (r *CleanupRepository) PurgeEmailOutbox(ctx context.Context, before time.Time) (int64, error) {
	return r.deleteWhere(ctx, "email_outbox", "created_at < ?", before)
}

func (r *CleanupRepository) deleteWhere(ctx context.Context, table, where string, arg time.Time) (int64, error) {
	var total int64
	for {
		result := r.db.WithContext(ctx).Exec(
			fmt.Sprintf("DELETE FROM %s WHERE ctid IN (SELECT ctid FROM %s WHERE %s LIMIT %d)", table, table, where, cleanupBatchSize),
			arg,
		)
		if result.Error != nil {
			return total, result.Error
		}
		total += result.RowsAffected
		if result.RowsAffected < int64(cleanupBatchSize) {
			return total, nil
		}
	}
}

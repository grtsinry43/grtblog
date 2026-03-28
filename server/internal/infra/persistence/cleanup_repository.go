package persistence

import (
	"context"
	"time"

	"gorm.io/gorm"
)

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
	result := r.db.WithContext(ctx).Table(table).Where(where, arg).Delete(nil)
	return result.RowsAffected, result.Error
}

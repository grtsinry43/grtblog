package cleanup

import (
	"context"
	"log"
	"time"
)

// Repository abstracts the persistence operations needed by the cleanup worker.
type Repository interface {
	PurgeContentHourlyStats(ctx context.Context, before time.Time) (int64, error)
	PurgeOnlineHourlyStats(ctx context.Context, before time.Time) (int64, error)
	PurgeRSSAccessHourlyStats(ctx context.Context, before time.Time) (int64, error)
	PurgeStaleVisitorViews(ctx context.Context, lastViewBefore time.Time) (int64, error)
	PurgeAITaskLogs(ctx context.Context, before time.Time) (int64, error)
	PurgeEmailOutbox(ctx context.Context, before time.Time) (int64, error)
}

// Service periodically purges stale analytics, logs and transactional records
// to prevent unbounded table growth.
type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// Run starts a blocking loop that executes cleanup once on startup and then
// every interval until ctx is cancelled.
func (s *Service) Run(ctx context.Context, interval time.Duration) {
	log.Printf("[cleanup] worker started, interval=%s", interval)
	s.runOnce(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("[cleanup] worker stopped")
			return
		case <-ticker.C:
			s.runOnce(ctx)
		}
	}
}

func (s *Service) runOnce(ctx context.Context) {
	now := s.now().UTC()

	type purgeTask struct {
		name string
		fn   func() (int64, error)
	}

	tasks := []purgeTask{
		{"content_hourly_stats", func() (int64, error) {
			return s.repo.PurgeContentHourlyStats(ctx, now.AddDate(0, 0, -90))
		}},
		{"online_hourly_stats", func() (int64, error) {
			return s.repo.PurgeOnlineHourlyStats(ctx, now.AddDate(0, 0, -90))
		}},
		{"rss_access_hourly_stats", func() (int64, error) {
			return s.repo.PurgeRSSAccessHourlyStats(ctx, now.AddDate(0, 0, -90))
		}},
		{"stale_visitor_views", func() (int64, error) {
			return s.repo.PurgeStaleVisitorViews(ctx, now.AddDate(0, 0, -180))
		}},
		{"ai_task_logs", func() (int64, error) {
			return s.repo.PurgeAITaskLogs(ctx, now.AddDate(0, 0, -30))
		}},
		{"email_outbox", func() (int64, error) {
			return s.repo.PurgeEmailOutbox(ctx, now.AddDate(0, 0, -30))
		}},
	}

	for _, t := range tasks {
		rows, err := t.fn()
		if err != nil {
			log.Printf("[cleanup] %s error: %v", t.name, err)
		} else if rows > 0 {
			log.Printf("[cleanup] %s purged %d rows", t.name, rows)
		}
	}
}

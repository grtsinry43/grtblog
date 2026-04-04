package telemetry

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/htmlsnapshot"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/metrics"
	"github.com/grtsinry43/grtblog-v2/server/internal/ws"
)

// Service ties together error collection and runtime metrics into a
// unified telemetry snapshot. It reads from existing subsystems
// (HTTPStats, htmlsnapshot, sysconfig, DB) without duplicating their logic.
type Service struct {
	collector    *Collector
	db           *gorm.DB
	httpStats    *metrics.HTTPStats
	renderer     *htmlsnapshot.Service
	wsManager    atomic.Pointer[ws.Manager]
	sysCfg       *sysconfig.Service
	startedAt    time.Time
	reporter     *Reporter
	reporterOnce sync.Once
}

// NewService creates a telemetry Service. All dependencies are optional;
// missing ones simply omit the corresponding section from the snapshot.
func NewService(
	collector *Collector,
	db *gorm.DB,
	httpStats *metrics.HTTPStats,
	renderer *htmlsnapshot.Service,
	wsManager *ws.Manager,
	sysCfg *sysconfig.Service,
) *Service {
	svc := &Service{
		collector: collector,
		db:        db,
		httpStats: httpStats,
		renderer:  renderer,
		sysCfg:    sysCfg,
		startedAt: time.Now(),
	}
	if wsManager != nil {
		svc.wsManager.Store(wsManager)
	}
	return svc
}

// Collector returns the underlying error collector.
func (s *Service) Collector() *Collector {
	if s == nil {
		return nil
	}
	return s.collector
}

// Reporter returns the background reporter (created lazily on first call).
func (s *Service) Reporter() *Reporter {
	if s == nil {
		return nil
	}
	s.reporterOnce.Do(func() {
		s.reporter = NewReporter(s)
	})
	return s.reporter
}

// SetWSManager allows late injection of the WebSocket manager, which may be
// created after the telemetry service (e.g. inside router.Register).
// Safe for concurrent use via atomic.Pointer.
func (s *Service) SetWSManager(m *ws.Manager) {
	if s != nil && m != nil {
		s.wsManager.Store(m)
	}
}

// FullSnapshot builds a comprehensive telemetry snapshot containing
// environment info, feature flags, runtime metrics, and error digests.
func (s *Service) FullSnapshot(ctx context.Context) *FullTelemetrySnapshot {
	// Guard against nil collector — degrade gracefully.
	if s.collector == nil {
		now := time.Now().UTC()
		return &FullTelemetrySnapshot{
			GeneratedAt: now,
			Instance:    s.buildInstanceInfo(ctx),
		}
	}

	now := s.collector.now().UTC()

	snap := &FullTelemetrySnapshot{
		GeneratedAt: now,
		Instance:    s.buildInstanceInfo(ctx),
		Metrics:     s.buildMetrics(ctx, now),
	}

	// Error digests from P0 collector.
	all := s.collector.Snapshot()
	for _, d := range all {
		if d.Kind == KindPanic {
			snap.Panics = append(snap.Panics, d)
		} else {
			snap.Errors = append(snap.Errors, d)
		}
	}
	for _, d := range snap.Errors {
		snap.Summary.TotalErrors += d.Count
	}
	for _, d := range snap.Panics {
		snap.Summary.TotalPanics += d.Count
	}
	snap.Summary.UniqueErrors = len(snap.Errors)
	snap.Summary.UniquePanics = len(snap.Panics)

	return snap
}

// buildInstanceInfo gathers anonymous environment metadata + feature flags.
func (s *Service) buildInstanceInfo(ctx context.Context) InstanceInfo {
	info := buildInstanceInfo() // reuse base from snapshot.go
	info.UptimeSeconds = int64(time.Since(s.startedAt).Seconds())
	info.DeployMode = detectDeployMode()
	info.Features = s.detectFeatures(ctx)
	return info
}

// buildMetrics aggregates runtime metrics from all subsystems.
func (s *Service) buildMetrics(ctx context.Context, now time.Time) RuntimeMetrics {
	m := RuntimeMetrics{}

	// Traffic: 24h window from HTTPStats.
	if s.httpStats != nil {
		window := 24 * time.Hour
		snap := s.httpStats.Snapshot(window)
		m.Traffic = TrafficMetrics{
			Window:       window.String(),
			RequestTotal: snap.Requests,
			ErrorRate5xx: snap.ErrorRate,
			P95LatencyMS: snap.P95LatencyMS,
		}
	}

	// ISR / rendering metrics.
	if s.renderer != nil {
		rs := s.renderer.MetricsSnapshot()
		m.ISR = ISRMetrics{
			RenderTotal:   rs.TotalJobs,
			RenderSuccess: rs.SuccessJobs,
			RenderFailed:  rs.FailedJobs,
			AvgRenderMS:   rs.AverageDurationMS,
			P95RenderMS:   rs.P95DurationMS,
		}
	}

	// Realtime / WebSocket (atomic load, safe for concurrent access).
	if mgr := s.wsManager.Load(); mgr != nil {
		wsSnap := mgr.Snapshot()
		m.Realtime = RealtimeMetrics{
			WSConnectionsCurrent: wsSnap.CurrentOnline,
			WSRooms:              wsSnap.Rooms,
			BroadcastTotal:       wsSnap.BroadcastTotal,
		}
	}

	// Content + federation counts from DB.
	if s.db != nil {
		m.Content = s.queryContentCounts(ctx)
		m.Federation = s.queryFederationCounts(ctx, now)
	}

	return m
}

// detectFeatures reads sysconfig to determine which features are enabled.
// Only boolean on/off flags; no secrets or PII.
func (s *Service) detectFeatures(ctx context.Context) FeatureFlags {
	flags := FeatureFlags{}
	if s.sysCfg == nil {
		return flags
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	fed, err := s.sysCfg.FederationSettings(ctx)
	if err == nil {
		flags.FederationEnabled = fed.Enabled
	}

	ap, err := s.sysCfg.ActivityPubSettings(ctx)
	if err == nil {
		flags.ActivityPubEnabled = ap.Enabled
	}

	cs := s.sysCfg.CommentSettings(ctx)
	flags.CommentsDisabled = cs.Disabled

	es, err := s.sysCfg.EmailSettings(ctx)
	if err == nil {
		flags.EmailEnabled = es.Enabled
	}

	ts, err := s.sysCfg.Turnstile(ctx)
	if err == nil {
		flags.TurnstileEnabled = ts.Enabled
	}

	return flags
}

// queryContentCounts runs lightweight COUNT queries against the DB.
func (s *Service) queryContentCounts(ctx context.Context) ContentMetrics {
	var m ContentMetrics
	if err := s.db.WithContext(ctx).Table("article").Where("deleted_at IS NULL").Count(&m.ArticlesTotal).Error; err != nil {
		log.Printf("[telemetry] article count query failed: %v", err)
	}
	if err := s.db.WithContext(ctx).Table("moment").Where("deleted_at IS NULL").Count(&m.MomentsTotal).Error; err != nil {
		log.Printf("[telemetry] moment count query failed: %v", err)
	}
	if err := s.db.WithContext(ctx).Table("comment").Where("deleted_at IS NULL").Count(&m.CommentsTotal).Error; err != nil {
		log.Printf("[telemetry] comment count query failed: %v", err)
	}
	if err := s.db.WithContext(ctx).Table("friend_link").Where("deleted_at IS NULL AND is_active = ?", true).Count(&m.FriendLinksTotal).Error; err != nil {
		log.Printf("[telemetry] friend_link count query failed: %v", err)
	}
	return m
}

// queryFederationCounts counts outbound deliveries in the last 24h.
func (s *Service) queryFederationCounts(ctx context.Context, now time.Time) FederationMetrics {
	var m FederationMetrics
	since := now.Add(-24 * time.Hour)

	// Single query for total + failure counts.
	type deliveryCounts struct {
		Total    int64
		Failures int64
	}
	var counts deliveryCounts
	if err := s.db.WithContext(ctx).Table("federation_outbound_delivery").
		Select("COUNT(*) AS total, COUNT(*) FILTER (WHERE status NOT IN ('accepted', 'approved')) AS failures").
		Where("created_at >= ?", since).
		Scan(&counts).Error; err != nil {
		log.Printf("[telemetry] federation delivery count query failed: %v", err)
	} else {
		m.OutboundTotal = counts.Total
		m.OutboundFailures = counts.Failures
	}

	if err := s.db.WithContext(ctx).Table("federation_instance").
		Where("status = ?", "active").
		Count(&m.ActiveInstances).Error; err != nil {
		log.Printf("[telemetry] federation instance count query failed: %v", err)
	}

	return m
}

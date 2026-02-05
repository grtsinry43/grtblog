package observability

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/htmlsnapshot"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/metrics"
	"github.com/grtsinry43/grtblog-v2/server/internal/ws"
)

type Service struct {
	db          *gorm.DB
	redis       *redis.Client
	redisPrefix string
	events      appEvent.Bus
	httpStats   *metrics.HTTPStats
	wsManager   *ws.Manager
	renderer    *htmlsnapshot.Service
	startedAt   time.Time

	fedCounter *federationCounter
}

type Overview struct {
	GeneratedAt time.Time         `json:"generatedAt"`
	UptimeSec   int64             `json:"uptimeSec"`
	API         APISummary        `json:"api"`
	Realtime    RealtimeSummary   `json:"realtime"`
	Federation  FederationSummary `json:"federation"`
	Render      RenderSummary     `json:"render"`
}

type APISummary struct {
	Window       string  `json:"window"`
	Requests     int64   `json:"requests"`
	ErrorRate    float64 `json:"errorRate"`
	P95LatencyMS float64 `json:"p95LatencyMs"`
}

type RealtimeSummary struct {
	CurrentOnline int64   `json:"currentOnline"`
	WSRooms       int     `json:"wsRooms"`
	FanoutP95MS   float64 `json:"fanoutP95Ms"`
}

type FederationSummary struct {
	Window              string  `json:"window"`
	DeliveryTotal       int64   `json:"deliveryTotal"`
	DeliverySuccessRate float64 `json:"deliverySuccessRate"`
	VerifyFailedTotal   int64   `json:"verifyFailedTotal"`
	RateLimitedTotal    int64   `json:"rateLimitedTotal"`
}

type RenderSummary struct {
	SuccessJobs       int64   `json:"successJobs"`
	FailedJobs        int64   `json:"failedJobs"`
	LastDurationMS    int64   `json:"lastDurationMs"`
	P95DurationMS     float64 `json:"p95DurationMs"`
	LastRenderedFiles int64   `json:"lastRenderedFiles"`
}

type ControlPlane struct {
	GeneratedAt time.Time                  `json:"generatedAt"`
	API         metrics.HTTPWindowSnapshot `json:"api"`
	Database    DatabasePool               `json:"database"`
	GoRuntime   GoRuntime                  `json:"goRuntime"`
}

type DatabasePool struct {
	Status             string `json:"status"`
	MaxOpenConnections int    `json:"maxOpenConnections"`
	OpenConnections    int    `json:"openConnections"`
	InUse              int    `json:"inUse"`
	Idle               int    `json:"idle"`
	WaitCount          int64  `json:"waitCount"`
}

type GoRuntime struct {
	NumGoroutine int    `json:"numGoroutine"`
	GoVersion    string `json:"goVersion"`
}

type RenderPlane struct {
	GeneratedAt time.Time                    `json:"generatedAt"`
	Snapshot    htmlsnapshot.MetricsSnapshot `json:"snapshot"`
}

type RealtimePlane struct {
	GeneratedAt time.Time   `json:"generatedAt"`
	Snapshot    ws.Snapshot `json:"snapshot"`
}

type FederationPlane struct {
	GeneratedAt        time.Time        `json:"generatedAt"`
	Window             string           `json:"window"`
	OutboundByStatus   map[string]int64 `json:"outboundByStatus"`
	OutboundTotal      int64            `json:"outboundTotal"`
	SuccessRate        float64          `json:"successRate"`
	RetryReadyCount    int64            `json:"retryReadyCount"`
	DeadLetterCount    int64            `json:"deadLetterCount"`
	PendingCitations   int64            `json:"pendingCitations"`
	PendingMentions    int64            `json:"pendingMentions"`
	InstancesActive    int64            `json:"instancesActive"`
	InstancesBlocked   int64            `json:"instancesBlocked"`
	VerifyFailedTotal  int64            `json:"verifyFailedTotal"`
	RateLimitedTotal   int64            `json:"rateLimitedTotal"`
	InboundEventTotals map[string]int64 `json:"inboundEventTotals"`
}

type StoragePlane struct {
	GeneratedAt time.Time     `json:"generatedAt"`
	StorageHTML DirectoryStat `json:"storageHtml"`
	StorageLogs DirectoryStat `json:"storageLogs"`
	Redis       RedisStat     `json:"redis"`
}

type DirectoryStat struct {
	Path  string `json:"path"`
	Size  uint64 `json:"size"`
	Files int64  `json:"files"`
}

type RedisStat struct {
	Status              string `json:"status"`
	UsedMemory          string `json:"usedMemory,omitempty"`
	ConnectedClients    int64  `json:"connectedClients,omitempty"`
	AnalyticsQueueDepth int64  `json:"analyticsQueueDepth,omitempty"`
}

type Timeline struct {
	GeneratedAt time.Time      `json:"generatedAt"`
	GroupBy     string         `json:"groupBy"`
	Since       time.Time      `json:"since"`
	Until       time.Time      `json:"until"`
	Series      []SeriesBucket `json:"series"`
}

type SeriesBucket struct {
	Metric    string            `json:"metric"`
	Timestamp time.Time         `json:"timestamp"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags,omitempty"`
}

type Alerts struct {
	GeneratedAt time.Time   `json:"generatedAt"`
	Items       []AlertItem `json:"items"`
}

type AlertItem struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsRead    bool      `json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewService(db *gorm.DB, redisClient *redis.Client, redisPrefix string, events appEvent.Bus, httpStats *metrics.HTTPStats, wsManager *ws.Manager, renderer *htmlsnapshot.Service) *Service {
	svc := &Service{
		db:          db,
		redis:       redisClient,
		redisPrefix: redisPrefix,
		events:      events,
		httpStats:   httpStats,
		wsManager:   wsManager,
		renderer:    renderer,
		startedAt:   time.Now(),
	}
	if events != nil {
		svc.fedCounter = newFederationCounter(7 * 24 * time.Hour)
		svc.fedCounter.register(events)
	}
	return svc
}

func (s *Service) GetOverview(ctx context.Context) (*Overview, error) {
	window := 5 * time.Minute
	apiSnapshot := snapshotHTTP(s.httpStats, window)
	realtime := snapshotWS(s.wsManager)
	render := snapshotRender(s.renderer)
	fed, err := s.GetFederation(ctx, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &Overview{
		GeneratedAt: time.Now().UTC(),
		UptimeSec:   int64(time.Since(s.startedAt).Seconds()),
		API: APISummary{
			Window:       window.String(),
			Requests:     apiSnapshot.Requests,
			ErrorRate:    apiSnapshot.ErrorRate,
			P95LatencyMS: apiSnapshot.P95LatencyMS,
		},
		Realtime: RealtimeSummary{
			CurrentOnline: realtime.CurrentOnline,
			WSRooms:       realtime.Rooms,
			FanoutP95MS:   realtime.BroadcastP95MS,
		},
		Federation: FederationSummary{
			Window:              fed.Window,
			DeliveryTotal:       fed.OutboundTotal,
			DeliverySuccessRate: fed.SuccessRate,
			VerifyFailedTotal:   fed.VerifyFailedTotal,
			RateLimitedTotal:    fed.RateLimitedTotal,
		},
		Render: RenderSummary{
			SuccessJobs:       render.SuccessJobs,
			FailedJobs:        render.FailedJobs,
			LastDurationMS:    render.LastDurationMS,
			P95DurationMS:     render.P95DurationMS,
			LastRenderedFiles: render.LastRenderedFiles,
		},
	}, nil
}

func (s *Service) GetControlPlane(_ context.Context, window time.Duration) (*ControlPlane, error) {
	apiSnapshot := snapshotHTTP(s.httpStats, window)
	dbPool := DatabasePool{Status: "unknown"}
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err == nil {
			dbPool.Status = "connected"
			stats := sqlDB.Stats()
			dbPool.MaxOpenConnections = stats.MaxOpenConnections
			dbPool.OpenConnections = stats.OpenConnections
			dbPool.InUse = stats.InUse
			dbPool.Idle = stats.Idle
			dbPool.WaitCount = stats.WaitCount
		}
	}
	return &ControlPlane{
		GeneratedAt: time.Now().UTC(),
		API:         apiSnapshot,
		Database:    dbPool,
		GoRuntime: GoRuntime{
			NumGoroutine: runtime.NumGoroutine(),
			GoVersion:    runtime.Version(),
		},
	}, nil
}

func (s *Service) GetRenderPlane(_ context.Context) (*RenderPlane, error) {
	return &RenderPlane{
		GeneratedAt: time.Now().UTC(),
		Snapshot:    snapshotRender(s.renderer),
	}, nil
}

func (s *Service) GetRealtime(_ context.Context) (*RealtimePlane, error) {
	return &RealtimePlane{
		GeneratedAt: time.Now().UTC(),
		Snapshot:    snapshotWS(s.wsManager),
	}, nil
}

func (s *Service) GetFederation(ctx context.Context, window time.Duration) (*FederationPlane, error) {
	if window <= 0 {
		window = 24 * time.Hour
	}
	result := &FederationPlane{
		GeneratedAt:        time.Now().UTC(),
		Window:             window.String(),
		OutboundByStatus:   map[string]int64{},
		InboundEventTotals: map[string]int64{},
	}
	if s.db == nil {
		return result, nil
	}

	type statusCount struct {
		Status string
		Count  int64
	}
	var counts []statusCount
	since := time.Now().UTC().Add(-window)
	if err := s.db.WithContext(ctx).
		Table("federation_outbound_delivery").
		Select("status, COUNT(*) AS count").
		Where("created_at >= ?", since).
		Group("status").
		Scan(&counts).Error; err != nil {
		return nil, err
	}
	var total int64
	var success int64
	for _, item := range counts {
		result.OutboundByStatus[item.Status] = item.Count
		total += item.Count
		if item.Status == "accepted" || item.Status == "approved" {
			success += item.Count
		}
	}
	result.OutboundTotal = total
	if total > 0 {
		result.SuccessRate = float64(success) / float64(total)
	}

	if err := s.db.WithContext(ctx).Table("federation_outbound_delivery").
		Where("status IN ?", []string{"queued", "failed", "timeout"}).
		Where("(next_retry_at IS NULL OR next_retry_at <= ?)", time.Now().UTC()).
		Count(&result.RetryReadyCount).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Table("federation_outbound_delivery").
		Where("status = ?", "dead").
		Count(&result.DeadLetterCount).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Table("federated_citation").
		Where("status = ?", "pending").
		Count(&result.PendingCitations).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Table("federated_mention").
		Where("status = ?", "pending").
		Count(&result.PendingMentions).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Table("federation_instance").
		Where("status = ?", "active").
		Count(&result.InstancesActive).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Table("federation_instance").
		Where("status = ?", "blocked").
		Count(&result.InstancesBlocked).Error; err != nil {
		return nil, err
	}
	if s.fedCounter != nil {
		result.VerifyFailedTotal = s.fedCounter.sumSince("federation.signature.verify_failed", since)
		result.RateLimitedTotal = s.fedCounter.sumSince("federation.inbound.rate_limited", since)
		result.InboundEventTotals["friendlink"] = s.fedCounter.sumSince("federation.friendlink.received", since)
		result.InboundEventTotals["citation"] = s.fedCounter.sumSince("federation.citation.received", since)
		result.InboundEventTotals["mention"] = s.fedCounter.sumSince("federation.mention.received", since)
	}
	return result, nil
}

func (s *Service) GetStorage(ctx context.Context) (*StoragePlane, error) {
	result := &StoragePlane{
		GeneratedAt: time.Now().UTC(),
		StorageHTML: directoryStat("storage/html"),
		StorageLogs: directoryStat("storage/logs"),
		Redis:       RedisStat{Status: "not_configured"},
	}
	if s.redis != nil {
		pingCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
		defer cancel()
		if err := s.redis.Ping(pingCtx).Err(); err != nil {
			result.Redis.Status = "error"
		} else {
			result.Redis.Status = "connected"
			info, err := s.redis.Info(ctx, "memory", "clients").Result()
			if err == nil {
				for _, line := range strings.Split(info, "\r\n") {
					switch {
					case strings.HasPrefix(line, "used_memory_human:"):
						result.Redis.UsedMemory = strings.TrimPrefix(line, "used_memory_human:")
					case strings.HasPrefix(line, "connected_clients:"):
						val := strings.TrimPrefix(line, "connected_clients:")
						fmt.Sscan(val, &result.Redis.ConnectedClients)
					}
				}
			}
			queueKey := s.redisPrefix + "analytics:view:queue"
			depth, err := s.redis.LLen(ctx, queueKey).Result()
			if err == nil {
				result.Redis.AnalyticsQueueDepth = depth
			}
		}
	}
	return result, nil
}

func (s *Service) GetTimeline(ctx context.Context, since, until time.Time, groupBy string) (*Timeline, error) {
	if since.IsZero() || until.IsZero() {
		until = time.Now().UTC()
		since = until.Add(-24 * time.Hour)
	}
	if !since.Before(until) {
		return nil, errors.New("invalid timeline range")
	}
	if groupBy == "" {
		groupBy = "hour"
	}
	sqlBucket := "hour"
	switch groupBy {
	case "minute":
		sqlBucket = "minute"
	case "day":
		sqlBucket = "day"
	default:
		groupBy = "hour"
		sqlBucket = "hour"
	}

	out := &Timeline{
		GeneratedAt: time.Now().UTC(),
		GroupBy:     groupBy,
		Since:       since.UTC(),
		Until:       until.UTC(),
		Series:      make([]SeriesBucket, 0, 256),
	}
	if s.db == nil {
		return out, nil
	}

	type aggRow struct {
		Timestamp time.Time
		Value     float64
	}
	var pvRows []aggRow
	if err := s.db.WithContext(ctx).
		Raw("SELECT date_trunc(?, hour_bucket) AS timestamp, SUM(pv)::double precision AS value FROM analytics_content_hourly WHERE hour_bucket BETWEEN ? AND ? GROUP BY 1 ORDER BY 1 ASC", sqlBucket, since, until).
		Scan(&pvRows).Error; err != nil {
		return nil, err
	}
	for _, row := range pvRows {
		out.Series = append(out.Series, SeriesBucket{Metric: "pv", Timestamp: row.Timestamp.UTC(), Value: row.Value})
	}

	var onlineRows []aggRow
	if err := s.db.WithContext(ctx).
		Raw("SELECT date_trunc(?, hour_bucket) AS timestamp, AVG(peak_online)::double precision AS value FROM analytics_online_hourly WHERE hour_bucket BETWEEN ? AND ? GROUP BY 1 ORDER BY 1 ASC", sqlBucket, since, until).
		Scan(&onlineRows).Error; err != nil {
		return nil, err
	}
	for _, row := range onlineRows {
		out.Series = append(out.Series, SeriesBucket{Metric: "online_peak_avg", Timestamp: row.Timestamp.UTC(), Value: row.Value})
	}

	var outboundRows []aggRow
	if err := s.db.WithContext(ctx).
		Raw("SELECT date_trunc(?, created_at) AS timestamp, COUNT(*)::double precision AS value FROM federation_outbound_delivery WHERE created_at BETWEEN ? AND ? GROUP BY 1 ORDER BY 1 ASC", sqlBucket, since, until).
		Scan(&outboundRows).Error; err != nil {
		return nil, err
	}
	for _, row := range outboundRows {
		out.Series = append(out.Series, SeriesBucket{Metric: "federation_outbound_total", Timestamp: row.Timestamp.UTC(), Value: row.Value})
	}
	sort.Slice(out.Series, func(i, j int) bool {
		if out.Series[i].Timestamp.Equal(out.Series[j].Timestamp) {
			return out.Series[i].Metric < out.Series[j].Metric
		}
		return out.Series[i].Timestamp.Before(out.Series[j].Timestamp)
	})
	return out, nil
}

func (s *Service) GetAlerts(ctx context.Context, limit int) (*Alerts, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	result := &Alerts{
		GeneratedAt: time.Now().UTC(),
		Items:       make([]AlertItem, 0, limit),
	}
	if s.db == nil {
		return result, nil
	}
	if err := s.db.WithContext(ctx).
		Table("admin_notification").
		Select("id, notif_type AS type, title, content, is_read, created_at").
		Where("notif_type LIKE ? OR notif_type LIKE ? OR notif_type = ?", "system.%", "federation.%", "system.monitor.alert").
		Order("created_at DESC").
		Limit(limit).
		Scan(&result.Items).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func snapshotHTTP(stats *metrics.HTTPStats, window time.Duration) metrics.HTTPWindowSnapshot {
	if stats == nil {
		return metrics.HTTPWindowSnapshot{Window: window}
	}
	return stats.Snapshot(window)
}

func snapshotWS(manager *ws.Manager) ws.Snapshot {
	if manager == nil {
		return ws.Snapshot{ByRoomType: map[string]int64{}}
	}
	return manager.Snapshot()
}

func snapshotRender(renderer *htmlsnapshot.Service) htmlsnapshot.MetricsSnapshot {
	if renderer == nil {
		return htmlsnapshot.MetricsSnapshot{}
	}
	return renderer.MetricsSnapshot()
}

func directoryStat(path string) DirectoryStat {
	stat := DirectoryStat{Path: path}
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		stat.Files++
		stat.Size += uint64(info.Size())
		return nil
	})
	return stat
}

type eventHandlerFunc func(ctx context.Context, event appEvent.Event) error

func (f eventHandlerFunc) Handle(ctx context.Context, event appEvent.Event) error {
	return f(ctx, event)
}

type federationCounter struct {
	mu        sync.Mutex
	retention time.Duration
	counts    map[string]map[int64]int64
}

func newFederationCounter(retention time.Duration) *federationCounter {
	if retention <= 0 {
		retention = 24 * time.Hour
	}
	return &federationCounter{
		retention: retention,
		counts:    make(map[string]map[int64]int64),
	}
}

func (c *federationCounter) register(bus appEvent.Bus) {
	events := []string{
		"federation.friendlink.received",
		"federation.citation.received",
		"federation.mention.received",
		"federation.signature.verify_failed",
		"federation.inbound.rate_limited",
	}
	for _, name := range events {
		eventName := name
		bus.Subscribe(eventName, eventHandlerFunc(func(_ context.Context, _ appEvent.Event) error {
			c.inc(eventName, time.Now().UTC())
			return nil
		}))
	}
}

func (c *federationCounter) inc(name string, at time.Time) {
	minute := at.UTC().Truncate(time.Minute).Unix()
	c.mu.Lock()
	defer c.mu.Unlock()
	rows := c.counts[name]
	if rows == nil {
		rows = make(map[int64]int64)
		c.counts[name] = rows
	}
	rows[minute]++
	expireBefore := at.UTC().Add(-c.retention).Truncate(time.Minute).Unix()
	for key := range rows {
		if key < expireBefore {
			delete(rows, key)
		}
	}
}

func (c *federationCounter) sumSince(name string, since time.Time) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	rows := c.counts[name]
	if rows == nil {
		return 0
	}
	minute := since.UTC().Truncate(time.Minute).Unix()
	var total int64
	for key, value := range rows {
		if key >= minute {
			total += value
		}
	}
	return total
}

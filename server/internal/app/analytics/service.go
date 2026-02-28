package analytics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/clientinfo"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/geoip"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

const (
	ContentTypeArticle = "article"
	ContentTypeMoment  = "moment"
	ContentTypePage    = "page"
)

const (
	defaultQueueKey       = "analytics:view:queue"
	defaultQueueMaxLength = 200000
	defaultUVTTL          = 72 * time.Hour
	defaultViewDedupTTL   = 30 * time.Minute
)

type Service struct {
	db    *gorm.DB
	redis *redis.Client

	queueKey       string
	queueMaxLength int64
	redisPrefix    string

	uaParser *clientinfo.UAParser
	geo      *geoip.Resolver
	now      func() time.Time
}

type ViewTrackInput struct {
	ContentType string
	ContentID   int64
	VisitorID   string
	IP          string
	UserAgent   string
	At          time.Time
}

type ViewTrackResult struct {
	VisitorID string `json:"visitorId"`
	Queued    bool   `json:"queued"`
}

type ViewTrackEvent struct {
	ContentType string    `json:"contentType"`
	ContentID   int64     `json:"contentId"`
	VisitorID   string    `json:"visitorId"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"userAgent"`
	At          time.Time `json:"at"`
}

func NewService(cfg config.Config, db *gorm.DB, redisClient *redis.Client) *Service {
	geoip.EnsureDatabasesAsync(
		context.Background(),
		cfg.GeoIP.DBPath,
		cfg.GeoIP.DownloadURL,
		cfg.GeoIP.ASNPath,
		cfg.GeoIP.ASNURL,
		nil,
	)

	geoResolver := geoip.NewLazyResolver(cfg.GeoIP.DBPath, cfg.GeoIP.ASNPath)

	return &Service{
		db:             db,
		redis:          redisClient,
		queueKey:       cfg.Redis.Prefix + defaultQueueKey,
		queueMaxLength: defaultQueueMaxLength,
		redisPrefix:    cfg.Redis.Prefix,
		uaParser:       clientinfo.NewUAParser(),
		geo:            geoResolver,
		now:            time.Now,
	}
}

func (s *Service) TrackView(ctx context.Context, in ViewTrackInput) (*ViewTrackResult, error) {
	event, err := s.normalizeViewTrackInput(in)
	if err != nil {
		return nil, err
	}
	if err := s.ensureContentExists(ctx, event.ContentType, event.ContentID); err != nil {
		return nil, err
	}
	if duplicated := s.isDuplicateView(ctx, event); duplicated {
		return &ViewTrackResult{VisitorID: event.VisitorID, Queued: false}, nil
	}

	if s.redis == nil {
		if err := s.processViewEvent(ctx, event); err != nil {
			return nil, err
		}
		return &ViewTrackResult{VisitorID: event.VisitorID, Queued: false}, nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	pipe := s.redis.TxPipeline()
	pipe.LPush(ctx, s.queueKey, payload)
	pipe.LTrim(ctx, s.queueKey, 0, s.queueMaxLength-1)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	return &ViewTrackResult{VisitorID: event.VisitorID, Queued: true}, nil
}

func (s *Service) isDuplicateView(ctx context.Context, event ViewTrackEvent) bool {
	if s.redis == nil {
		return false
	}
	key := fmt.Sprintf(
		"%sanalytics:view:dedupe:%s:%d:%s",
		s.redisPrefix,
		event.ContentType,
		event.ContentID,
		event.VisitorID,
	)
	added, err := s.redis.SetNX(ctx, key, "1", defaultViewDedupTTL).Result()
	if err != nil {
		return false
	}
	return !added
}

func (s *Service) RunViewEventWorker(ctx context.Context) {
	if s.redis == nil {
		return
	}
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		res, err := s.redis.BRPop(ctx, time.Second, s.queueKey).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			continue
		}
		if len(res) != 2 {
			continue
		}
		var event ViewTrackEvent
		if err := json.Unmarshal([]byte(res[1]), &event); err != nil {
			continue
		}
		_ = s.processViewEvent(context.Background(), event)
	}
}

func (s *Service) TrackOnlineSample(ctx context.Context, current int64) error {
	if current < 0 {
		current = 0
	}
	hourBucket := s.now().UTC().Truncate(time.Hour)

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "hour_bucket"}},
		DoUpdates: clause.Assignments(map[string]any{
			"peak_online":  gorm.Expr("GREATEST(analytics_online_hourly.peak_online, EXCLUDED.peak_online)"),
			"sample_total": gorm.Expr("analytics_online_hourly.sample_total + EXCLUDED.sample_total"),
			"sample_count": gorm.Expr("analytics_online_hourly.sample_count + EXCLUDED.sample_count"),
			"updated_at":   gorm.Expr("NOW()"),
		}),
	}).Create(&model.AnalyticsOnlineHourly{
		HourBucket:  hourBucket,
		PeakOnline:  current,
		SampleTotal: current,
		SampleCount: 1,
	}).Error
}

func (s *Service) normalizeViewTrackInput(in ViewTrackInput) (ViewTrackEvent, error) {
	ct := strings.ToLower(strings.TrimSpace(in.ContentType))
	switch ct {
	case ContentTypeArticle, ContentTypeMoment, ContentTypePage:
	default:
		return ViewTrackEvent{}, fmt.Errorf("invalid content type")
	}
	if in.ContentID <= 0 {
		return ViewTrackEvent{}, fmt.Errorf("invalid content id")
	}
	visitorID := strings.TrimSpace(in.VisitorID)
	if visitorID == "" {
		visitorID = fallbackVisitorID(in.IP, in.UserAgent)
	}
	at := in.At
	if at.IsZero() {
		at = s.now().UTC()
	} else {
		at = at.UTC()
	}
	return ViewTrackEvent{
		ContentType: ct,
		ContentID:   in.ContentID,
		VisitorID:   visitorID,
		IP:          strings.TrimSpace(in.IP),
		UserAgent:   strings.TrimSpace(in.UserAgent),
		At:          at,
	}, nil
}

func (s *Service) ensureContentExists(ctx context.Context, contentType string, contentID int64) error {
	var count int64
	q := s.db.WithContext(ctx)
	switch contentType {
	case ContentTypeArticle:
		if err := q.Model(&model.Article{}).Where("id = ? AND is_published = ?", contentID, true).Count(&count).Error; err != nil {
			return err
		}
	case ContentTypeMoment:
		if err := q.Model(&model.Moment{}).Where("id = ? AND is_published = ?", contentID, true).Count(&count).Error; err != nil {
			return err
		}
	case ContentTypePage:
		if err := q.Model(&model.Page{}).Where("id = ? AND is_enabled = ?", contentID, true).Count(&count).Error; err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid content type")
	}
	if count == 0 {
		return fmt.Errorf("content not found")
	}
	return nil
}

func (s *Service) processViewEvent(ctx context.Context, event ViewTrackEvent) error {
	hourBucket := event.At.UTC().Truncate(time.Hour)
	uvBucket := event.At.UTC().Format("20060102")
	uvKey := fmt.Sprintf("%sanalytics:uv:%s:%d:%s", s.redisPrefix, event.ContentType, event.ContentID, uvBucket)

	isNewVisitor := int64(1)
	if s.redis != nil {
		added, err := s.redis.SAdd(ctx, uvKey, event.VisitorID).Result()
		if err == nil {
			isNewVisitor = added
			_ = s.redis.Expire(ctx, uvKey, defaultUVTTL).Err()
		}
	}

	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "content_type"}, {Name: "content_id"}, {Name: "hour_bucket"}},
		DoUpdates: clause.Assignments(map[string]any{
			"pv":         gorm.Expr("analytics_content_hourly.pv + 1"),
			"uv":         gorm.Expr("analytics_content_hourly.uv + ?", isNewVisitor),
			"updated_at": gorm.Expr("NOW()"),
		}),
	}).Create(&model.AnalyticsContentHourly{
		ContentType: event.ContentType,
		ContentID:   event.ContentID,
		HourBucket:  hourBucket,
		PV:          1,
		UV:          isNewVisitor,
	}).Error; err != nil {
		return err
	}

	if err := s.incrementContentViews(ctx, event.ContentType, event.ContentID); err != nil {
		return err
	}

	if err := s.recordVisitorView(ctx, event); err != nil {
		return err
	}

	s.recordSourceStats(ctx, event)

	return nil
}

func (s *Service) recordVisitorView(ctx context.Context, event ViewTrackEvent) error {
	if strings.TrimSpace(event.VisitorID) == "" {
		return nil
	}
	nowAt := event.At.UTC()
	ip := strings.TrimSpace(event.IP)
	info := s.uaParser.Resolve(event.UserAgent)
	platform := strings.TrimSpace(info.Platform)
	if platform == "" {
		platform = "Unknown"
	}
	browser := strings.TrimSpace(info.Browser)
	if browser == "" {
		browser = "Unknown"
	}
	location := ""
	if s.geo != nil {
		location = strings.TrimSpace(s.geo.Resolve(event.IP))
	}
	if location == "" {
		location = "Unknown"
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "visitor_id"},
			{Name: "content_type"},
			{Name: "content_id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"last_view_at": gorm.Expr("EXCLUDED.last_view_at"),
			"view_count":   gorm.Expr("analytics_visitor_view.view_count + 1"),
			"last_ip":      gorm.Expr("CASE WHEN EXCLUDED.last_ip <> '' THEN EXCLUDED.last_ip ELSE analytics_visitor_view.last_ip END"),
			"platform":     gorm.Expr("EXCLUDED.platform"),
			"browser":      gorm.Expr("EXCLUDED.browser"),
			"location":     gorm.Expr("EXCLUDED.location"),
			"updated_at":   gorm.Expr("NOW()"),
		}),
	}).Create(&model.AnalyticsVisitorView{
		VisitorID:   event.VisitorID,
		ContentType: event.ContentType,
		ContentID:   event.ContentID,
		LastIP:      ip,
		Platform:    platform,
		Browser:     browser,
		Location:    location,
		FirstViewAt: nowAt,
		LastViewAt:  nowAt,
		ViewCount:   1,
	}).Error
}

func (s *Service) recordSourceStats(ctx context.Context, event ViewTrackEvent) {
	if s.redis == nil {
		return
	}
	day := event.At.UTC().Format("20060102")
	info := s.uaParser.Resolve(event.UserAgent)
	platform := strings.TrimSpace(info.Platform)
	if platform == "" {
		platform = "Unknown"
	}
	browser := strings.TrimSpace(info.Browser)
	if browser == "" {
		browser = "Unknown"
	}
	location := ""
	if s.geo != nil {
		location = strings.TrimSpace(s.geo.Resolve(event.IP))
	}
	if location == "" {
		location = "Unknown"
	}

	platformKey := fmt.Sprintf("%sanalytics:source:platform:%s", s.redisPrefix, day)
	browserKey := fmt.Sprintf("%sanalytics:source:browser:%s", s.redisPrefix, day)
	locationKey := fmt.Sprintf("%sanalytics:source:location:%s", s.redisPrefix, day)
	ttl := 14 * 24 * time.Hour

	pipe := s.redis.TxPipeline()
	pipe.HIncrBy(ctx, platformKey, platform, 1)
	pipe.HIncrBy(ctx, browserKey, browser, 1)
	pipe.HIncrBy(ctx, locationKey, location, 1)
	pipe.Expire(ctx, platformKey, ttl)
	pipe.Expire(ctx, browserKey, ttl)
	pipe.Expire(ctx, locationKey, ttl)
	_, _ = pipe.Exec(ctx)
}

func (s *Service) incrementContentViews(ctx context.Context, contentType string, contentID int64) error {
	switch contentType {
	case ContentTypeArticle:
		return s.db.WithContext(ctx).Exec(`
			INSERT INTO article_metrics (article_id, views, likes, comments, updated_at)
			VALUES (?, 1, 0, 0, NOW())
			ON CONFLICT (article_id)
			DO UPDATE SET views = article_metrics.views + 1, updated_at = NOW()
		`, contentID).Error
	case ContentTypeMoment:
		return s.db.WithContext(ctx).Exec(`
			INSERT INTO moment_metrics (moment_id, views, likes, comments, updated_at)
			VALUES (?, 1, 0, 0, NOW())
			ON CONFLICT (moment_id)
			DO UPDATE SET views = moment_metrics.views + 1, updated_at = NOW()
		`, contentID).Error
	case ContentTypePage:
		return s.db.WithContext(ctx).Exec(`
			INSERT INTO page_metrics (page_id, views, likes, comments, updated_at)
			VALUES (?, 1, 0, 0, NOW())
			ON CONFLICT (page_id)
			DO UPDATE SET views = page_metrics.views + 1, updated_at = NOW()
		`, contentID).Error
	default:
		return fmt.Errorf("invalid content type")
	}
}

func fallbackVisitorID(ip, ua string) string {
	raw := strings.TrimSpace(ip) + "|" + strings.TrimSpace(ua)
	if raw == "|" {
		raw = fmt.Sprintf("anonymous-%d", time.Now().UnixNano())
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:16])
}

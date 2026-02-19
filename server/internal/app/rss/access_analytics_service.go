package rss

import (
	"context"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/grtsinry43/grtblog-v2/server/internal/infra/clientinfo"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/geoip"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type AccessMeta struct {
	Path       string
	IP         string
	UserAgent  string
	ClientHint string
	At         time.Time
}

type AccessStats struct {
	Days          int            `json:"days"`
	GeneratedAt   time.Time      `json:"generatedAt"`
	Total         int64          `json:"total"`
	UniqueIP      int64          `json:"uniqueIp"`
	Trend         []AccessTrend  `json:"trend"`
	TopClients    []AccessBucket `json:"topClients"`
	TopIPs        []AccessBucket `json:"topIps"`
	TopPlatforms  []AccessBucket `json:"topPlatforms"`
	TopBrowsers   []AccessBucket `json:"topBrowsers"`
	TopLocations  []AccessBucket `json:"topLocations"`
	TopUserAgents []AccessBucket `json:"topUserAgents"`
	TopHints      []AccessBucket `json:"topHints"`
}

type AccessTrend struct {
	Hour     string `json:"hour"`
	Requests int64  `json:"requests"`
	UniqueIP int64  `json:"uniqueIp"`
}

type AccessBucket struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type AccessAnalyticsService struct {
	db       *gorm.DB
	uaParser *clientinfo.UAParser
	geo      *geoip.Resolver
	now      func() time.Time
}

func NewAccessAnalyticsService(db *gorm.DB, geoResolver *geoip.Resolver) *AccessAnalyticsService {
	return &AccessAnalyticsService{
		db:       db,
		uaParser: clientinfo.NewUAParser(),
		geo:      geoResolver,
		now:      time.Now,
	}
}

func (s *AccessAnalyticsService) RecordAccess(ctx context.Context, in AccessMeta) error {
	if s == nil || s.db == nil {
		return nil
	}
	ip := strings.TrimSpace(in.IP)
	if ip == "" {
		ip = "unknown"
	}
	path := strings.TrimSpace(in.Path)
	if path == "" {
		path = "/feed"
	}
	if len(path) > 64 {
		path = path[:64]
	}

	at := in.At
	if at.IsZero() {
		at = s.now()
	}
	hour := at.UTC().Truncate(time.Hour)

	ua := strings.TrimSpace(in.UserAgent)
	if len(ua) > 512 {
		ua = ua[:512]
	}
	clientHint := strings.TrimSpace(in.ClientHint)
	if len(clientHint) > 128 {
		clientHint = clientHint[:128]
	}
	clientName := detectRSSClient(ua, clientHint)
	info := s.uaParser.Resolve(ua)
	platform := strings.TrimSpace(info.Platform)
	browser := strings.TrimSpace(info.Browser)
	location := ""
	if s.geo != nil {
		location = strings.TrimSpace(s.geo.Resolve(ip))
	}

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "hour_bucket"},
			{Name: "request_path"},
			{Name: "ip"},
			{Name: "client_name"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"requests":    gorm.Expr("analytics_rss_access_hourly.requests + 1"),
			"client_hint": gorm.Expr("EXCLUDED.client_hint"),
			"user_agent":  gorm.Expr("EXCLUDED.user_agent"),
			"platform":    gorm.Expr("EXCLUDED.platform"),
			"browser":     gorm.Expr("EXCLUDED.browser"),
			"location":    gorm.Expr("EXCLUDED.location"),
			"updated_at":  gorm.Expr("NOW()"),
		}),
	}).Create(&model.AnalyticsRSSAccessHourly{
		HourBucket:  hour,
		RequestPath: path,
		IP:          ip,
		ClientName:  clientName,
		ClientHint:  clientHint,
		UserAgent:   ua,
		Platform:    platform,
		Browser:     browser,
		Location:    location,
		Requests:    1,
	}).Error
}

func (s *AccessAnalyticsService) GetStats(ctx context.Context, days, topN int) (*AccessStats, error) {
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}
	if topN <= 0 {
		topN = 12
	}
	if topN > 50 {
		topN = 50
	}
	start := s.now().UTC().AddDate(0, 0, -(days - 1)).Truncate(24 * time.Hour)

	type sumRow struct {
		Total    int64 `gorm:"column:total"`
		UniqueIP int64 `gorm:"column:unique_ip"`
	}
	var sum sumRow
	if err := s.db.WithContext(ctx).Model(&model.AnalyticsRSSAccessHourly{}).
		Select("COALESCE(SUM(requests),0) AS total, COUNT(DISTINCT ip) AS unique_ip").
		Where("hour_bucket >= ?", start).
		Scan(&sum).Error; err != nil {
		return nil, err
	}

	type trendRow struct {
		Hour     time.Time `gorm:"column:hour"`
		Requests int64     `gorm:"column:requests"`
		UniqueIP int64     `gorm:"column:unique_ip"`
	}
	var trendRows []trendRow
	if err := s.db.WithContext(ctx).Model(&model.AnalyticsRSSAccessHourly{}).
		Select("hour_bucket AS hour, COALESCE(SUM(requests),0) AS requests, COUNT(DISTINCT ip) AS unique_ip").
		Where("hour_bucket >= ?", start).
		Group("hour_bucket").
		Order("hour_bucket ASC").
		Scan(&trendRows).Error; err != nil {
		return nil, err
	}
	trend := make([]AccessTrend, 0, len(trendRows))
	for _, row := range trendRows {
		trend = append(trend, AccessTrend{
			Hour:     row.Hour.Format("2006-01-02 15:00"),
			Requests: row.Requests,
			UniqueIP: row.UniqueIP,
		})
	}

	topClients, err := s.queryTop(ctx, start, "client_name", topN)
	if err != nil {
		return nil, err
	}
	topIPs, err := s.queryTop(ctx, start, "ip", topN)
	if err != nil {
		return nil, err
	}
	topPlatforms, err := s.queryTop(ctx, start, "platform", topN)
	if err != nil {
		return nil, err
	}
	topBrowsers, err := s.queryTop(ctx, start, "browser", topN)
	if err != nil {
		return nil, err
	}
	topLocations, err := s.queryTop(ctx, start, "location", topN)
	if err != nil {
		return nil, err
	}
	topUserAgents, err := s.queryTop(ctx, start, "user_agent", topN)
	if err != nil {
		return nil, err
	}
	topHints, err := s.queryTop(ctx, start, "client_hint", topN)
	if err != nil {
		return nil, err
	}

	return &AccessStats{
		Days:          days,
		GeneratedAt:   s.now().UTC(),
		Total:         sum.Total,
		UniqueIP:      sum.UniqueIP,
		Trend:         trend,
		TopClients:    topClients,
		TopIPs:        topIPs,
		TopPlatforms:  topPlatforms,
		TopBrowsers:   topBrowsers,
		TopLocations:  topLocations,
		TopUserAgents: topUserAgents,
		TopHints:      topHints,
	}, nil
}

func (s *AccessAnalyticsService) queryTop(ctx context.Context, start time.Time, column string, topN int) ([]AccessBucket, error) {
	type row struct {
		Name  string `gorm:"column:name"`
		Count int64  `gorm:"column:count"`
	}
	var rows []row
	selectExpr := "COALESCE(NULLIF(TRIM(" + column + "),''), 'Unknown') AS name, COALESCE(SUM(requests),0) AS count"
	if err := s.db.WithContext(ctx).Model(&model.AnalyticsRSSAccessHourly{}).
		Select(selectExpr).
		Where("hour_bucket >= ?", start).
		Group("name").
		Order("count DESC").
		Limit(topN).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]AccessBucket, 0, len(rows))
	for _, row := range rows {
		items = append(items, AccessBucket{Name: row.Name, Count: row.Count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Name < items[j].Name
		}
		return items[i].Count > items[j].Count
	})
	return items, nil
}

func detectRSSClient(userAgent string, clientHint string) string {
	hint := strings.ToLower(strings.TrimSpace(clientHint))
	if hint != "" {
		switch {
		case strings.Contains(hint, "feedly"):
			return "Feedly"
		case strings.Contains(hint, "inoreader"):
			return "Inoreader"
		case strings.Contains(hint, "newsblur"):
			return "NewsBlur"
		case strings.Contains(hint, "miniflux"):
			return "Miniflux"
		case strings.Contains(hint, "netnewswire"):
			return "NetNewsWire"
		case strings.Contains(hint, "feeder"):
			return "Feeder"
		}
	}

	ua := strings.TrimSpace(userAgent)
	if ua == "" {
		return "Unknown"
	}
	l := strings.ToLower(ua)
	switch {
	case strings.Contains(l, "feedly"):
		return "Feedly"
	case strings.Contains(l, "inoreader"):
		return "Inoreader"
	case strings.Contains(l, "newsblur"):
		return "NewsBlur"
	case strings.Contains(l, "miniflux"):
		return "Miniflux"
	case strings.Contains(l, "netnewswire"):
		return "NetNewsWire"
	case strings.Contains(l, "feeder"):
		return "Feeder"
	case strings.Contains(l, "rss"):
		return "RSS Reader"
	case strings.Contains(l, "bot") || strings.Contains(l, "spider") || strings.Contains(l, "crawler"):
		return "Bot"
	default:
		return "Unknown"
	}
}

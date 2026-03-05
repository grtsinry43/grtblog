package adminstats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

const (
	defaultDashboardCacheTTL = 2 * time.Minute
	dashboardCacheVersion    = "v2"
)

type OnlineProvider interface {
	CurrentConnections() int64
}

type Service struct {
	db       *gorm.DB
	redis    *redis.Client
	online   OnlineProvider
	prefix   string
	cacheKey string
	cacheTTL time.Duration
	now      func() time.Time
}

type DashboardStats struct {
	GeneratedAt     time.Time           `json:"generatedAt"`
	Cached          bool                `json:"cached"`
	Overview        OverviewStats       `json:"overview"`
	Interaction     InteractionStats    `json:"interaction"`
	Words           WordCountStats      `json:"words"`
	Pending         PendingStats        `json:"pending"`
	Trend           []PublishTrendPoint `json:"trend"`
	ViewTrend       []DayCountPoint     `json:"viewTrend"`
	CommentTrend    []DayCountPoint     `json:"commentTrend"`
	Online24H       []OnlineTrendPoint  `json:"online24h"`
	CurrentOnline   int64               `json:"currentOnline"`
	TodayPeakOnline int64               `json:"todayPeakOnline"`
	Categories      []DistributionItem  `json:"categories"`
	Columns         []DistributionItem  `json:"columns"`
	TagTop          []DistributionItem  `json:"tagTop"`
	PlatformTop     []DistributionItem  `json:"platformTop"`
	BrowserTop      []DistributionItem  `json:"browserTop"`
	LocationTop     []DistributionItem  `json:"locationTop"`
	TopArticles     []HotContentItem    `json:"topArticles"`
	TopMoments      []HotContentItem    `json:"topMoments"`
	TopPages        []HotContentItem    `json:"topPages"`
	TopThinkings    []HotContentItem    `json:"topThinkings"`
}

type OverviewStats struct {
	Users             int64 `json:"users"`
	ArticlesTotal     int64 `json:"articlesTotal"`
	ArticlesPublished int64 `json:"articlesPublished"`
	ArticlesDraft     int64 `json:"articlesDraft"`
	MomentsTotal      int64 `json:"momentsTotal"`
	MomentsPublished  int64 `json:"momentsPublished"`
	MomentsDraft      int64 `json:"momentsDraft"`
	PagesTotal        int64 `json:"pagesTotal"`
	PagesEnabled      int64 `json:"pagesEnabled"`
	ThinkingsTotal    int64 `json:"thinkingsTotal"`
	CategoriesTotal   int64 `json:"categoriesTotal"`
	ColumnsTotal      int64 `json:"columnsTotal"`
	TagsTotal         int64 `json:"tagsTotal"`
}

type InteractionStats struct {
	ViewsTotal    int64 `json:"viewsTotal"`
	LikesTotal    int64 `json:"likesTotal"`
	CommentsTotal int64 `json:"commentsTotal"`

	ArticleViews    int64 `json:"articleViews"`
	ArticleLikes    int64 `json:"articleLikes"`
	ArticleComments int64 `json:"articleComments"`

	MomentViews    int64 `json:"momentViews"`
	MomentLikes    int64 `json:"momentLikes"`
	MomentComments int64 `json:"momentComments"`

	PageViews    int64 `json:"pageViews"`
	PageLikes    int64 `json:"pageLikes"`
	PageComments int64 `json:"pageComments"`

	ThinkingViews    int64 `json:"thinkingViews"`
	ThinkingLikes    int64 `json:"thinkingLikes"`
	ThinkingComments int64 `json:"thinkingComments"`
}

type WordCountStats struct {
	Total     int64 `json:"total"`
	Articles  int64 `json:"articles"`
	Moments   int64 `json:"moments"`
	Pages     int64 `json:"pages"`
	Thinkings int64 `json:"thinkings"`
}

type PendingStats struct {
	UnviewedComments       int64 `json:"unviewedComments"`
	FriendLinkApplications int64 `json:"friendLinkApplications"`
}

type PublishTrendPoint struct {
	Date      string `json:"date"`
	Articles  int64  `json:"articles"`
	Moments   int64  `json:"moments"`
	Pages     int64  `json:"pages"`
	Thinkings int64  `json:"thinkings"`
}

type DayCountPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type OnlineTrendPoint struct {
	Hour string  `json:"hour"`
	Peak int64   `json:"peak"`
	Avg  float64 `json:"avg"`
}

type DistributionItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type HotContentItem struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	ShortURL  string    `json:"shortUrl"`
	Views     int64     `json:"views"`
	Likes     int64     `json:"likes"`
	Comments  int64     `json:"comments"`
	Score     int64     `json:"score"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewService(db *gorm.DB, redisClient *redis.Client, redisPrefix string, online OnlineProvider) *Service {
	return &Service{
		db:       db,
		redis:    redisClient,
		online:   online,
		prefix:   redisPrefix,
		cacheKey: fmt.Sprintf("%sadmin:stats:dashboard:%s", redisPrefix, dashboardCacheVersion),
		cacheTTL: defaultDashboardCacheTTL,
		now:      time.Now,
	}
}

func (s *Service) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	if cached, ok := s.getCachedDashboard(ctx); ok {
		cached.Cached = true
		if s.online != nil {
			cached.CurrentOnline = s.online.CurrentConnections()
		}
		return cached, nil
	}

	stats, err := s.computeDashboardStats(ctx)
	if err != nil {
		return nil, err
	}
	stats.Cached = false
	stats.GeneratedAt = s.now().UTC()
	if s.online != nil {
		stats.CurrentOnline = s.online.CurrentConnections()
	}

	s.setCachedDashboard(ctx, stats)
	return stats, nil
}

func (s *Service) getCachedDashboard(ctx context.Context) (*DashboardStats, bool) {
	if s.redis == nil {
		return nil, false
	}
	payload, err := s.redis.Get(ctx, s.cacheKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false
		}
		return nil, false
	}
	var stats DashboardStats
	if err := json.Unmarshal(payload, &stats); err != nil {
		return nil, false
	}
	return &stats, true
}

func (s *Service) setCachedDashboard(ctx context.Context, stats *DashboardStats) {
	if s.redis == nil || stats == nil {
		return
	}
	payload, err := json.Marshal(stats)
	if err != nil {
		return
	}
	cacheCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	_ = s.redis.Set(cacheCtx, s.cacheKey, payload, s.cacheTTL).Err()
}

func (s *Service) computeDashboardStats(ctx context.Context) (*DashboardStats, error) {
	overview, err := s.queryOverviewStats(ctx)
	if err != nil {
		return nil, err
	}
	interaction, err := s.queryInteractionStats(ctx)
	if err != nil {
		return nil, err
	}
	words, err := s.queryWordStats(ctx)
	if err != nil {
		return nil, err
	}
	pending, err := s.queryPendingStats(ctx)
	if err != nil {
		return nil, err
	}

	trend, err := s.queryPublishTrend(ctx, 30)
	if err != nil {
		return nil, err
	}
	viewTrend, err := s.queryViewTrend(ctx, 30)
	if err != nil {
		return nil, err
	}
	commentTrend, err := s.queryCommentTrend(ctx, 30)
	if err != nil {
		return nil, err
	}
	online24h, todayPeak, err := s.queryOnlineTrend(ctx)
	if err != nil {
		return nil, err
	}
	categories, err := s.queryCategoryDistribution(ctx)
	if err != nil {
		return nil, err
	}
	columns, err := s.queryColumnDistribution(ctx)
	if err != nil {
		return nil, err
	}
	tagTop, err := s.queryTagTop(ctx, 20)
	if err != nil {
		return nil, err
	}
	platformTop, browserTop, locationTop, err := s.querySourceTopFromRedis(ctx, 7, 8)
	if err != nil {
		return nil, err
	}
	topArticles, err := s.queryTopArticles(ctx, 10)
	if err != nil {
		return nil, err
	}
	topMoments, err := s.queryTopMoments(ctx, 10)
	if err != nil {
		return nil, err
	}
	topPages, err := s.queryTopPages(ctx, 10)
	if err != nil {
		return nil, err
	}
	topThinkings, err := s.queryTopThinkings(ctx, 10)
	if err != nil {
		return nil, err
	}

	return &DashboardStats{
		Overview:        overview,
		Interaction:     interaction,
		Words:           words,
		Pending:         pending,
		Trend:           trend,
		ViewTrend:       viewTrend,
		CommentTrend:    commentTrend,
		Online24H:       online24h,
		TodayPeakOnline: todayPeak,
		Categories:      categories,
		Columns:         columns,
		TagTop:          tagTop,
		PlatformTop:     platformTop,
		BrowserTop:      browserTop,
		LocationTop:     locationTop,
		TopArticles:     topArticles,
		TopMoments:      topMoments,
		TopPages:        topPages,
		TopThinkings:    topThinkings,
	}, nil
}

func (s *Service) queryOverviewStats(ctx context.Context) (OverviewStats, error) {
	var out OverviewStats
	var err error
	if out.Users, err = s.count(ctx, &model.User{}); err != nil {
		return out, err
	}
	if out.ArticlesTotal, err = s.count(ctx, &model.Article{}); err != nil {
		return out, err
	}
	if out.ArticlesPublished, err = s.countWhere(ctx, &model.Article{}, "is_published = ?", true); err != nil {
		return out, err
	}
	out.ArticlesDraft = maxInt64(out.ArticlesTotal-out.ArticlesPublished, 0)
	if out.MomentsTotal, err = s.count(ctx, &model.Moment{}); err != nil {
		return out, err
	}
	if out.MomentsPublished, err = s.countWhere(ctx, &model.Moment{}, "is_published = ?", true); err != nil {
		return out, err
	}
	out.MomentsDraft = maxInt64(out.MomentsTotal-out.MomentsPublished, 0)
	if out.PagesTotal, err = s.count(ctx, &model.Page{}); err != nil {
		return out, err
	}
	if out.PagesEnabled, err = s.countWhere(ctx, &model.Page{}, "is_enabled = ?", true); err != nil {
		return out, err
	}
	if out.ThinkingsTotal, err = s.count(ctx, &model.Thinking{}); err != nil {
		return out, err
	}
	if out.CategoriesTotal, err = s.count(ctx, &model.ArticleCategory{}); err != nil {
		return out, err
	}
	if out.ColumnsTotal, err = s.count(ctx, &model.MomentColumn{}); err != nil {
		return out, err
	}
	if out.TagsTotal, err = s.count(ctx, &model.Tag{}); err != nil {
		return out, err
	}
	return out, nil
}

func (s *Service) queryInteractionStats(ctx context.Context) (InteractionStats, error) {
	var out InteractionStats
	article, err := s.sumMetrics(ctx, &model.ArticleMetrics{})
	if err != nil {
		return out, err
	}
	moment, err := s.sumMetrics(ctx, &model.MomentMetrics{})
	if err != nil {
		return out, err
	}
	page, err := s.sumMetrics(ctx, &model.PageMetrics{})
	if err != nil {
		return out, err
	}
	thinking, err := s.sumMetrics(ctx, &model.ThinkingMetrics{})
	if err != nil {
		return out, err
	}
	out.ArticleViews, out.ArticleLikes, out.ArticleComments = article.views, article.likes, article.comments
	out.MomentViews, out.MomentLikes, out.MomentComments = moment.views, moment.likes, moment.comments
	out.PageViews, out.PageLikes, out.PageComments = page.views, page.likes, page.comments
	out.ThinkingViews, out.ThinkingLikes, out.ThinkingComments = thinking.views, thinking.likes, thinking.comments
	out.ViewsTotal = out.ArticleViews + out.MomentViews + out.PageViews + out.ThinkingViews
	out.LikesTotal = out.ArticleLikes + out.MomentLikes + out.PageLikes + out.ThinkingLikes
	out.CommentsTotal = out.ArticleComments + out.MomentComments + out.PageComments + out.ThinkingComments
	return out, nil
}

func (s *Service) queryWordStats(ctx context.Context) (WordCountStats, error) {
	var out WordCountStats
	var err error
	if out.Articles, err = s.sumContentLength(ctx, "article", "content"); err != nil {
		return out, err
	}
	if out.Moments, err = s.sumContentLength(ctx, "moment", "content"); err != nil {
		return out, err
	}
	if out.Pages, err = s.sumContentLength(ctx, "page", "content"); err != nil {
		return out, err
	}
	if out.Thinkings, err = s.sumContentLength(ctx, "thinking", "content"); err != nil {
		return out, err
	}
	out.Total = out.Articles + out.Moments + out.Pages + out.Thinkings
	return out, nil
}

func (s *Service) queryPendingStats(ctx context.Context) (PendingStats, error) {
	var out PendingStats
	var err error
	if out.UnviewedComments, err = s.countWhere(ctx, &model.Comment{}, "is_viewed = ?", false); err != nil {
		return out, err
	}
	if out.FriendLinkApplications, err = s.countWhere(ctx, &model.FriendLinkApplication{}, "status = ?", "pending"); err != nil {
		return out, err
	}
	return out, nil
}

func (s *Service) queryPublishTrend(ctx context.Context, days int) ([]PublishTrendPoint, error) {
	if days <= 0 {
		days = 30
	}
	start := s.now().UTC().AddDate(0, 0, -(days - 1)).Truncate(24 * time.Hour)
	articleMap, err := s.queryDayCountMap(ctx, s.db.WithContext(ctx).Model(&model.Article{}).Where("is_published = ?", true), start)
	if err != nil {
		return nil, err
	}
	momentMap, err := s.queryDayCountMap(ctx, s.db.WithContext(ctx).Model(&model.Moment{}).Where("is_published = ?", true), start)
	if err != nil {
		return nil, err
	}
	pageMap, err := s.queryDayCountMap(ctx, s.db.WithContext(ctx).Model(&model.Page{}).Where("is_enabled = ?", true), start)
	if err != nil {
		return nil, err
	}
	thinkingMap, err := s.queryDayCountMap(ctx, s.db.WithContext(ctx).Model(&model.Thinking{}), start)
	if err != nil {
		return nil, err
	}

	points := make([]PublishTrendPoint, 0, days)
	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i).Format("2006-01-02")
		points = append(points, PublishTrendPoint{Date: day, Articles: articleMap[day], Moments: momentMap[day], Pages: pageMap[day], Thinkings: thinkingMap[day]})
	}
	return points, nil
}

func (s *Service) queryViewTrend(ctx context.Context, days int) ([]DayCountPoint, error) {
	return s.queryAnalyticsCountByDay(ctx, days, "COALESCE(SUM(pv),0)")
}

func (s *Service) queryCommentTrend(ctx context.Context, days int) ([]DayCountPoint, error) {
	if days <= 0 {
		days = 30
	}
	start := s.now().UTC().AddDate(0, 0, -(days - 1)).Truncate(24 * time.Hour)
	q := s.db.WithContext(ctx).Model(&model.Comment{}).Where("created_at >= ?", start).Select("DATE(created_at) AS day, COUNT(1) AS count").Group("DATE(created_at)")
	dayMap, err := scanDayCountMap(q)
	if err != nil {
		return nil, err
	}
	return fillDaySeries(start, days, dayMap), nil
}

func (s *Service) queryOnlineTrend(ctx context.Context) ([]OnlineTrendPoint, int64, error) {
	start := s.now().UTC().Add(-23 * time.Hour).Truncate(time.Hour)
	type row struct {
		HourBucket  time.Time `gorm:"column:hour_bucket"`
		PeakOnline  int64     `gorm:"column:peak_online"`
		SampleTotal int64     `gorm:"column:sample_total"`
		SampleCount int64     `gorm:"column:sample_count"`
	}
	var rows []row
	if err := s.db.WithContext(ctx).Model(&model.AnalyticsOnlineHourly{}).
		Where("hour_bucket >= ?", start).
		Order("hour_bucket ASC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	data := make(map[string]row, len(rows))
	for _, r := range rows {
		data[r.HourBucket.Format("2006-01-02 15:00")] = r
	}
	out := make([]OnlineTrendPoint, 0, 24)
	var todayPeak int64
	todayStart := s.now().UTC().Truncate(24 * time.Hour)
	for i := 0; i < 24; i++ {
		h := start.Add(time.Duration(i) * time.Hour)
		key := h.Format("2006-01-02 15:00")
		r := data[key]
		avg := 0.0
		if r.SampleCount > 0 {
			avg = float64(r.SampleTotal) / float64(r.SampleCount)
		}
		out = append(out, OnlineTrendPoint{Hour: key, Peak: r.PeakOnline, Avg: avg})
		if h.After(todayStart) || h.Equal(todayStart) {
			if r.PeakOnline > todayPeak {
				todayPeak = r.PeakOnline
			}
		}
	}
	if s.online != nil && s.online.CurrentConnections() > todayPeak {
		todayPeak = s.online.CurrentConnections()
	}
	return out, todayPeak, nil
}

func (s *Service) queryCategoryDistribution(ctx context.Context) ([]DistributionItem, error) {
	type row struct {
		Name  string `gorm:"column:name"`
		Count int64  `gorm:"column:count"`
	}
	var rows []row
	err := s.db.WithContext(ctx).
		Model(&model.Article{}).
		Select("COALESCE(article_category.name, ?) AS name, COUNT(article.id) AS count", "未分类").
		Joins("LEFT JOIN article_category ON article_category.id = article.category_id").
		Where("article.is_published = ?", true).
		Group("article_category.name").
		Order("count DESC, name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]DistributionItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, DistributionItem{Name: r.Name, Count: r.Count})
	}
	return items, nil
}

func (s *Service) queryColumnDistribution(ctx context.Context) ([]DistributionItem, error) {
	type row struct {
		Name  string `gorm:"column:name"`
		Count int64  `gorm:"column:count"`
	}
	var rows []row
	err := s.db.WithContext(ctx).
		Model(&model.Moment{}).
		Select("COALESCE(moment_column.name, ?) AS name, COUNT(moment.id) AS count", "未分栏").
		Joins("LEFT JOIN moment_column ON moment_column.id = moment.column_id").
		Where("moment.is_published = ?", true).
		Group("moment_column.name").
		Order("count DESC, name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]DistributionItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, DistributionItem{Name: r.Name, Count: r.Count})
	}
	return items, nil
}

func (s *Service) queryTagTop(ctx context.Context, limit int) ([]DistributionItem, error) {
	if limit <= 0 {
		limit = 20
	}
	type row struct {
		Name  string `gorm:"column:name"`
		Count int64  `gorm:"column:count"`
	}
	var rows []row
	err := s.db.WithContext(ctx).Raw(`
		SELECT t.name AS name, COUNT(*) AS count
		FROM tag t
		JOIN (
			SELECT tag_id FROM article_tag
			UNION ALL
			SELECT tag_id FROM moment_topic
		) x ON x.tag_id = t.id
		GROUP BY t.name
		ORDER BY count DESC, t.name ASC
		LIMIT ?
	`, limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]DistributionItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, DistributionItem{Name: r.Name, Count: r.Count})
	}
	return items, nil
}

func (s *Service) querySourceTopFromRedis(ctx context.Context, days, topN int) ([]DistributionItem, []DistributionItem, []DistributionItem, error) {
	if days <= 0 {
		days = 7
	}
	platformMap := map[string]int64{}
	browserMap := map[string]int64{}
	locationMap := map[string]int64{}
	if s.redis != nil {
		for i := 0; i < days; i++ {
			dateKey := s.now().UTC().AddDate(0, 0, -i).Format("20060102")
			platformKey := fmt.Sprintf("%sanalytics:source:platform:%s", s.prefix, dateKey)
			browserKey := fmt.Sprintf("%sanalytics:source:browser:%s", s.prefix, dateKey)
			locationKey := fmt.Sprintf("%sanalytics:source:location:%s", s.prefix, dateKey)

			mergeRedisHash(ctx, s.redis, platformKey, platformMap)
			mergeRedisHash(ctx, s.redis, browserKey, browserMap)
			mergeRedisHash(ctx, s.redis, locationKey, locationMap)
		}
	}
	platformTop := topMap(platformMap, topN)
	browserTop := topMap(browserMap, topN)
	locationTop := topMap(locationMap, topN)
	if len(platformTop) > 0 && len(browserTop) > 0 && len(locationTop) > 0 {
		return platformTop, browserTop, locationTop, nil
	}

	dbPlatformTop, dbBrowserTop, dbLocationTop, err := s.querySourceTopFromDB(ctx, days, topN)
	if err != nil {
		// Keep dashboard available even if source fallback query fails.
		return platformTop, browserTop, locationTop, nil
	}
	if len(platformTop) == 0 {
		platformTop = dbPlatformTop
	}
	if len(browserTop) == 0 {
		browserTop = dbBrowserTop
	}
	if len(locationTop) == 0 {
		locationTop = dbLocationTop
	}
	return platformTop, browserTop, locationTop, nil
}

func (s *Service) querySourceTopFromDB(ctx context.Context, days, topN int) ([]DistributionItem, []DistributionItem, []DistributionItem, error) {
	start := s.now().UTC().AddDate(0, 0, -(days - 1)).Truncate(24 * time.Hour)
	platformTop, err := s.querySourceDimensionFromDB(ctx, start, "platform", topN)
	if err != nil {
		return nil, nil, nil, err
	}
	browserTop, err := s.querySourceDimensionFromDB(ctx, start, "browser", topN)
	if err != nil {
		return nil, nil, nil, err
	}
	locationTop, err := s.querySourceDimensionFromDB(ctx, start, "location", topN)
	if err != nil {
		return nil, nil, nil, err
	}
	return platformTop, browserTop, locationTop, nil
}

func (s *Service) querySourceDimensionFromDB(ctx context.Context, start time.Time, column string, topN int) ([]DistributionItem, error) {
	switch column {
	case "platform", "browser", "location":
	default:
		return nil, fmt.Errorf("invalid source dimension: %s", column)
	}

	type row struct {
		Name  string `gorm:"column:name"`
		Count int64  `gorm:"column:count"`
	}
	sql := fmt.Sprintf(`
		SELECT COALESCE(NULLIF(TRIM(%s), ''), 'Unknown') AS name, COALESCE(SUM(view_count), 0) AS count
		FROM analytics_visitor_view
		WHERE last_view_at >= ?
		GROUP BY name
		ORDER BY count DESC, name ASC
	`, column)

	var rows []row
	query := s.db.WithContext(ctx).Raw(sql, start)
	if topN > 0 {
		query = s.db.WithContext(ctx).Raw(sql+" LIMIT ?", start, topN)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]DistributionItem, 0, len(rows))
	for _, r := range rows {
		name := strings.TrimSpace(r.Name)
		if name == "" || r.Count <= 0 {
			continue
		}
		items = append(items, DistributionItem{Name: name, Count: r.Count})
	}
	return items, nil
}

func (s *Service) queryTopArticles(ctx context.Context, limit int) ([]HotContentItem, error) {
	if limit <= 0 {
		limit = 10
	}
	type row struct {
		ID        int64     `gorm:"column:id"`
		Title     string    `gorm:"column:title"`
		ShortURL  string    `gorm:"column:short_url"`
		Views     int64     `gorm:"column:views"`
		Likes     int64     `gorm:"column:likes"`
		Comments  int64     `gorm:"column:comments"`
		Score     int64     `gorm:"column:score"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var rows []row
	err := s.db.WithContext(ctx).Model(&model.Article{}).Select(`
		article.id, article.title, article.short_url, article.created_at,
		COALESCE(article_metrics.views, 0) AS views,
		COALESCE(article_metrics.likes, 0) AS likes,
		COALESCE(article_metrics.comments, 0) AS comments,
		(COALESCE(article_metrics.views, 0) + COALESCE(article_metrics.likes, 0) * 5 + COALESCE(article_metrics.comments, 0) * 8) AS score
	`).Joins("LEFT JOIN article_metrics ON article_metrics.article_id = article.id").
		Where("article.is_published = ?", true).
		Order("score DESC, article.created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]HotContentItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, HotContentItem{
			ID:        r.ID,
			Title:     r.Title,
			ShortURL:  r.ShortURL,
			Views:     r.Views,
			Likes:     r.Likes,
			Comments:  r.Comments,
			Score:     r.Score,
			CreatedAt: r.CreatedAt,
		})
	}
	return items, nil
}

func (s *Service) queryTopMoments(ctx context.Context, limit int) ([]HotContentItem, error) {
	if limit <= 0 {
		limit = 10
	}
	type row struct {
		ID        int64     `gorm:"column:id"`
		Title     string    `gorm:"column:title"`
		ShortURL  string    `gorm:"column:short_url"`
		Views     int64     `gorm:"column:views"`
		Likes     int64     `gorm:"column:likes"`
		Comments  int64     `gorm:"column:comments"`
		Score     int64     `gorm:"column:score"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var rows []row
	err := s.db.WithContext(ctx).Model(&model.Moment{}).Select(`
		moment.id, moment.title, moment.short_url, moment.created_at,
		COALESCE(moment_metrics.views, 0) AS views,
		COALESCE(moment_metrics.likes, 0) AS likes,
		COALESCE(moment_metrics.comments, 0) AS comments,
		(COALESCE(moment_metrics.views, 0) + COALESCE(moment_metrics.likes, 0) * 5 + COALESCE(moment_metrics.comments, 0) * 8) AS score
	`).Joins("LEFT JOIN moment_metrics ON moment_metrics.moment_id = moment.id").
		Where("moment.is_published = ?", true).
		Order("score DESC, moment.created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]HotContentItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, HotContentItem{
			ID:        r.ID,
			Title:     r.Title,
			ShortURL:  r.ShortURL,
			Views:     r.Views,
			Likes:     r.Likes,
			Comments:  r.Comments,
			Score:     r.Score,
			CreatedAt: r.CreatedAt,
		})
	}
	return items, nil
}

func (s *Service) queryTopPages(ctx context.Context, limit int) ([]HotContentItem, error) {
	if limit <= 0 {
		limit = 10
	}
	type row struct {
		ID        int64     `gorm:"column:id"`
		Title     string    `gorm:"column:title"`
		ShortURL  string    `gorm:"column:short_url"`
		Views     int64     `gorm:"column:views"`
		Likes     int64     `gorm:"column:likes"`
		Comments  int64     `gorm:"column:comments"`
		Score     int64     `gorm:"column:score"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var rows []row
	err := s.db.WithContext(ctx).Model(&model.Page{}).Select(`
		page.id, page.title, page.short_url, page.created_at,
		COALESCE(page_metrics.views, 0) AS views,
		COALESCE(page_metrics.likes, 0) AS likes,
		COALESCE(page_metrics.comments, 0) AS comments,
		(COALESCE(page_metrics.views, 0) + COALESCE(page_metrics.likes, 0) * 5 + COALESCE(page_metrics.comments, 0) * 8) AS score
	`).Joins("LEFT JOIN page_metrics ON page_metrics.page_id = page.id").
		Where("page.is_enabled = ?", true).
		Order("score DESC, page.created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]HotContentItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, HotContentItem{
			ID:        r.ID,
			Title:     r.Title,
			ShortURL:  r.ShortURL,
			Views:     r.Views,
			Likes:     r.Likes,
			Comments:  r.Comments,
			Score:     r.Score,
			CreatedAt: r.CreatedAt,
		})
	}
	return items, nil
}

func (s *Service) queryTopThinkings(ctx context.Context, limit int) ([]HotContentItem, error) {
	if limit <= 0 {
		limit = 10
	}
	type row struct {
		ID        int64     `gorm:"column:id"`
		Content   string    `gorm:"column:content"`
		Views     int64     `gorm:"column:views"`
		Likes     int64     `gorm:"column:likes"`
		Comments  int64     `gorm:"column:comments"`
		Score     int64     `gorm:"column:score"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var rows []row
	err := s.db.WithContext(ctx).Model(&model.Thinking{}).Select(`
		thinking.id, thinking.content, thinking.created_at,
		COALESCE(thinking_metrics.views, 0) AS views,
		COALESCE(thinking_metrics.likes, 0) AS likes,
		COALESCE(thinking_metrics.comments, 0) AS comments,
		(COALESCE(thinking_metrics.views, 0) + COALESCE(thinking_metrics.likes, 0) * 5 + COALESCE(thinking_metrics.comments, 0) * 8) AS score
	`).Joins("LEFT JOIN thinking_metrics ON thinking_metrics.thinking_id = thinking.id").
		Order("score DESC, thinking.created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]HotContentItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, HotContentItem{
			ID:        r.ID,
			Title:     buildThinkingTitle(r.ID, r.Content),
			ShortURL:  "",
			Views:     r.Views,
			Likes:     r.Likes,
			Comments:  r.Comments,
			Score:     r.Score,
			CreatedAt: r.CreatedAt,
		})
	}
	return items, nil
}

func buildThinkingTitle(id int64, content string) string {
	txt := strings.TrimSpace(content)
	if txt == "" {
		return fmt.Sprintf("Thinking #%d", id)
	}
	rs := []rune(txt)
	if len(rs) > 60 {
		return string(rs[:60]) + "..."
	}
	return txt
}

type metricSum struct {
	views    int64
	likes    int64
	comments int64
}

func (s *Service) sumMetrics(ctx context.Context, modelRef any) (metricSum, error) {
	type row struct {
		Views    int64 `gorm:"column:views"`
		Likes    int64 `gorm:"column:likes"`
		Comments int64 `gorm:"column:comments"`
	}
	var r row
	err := s.db.WithContext(ctx).Model(modelRef).
		Select("COALESCE(SUM(views),0) AS views, COALESCE(SUM(likes),0) AS likes, COALESCE(SUM(comments),0) AS comments").
		Scan(&r).Error
	if err != nil {
		return metricSum{}, err
	}
	return metricSum{views: r.Views, likes: r.Likes, comments: r.Comments}, nil
}

func (s *Service) queryDayCountMap(ctx context.Context, query *gorm.DB, start time.Time) (map[string]int64, error) {
	return scanDayCountMap(query.Where("created_at >= ?", start).Select("DATE(created_at) AS day, COUNT(1) AS count").Group("DATE(created_at)"))
}

func scanDayCountMap(query *gorm.DB) (map[string]int64, error) {
	type row struct {
		Day   string `gorm:"column:day"`
		Count int64  `gorm:"column:count"`
	}
	var rows []row
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, r := range rows {
		if strings.TrimSpace(r.Day) != "" {
			out[r.Day] = r.Count
		}
	}
	return out, nil
}

func fillDaySeries(start time.Time, days int, dayMap map[string]int64) []DayCountPoint {
	points := make([]DayCountPoint, 0, days)
	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i).Format("2006-01-02")
		points = append(points, DayCountPoint{Date: day, Count: dayMap[day]})
	}
	return points
}

func (s *Service) queryAnalyticsCountByDay(ctx context.Context, days int, aggregateExpr string) ([]DayCountPoint, error) {
	if days <= 0 {
		days = 30
	}
	start := s.now().UTC().AddDate(0, 0, -(days - 1)).Truncate(24 * time.Hour)
	type row struct {
		Day   string `gorm:"column:day"`
		Count int64  `gorm:"column:count"`
	}
	var rows []row
	err := s.db.WithContext(ctx).Model(&model.AnalyticsContentHourly{}).
		Where("hour_bucket >= ?", start).
		Select("DATE(hour_bucket) AS day, " + aggregateExpr + " AS count").
		Group("DATE(hour_bucket)").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	dayMap := make(map[string]int64, len(rows))
	for _, r := range rows {
		if r.Day != "" {
			dayMap[r.Day] = r.Count
		}
	}
	return fillDaySeries(start, days, dayMap), nil
}

func (s *Service) sumContentLength(ctx context.Context, tableName, columnName string) (int64, error) {
	type row struct {
		Val int64 `gorm:"column:val"`
	}
	var out row
	query := fmt.Sprintf("COALESCE(SUM(CHAR_LENGTH(%s)), 0) AS val", columnName)
	err := s.db.WithContext(ctx).Table(tableName).Select(query).Scan(&out).Error
	return out.Val, err
}

func (s *Service) count(ctx context.Context, modelRef any) (int64, error) {
	var total int64
	return total, s.db.WithContext(ctx).Model(modelRef).Count(&total).Error
}

func (s *Service) countWhere(ctx context.Context, modelRef any, query string, args ...any) (int64, error) {
	var total int64
	return total, s.db.WithContext(ctx).Model(modelRef).Where(query, args...).Count(&total).Error
}

func topMap(m map[string]int64, topN int) []DistributionItem {
	items := make([]DistributionItem, 0, len(m))
	for k, v := range m {
		if strings.TrimSpace(k) == "" || v <= 0 {
			continue
		}
		items = append(items, DistributionItem{Name: k, Count: v})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Name < items[j].Name
		}
		return items[i].Count > items[j].Count
	})
	if topN > 0 && len(items) > topN {
		items = items[:topN]
	}
	return items
}

func mergeRedisHash(ctx context.Context, cli *redis.Client, key string, target map[string]int64) {
	vals, err := cli.HGetAll(ctx, key).Result()
	if err != nil {
		return
	}
	for k, v := range vals {
		count, err := parseInt64(v)
		if err != nil {
			continue
		}
		target[k] += count
	}
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

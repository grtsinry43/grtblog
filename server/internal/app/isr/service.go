package isr

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/htmlsnapshot"
	domainalbum "github.com/grtsinry43/grtblog-v2/server/internal/domain/album"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	domainthinking "github.com/grtsinry43/grtblog-v2/server/internal/domain/thinking"
)

const (
	defaultDebounce      = 2 * time.Second
	defaultTick          = time.Second
	defaultBatch         = 20
	defaultRecentLimit   = 30
	defaultTrackedLimit  = 200
	discoveryPageSize    = 200
	defaultTrackedPages  = 3
	defaultMomentPerPage = 20

	bootstrapVersionFile = "storage/html/.bootstrap-version"
)

var errStopWalk = errors.New("stop-walk")

type RenderRecord struct {
	URLPath       string   `json:"urlPath"`
	Trigger       string   `json:"trigger"`
	Status        string   `json:"status"`
	Deps          []string `json:"deps,omitempty"`
	UpdatedFiles  []string `json:"updatedFiles,omitempty"`
	RemovedFiles  []string `json:"removedFiles,omitempty"`
	DurationMS    int64    `json:"durationMs"`
	Error         string   `json:"error,omitempty"`
	RenderedCount int64    `json:"renderedCount"`
}

type InvalidateReport struct {
	GeneratedAt    time.Time      `json:"generatedAt"`
	Source         string         `json:"source"`
	DepKeys        []string       `json:"depKeys"`
	DirectURLs     []string       `json:"directUrls"`
	MatchedURLs    []string       `json:"matchedUrls"`
	CandidateURLs  []string       `json:"candidateUrls"`
	EnqueuedURLs   []string       `json:"enqueuedUrls"`
	Rendered       []RenderRecord `json:"rendered"`
	QueueDepth     int64          `json:"queueDepth"`
	TrackedURLKeys int64          `json:"trackedUrlKeys"`
}

type BootstrapReport struct {
	GeneratedAt   time.Time      `json:"generatedAt"`
	StartedAt     time.Time      `json:"startedAt"`
	FinishedAt    time.Time      `json:"finishedAt"`
	DurationMS    int64          `json:"durationMs"`
	TotalRoutes   int            `json:"totalRoutes"`
	RenderedCount int64          `json:"renderedCount"`
	Routes        []string       `json:"routes"`
	Rendered      []RenderRecord `json:"rendered"`
	Failed        []RenderRecord `json:"failed"`
}

type TrackedPage struct {
	URLPath string   `json:"urlPath"`
	Deps    []string `json:"deps"`
}

type InvalidationActivity struct {
	GeneratedAt   time.Time `json:"generatedAt"`
	Source        string    `json:"source"`
	DepKeys       []string  `json:"depKeys"`
	CandidateURLs []string  `json:"candidateUrls"`
	EnqueuedURLs  []string  `json:"enqueuedUrls"`
	RenderedURLs  []string  `json:"renderedUrls"`
}

type StateSnapshot struct {
	GeneratedAt          time.Time                     `json:"generatedAt"`
	QueueDepth           int64                         `json:"queueDepth"`
	DepKeyCount          int64                         `json:"depKeyCount"`
	URLKeyCount          int64                         `json:"urlKeyCount"`
	TrackedPages         []TrackedPage                 `json:"trackedPages"`
	RecentInvalidations  []InvalidationActivity        `json:"recentInvalidations"`
	RecentRenderActivity []htmlsnapshot.RenderActivity `json:"recentRenderActivity"`
	LastBootstrap        *BootstrapReport              `json:"lastBootstrap,omitempty"`
}

type Service struct {
	redis        *redis.Client
	redisPrefix  string
	renderer     *htmlsnapshot.Service
	contentRepo  content.Repository
	albumRepo    domainalbum.Repository
	thinkingRepo domainthinking.ThinkingRepository
	debounce     time.Duration

	activityMu          sync.Mutex
	recentInvalidations []InvalidationActivity
	lastBootstrap       *BootstrapReport
}

func NewService(redisClient *redis.Client, redisPrefix string, renderer *htmlsnapshot.Service, contentRepo content.Repository, albumRepo domainalbum.Repository, thinkingRepo domainthinking.ThinkingRepository) *Service {
	return &Service{
		redis:        redisClient,
		redisPrefix:  redisPrefix,
		renderer:     renderer,
		contentRepo:  contentRepo,
		albumRepo:    albumRepo,
		thinkingRepo: thinkingRepo,
		debounce:     defaultDebounce,
	}
}

func (s *Service) Invalidate(ctx context.Context, depKeys []string, urls []string) error {
	_, err := s.InvalidateWithReport(ctx, depKeys, urls, "event", false)
	return err
}

func (s *Service) InvalidateWithReport(ctx context.Context, depKeys []string, urls []string, source string, syncRender bool) (*InvalidateReport, error) {
	report := &InvalidateReport{
		GeneratedAt: time.Now().UTC(),
		Source:      strings.TrimSpace(source),
		DepKeys:     normalizeDeps(depKeys),
		DirectURLs:  normalizeURLs(urls),
	}
	if report.Source == "" {
		report.Source = "manual"
	}

	matched, err := s.urlsByDeps(ctx, report.DepKeys)
	if err != nil {
		return nil, err
	}
	report.MatchedURLs = matched
	report.CandidateURLs = normalizeURLs(append(append([]string(nil), matched...), report.DirectURLs...))

	if syncRender {
		rendered, err := s.renderNow(ctx, report.CandidateURLs, fmt.Sprintf("isr:invalidate:%s", report.Source))
		if err != nil {
			return nil, err
		}
		report.Rendered = rendered
		report.EnqueuedURLs = nil
	} else {
		enqueued, err := s.enqueueURLs(ctx, report.CandidateURLs)
		if err != nil {
			return nil, err
		}
		report.EnqueuedURLs = enqueued
	}

	depCount, urlCount, depth, err := s.stats(ctx)
	if err == nil {
		report.TrackedURLKeys = urlCount
		report.QueueDepth = depth
		_ = depCount
	}

	s.recordInvalidation(report)
	return report, nil
}

func (s *Service) Bootstrap(ctx context.Context) (*BootstrapReport, error) {
	startedAt := time.Now().UTC()
	routes, err := s.DiscoverRoutes(ctx)
	if err != nil {
		return nil, err
	}

	report := &BootstrapReport{
		GeneratedAt: startedAt,
		StartedAt:   startedAt,
		Routes:      routes,
		TotalRoutes: len(routes),
		Rendered:    make([]RenderRecord, 0, len(routes)),
		Failed:      make([]RenderRecord, 0, len(routes)/10+1),
	}

	rendered, err := s.renderNow(ctx, routes, "isr:bootstrap")
	if err != nil {
		return nil, err
	}
	report.Rendered = rendered
	for _, item := range rendered {
		if item.Status == "error" {
			report.Failed = append(report.Failed, item)
		} else {
			report.RenderedCount += item.RenderedCount
		}
	}

	report.FinishedAt = time.Now().UTC()
	report.DurationMS = report.FinishedAt.Sub(startedAt).Milliseconds()
	s.setLastBootstrap(report)
	return report, nil
}

func (s *Service) DiscoverRoutes(ctx context.Context) ([]string, error) {
	routes := []string{
		"/",
		"/albums",
		"/friends",
		"/friends-timeline",
		"/moments",
		"/tags",
		"/thinkings",
		"/timeline",
	}
	if s.contentRepo == nil && s.albumRepo == nil && s.thinkingRepo == nil {
		return normalizeURLs(routes), nil
	}

	if s.contentRepo != nil {
		pageRoutes, err := s.discoverPageRoutes(ctx)
		if err != nil {
			return nil, err
		}
		routes = append(routes, pageRoutes...)

		categoryRoutes, err := s.discoverCategoryRoutes(ctx)
		if err != nil {
			return nil, err
		}
		routes = append(routes, categoryRoutes...)

		columnRoutes, err := s.discoverColumnRoutes(ctx)
		if err != nil {
			return nil, err
		}
		routes = append(routes, columnRoutes...)

		articleTotalPages, articleRoutes, err := s.discoverArticleRoutes(ctx)
		if err != nil {
			return nil, err
		}
		routes = append(routes, "/posts")
		routes = append(routes, articleRoutes...)
		for page := int64(1); page <= articleTotalPages; page++ {
			routes = append(routes, fmt.Sprintf("/posts/page/%d", page))
		}

		momentTotalPages, momentRoutes, err := s.discoverMomentRoutes(ctx)
		if err != nil {
			return nil, err
		}
		routes = append(routes, momentRoutes...)
		for page := int64(1); page <= momentTotalPages; page++ {
			if page == 1 {
				continue
			}
			routes = append(routes, fmt.Sprintf("/moments/page/%d", page))
		}
	}

	if s.thinkingRepo != nil {
		thinkingTotalPages, err := s.discoverThinkingTotalPages(ctx)
		if err != nil {
			return nil, err
		}
		for page := int64(1); page <= thinkingTotalPages; page++ {
			if page == 1 {
				continue
			}
			routes = append(routes, fmt.Sprintf("/thinkings/page/%d", page))
		}
	}

	if s.albumRepo != nil {
		albumRoutes, albumTotalPages, err := s.discoverAlbumRoutes(ctx)
		if err != nil {
			return nil, err
		}
		routes = append(routes, albumRoutes...)
		for page := int64(1); page <= albumTotalPages; page++ {
			if page == 1 {
				continue
			}
			routes = append(routes, fmt.Sprintf("/albums/page/%d", page))
		}
	}

	for p := 1; p <= defaultTrackedPages; p++ {
		if p > 1 {
			routes = append(routes, fmt.Sprintf("/friends-timeline/page/%d", p))
		}
	}

	return normalizeURLs(routes), nil
}

func (s *Service) Snapshot(ctx context.Context, trackedLimit int, recentLimit int) (*StateSnapshot, error) {
	if trackedLimit <= 0 {
		trackedLimit = defaultTrackedLimit
	}
	if recentLimit <= 0 {
		recentLimit = defaultRecentLimit
	}

	depCount, urlCount, depth, err := s.stats(ctx)
	if err != nil {
		return nil, err
	}
	trackedPages, err := s.loadTrackedPages(ctx, trackedLimit)
	if err != nil {
		return nil, err
	}

	s.activityMu.Lock()
	recentInvalidations := copyRecentInvalidations(s.recentInvalidations, recentLimit)
	lastBootstrap := s.lastBootstrap
	s.activityMu.Unlock()

	recentRenders := make([]htmlsnapshot.RenderActivity, 0)
	if s.renderer != nil {
		recentRenders = s.renderer.RecentActivities(recentLimit)
	}

	return &StateSnapshot{
		GeneratedAt:          time.Now().UTC(),
		QueueDepth:           depth,
		DepKeyCount:          depCount,
		URLKeyCount:          urlCount,
		TrackedPages:         trackedPages,
		RecentInvalidations:  recentInvalidations,
		RecentRenderActivity: recentRenders,
		LastBootstrap:        lastBootstrap,
	}, nil
}

func (s *Service) NeedsBootstrap(ctx context.Context) (bool, error) {
	if s.renderer == nil {
		return false, nil
	}
	if s.redis == nil {
		return true, nil
	}
	_, urlCount, _, err := s.stats(ctx)
	if err != nil {
		return false, err
	}
	return urlCount == 0, nil
}

func (s *Service) NeedsBootstrapForVersion(ctx context.Context, version string) (bool, string, error) {
	if s.renderer == nil {
		return false, "renderer_not_configured", nil
	}
	normalizedVersion := normalizeVersion(version)

	if s.redis != nil {
		_, urlCount, _, err := s.stats(ctx)
		if err != nil {
			return false, "", err
		}
		if urlCount == 0 {
			return true, "empty_isr_index", nil
		}
		currentVersion, err := s.redis.Get(ctx, s.bootstrapVersionKey()).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return false, "", err
		}
		currentVersion = normalizeVersion(currentVersion)
		if currentVersion != normalizedVersion {
			return true, fmt.Sprintf("version_changed:%s->%s", currentVersion, normalizedVersion), nil
		}
		return false, "version_unchanged", nil
	}

	hasHTML, err := hasRenderedHTML("storage/html")
	if err != nil {
		return false, "", err
	}
	if !hasHTML {
		return true, "missing_html_snapshot", nil
	}
	currentVersion, err := s.readBootstrapVersionFile()
	if err != nil {
		return false, "", err
	}
	currentVersion = normalizeVersion(currentVersion)
	if currentVersion != normalizedVersion {
		return true, fmt.Sprintf("version_changed:%s->%s", currentVersion, normalizedVersion), nil
	}
	return false, "version_unchanged", nil
}

func (s *Service) MarkBootstrapVersion(ctx context.Context, version string) error {
	normalizedVersion := normalizeVersion(version)
	var errs []error

	if s.redis != nil {
		if err := s.redis.Set(ctx, s.bootstrapVersionKey(), normalizedVersion, 0).Err(); err != nil {
			errs = append(errs, err)
		}
	}
	if err := s.writeBootstrapVersionFile(normalizedVersion); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (s *Service) ProcessDue(ctx context.Context, maxJobs int) error {
	if s.redis == nil || s.renderer == nil {
		return nil
	}
	if maxJobs <= 0 {
		maxJobs = defaultBatch
	}

	now := float64(time.Now().UnixMilli())
	items, err := s.redis.ZRangeByScore(ctx, s.queueKey(), &redis.ZRangeBy{
		Min:   "-inf",
		Max:   fmt.Sprintf("%.0f", now),
		Count: int64(maxJobs),
	}).Result()
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	for _, rawURL := range items {
		removed, err := s.redis.ZRem(ctx, s.queueKey(), rawURL).Result()
		if err != nil {
			log.Printf("[isr] dequeue failed url=%s err=%v", rawURL, err)
			continue
		}
		if removed == 0 {
			continue
		}

		detail, err := s.renderer.RenderURLDetailed(htmlsnapshot.WithRenderTrigger(ctx, "isr:worker"), rawURL)
		if err != nil {
			log.Printf("[isr] render failed url=%s err=%v", rawURL, err)
			retryAt := float64(time.Now().Add(5 * time.Second).UnixMilli())
			_ = s.redis.ZAdd(ctx, s.queueKey(), redis.Z{Score: retryAt, Member: rawURL}).Err()
			continue
		}
		_ = detail
	}
	return nil
}

func (s *Service) RunWorker(ctx context.Context, maxJobs int, tick time.Duration) {
	if s.redis == nil || s.renderer == nil {
		return
	}
	if maxJobs <= 0 {
		maxJobs = defaultBatch
	}
	if tick <= 0 {
		tick = defaultTick
	}

	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.ProcessDue(ctx, maxJobs); err != nil {
				log.Printf("[isr] process due failed: %v", err)
			}
		}
	}
}

// RenderErrorPage delegates to the htmlsnapshot renderer to capture the
// SvelteKit 404 error page as a static file for nginx fallback.
func (s *Service) RenderErrorPage(ctx context.Context) error {
	if s.renderer == nil {
		return nil
	}
	return s.renderer.RenderErrorPage(ctx)
}

func (s *Service) discoverArticleRoutes(ctx context.Context) (int64, []string, error) {
	paths := make([]string, 0, 256)
	page := 1
	var totalArticles int64
	for {
		items, total, err := s.contentRepo.ListPublicArticles(ctx, content.ArticleListOptions{
			Page:     page,
			PageSize: discoveryPageSize,
		})
		if err != nil {
			return 0, nil, err
		}
		totalArticles = total
		for _, item := range items {
			shortURL := strings.TrimSpace(item.ShortURL)
			if shortURL == "" {
				continue
			}
			paths = append(paths, fmt.Sprintf("/posts/%s", shortURL))
		}
		if len(items) == 0 || int64(page*discoveryPageSize) >= total {
			break
		}
		page++
	}

	totalPages := int64(1)
	if totalArticles > 0 {
		totalPages = (totalArticles + 10 - 1) / 10
	}
	return totalPages, paths, nil
}

func (s *Service) discoverMomentRoutes(ctx context.Context) (int64, []string, error) {
	paths := make([]string, 0, 256)
	page := 1
	var totalMoments int64
	for {
		items, total, err := s.contentRepo.ListPublicMoments(ctx, content.MomentListOptions{
			Page:     page,
			PageSize: discoveryPageSize,
		})
		if err != nil {
			return 0, nil, err
		}
		totalMoments = total
		for _, item := range items {
			shortURL := strings.TrimSpace(item.ShortURL)
			if shortURL == "" {
				continue
			}
			year, month, day := item.CreatedAt.Date()
			paths = append(paths, fmt.Sprintf("/moments/%04d/%02d/%02d/%s", year, int(month), day, shortURL))
		}
		if len(items) == 0 || int64(page*discoveryPageSize) >= total {
			break
		}
		page++
	}
	totalPages := int64(1)
	if totalMoments > 0 {
		totalPages = (totalMoments + defaultMomentPerPage - 1) / defaultMomentPerPage
	}
	return totalPages, paths, nil
}

func (s *Service) discoverPageRoutes(ctx context.Context) ([]string, error) {
	paths := make([]string, 0, 64)
	enabled := true
	page := 1
	for {
		items, total, err := s.contentRepo.ListPublicPages(ctx, content.PageListOptions{
			Page:     page,
			PageSize: discoveryPageSize,
			Enabled:  &enabled,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			shortURL := strings.TrimSpace(item.ShortURL)
			if shortURL == "" {
				continue
			}
			paths = append(paths, fmt.Sprintf("/%s", shortURL))
		}
		if len(items) == 0 || int64(page*discoveryPageSize) >= total {
			break
		}
		page++
	}
	return paths, nil
}

func (s *Service) discoverThinkingTotalPages(ctx context.Context) (int64, error) {
	if s.thinkingRepo == nil {
		return 1, nil
	}
	_, total, err := s.thinkingRepo.List(ctx, 1, 0)
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 1, nil
	}
	return (total + 20 - 1) / 20, nil
}

func (s *Service) discoverAlbumRoutes(ctx context.Context) ([]string, int64, error) {
	if s.albumRepo == nil {
		return nil, 1, nil
	}

	paths := make([]string, 0, 256)
	_, total, err := s.albumRepo.ListPublicAlbums(ctx, domainalbum.AlbumListOptions{
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		return nil, 0, err
	}

	shortURLs, err := s.albumRepo.ListPublishedAlbumShortURLs(ctx)
	if err != nil {
		return nil, 0, err
	}
	for _, shortURL := range shortURLs {
		shortURL = strings.TrimSpace(shortURL)
		if shortURL == "" {
			continue
		}
		albumPath := fmt.Sprintf("/albums/%s", shortURL)
		paths = append(paths, albumPath)

		albumItem, err := s.albumRepo.GetAlbumByShortURL(ctx, shortURL)
		if err != nil {
			return nil, 0, err
		}
		photos, err := s.albumRepo.ListPhotosByAlbumID(ctx, albumItem.ID)
		if err != nil {
			return nil, 0, err
		}
		for _, photo := range photos {
			paths = append(paths, fmt.Sprintf("%s/photo/%d", albumPath, photo.ID))
		}
	}

	totalPages := int64(1)
	if total > 0 {
		totalPages = (total + 20 - 1) / 20
	}
	return paths, totalPages, nil
}

func (s *Service) discoverCategoryRoutes(ctx context.Context) ([]string, error) {
	if s.contentRepo == nil {
		return nil, nil
	}

	categories, err := s.contentRepo.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(categories)*2)
	for _, category := range categories {
		if category == nil || category.ShortURL == nil || strings.TrimSpace(*category.ShortURL) == "" {
			continue
		}
		basePath := fmt.Sprintf("/categories/%s", strings.TrimSpace(*category.ShortURL))
		paths = append(paths, basePath)
		_, total, err := s.contentRepo.ListPublicArticles(ctx, content.ArticleListOptions{
			Page:       1,
			PageSize:   1,
			CategoryID: &category.ID,
		})
		if err != nil {
			return nil, err
		}
		totalPages := int64(1)
		if total > 0 {
			totalPages = (total + 10 - 1) / 10
		}
		for page := int64(2); page <= totalPages; page++ {
			paths = append(paths, fmt.Sprintf("%s/page/%d", basePath, page))
		}
	}
	return paths, nil
}

func (s *Service) discoverColumnRoutes(ctx context.Context) ([]string, error) {
	if s.contentRepo == nil {
		return nil, nil
	}

	columns, err := s.contentRepo.ListColumns(ctx)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(columns)*2)
	for _, column := range columns {
		if column == nil || column.ShortURL == nil || strings.TrimSpace(*column.ShortURL) == "" {
			continue
		}
		basePath := fmt.Sprintf("/columns/%s", strings.TrimSpace(*column.ShortURL))
		paths = append(paths, basePath)
		_, total, err := s.contentRepo.ListPublicMoments(ctx, content.MomentListOptions{
			Page:     1,
			PageSize: 1,
			ColumnID: &column.ID,
		})
		if err != nil {
			return nil, err
		}
		totalPages := int64(1)
		if total > 0 {
			totalPages = (total + defaultMomentPerPage - 1) / defaultMomentPerPage
		}
		for page := int64(2); page <= totalPages; page++ {
			paths = append(paths, fmt.Sprintf("%s/page/%d", basePath, page))
		}
	}
	return paths, nil
}

func (s *Service) renderNow(ctx context.Context, urls []string, trigger string) ([]RenderRecord, error) {
	if s.renderer == nil {
		return nil, nil
	}
	normalized := normalizeURLs(urls)
	records := make([]RenderRecord, 0, len(normalized))
	for _, rawURL := range normalized {
		detail, err := s.renderer.RenderURLDetailed(htmlsnapshot.WithRenderTrigger(ctx, trigger), rawURL)
		record := RenderRecord{
			URLPath:       detail.URLPath,
			Trigger:       detail.Trigger,
			Status:        detail.Status,
			Deps:          append([]string(nil), detail.Deps...),
			UpdatedFiles:  append([]string(nil), detail.UpdatedFiles...),
			RemovedFiles:  append([]string(nil), detail.RemovedFiles...),
			DurationMS:    detail.DurationMS,
			RenderedCount: int64(len(detail.UpdatedFiles)),
		}
		if err != nil {
			record.Status = "error"
			record.Error = err.Error()
		}
		records = append(records, record)
	}
	return records, nil
}

func (s *Service) enqueueURLs(ctx context.Context, urls []string) ([]string, error) {
	normalized := normalizeURLs(urls)
	if len(normalized) == 0 {
		return nil, nil
	}

	if s.redis == nil {
		records, err := s.renderNow(ctx, normalized, "isr:direct")
		if err != nil {
			return nil, err
		}
		enqueued := make([]string, 0, len(records))
		for _, item := range records {
			enqueued = append(enqueued, item.URLPath)
		}
		return enqueued, nil
	}

	queueKey := s.queueKey()
	enqueued := make([]string, 0, len(normalized))
	for _, rawURL := range normalized {
		lockKey := s.lockKey(rawURL)
		added, err := s.redis.SetNX(ctx, lockKey, "1", s.debounce).Result()
		if err != nil {
			return nil, err
		}
		if !added {
			continue
		}
		score := float64(time.Now().Add(s.debounce).UnixMilli())
		if err := s.redis.ZAdd(ctx, queueKey, redis.Z{Score: score, Member: rawURL}).Err(); err != nil {
			return nil, err
		}
		enqueued = append(enqueued, rawURL)
	}
	return enqueued, nil
}

func (s *Service) urlsByDeps(ctx context.Context, depKeys []string) ([]string, error) {
	if s.redis == nil {
		return nil, nil
	}

	normalizedDeps := normalizeDeps(depKeys)
	if len(normalizedDeps) == 0 {
		return nil, nil
	}

	urlSet := make(map[string]struct{}, 32)
	for _, dep := range normalizedDeps {
		members, err := s.redis.SMembers(ctx, s.depKey(dep)).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, err
		}
		for _, item := range members {
			normalized, err := htmlsnapshot.NormalizeURLPath(item)
			if err != nil {
				continue
			}
			urlSet[normalized] = struct{}{}
		}
	}

	out := make([]string, 0, len(urlSet))
	for item := range urlSet {
		out = append(out, item)
	}
	sort.Strings(out)
	return out, nil
}

func (s *Service) loadTrackedPages(ctx context.Context, limit int) ([]TrackedPage, error) {
	if s.redis == nil || limit <= 0 {
		return nil, nil
	}
	keys, err := s.scanKeys(ctx, s.urlPrefix()+"*", limit)
	if err != nil {
		return nil, err
	}
	items := make([]TrackedPage, 0, len(keys))
	for _, key := range keys {
		encoded := strings.TrimPrefix(key, s.urlPrefix())
		rawURL, err := url.QueryUnescape(encoded)
		if err != nil {
			continue
		}
		deps, err := s.redis.SMembers(ctx, key).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, err
		}
		items = append(items, TrackedPage{
			URLPath: rawURL,
			Deps:    normalizeDeps(deps),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].URLPath < items[j].URLPath })
	return items, nil
}

func (s *Service) stats(ctx context.Context) (depCount int64, urlCount int64, queueDepth int64, err error) {
	if s.redis == nil {
		return 0, 0, 0, nil
	}
	depCount, err = s.countKeys(ctx, s.depPrefix()+"*")
	if err != nil {
		return 0, 0, 0, err
	}
	urlCount, err = s.countKeys(ctx, s.urlPrefix()+"*")
	if err != nil {
		return 0, 0, 0, err
	}
	queueDepth, err = s.redis.ZCard(ctx, s.queueKey()).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, 0, 0, err
	}
	return depCount, urlCount, queueDepth, nil
}

func (s *Service) countKeys(ctx context.Context, pattern string) (int64, error) {
	if s.redis == nil {
		return 0, nil
	}
	var cursor uint64
	var count int64
	for {
		keys, next, err := s.redis.Scan(ctx, cursor, pattern, 500).Result()
		if err != nil {
			return 0, err
		}
		count += int64(len(keys))
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return count, nil
}

func (s *Service) scanKeys(ctx context.Context, pattern string, limit int) ([]string, error) {
	if s.redis == nil || limit <= 0 {
		return nil, nil
	}
	var cursor uint64
	out := make([]string, 0, limit)
	for {
		keys, next, err := s.redis.Scan(ctx, cursor, pattern, 200).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			out = append(out, key)
			if len(out) >= limit {
				return out, nil
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return out, nil
}

func (s *Service) recordInvalidation(report *InvalidateReport) {
	if report == nil {
		return
	}
	activity := InvalidationActivity{
		GeneratedAt:   report.GeneratedAt,
		Source:        report.Source,
		DepKeys:       cloneStrings(report.DepKeys),
		CandidateURLs: cloneStrings(report.CandidateURLs),
		EnqueuedURLs:  cloneStrings(report.EnqueuedURLs),
		RenderedURLs:  cloneStrings(renderedURLs(report.Rendered)),
	}
	s.activityMu.Lock()
	defer s.activityMu.Unlock()
	s.recentInvalidations = append(s.recentInvalidations, activity)
	if len(s.recentInvalidations) > 200 {
		s.recentInvalidations = s.recentInvalidations[len(s.recentInvalidations)-200:]
	}
}

func (s *Service) setLastBootstrap(report *BootstrapReport) {
	if report == nil {
		return
	}
	s.activityMu.Lock()
	defer s.activityMu.Unlock()
	copied := *report
	copied.Routes = append([]string(nil), report.Routes...)
	copied.Rendered = append([]RenderRecord(nil), report.Rendered...)
	copied.Failed = append([]RenderRecord(nil), report.Failed...)
	s.lastBootstrap = &copied
}

func renderedURLs(items []RenderRecord) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item.Status == "error" {
			continue
		}
		out = append(out, item.URLPath)
	}
	return out
}

func copyRecentInvalidations(items []InvalidationActivity, limit int) []InvalidationActivity {
	if len(items) == 0 || limit <= 0 {
		return nil
	}
	start := 0
	if len(items) > limit {
		start = len(items) - limit
	}
	out := make([]InvalidationActivity, 0, len(items)-start)
	for i := len(items) - 1; i >= start; i-- {
		item := items[i]
		item.DepKeys = cloneStrings(item.DepKeys)
		item.CandidateURLs = cloneStrings(item.CandidateURLs)
		item.EnqueuedURLs = cloneStrings(item.EnqueuedURLs)
		item.RenderedURLs = cloneStrings(item.RenderedURLs)
		out = append(out, item)
	}
	return out
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return make([]string, 0)
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

func (s *Service) queueKey() string {
	return fmt.Sprintf("%sisr:queue:zset", s.redisPrefix)
}

func (s *Service) lockKey(rawURL string) string {
	return fmt.Sprintf("%sisr:lock:url:%s", s.redisPrefix, url.QueryEscape(rawURL))
}

func (s *Service) depKey(dep string) string {
	return fmt.Sprintf("%sisr:dep:%s", s.redisPrefix, dep)
}

func (s *Service) depPrefix() string {
	return fmt.Sprintf("%sisr:dep:", s.redisPrefix)
}

func (s *Service) urlPrefix() string {
	return fmt.Sprintf("%sisr:url:", s.redisPrefix)
}

func (s *Service) bootstrapVersionKey() string {
	return fmt.Sprintf("%sisr:bootstrap:version", s.redisPrefix)
}

func (s *Service) readBootstrapVersionFile() (string, error) {
	content, err := os.ReadFile(bootstrapVersionFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

func (s *Service) writeBootstrapVersionFile(version string) error {
	if err := os.MkdirAll(filepath.Dir(bootstrapVersionFile), 0o755); err != nil {
		return err
	}
	return os.WriteFile(bootstrapVersionFile, []byte(version+"\n"), 0o644)
}

func normalizeDeps(deps []string) []string {
	set := make(map[string]struct{}, len(deps))
	out := make([]string, 0, len(deps))
	for _, dep := range deps {
		normalized := strings.TrimSpace(dep)
		if normalized == "" {
			continue
		}
		if _, exists := set[normalized]; exists {
			continue
		}
		set[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}

func normalizeURLs(urls []string) []string {
	set := make(map[string]struct{}, len(urls))
	out := make([]string, 0, len(urls))
	for _, item := range urls {
		normalized, err := htmlsnapshot.NormalizeURLPath(item)
		if err != nil {
			continue
		}
		if _, exists := set[normalized]; exists {
			continue
		}
		set[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}

func normalizeVersion(version string) string {
	normalized := strings.TrimSpace(version)
	if normalized == "" {
		return "dev"
	}
	return normalized
}

func hasRenderedHTML(root string) (bool, error) {
	if _, err := os.Stat(root); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	found := false
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".html") {
			found = true
			return errStopWalk
		}
		return nil
	})
	if err != nil && !errors.Is(err, errStopWalk) {
		return false, err
	}
	return found, nil
}

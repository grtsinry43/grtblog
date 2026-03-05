package htmlsnapshot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
)

const (
	defaultBaseURL = "http://localhost:3000"
	pageSize       = 100
	listPageSize   = 10

	depsHeader = "x-grt-deps"

	defaultRecentActivityLimit = 200
)

var ErrRemoteNotFound = errors.New("remote page not found")

type contextKey string

const renderTriggerContextKey contextKey = "htmlsnapshot:render-trigger"

type Service struct {
	contentRepo content.Repository
	baseURL     string
	client      *http.Client
	redis       *redis.Client
	redisPrefix string
	metricsMu   sync.Mutex
	metrics     MetricsSnapshot
	activityMu  sync.Mutex
	activities  []RenderActivity
	ogSem       chan struct{} // limits concurrent OG image fetch goroutines
}

type MetricsSnapshot struct {
	TotalJobs           int64     `json:"totalJobs"`
	SuccessJobs         int64     `json:"successJobs"`
	FailedJobs          int64     `json:"failedJobs"`
	LastDurationMS      int64     `json:"lastDurationMs"`
	LastRenderedFiles   int64     `json:"lastRenderedFiles"`
	TotalRenderedFiles  int64     `json:"totalRenderedFiles"`
	LastSuccessAt       time.Time `json:"lastSuccessAt,omitempty"`
	LastFailureAt       time.Time `json:"lastFailureAt,omitempty"`
	AverageDurationMS   float64   `json:"averageDurationMs"`
	P95DurationMS       float64   `json:"p95DurationMs"`
	lastDurationSamples []int64
}

type RenderDetail struct {
	URLPath      string   `json:"urlPath"`
	Trigger      string   `json:"trigger"`
	Status       string   `json:"status"`
	Deps         []string `json:"deps,omitempty"`
	UpdatedFiles []string `json:"updatedFiles,omitempty"`
	RemovedFiles []string `json:"removedFiles,omitempty"`
	DurationMS   int64    `json:"durationMs"`
	Error        string   `json:"error,omitempty"`
}

type RenderActivity struct {
	GeneratedAt  time.Time `json:"generatedAt"`
	URLPath      string    `json:"urlPath"`
	Trigger      string    `json:"trigger"`
	Status       string    `json:"status"`
	DurationMS   int64     `json:"durationMs"`
	Deps         []string  `json:"deps,omitempty"`
	UpdatedFiles []string  `json:"updatedFiles,omitempty"`
	RemovedFiles []string  `json:"removedFiles,omitempty"`
	Error        string    `json:"error,omitempty"`
}

func NewService(contentRepo content.Repository, baseURL string, redisClient *redis.Client, redisPrefix string) *Service {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Service{
		contentRepo: contentRepo,
		baseURL:     strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		redis:       redisClient,
		redisPrefix: redisPrefix,
		metrics: MetricsSnapshot{
			lastDurationSamples: make([]int64, 0, 256),
		},
		activities: make([]RenderActivity, 0, defaultRecentActivityLimit),
		ogSem:      make(chan struct{}, 3),
	}
}

func (s *Service) RefreshPostsHTML(ctx context.Context) error {
	start := time.Now()
	renderedFiles := int64(0)
	var runErr error
	defer func() {
		durationMS := time.Since(start).Milliseconds()
		s.recordMetrics(durationMS, renderedFiles, runErr)
	}()

	paths := make([]string, 0, 256)
	paths = append(paths, "/", "/posts", "/posts/page/1")

	page := 1
	var totalArticles int64
	for {
		articles, total, err := s.contentRepo.ListPublicArticles(ctx, content.ArticleListOptions{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			runErr = err
			return err
		}
		totalArticles = total

		for _, article := range articles {
			shortURL := strings.TrimSpace(article.ShortURL)
			if shortURL == "" {
				continue
			}
			if strings.Contains(shortURL, "/") || strings.Contains(shortURL, "\\") {
				continue
			}
			paths = append(paths, fmt.Sprintf("/posts/%s", shortURL))
		}

		if len(articles) == 0 || int64(page*pageSize) >= total {
			break
		}
		page++
	}

	totalPages := int64(1)
	if listPageSize > 0 && totalArticles > 0 {
		totalPages = (totalArticles + listPageSize - 1) / listPageSize
	}
	for page := int64(2); page <= totalPages; page++ {
		paths = append(paths, fmt.Sprintf("/posts/page/%d", page))
	}

	for _, routePath := range uniquePaths(paths) {
		detail, err := s.renderURLDetailed(WithRenderTrigger(ctx, "snapshot:refresh"), routePath, false)
		if err != nil {
			runErr = err
			return fmt.Errorf("render %s: %w", routePath, err)
		}
		renderedFiles += int64(len(detail.UpdatedFiles))
	}

	log.Printf("[html-snapshot] done files=%d duration=%s", renderedFiles, time.Since(start))
	return nil
}

func (s *Service) RenderURL(ctx context.Context, rawURLPath string) (int64, error) {
	detail, err := s.RenderURLDetailed(ctx, rawURLPath)
	return int64(len(detail.UpdatedFiles)), err
}

func (s *Service) RenderURLDetailed(ctx context.Context, rawURLPath string) (RenderDetail, error) {
	return s.renderURLDetailed(ctx, rawURLPath, true)
}

func (s *Service) renderURLDetailed(ctx context.Context, rawURLPath string, trackMetrics bool) (RenderDetail, error) {
	start := time.Now()
	detail := RenderDetail{
		URLPath: strings.TrimSpace(rawURLPath),
		Trigger: renderTriggerFromContext(ctx),
		Status:  "success",
	}

	finalize := func(err error) (RenderDetail, error) {
		detail.DurationMS = time.Since(start).Milliseconds()
		if err != nil {
			detail.Status = "error"
			detail.Error = err.Error()
		}
		s.recordActivity(detail)
		if trackMetrics {
			s.recordMetrics(detail.DurationMS, int64(len(detail.UpdatedFiles)), err)
		}
		return detail, err
	}

	urlPath, err := NormalizeURLPath(rawURLPath)
	if err != nil {
		return finalize(err)
	}
	detail.URLPath = urlPath

	htmlPath, dataPath := resolveOutputPaths(urlPath)
	htmlURL := s.pageURL(urlPath)
	htmlBody, deps, err := s.fetchRequiredWithDeps(ctx, htmlURL)
	if err != nil {
		if !errors.Is(err, ErrRemoteNotFound) {
			return finalize(err)
		}

		detail.Status = "not_found"
		removed, rmErr := removeIfExists(htmlPath)
		if rmErr != nil {
			return finalize(rmErr)
		}
		if removed {
			detail.RemovedFiles = append(detail.RemovedFiles, filepath.ToSlash(htmlPath))
		}

		removed, rmErr = removeIfExists(dataPath)
		if rmErr != nil {
			return finalize(rmErr)
		}
		if removed {
			detail.RemovedFiles = append(detail.RemovedFiles, filepath.ToSlash(dataPath))
		}

		ogImageFilePath := resolveOgImagePath(urlPath)
		removed, rmErr = removeIfExists(ogImageFilePath)
		if rmErr != nil {
			return finalize(rmErr)
		}
		if removed {
			detail.RemovedFiles = append(detail.RemovedFiles, filepath.ToSlash(ogImageFilePath))
		}

		if depErr := s.syncURLDependencies(ctx, urlPath, nil); depErr != nil {
			log.Printf("[html-snapshot] clear deps failed url=%s err=%v", urlPath, depErr)
		}
		return finalize(nil)
	}
	detail.Deps = append([]string(nil), deps...)

	if err := writeFileAtomically(htmlPath, htmlBody); err != nil {
		return finalize(err)
	}
	detail.UpdatedFiles = append(detail.UpdatedFiles, filepath.ToSlash(htmlPath))

	if err := s.syncURLDependencies(ctx, urlPath, deps); err != nil {
		log.Printf("[html-snapshot] sync deps failed url=%s err=%v", urlPath, err)
	}

	// Background goroutine: fetch and cache OG image from the SvelteKit fallback route.
	ogImageFilePath := resolveOgImagePath(urlPath)
	go func() {
		s.ogSem <- struct{}{}
		defer func() { <-s.ogSem }()

		ogCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		ogBody, hasOg, fetchErr := s.fetchOptional(ogCtx, s.ogImageURL(urlPath))
		if fetchErr != nil {
			log.Printf("[html-snapshot] og-image fetch failed url=%s err=%v", urlPath, fetchErr)
			return
		}
		if !hasOg {
			// Page has a content image (cover), no generated OG image needed — clean up stale cache.
			_, _ = removeIfExists(ogImageFilePath)
			return
		}
		if writeErr := writeFileAtomically(ogImageFilePath, ogBody); writeErr != nil {
			log.Printf("[html-snapshot] og-image write failed url=%s err=%v", urlPath, writeErr)
			return
		}
		log.Printf("[html-snapshot] og-image cached url=%s", urlPath)
	}()

	dataURL := s.dataURL(urlPath)
	dataBody, hasData, err := s.fetchOptional(ctx, dataURL)
	if err != nil {
		return finalize(err)
	}
	if !hasData {
		removed, rmErr := removeIfExists(dataPath)
		if rmErr != nil {
			return finalize(rmErr)
		}
		if removed {
			detail.RemovedFiles = append(detail.RemovedFiles, filepath.ToSlash(dataPath))
		}
		return finalize(nil)
	}
	if err := writeFileAtomically(dataPath, dataBody); err != nil {
		return finalize(err)
	}
	detail.UpdatedFiles = append(detail.UpdatedFiles, filepath.ToSlash(dataPath))

	return finalize(nil)
}

// RenderErrorPage fetches a non-existent path from the renderer to capture
// its rendered 404 error page, then saves the HTML to storage/html/404.html.
// This allows nginx to serve the same styled page when the renderer is down.
func (s *Service) RenderErrorPage(ctx context.Context) error {
	errorPageURL := s.baseURL + "/___error_page_render___"

	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, errorPageURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch error page: %w", err)
	}
	defer resp.Body.Close()

	// SvelteKit renders a styled 404 page with status 404 — that's exactly what we want.
	if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status %d (expected 404)", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return fmt.Errorf("empty response body")
	}

	outPath := filepath.Join("storage", "html", "404.html")
	if err := writeFileAtomically(outPath, body); err != nil {
		return fmt.Errorf("write 404.html: %w", err)
	}
	log.Printf("[html-snapshot] error page cached path=%s", outPath)
	return nil
}

func NormalizeURLPath(rawURLPath string) (string, error) {
	candidate := strings.TrimSpace(rawURLPath)
	if candidate == "" {
		return "/", nil
	}

	if strings.HasPrefix(candidate, "http://") || strings.HasPrefix(candidate, "https://") {
		parsed, err := url.Parse(candidate)
		if err != nil {
			return "", err
		}
		candidate = parsed.EscapedPath()
	}

	if idx := strings.IndexAny(candidate, "?#"); idx >= 0 {
		candidate = candidate[:idx]
	}
	if !strings.HasPrefix(candidate, "/") {
		candidate = "/" + candidate
	}

	cleaned := path.Clean(candidate)
	if cleaned == "." {
		return "/", nil
	}
	return cleaned, nil
}

func (s *Service) MetricsSnapshot() MetricsSnapshot {
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()
	out := s.metrics
	out.lastDurationSamples = nil
	return out
}

func (s *Service) RecentActivities(limit int) []RenderActivity {
	if limit <= 0 {
		limit = 20
	}
	s.activityMu.Lock()
	defer s.activityMu.Unlock()
	if len(s.activities) == 0 {
		return nil
	}

	start := 0
	if len(s.activities) > limit {
		start = len(s.activities) - limit
	}
	out := make([]RenderActivity, 0, len(s.activities)-start)
	for idx := len(s.activities) - 1; idx >= start; idx-- {
		item := s.activities[idx]
		item.Deps = append([]string(nil), item.Deps...)
		item.UpdatedFiles = append([]string(nil), item.UpdatedFiles...)
		item.RemovedFiles = append([]string(nil), item.RemovedFiles...)
		out = append(out, item)
	}
	return out
}

func WithRenderTrigger(ctx context.Context, trigger string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	trigger = strings.TrimSpace(trigger)
	if trigger == "" {
		return ctx
	}
	return context.WithValue(ctx, renderTriggerContextKey, trigger)
}

func renderTriggerFromContext(ctx context.Context) string {
	if ctx == nil {
		return "manual"
	}
	trigger, _ := ctx.Value(renderTriggerContextKey).(string)
	trigger = strings.TrimSpace(trigger)
	if trigger == "" {
		return "manual"
	}
	return trigger
}

func (s *Service) pageURL(urlPath string) string {
	if urlPath == "/" {
		return s.baseURL + "/"
	}
	return s.baseURL + urlPath
}

func (s *Service) dataURL(urlPath string) string {
	if urlPath == "/" {
		return s.baseURL + "/__data.json"
	}
	return s.baseURL + urlPath + "/__data.json"
}

func resolveOgImagePath(urlPath string) string {
	root := filepath.Join("storage", "html")
	if urlPath == "/" {
		return filepath.Join(root, "og-image.png")
	}
	relative := filepath.FromSlash(strings.TrimPrefix(urlPath, "/"))
	return filepath.Join(root, relative, "og-image.png")
}

func (s *Service) ogImageURL(urlPath string) string {
	if urlPath == "/" {
		return s.baseURL + "/og-image.png"
	}
	return s.baseURL + urlPath + "/og-image.png"
}

func resolveOutputPaths(urlPath string) (htmlPath string, dataPath string) {
	root := filepath.Join("storage", "html")
	if urlPath == "/" {
		return filepath.Join(root, "index.html"), filepath.Join(root, "__data.json")
	}

	relative := filepath.FromSlash(strings.TrimPrefix(urlPath, "/"))
	dir := filepath.Join(root, relative)
	return filepath.Join(dir, "index.html"), filepath.Join(dir, "__data.json")
}

func (s *Service) fetchRequiredWithDeps(ctx context.Context, pageURL string) ([]byte, []string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return nil, nil, ErrRemoteNotFound
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	deps := parseDepsHeader(resp.Header.Get(depsHeader))
	return data, deps, nil
}

func (s *Service) fetchOptional(ctx context.Context, pageURL string) ([]byte, bool, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, false, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusNoContent {
		return nil, false, nil
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, false, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, false, nil
	}
	return data, true, nil
}

func (s *Service) recordActivity(detail RenderDetail) {
	activity := RenderActivity{
		GeneratedAt:  time.Now().UTC(),
		URLPath:      detail.URLPath,
		Trigger:      detail.Trigger,
		Status:       detail.Status,
		DurationMS:   detail.DurationMS,
		Deps:         append([]string(nil), detail.Deps...),
		UpdatedFiles: append([]string(nil), detail.UpdatedFiles...),
		RemovedFiles: append([]string(nil), detail.RemovedFiles...),
		Error:        detail.Error,
	}

	s.activityMu.Lock()
	defer s.activityMu.Unlock()
	s.activities = append(s.activities, activity)
	if len(s.activities) > defaultRecentActivityLimit {
		s.activities = s.activities[len(s.activities)-defaultRecentActivityLimit:]
	}
}

func (s *Service) recordMetrics(durationMS int64, renderedFiles int64, runErr error) {
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()

	s.metrics.TotalJobs++
	s.metrics.LastDurationMS = durationMS
	s.metrics.LastRenderedFiles = renderedFiles
	s.metrics.TotalRenderedFiles += renderedFiles
	s.metrics.lastDurationSamples = append(s.metrics.lastDurationSamples, durationMS)
	if len(s.metrics.lastDurationSamples) > 256 {
		s.metrics.lastDurationSamples = s.metrics.lastDurationSamples[len(s.metrics.lastDurationSamples)-256:]
	}
	if runErr != nil {
		s.metrics.FailedJobs++
		s.metrics.LastFailureAt = time.Now().UTC()
	} else {
		s.metrics.SuccessJobs++
		s.metrics.LastSuccessAt = time.Now().UTC()
	}

	var total int64
	samples := append([]int64(nil), s.metrics.lastDurationSamples...)
	for _, sample := range samples {
		total += sample
	}
	if len(samples) > 0 {
		s.metrics.AverageDurationMS = float64(total) / float64(len(samples))
		s.metrics.P95DurationMS = percentile95(samples)
	}
}

func writeFileAtomically(filePath string, body []byte) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".snapshot-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	closed := false
	defer func() {
		if !closed {
			_ = tmp.Close()
		}
		_ = os.Remove(tmpName)
	}()

	if _, err := tmp.Write(body); err != nil {
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	closed = true
	return os.Rename(tmpName, filePath)
}

func removeIfExists(filePath string) (bool, error) {
	if err := os.Remove(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func parseDepsHeader(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	if strings.HasPrefix(raw, "[") {
		var items []string
		if err := json.Unmarshal([]byte(raw), &items); err == nil {
			return normalizeDeps(items)
		}
	}

	return normalizeDeps(strings.Split(raw, ","))
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

func (s *Service) syncURLDependencies(ctx context.Context, urlPath string, deps []string) error {
	if s.redis == nil {
		return nil
	}

	urlKey := s.urlDepsKey(urlPath)
	oldDeps, err := s.redis.SMembers(ctx, urlKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	oldSet := make(map[string]struct{}, len(oldDeps))
	for _, item := range oldDeps {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			continue
		}
		oldSet[normalized] = struct{}{}
	}

	newDeps := normalizeDeps(deps)
	newSet := make(map[string]struct{}, len(newDeps))
	for _, dep := range newDeps {
		newSet[dep] = struct{}{}
	}

	pipe := s.redis.TxPipeline()
	for dep := range oldSet {
		if _, keep := newSet[dep]; keep {
			continue
		}
		pipe.SRem(ctx, s.depURLsKey(dep), urlPath)
	}

	pipe.Del(ctx, urlKey)
	if len(newDeps) > 0 {
		members := make([]any, 0, len(newDeps))
		for _, dep := range newDeps {
			members = append(members, dep)
		}
		pipe.SAdd(ctx, urlKey, members...)
		for _, dep := range newDeps {
			pipe.SAdd(ctx, s.depURLsKey(dep), urlPath)
		}
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (s *Service) depURLsKey(dep string) string {
	return fmt.Sprintf("%sisr:dep:%s", s.redisPrefix, dep)
}

func (s *Service) urlDepsKey(urlPath string) string {
	return fmt.Sprintf("%sisr:url:%s", s.redisPrefix, url.QueryEscape(urlPath))
}

func uniquePaths(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		normalized, err := NormalizeURLPath(item)
		if err != nil {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}

func percentile95(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}
	copied := append([]int64(nil), values...)
	sort.Slice(copied, func(i, j int) bool { return copied[i] < copied[j] })
	idx := int(float64(len(copied)-1) * 0.95)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(copied) {
		idx = len(copied) - 1
	}
	return float64(copied[idx])
}

package htmlsnapshot

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
)

const (
	defaultBaseURL = "http://localhost:3000"
	pageSize       = 100
	listPageSize   = 10
)

type Service struct {
	contentRepo content.Repository
	baseURL     string
	client      *http.Client
	metricsMu   sync.Mutex
	metrics     MetricsSnapshot
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

func NewService(contentRepo content.Repository, baseURL string) *Service {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Service{
		contentRepo: contentRepo,
		baseURL:     baseURL,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		metrics: MetricsSnapshot{
			lastDurationSamples: make([]int64, 0, 256),
		},
	}
}

func (s *Service) RefreshPostsHTML(ctx context.Context) error {
	start := time.Now()
	successCount := int64(0)
	var runErr error
	s.metricsMu.Lock()
	s.metrics.TotalJobs++
	s.metricsMu.Unlock()
	defer func() {
		durationMS := time.Since(start).Milliseconds()
		s.metricsMu.Lock()
		defer s.metricsMu.Unlock()
		s.metrics.LastDurationMS = durationMS
		s.metrics.LastRenderedFiles = successCount
		s.metrics.TotalRenderedFiles += successCount
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
	}()

	outputDir := filepath.Join("storage", "html", "posts")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		runErr = err
		return err
	}

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

			escaped := url.PathEscape(shortURL)
			pageURL := fmt.Sprintf("%s/posts/%s", s.baseURL, escaped)
			pageDir := filepath.Join(outputDir, shortURL)
			if err := os.MkdirAll(pageDir, 0o755); err != nil {
				runErr = err
				return err
			}
			dirIndexPath := filepath.Join(pageDir, "index.html")
			if err := s.fetchAndSave(ctx, pageURL, dirIndexPath); err == nil {
				successCount++
			}
			dataURL := fmt.Sprintf("%s/posts/%s/__data.json", s.baseURL, escaped)
			dataPath := filepath.Join(pageDir, "__data.json")
			if err := s.fetchAndSaveOptional(ctx, dataURL, dataPath); err == nil {
				successCount++
			}
		}

		if len(articles) == 0 || int64(page*pageSize) >= total {
			break
		}
		page++
	}

	indexURL := fmt.Sprintf("%s/posts/", s.baseURL)
	indexPath := filepath.Join(outputDir, "index.html")
	if err := s.fetchAndSave(ctx, indexURL, indexPath); err != nil {
		runErr = err
		return fmt.Errorf("fetch index: %w", err)
	}
	successCount++
	indexDataURL := fmt.Sprintf("%s/posts/__data.json", s.baseURL)
	indexDataPath := filepath.Join(outputDir, "__data.json")
	if err := s.fetchAndSaveOptional(ctx, indexDataURL, indexDataPath); err == nil {
		successCount++
	}

	totalPages := int64(1)
	if listPageSize > 0 && totalArticles > 0 {
		totalPages = (totalArticles + listPageSize - 1) / listPageSize
	}
	for page := int64(1); page <= totalPages; page++ {
		pageDir := filepath.Join(outputDir, "page", fmt.Sprintf("%d", page))
		if err := os.MkdirAll(pageDir, 0o755); err != nil {
			runErr = err
			return err
		}
		pageURL := fmt.Sprintf("%s/posts/page/%d/", s.baseURL, page)
		pageIndexPath := filepath.Join(pageDir, "index.html")
		if err := s.fetchAndSave(ctx, pageURL, pageIndexPath); err == nil {
			successCount++
		}
		pageDataURL := fmt.Sprintf("%s/posts/page/%d/__data.json", s.baseURL, page)
		pageDataPath := filepath.Join(pageDir, "__data.json")
		if err := s.fetchAndSaveOptional(ctx, pageDataURL, pageDataPath); err == nil {
			successCount++
		}
	}

	rootURL := fmt.Sprintf("%s/", s.baseURL)
	rootPath := filepath.Join("storage", "html", "index.html")
	if err := s.fetchAndSave(ctx, rootURL, rootPath); err != nil {
		runErr = err
		return fmt.Errorf("fetch root: %w", err)
	}
	successCount++
	rootDataURL := fmt.Sprintf("%s/__data.json", s.baseURL)
	rootDataPath := filepath.Join("storage", "html", "__data.json")
	if err := s.fetchAndSaveOptional(ctx, rootDataURL, rootDataPath); err == nil {
		successCount++
	}
	log.Printf("[html-snapshot] done success=%d duration=%s", successCount, time.Since(start))

	return nil
}

func (s *Service) MetricsSnapshot() MetricsSnapshot {
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()
	out := s.metrics
	out.lastDurationSamples = nil
	return out
}

func (s *Service) fetchAndSave(ctx context.Context, pageURL, filePath string) error {
	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, pageURL, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0o644)
}

func (s *Service) fetchAndSaveOptional(ctx context.Context, pageURL, filePath string) error {
	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, pageURL, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusNoContent {
		_ = os.Remove(filePath)
		return nil
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		_ = os.Remove(filePath)
		return nil
	}

	return os.WriteFile(filePath, data, 0o644)
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

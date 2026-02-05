package search

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domainsearch "github.com/grtsinry43/grtblog-v2/server/internal/domain/search"
	"github.com/redis/go-redis/v9"
)

const (
	defaultLimitPerGroup = 8
	maxLimitPerGroup     = 20
	cacheTTL             = 5 * time.Minute
)

type Service struct {
	repo   domainsearch.Repository
	redis  *redis.Client
	prefix string
}

type Result struct {
	Query     string             `json:"query"`
	Keywords  []string           `json:"keywords"`
	Articles  []domainsearch.Hit `json:"articles"`
	Moments   []domainsearch.Hit `json:"moments"`
	Pages     []domainsearch.Hit `json:"pages"`
	Thinkings []domainsearch.Hit `json:"thinkings"`
	Cached    bool               `json:"cached"`
}

type cachedResult struct {
	Query     string             `json:"query"`
	Keywords  []string           `json:"keywords"`
	Articles  []domainsearch.Hit `json:"articles"`
	Moments   []domainsearch.Hit `json:"moments"`
	Pages     []domainsearch.Hit `json:"pages"`
	Thinkings []domainsearch.Hit `json:"thinkings"`
}

func NewService(repo domainsearch.Repository, redisClient *redis.Client, redisPrefix string) *Service {
	return &Service{
		repo:   repo,
		redis:  redisClient,
		prefix: redisPrefix,
	}
}

func (s *Service) SearchSite(ctx context.Context, query string, limitPerGroup int) (Result, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return Result{}, ErrEmptyQuery
	}
	limitPerGroup = normalizeLimit(limitPerGroup)

	cacheKey := s.cacheKey(query, limitPerGroup)
	if cached, ok := s.getCached(ctx, cacheKey); ok {
		cached.Cached = true
		return cached, nil
	}

	hits, err := s.repo.SearchSite(ctx, query, limitPerGroup)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Query:    query,
		Keywords: extractKeywords(query),
	}
	for _, hit := range hits {
		switch hit.Kind {
		case domainsearch.KindArticle:
			result.Articles = append(result.Articles, hit)
		case domainsearch.KindMoment:
			result.Moments = append(result.Moments, hit)
		case domainsearch.KindPage:
			result.Pages = append(result.Pages, hit)
		case domainsearch.KindThinking:
			result.Thinkings = append(result.Thinkings, hit)
		}
	}
	s.cacheResult(ctx, cacheKey, result)
	return result, nil
}

func (s *Service) getCached(ctx context.Context, cacheKey string) (Result, bool) {
	if s.redis == nil {
		return Result{}, false
	}
	raw, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err != nil || len(raw) == 0 {
		return Result{}, false
	}
	var cached cachedResult
	if err := json.Unmarshal(raw, &cached); err != nil {
		return Result{}, false
	}
	return Result{
		Query:     cached.Query,
		Keywords:  cached.Keywords,
		Articles:  cached.Articles,
		Moments:   cached.Moments,
		Pages:     cached.Pages,
		Thinkings: cached.Thinkings,
	}, true
}

func (s *Service) cacheResult(ctx context.Context, cacheKey string, result Result) {
	if s.redis == nil {
		return
	}
	payload := cachedResult{
		Query:     result.Query,
		Keywords:  result.Keywords,
		Articles:  result.Articles,
		Moments:   result.Moments,
		Pages:     result.Pages,
		Thinkings: result.Thinkings,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	cacheCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	_ = s.redis.Set(cacheCtx, cacheKey, body, cacheTTL).Err()
}

func (s *Service) cacheKey(query string, limitPerGroup int) string {
	h := sha1.Sum([]byte(strings.ToLower(strings.TrimSpace(query))))
	return fmt.Sprintf("%ssearch:site:%s:%d", s.prefix, hex.EncodeToString(h[:]), limitPerGroup)
}

func normalizeLimit(limitPerGroup int) int {
	if limitPerGroup <= 0 {
		return defaultLimitPerGroup
	}
	if limitPerGroup > maxLimitPerGroup {
		return maxLimitPerGroup
	}
	return limitPerGroup
}

func extractKeywords(query string) []string {
	parts := strings.Fields(strings.TrimSpace(query))
	if len(parts) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(parts))
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key := strings.ToLower(part)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, part)
		if len(result) >= 8 {
			break
		}
	}
	return result
}

package friendtimeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domainfed "github.com/grtsinry43/grtblog-v2/server/internal/domain/federation"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo   domainfed.FederatedPostCacheRepository
	redis  *redis.Client
	prefix string
}

type ListResult struct {
	Items []domainfed.FederatedPostCache `json:"items"`
	Total int64                          `json:"total"`
	Page  int                            `json:"page"`
	Size  int                            `json:"size"`
}

func NewService(repo domainfed.FederatedPostCacheRepository, redisClient *redis.Client, redisPrefix string) *Service {
	return &Service{
		repo:   repo,
		redis:  redisClient,
		prefix: redisPrefix,
	}
}

func (s *Service) List(ctx context.Context, page, pageSize int) (ListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	cacheKey := s.cacheKey(page, pageSize)
	if s.redis != nil {
		if raw, err := s.redis.Get(ctx, cacheKey).Result(); err == nil && raw != "" {
			var cached ListResult
			if err := json.Unmarshal([]byte(raw), &cached); err == nil {
				return cached, nil
			}
		}
	}
	items, total, err := s.repo.ListTimeline(ctx, page, pageSize)
	if err != nil {
		return ListResult{}, err
	}
	result := ListResult{
		Items: items,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}
	if s.redis != nil {
		if encoded, err := json.Marshal(result); err == nil {
			_ = s.redis.Set(ctx, cacheKey, encoded, 60*time.Second).Err()
		}
	}
	return result, nil
}

func (s *Service) cacheKey(page, pageSize int) string {
	return fmt.Sprintf("%sfriend_timeline:%d:%d", s.prefix, page, pageSize)
}

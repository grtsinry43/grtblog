package hitokoto

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultBaseURL = "https://v1.hitokoto.cn"
	cacheTTL       = 60 * time.Second
)

type Service struct {
	client    *http.Client
	redis     *redis.Client
	keyPrefix string
	baseURL   string
}

type Sentence struct {
	ID         int64  `json:"id"`
	Hitokoto   string `json:"hitokoto"`
	Type       string `json:"type"`
	From       string `json:"from"`
	FromWho    string `json:"from_who"`
	Creator    string `json:"creator"`
	CreatorUID int64  `json:"creator_uid"`
	Reviewer   int64  `json:"reviewer"`
	UUID       string `json:"uuid"`
	CreatedAt  string `json:"created_at"`
	Length     int    `json:"length"`
}

type Query struct {
	Categories []string
	MinLength  string
	MaxLength  string
	Charset    string
}

type Result struct {
	Sentence Sentence `json:"sentence"`
	Cached   bool     `json:"cached"`
}

func NewService(redisClient *redis.Client, redisPrefix string) *Service {
	return &Service{
		client:    &http.Client{Timeout: 4 * time.Second},
		redis:     redisClient,
		keyPrefix: fmt.Sprintf("%sadmin:hitokoto:", redisPrefix),
		baseURL:   defaultBaseURL,
	}
}

func (s *Service) GetSentence(ctx context.Context, query Query) (*Result, error) {
	cacheKey := s.cacheKey(query)
	if s.redis != nil {
		if raw, err := s.redis.Get(ctx, cacheKey).Bytes(); err == nil && len(raw) > 0 {
			var sentence Sentence
			if err := json.Unmarshal(raw, &sentence); err == nil {
				return &Result{Sentence: sentence, Cached: true}, nil
			}
		}
	}

	sentence, err := s.fetchSentence(ctx, query)
	if err != nil {
		return nil, err
	}

	if s.redis != nil {
		if payload, err := json.Marshal(sentence); err == nil {
			_ = s.redis.Set(ctx, cacheKey, payload, cacheTTL).Err()
		}
	}
	return &Result{Sentence: *sentence, Cached: false}, nil
}

func (s *Service) fetchSentence(ctx context.Context, query Query) (*Sentence, error) {
	endpoint := s.buildURL(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "grtblog-admin/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("hitokoto upstream status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var sentence Sentence
	if err := json.NewDecoder(resp.Body).Decode(&sentence); err != nil {
		return nil, err
	}
	return &sentence, nil
}

func (s *Service) buildURL(query Query) string {
	values := url.Values{}
	values.Set("encode", "json")
	for _, c := range normalizeCategories(query.Categories) {
		values.Add("c", c)
	}
	if query.MinLength != "" {
		values.Set("min_length", strings.TrimSpace(query.MinLength))
	}
	if query.MaxLength != "" {
		values.Set("max_length", strings.TrimSpace(query.MaxLength))
	}
	if query.Charset != "" {
		values.Set("charset", strings.TrimSpace(query.Charset))
	}
	return s.baseURL + "?" + values.Encode()
}

func (s *Service) cacheKey(query Query) string {
	base := s.buildURL(query)
	h := sha1.Sum([]byte(base))
	return s.keyPrefix + hex.EncodeToString(h[:])
}

func normalizeCategories(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, item := range in {
		c := strings.ToLower(strings.TrimSpace(item))
		if len(c) != 1 || c[0] < 'a' || c[0] > 'l' {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

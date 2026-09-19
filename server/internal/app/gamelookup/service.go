// Package gamelookup resolves a game name (reported by the owner status
// reporter) to public store metadata. It prefers the Steam store public
// endpoints and caches both positive and negative results in Redis.
package gamelookup

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	positiveCacheTTL = 24 * time.Hour
	negativeCacheTTL = time.Hour
	maxNameLength    = 64
	requestTimeout   = 6 * time.Second
	maxBodyBytes     = 1 << 20

	storeSearchURL  = "https://store.steampowered.com/api/storesearch/?term=%s&cc=us&l=english"
	appDetailsURL   = "https://store.steampowered.com/api/appdetails?appids=%d&cc=us&l=english"
	capsuleImageURL = "https://cdn.cloudflare.steamstatic.com/steam/apps/%d/library_600x900_2x.jpg"
	storePageURL    = "https://store.steampowered.com/app/%d/"
	httpUserAgent   = "grtblog-v2/gamelookup"
)

// GameInfo describes a resolved game for the owner-status card.
type GameInfo struct {
	Found            bool   `json:"found"`
	Source           string `json:"source,omitempty"`
	AppID            int64  `json:"appId,omitempty"`
	Name             string `json:"name,omitempty"`
	ShortDescription string `json:"shortDescription,omitempty"`
	HeaderImage      string `json:"headerImage,omitempty"`
	CapsuleImage     string `json:"capsuleImage,omitempty"`
	StoreURL         string `json:"storeUrl,omitempty"`
}

type Service struct {
	redis       *redis.Client
	redisPrefix string
	httpClient  *http.Client
}

func NewService(redisClient *redis.Client, redisPrefix string) *Service {
	return &Service{
		redis:       redisClient,
		redisPrefix: redisPrefix,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// Lookup resolves a human readable game name into store metadata.
// A missing/unknown game is a normal result (Found=false), not an error.
func (s *Service) Lookup(ctx context.Context, name string) (*GameInfo, error) {
	normalized := normalizeName(name)
	if normalized == "" {
		return &GameInfo{Found: false}, nil
	}

	if cached, err := s.getCache(ctx, normalized); err == nil && cached != nil {
		return cached, nil
	}

	info, err := s.lookupSteam(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if info == nil {
		info = &GameInfo{Found: false}
	}

	_ = s.setCache(ctx, normalized, info)
	return info, nil
}

func (s *Service) lookupSteam(ctx context.Context, name string) (*GameInfo, error) {
	appID, err := s.searchAppID(ctx, name)
	if err != nil {
		return nil, err
	}
	if appID <= 0 {
		return nil, nil
	}

	info := &GameInfo{
		Found:        true,
		Source:       "steam",
		AppID:        appID,
		CapsuleImage: fmt.Sprintf(capsuleImageURL, appID),
		StoreURL:     fmt.Sprintf(storePageURL, appID),
	}

	var details struct {
		Name             string `json:"name"`
		ShortDescription string `json:"short_description"`
		HeaderImage      string `json:"header_image"`
	}
	if err := s.getAppDetails(ctx, appID, &details); err == nil {
		info.Name = strings.TrimSpace(details.Name)
		info.ShortDescription = strings.TrimSpace(details.ShortDescription)
		info.HeaderImage = strings.TrimSpace(details.HeaderImage)
	}
	if info.Name == "" {
		info.Name = name
	}
	return info, nil
}

func (s *Service) searchAppID(ctx context.Context, name string) (int64, error) {
	searchURL := fmt.Sprintf(storeSearchURL, url.QueryEscape(name))
	var result struct {
		Items []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := s.getJSON(ctx, searchURL, &result); err != nil {
		return 0, err
	}
	if len(result.Items) == 0 {
		return 0, nil
	}
	return result.Items[0].ID, nil
}

func (s *Service) getAppDetails(ctx context.Context, appID int64, target *struct {
	Name             string `json:"name"`
	ShortDescription string `json:"short_description"`
	HeaderImage      string `json:"header_image"`
}) error {
	detailsURL := fmt.Sprintf(appDetailsURL, appID)
	var envelope map[string]struct {
		Success bool             `json:"success"`
		Data    *json.RawMessage `json:"data"`
	}
	if err := s.getJSON(ctx, detailsURL, &envelope); err != nil {
		return err
	}
	entry, ok := envelope[strconv.FormatInt(appID, 10)]
	if !ok || !entry.Success || entry.Data == nil {
		return fmt.Errorf("appdetails unavailable for appid=%d", appID)
	}
	return json.Unmarshal(*entry.Data, target)
}

func (s *Service) getJSON(ctx context.Context, rawURL string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", httpUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upstream error: status=%d", resp.StatusCode)
	}

	decoder := json.NewDecoder(io.LimitReader(resp.Body, maxBodyBytes))
	return decoder.Decode(target)
}

func (s *Service) cacheKey(normalizedName string) string {
	sum := sha1.Sum([]byte(normalizedName))
	return fmt.Sprintf("%sgamelookup:%s", s.redisPrefix, hex.EncodeToString(sum[:]))
}

func (s *Service) getCache(ctx context.Context, normalizedName string) (*GameInfo, error) {
	if s.redis == nil {
		return nil, nil
	}
	val, err := s.redis.Get(ctx, s.cacheKey(normalizedName)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var info GameInfo
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (s *Service) setCache(ctx context.Context, normalizedName string, info *GameInfo) error {
	if s.redis == nil || info == nil {
		return nil
	}
	payload, err := json.Marshal(info)
	if err != nil {
		return err
	}
	ttl := positiveCacheTTL
	if !info.Found {
		ttl = negativeCacheTTL
	}
	return s.redis.Set(ctx, s.cacheKey(normalizedName), payload, ttl).Err()
}

func normalizeName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if len(name) > maxNameLength {
		name = name[:maxNameLength]
	}
	return name
}

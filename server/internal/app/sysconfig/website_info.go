package sysconfig

import (
	"context"
	"encoding/json"
	"strings"

	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
)

const siteKeyPrefix = "site."

// WebsiteInfo returns all site.* config keys as a flat map with the prefix stripped.
// e.g., "site.website_name" → "website_name": "My Blog"
func (s *Service) WebsiteInfo(ctx context.Context) (map[string]string, error) {
	items, err := s.repo.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, item := range items {
		if strings.HasPrefix(item.Key, siteKeyPrefix) {
			bareKey := strings.TrimPrefix(item.Key, siteKeyPrefix)
			result[bareKey] = item.Value
		}
	}
	return result, nil
}

// GetWebsiteInfoValue returns the value of a single website info key.
// The key should be the bare name (e.g., "website_name"), the site. prefix is added internally.
func (s *Service) GetWebsiteInfoValue(ctx context.Context, key string) (string, error) {
	cfg, err := s.repo.GetByKey(ctx, siteKeyPrefix+strings.TrimSpace(key))
	if err != nil {
		return "", err
	}
	return cfg.Value, nil
}

// ThemeExtendInfo returns the parsed JSON for the theme_extend_info config.
func (s *Service) ThemeExtendInfo(ctx context.Context) (json.RawMessage, error) {
	cfg, err := s.repo.GetByKey(ctx, siteKeyPrefix+"theme_extend_info")
	if err != nil {
		if err == domainconfig.ErrSysConfigNotFound {
			return json.RawMessage("{}"), nil
		}
		return nil, err
	}
	val := strings.TrimSpace(cfg.Value)
	if val == "" {
		return json.RawMessage("{}"), nil
	}
	return json.RawMessage(val), nil
}

// UpdateWebsiteInfoByKey is a convenience method for updating a single website info key.
// Used by the init page and the backward-compatible PUT /website-info/:key endpoint.
func (s *Service) UpdateWebsiteInfoByKey(ctx context.Context, key string, value string) error {
	fullKey := siteKeyPrefix + strings.TrimSpace(key)
	valRaw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	rawMsg := json.RawMessage(valRaw)
	_, err = s.UpdateConfigs(ctx, []UpdateItem{
		{Key: fullKey, Value: &rawMsg},
	})
	return err
}

// UpdateWebsiteInfoJSON is a convenience method for updating a JSON website info key (e.g., theme_extend_info).
func (s *Service) UpdateWebsiteInfoJSON(ctx context.Context, key string, jsonVal json.RawMessage) error {
	fullKey := siteKeyPrefix + strings.TrimSpace(key)
	_, err := s.UpdateConfigs(ctx, []UpdateItem{
		{Key: fullKey, Value: &jsonVal},
	})
	return err
}

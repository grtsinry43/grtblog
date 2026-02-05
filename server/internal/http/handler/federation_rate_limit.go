package handler

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
	fedinfra "github.com/grtsinry43/grtblog-v2/server/internal/infra/federation"
)

func enforceFederationInboundRateLimit(ctx context.Context, limiter fedinfra.RateLimiter, sourceBaseURL, action string, raw json.RawMessage) error {
	if limiter == nil {
		return nil
	}
	limit := parseRateLimit(raw, action)
	if limit <= 0 {
		return nil
	}
	key := strings.TrimSpace(action) + ":" + strings.TrimSpace(sourceBaseURL)
	ok, err := limiter.Allow(ctx, key, limit, time.Minute)
	if err != nil {
		return err
	}
	if !ok {
		return response.NewBizErrorWithMsg(response.TooManyRequests, "请求频率过高")
	}
	return nil
}

func parseRateLimit(raw json.RawMessage, action string) int {
	if len(raw) == 0 {
		return defaultActionLimit(action)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return defaultActionLimit(action)
	}
	key := action + "_per_minute"
	if val, ok := payload[key]; ok {
		return toInt(val, defaultActionLimit(action))
	}
	if val, ok := payload[action]; ok {
		return toInt(val, defaultActionLimit(action))
	}
	return defaultActionLimit(action)
}

func defaultActionLimit(action string) int {
	switch action {
	case "friendlink":
		return 30
	case "citation":
		return 60
	case "mention":
		return 120
	default:
		return 60
	}
}

func toInt(val any, fallback int) int {
	switch t := val.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	default:
		return fallback
	}
}

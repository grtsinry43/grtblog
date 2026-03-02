package handler

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/health"
	"github.com/grtsinry43/grtblog-v2/server/internal/buildinfo"
	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

// HealthHandler exposes lightweight probe endpoints for uptime monitoring.
type HealthHandler struct {
	cfg         config.AppConfig
	db          *gorm.DB
	redisClient *redis.Client
	version     string
	healthState *health.State
}

type ComponentHealthProbe struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Healthy bool   `json:"healthy"`
	Version string `json:"version,omitempty"`
}

func NewHealthHandler(cfg config.AppConfig, db *gorm.DB, redisClient *redis.Client, healthState *health.State) *HealthHandler {
	return &HealthHandler{
		cfg:         cfg,
		db:          db,
		redisClient: redisClient,
		version:     buildinfo.Version(),
		healthState: healthState,
	}
}

func (h *HealthHandler) Liveness(c *fiber.Ctx) error {
	isDev := strings.EqualFold(strings.TrimSpace(h.cfg.Env), "development")
	if !isDev {
		data := struct {
			Status string    `json:"status"`
			Time   time.Time `json:"time"`
		}{
			Status: "alive",
			Time:   time.Now().UTC(),
		}
		return response.Success(c, data)
	}

	data := struct {
		Status     string                 `json:"status"`
		App        string                 `json:"app"`
		Env        string                 `json:"env"`
		Version    string                 `json:"version"`
		Time       time.Time              `json:"time"`
		Components []ComponentHealthProbe `json:"components"`
	}{
		Status:  "alive",
		App:     h.cfg.Name,
		Env:     h.cfg.Env,
		Version: h.version,
		Time:    time.Now().UTC(),
		Components: []ComponentHealthProbe{
			{
				Name:    "api",
				Status:  "alive",
				Healthy: true,
				Version: h.version,
			},
		},
	}

	return response.Success(c, data)
}

func (h *HealthHandler) Readiness(c *fiber.Ctx) error {
	dbStatus, dbVersion := h.probeDatabase(c.UserContext())
	redisStatus, redisVersion := h.probeRedis(c.UserContext())
	apiHealthy := true
	dbHealthy := dbStatus == "connected"
	redisHealthy := redisStatus == "connected" || redisStatus == "not_configured"

	globalStatus := "ready"
	if !apiHealthy || !dbHealthy || !redisHealthy {
		globalStatus = "degraded"
	}

	components := []ComponentHealthProbe{
		{
			Name:    "api",
			Status:  "ready",
			Healthy: true,
			Version: h.version,
		},
		{
			Name:    "database",
			Status:  dbStatus,
			Healthy: dbHealthy,
			Version: dbVersion,
		},
		{
			Name:    "redis",
			Status:  redisStatus,
			Healthy: redisHealthy,
			Version: redisVersion,
		},
	}

	// Derive health state fields from the state machine (if available).
	var maintenance bool
	var healthBits uint8
	var healthMode string
	var stateIsDev bool
	if h.healthState != nil {
		snap := h.healthState.Snapshot()
		maintenance = snap.Maintenance
		healthBits = snap.HealthBits
		healthMode = string(snap.Mode)
		stateIsDev = snap.IsDev
	} else {
		healthMode = globalStatus
		stateIsDev = h.cfg.Env == "development"
	}

	isDev := strings.EqualFold(strings.TrimSpace(h.cfg.Env), "development")
	if !isDev {
		data := struct {
			Status      string    `json:"status"`
			Time        time.Time `json:"time"`
			Maintenance bool      `json:"maintenance"`
			HealthBits  uint8     `json:"healthBits"`
			HealthMode  string    `json:"healthMode"`
			IsDev       bool      `json:"isDev"`
			Components  []struct {
				Name    string `json:"name"`
				Status  string `json:"status"`
				Healthy bool   `json:"healthy"`
			} `json:"components"`
		}{
			Status:      globalStatus,
			Time:        time.Now().UTC(),
			Maintenance: maintenance,
			HealthBits:  healthBits,
			HealthMode:  healthMode,
			IsDev:       stateIsDev,
			Components: []struct {
				Name    string `json:"name"`
				Status  string `json:"status"`
				Healthy bool   `json:"healthy"`
			}{
				{Name: "api", Status: "ready", Healthy: true},
				{Name: "database", Status: dbStatus, Healthy: dbHealthy},
				{Name: "redis", Status: redisStatus, Healthy: redisHealthy},
			},
		}
		return response.SuccessWithMessage(c, data, globalStatus)
	}

	data := struct {
		Status      string                 `json:"status"`
		App         string                 `json:"app"`
		Env         string                 `json:"env"`
		Version     string                 `json:"version"`
		Time        time.Time              `json:"time"`
		Components  []ComponentHealthProbe `json:"components"`
		Maintenance bool                   `json:"maintenance"`
		HealthBits  uint8                  `json:"healthBits"`
		HealthMode  string                 `json:"healthMode"`
		IsDev       bool                   `json:"isDev"`
	}{
		Status:      globalStatus,
		App:         h.cfg.Name,
		Env:         h.cfg.Env,
		Version:     h.version,
		Time:        time.Now().UTC(),
		Components:  components,
		Maintenance: maintenance,
		HealthBits:  healthBits,
		HealthMode:  healthMode,
		IsDev:       stateIsDev,
	}

	return response.SuccessWithMessage(c, data, globalStatus)
}

func (h *HealthHandler) probeDatabase(ctx context.Context) (status string, version string) {
	if h.db == nil {
		return "not_configured", ""
	}
	sqlDB, err := h.db.DB()
	if err != nil {
		return "error", ""
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return "disconnected", ""
	}

	var dbVersion string
	queryCtx, queryCancel := context.WithTimeout(ctx, 2*time.Second)
	defer queryCancel()
	if err := h.db.WithContext(queryCtx).Raw("SELECT version()").Scan(&dbVersion).Error; err == nil {
		return "connected", strings.TrimSpace(dbVersion)
	}
	return "connected", ""
}

func (h *HealthHandler) probeRedis(ctx context.Context) (status string, version string) {
	if h.redisClient == nil {
		return "not_configured", ""
	}
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := h.redisClient.Ping(pingCtx).Err(); err != nil {
		return "error", ""
	}

	infoCtx, infoCancel := context.WithTimeout(ctx, 2*time.Second)
	defer infoCancel()
	info, err := h.redisClient.Info(infoCtx, "server").Result()
	if err != nil {
		return "connected", ""
	}
	for _, line := range strings.Split(info, "\r\n") {
		if strings.HasPrefix(line, "redis_version:") {
			return "connected", strings.TrimSpace(strings.TrimPrefix(line, "redis_version:"))
		}
	}
	return "connected", ""
}

package handler

import (
	"context"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sys/unix"
	"gorm.io/gorm"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/health"
	"github.com/grtsinry43/grtblog-v2/server/internal/buildinfo"
	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type SystemHandler struct {
	appCfg      config.AppConfig
	db          *gorm.DB
	redisClient *redis.Client
	events      appEvent.Bus
	healthState *health.State

	// 静态/缓存数据，减少分配
	version           string
	commit            string
	storageSize       uint64
	lastStorageUpdate time.Time
	mu                sync.RWMutex

	updateMu        sync.RWMutex
	lastUpdateCheck time.Time
	updateCache     SystemUpdateInfo
}

func NewSystemHandler(appCfg config.AppConfig, db *gorm.DB, redisClient *redis.Client, events appEvent.Bus, healthState *health.State) *SystemHandler {
	if events == nil {
		events = appEvent.NopBus{}
	}
	h := &SystemHandler{
		appCfg:      appCfg,
		db:          db,
		redisClient: redisClient,
		version:     buildinfo.Version(),
		commit:      buildinfo.Commit(),
		events:      events,
		healthState: healthState,
	}
	return h
}

type SystemStatus struct {
	App          AppInfo           `json:"app"`
	CPU          CPUInfo           `json:"cpu"`
	Memory       MemoryInfo        `json:"memory"`
	Disk         DiskInfo          `json:"disk"`
	Storage      StorageInfo       `json:"storage"`
	Database     DatabaseStatus    `json:"database"`
	Redis        RedisStatus       `json:"redis"`
	Platform     PlatformInfo      `json:"platform"`
	Components   []ComponentHealth `json:"components"`
	Update       SystemUpdateInfo  `json:"update"`
	HealthState  uint8             `json:"healthState"`
	HealthMode   string            `json:"healthMode"`
	Maintenance  bool              `json:"maintenance"`
}

type AppInfo struct {
	Version   string    `json:"version"`
	Commit    string    `json:"commit,omitempty"`
	GoVersion string    `json:"goVersion"`
	StartTime time.Time `json:"startTime"`
	Uptime    string    `json:"uptime"`
}

type CPUInfo struct {
	Cores int `json:"cores"`
}

type MemoryInfo struct {
	Alloc      uint64 `json:"alloc"`      // bytes
	TotalAlloc uint64 `json:"totalAlloc"` // bytes
	Sys        uint64 `json:"sys"`        // bytes
	NumGC      uint32 `json:"numGC"`
}

type DiskInfo struct {
	Path string `json:"path"`
	All  uint64 `json:"all"`  // bytes
	Used uint64 `json:"used"` // bytes
	Free uint64 `json:"free"` // bytes
}

type StorageInfo struct {
	Path string `json:"path"`
	Size uint64 `json:"size"` // bytes
}

type DBPoolStats struct {
	MaxOpenConnections int   `json:"maxOpenConnections"`
	OpenConnections    int   `json:"openConnections"`
	InUse              int   `json:"inUse"`
	Idle               int   `json:"idle"`
	WaitCount          int64 `json:"waitCount"`
	MaxIdleClosed      int64 `json:"maxIdleClosed"`
	MaxIdleTimeClosed  int64 `json:"maxIdleTimeClosed"`
	MaxLifetimeClosed  int64 `json:"maxLifetimeClosed"`
}

type DatabaseStatus struct {
	Status    string      `json:"status"`
	Driver    string      `json:"driver"`
	Version   string      `json:"version,omitempty"`
	PoolStats DBPoolStats `json:"poolStats"`
}

type RedisStatus struct {
	Status     string `json:"status"`
	UsedMemory string `json:"usedMemory,omitempty"`
	Version    string `json:"version,omitempty"`
}

type ComponentHealth struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Healthy   bool   `json:"healthy"`
	Version   string `json:"version,omitempty"`
	CheckedAt string `json:"checkedAt"`
}

type PlatformInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

type SystemStatusEnvelope struct {
	Code   int           `json:"code"`
	BizErr string        `json:"bizErr"`
	Msg    string        `json:"msg"`
	Data   SystemStatus  `json:"data"`
	Meta   response.Meta `json:"meta"`
}

var serverStartTime = time.Now()

// GetStatus godoc
// @Summary 获取系统运行状态
// @Description 获取包含 CPU、内存、磁盘、数据库和 Redis 在内的详细监控数据。存储占用大小每 5 分钟更新一次。
// @Tags Admin-System
// @Produce json
// @Success 200 {object} SystemStatusEnvelope
// @Security BearerAuth
// @Router /admin/system/status [get]
func (h *SystemHandler) GetStatus(c *fiber.Ctx) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	status := SystemStatus{
		App: AppInfo{
			Version:   h.version,
			Commit:    h.commit,
			GoVersion: runtime.Version(),
			StartTime: serverStartTime,
			Uptime:    time.Since(serverStartTime).Truncate(time.Second).String(),
		},
		CPU: CPUInfo{
			Cores: runtime.NumCPU(),
		},
		Memory: MemoryInfo{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			Sys:        m.Sys,
			NumGC:      m.NumGC,
		},
		Disk:     getDiskInfo("/"),
		Storage:  h.getCachedStorageInfo("storage"),
		Database: h.getDatabaseStatus(c.UserContext()),
		Redis:    h.getRedisStatus(c.UserContext()),
		Platform: PlatformInfo{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
	}
	status.Components = h.buildComponents(status)
	status.Update = h.peekCachedUpdateCheck()
	if h.healthState != nil {
		snap := h.healthState.Snapshot()
		status.HealthState = snap.HealthBits
		status.HealthMode = string(snap.Mode)
		status.Maintenance = snap.Maintenance
	}

	if status.Database.Status != "connected" || (status.Redis.Status != "connected" && status.Redis.Status != "not_configured") {
		_ = h.events.Publish(c.UserContext(), appEvent.Generic{
			EventName: "system.monitor.alert",
			At:        time.Now(),
			Payload: map[string]any{
				"DatabaseStatus": status.Database.Status,
				"RedisStatus":    status.Redis.Status,
			},
		})
	}

	return response.Success(c, status)
}

func (h *SystemHandler) GetUpdateCheck(c *fiber.Ctx) error {
	force := c.QueryBool("force", false)
	return response.Success(c, h.getCachedUpdateCheck(c.UserContext(), force))
}

func getDiskInfo(path string) DiskInfo {
	var stat unix.Statfs_t
	// 使用 Statfs 获取磁盘状态
	err := unix.Statfs(path, &stat)
	if err != nil {
		return DiskInfo{Path: path}
	}

	// 部分系统 Bsize 可能为 0，需降级使用 Frsize
	bsize := uint64(stat.Bsize)
	all := stat.Blocks * bsize
	free := stat.Bfree * bsize
	used := all - free

	return DiskInfo{
		Path: path,
		All:  all,
		Used: used,
		Free: free,
	}
}

func (h *SystemHandler) getCachedStorageInfo(path string) StorageInfo {
	h.mu.RLock()
	// 5 分钟缓存期
	if time.Since(h.lastStorageUpdate) < 5*time.Minute {
		defer h.mu.RUnlock()
		return StorageInfo{Path: path, Size: h.storageSize}
	}
	h.mu.RUnlock()

	h.mu.Lock()
	defer h.mu.Unlock()

	// 双重检查，防止并发穿透
	if time.Since(h.lastStorageUpdate) < 5*time.Minute {
		return StorageInfo{Path: path, Size: h.storageSize}
	}

	var totalSize uint64
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				totalSize += uint64(info.Size())
			}
		}
		return nil
	})

	h.storageSize = totalSize
	h.lastStorageUpdate = time.Now()

	return StorageInfo{
		Path: path,
		Size: totalSize,
	}
}

func (h *SystemHandler) getDatabaseStatus(ctx context.Context) DatabaseStatus {
	if h.db == nil {
		return DatabaseStatus{
			Status: "not_configured",
			Driver: "unknown",
		}
	}

	sqlDB, err := h.db.DB()
	if err != nil {
		return DatabaseStatus{Status: "error"}
	}

	status := "connected"
	// 设置 Ping 超时，防止数据库挂起导致 API 卡死

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		status = "disconnected"
	}

	var dbVersion string
	queryCtx, queryCancel := context.WithTimeout(ctx, 2*time.Second)
	defer queryCancel()
	if err := h.db.WithContext(queryCtx).Raw("SELECT version()").Scan(&dbVersion).Error; err != nil {
		dbVersion = ""
	}

	s := sqlDB.Stats()
	return DatabaseStatus{
		Status:  status,
		Driver:  h.db.Dialector.Name(),
		Version: strings.TrimSpace(dbVersion),
		PoolStats: DBPoolStats{
			MaxOpenConnections: s.MaxOpenConnections,
			OpenConnections:    s.OpenConnections,
			InUse:              s.InUse,
			Idle:               s.Idle,
			WaitCount:          s.WaitCount,
			MaxIdleClosed:      s.MaxIdleClosed,
			MaxIdleTimeClosed:  s.MaxIdleTimeClosed,
			MaxLifetimeClosed:  s.MaxLifetimeClosed,
		},
	}
}

func (h *SystemHandler) getRedisStatus(ctx context.Context) RedisStatus {
	if h.redisClient == nil {
		return RedisStatus{Status: "not_configured"}
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := h.redisClient.Ping(pingCtx).Err(); err != nil {
		return RedisStatus{Status: "error"}
	}

	res := RedisStatus{Status: "connected"}
	infoCtx, infoCancel := context.WithTimeout(ctx, 2*time.Second)
	defer infoCancel()
	info, err := h.redisClient.Info(infoCtx, "memory", "server").Result()
	if err == nil {
		for _, line := range strings.Split(info, "\r\n") {
			switch {
			case strings.HasPrefix(line, "used_memory_human:"):
				res.UsedMemory = strings.TrimPrefix(line, "used_memory_human:")
			case strings.HasPrefix(line, "redis_version:"):
				res.Version = strings.TrimSpace(strings.TrimPrefix(line, "redis_version:"))
			}
		}
	}
	return res
}

func (h *SystemHandler) buildComponents(status SystemStatus) []ComponentHealth {
	checkedAt := time.Now().UTC().Format(time.RFC3339)
	redisHealthy := status.Redis.Status == "connected" || status.Redis.Status == "not_configured"

	return []ComponentHealth{
		{
			Name:      "api",
			Status:    "running",
			Healthy:   true,
			Version:   h.version,
			CheckedAt: checkedAt,
		},
		{
			Name:      "database",
			Status:    status.Database.Status,
			Healthy:   status.Database.Status == "connected",
			Version:   status.Database.Version,
			CheckedAt: checkedAt,
		},
		{
			Name:      "redis",
			Status:    status.Redis.Status,
			Healthy:   redisHealthy,
			Version:   status.Redis.Version,
			CheckedAt: checkedAt,
		},
	}
}

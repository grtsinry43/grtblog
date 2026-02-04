package handler

import (
	"context"
	"io/fs"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sys/unix"
	"gorm.io/gorm"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type SystemHandler struct {
	db          *gorm.DB
	redisClient *redis.Client
	events      appEvent.Bus

	// 静态/缓存数据，减少分配
	version           string
	storageSize       uint64
	lastStorageUpdate time.Time
	mu                sync.RWMutex
}

func NewSystemHandler(db *gorm.DB, redisClient *redis.Client, events appEvent.Bus) *SystemHandler {
	if events == nil {
		events = appEvent.NopBus{}
	}
	h := &SystemHandler{
		db:          db,
		redisClient: redisClient,
		version:     initBuildVersion(),
		events:      events,
	}
	return h
}

type SystemStatus struct {
	App      AppInfo        `json:"app"`
	CPU      CPUInfo        `json:"cpu"`
	Memory   MemoryInfo     `json:"memory"`
	Disk     DiskInfo       `json:"disk"`
	Storage  StorageInfo    `json:"storage"`
	Database DatabaseStatus `json:"database"`
	Redis    RedisStatus    `json:"redis"`
	Platform PlatformInfo   `json:"platform"`
}

type AppInfo struct {
	Version   string    `json:"version"`
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
	PoolStats DBPoolStats `json:"poolStats"`
}

type RedisStatus struct {
	Status     string `json:"status"`
	UsedMemory string `json:"usedMemory,omitempty"`
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
	if status.Database.Status != "connected" || (status.Redis.Status != "ok" && status.Redis.Status != "not_configured") {
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

func initBuildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info == nil {
		return "dev"
	}
	if info.Main.Version != "" {
		return info.Main.Version
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}
	return "dev"
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

	s := sqlDB.Stats()
	return DatabaseStatus{
		Status: status,
		Driver: h.db.Dialector.Name(),
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
	info, err := h.redisClient.Info(ctx, "memory").Result()
	if err == nil {
		for _, line := range strings.Split(info, "\r\n") {
			if strings.HasPrefix(line, "used_memory_human:") {
				res.UsedMemory = strings.TrimPrefix(line, "used_memory_human:")
				break
			}
		}
	}
	return res
}

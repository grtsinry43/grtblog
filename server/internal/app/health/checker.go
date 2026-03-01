package health

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SysConfigReader is the subset of sysconfig.Service used by the checker.
type SysConfigReader interface {
	GetConfigValue(ctx context.Context, key string) (string, error)
}

// Checker is a background goroutine that periodically probes DB, Redis,
// the SvelteKit renderer, and the site.maintenance sysconfig flag,
// then publishes state changes.
type Checker struct {
	state       *State
	db          *gorm.DB
	redis       *redis.Client
	sysconf     SysConfigReader
	events      appEvent.Bus
	interval    time.Duration
	rendererURL string
	httpClient  *http.Client
}

// NewChecker creates a health checker. Interval defaults to 15s if zero.
func NewChecker(state *State, db *gorm.DB, redisClient *redis.Client, sysconf SysConfigReader, events appEvent.Bus, interval time.Duration, rendererURL string) *Checker {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	return &Checker{
		state:       state,
		db:          db,
		redis:       redisClient,
		sysconf:     sysconf,
		events:      events,
		interval:    interval,
		rendererURL: rendererURL,
		httpClient:  &http.Client{Timeout: 2 * time.Second},
	}
}

// Run starts the check loop. It blocks until ctx is cancelled.
func (c *Checker) Run(ctx context.Context) {
	log.Printf("[health] checker started (interval=%s)", c.interval)

	// Immediate first probe.
	c.probe(ctx)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[health] checker stopped")
			return
		case <-ticker.C:
			c.probe(ctx)
		}
	}
}

func (c *Checker) probe(ctx context.Context) {
	prevSnapshot := c.state.Snapshot()
	prevValue := prevSnapshot.HealthBits
	prevMaint := prevSnapshot.Maintenance

	// Backend is always healthy if we're running.
	c.state.SetBit(BitBackend, true)

	// Probe database.
	c.state.SetBit(BitDatabase, c.probeDB(ctx))

	// Probe Redis.
	c.state.SetBit(BitRedis, c.probeRedis(ctx))

	// Probe SvelteKit renderer.
	c.state.SetBit(BitRenderer, c.probeRenderer())

	// Read manual maintenance flag from sysconfig.
	c.readMaintenance(ctx)

	newSnapshot := c.state.Snapshot()

	// Publish event if anything changed.
	if newSnapshot.HealthBits != prevValue || newSnapshot.Maintenance != prevMaint {
		if c.events != nil {
			_ = c.events.Publish(ctx, StateChanged{
				Prev:     prevValue,
				Next:     newSnapshot.HealthBits,
				Snapshot: newSnapshot,
				At:       time.Now(),
			})
		}
		log.Printf("[health] state changed bits=%06b→%06b maintenance=%v mode=%s",
			prevValue, newSnapshot.HealthBits, newSnapshot.Maintenance, newSnapshot.Mode)
	}
}

func (c *Checker) probeDB(ctx context.Context) bool {
	if c.db == nil {
		return false
	}
	sqlDB, err := c.db.DB()
	if err != nil {
		return false
	}
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return sqlDB.PingContext(pingCtx) == nil
}

func (c *Checker) probeRedis(ctx context.Context) bool {
	if c.redis == nil {
		// Redis not configured is treated as healthy (optional component).
		return true
	}
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.redis.Ping(pingCtx).Err() == nil
}

func (c *Checker) probeRenderer() bool {
	if c.rendererURL == "" {
		return true // not configured → assume healthy
	}
	req, err := http.NewRequest(http.MethodHead, c.rendererURL, nil)
	if err != nil {
		return false
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode < 500
}

func (c *Checker) readMaintenance(ctx context.Context) {
	if c.sysconf == nil {
		return
	}
	val, err := c.sysconf.GetConfigValue(ctx, "site.maintenance")
	if err != nil {
		// Key doesn't exist yet — not in maintenance.
		c.state.SetMaintenance(false)
		return
	}
	on, err := strconv.ParseBool(strings.TrimSpace(val))
	if err != nil {
		c.state.SetMaintenance(false)
		return
	}
	c.state.SetMaintenance(on)
}

package router

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
	"github.com/grtsinry43/grtblog-v2/server/internal/ws"
)

// wsIPLimiter tracks per-IP WebSocket connection counts.
type wsIPLimiter struct {
	mu    sync.Mutex
	conns map[string]*int64
	max   int64
}

func newWSIPLimiter(max int64) *wsIPLimiter {
	return &wsIPLimiter{conns: make(map[string]*int64), max: max}
}

func (l *wsIPLimiter) acquire(ip string) bool {
	l.mu.Lock()
	counter, ok := l.conns[ip]
	if !ok {
		var n int64
		counter = &n
		l.conns[ip] = counter
	}
	l.mu.Unlock()
	cur := atomic.AddInt64(counter, 1)
	if cur > l.max {
		atomic.AddInt64(counter, -1)
		return false
	}
	return true
}

func (l *wsIPLimiter) Release(ip string) {
	l.mu.Lock()
	counter, ok := l.conns[ip]
	l.mu.Unlock()
	if ok {
		atomic.AddInt64(counter, -1)
	}
}

func registerWSRoutes(v2 fiber.Router, manager *ws.Manager, deps Dependencies) {
	contentRepo := persistence.NewContentRepository(deps.DB)
	thinkingRepo := persistence.NewThinkingRepository(deps.DB)
	presenceResolver := ws.NewPresenceTitleResolver(contentRepo, thinkingRepo)
	presenceHub := ws.NewPresenceHub(manager, presenceResolver)
	wsHandler := handler.NewWSHandler(manager, deps.Analytics, presenceHub, deps.OwnerStatus)
	userRepo := persistence.NewIdentityRepository(deps.DB)

	// Per-IP WebSocket connection limit (50 connections per IP).
	ipLimiter := newWSIPLimiter(50)

	// wsAcquire checks upgrade + per-IP limit. On success the caller MUST
	// store limiter/IP in Locals so the WS handler can release on disconnect.
	// Returns (acquired, error). When acquired is true the caller must ensure
	// Release is eventually called (via handler defer) even if a later
	// middleware step fails — so we always pass through to the handler and
	// let it release via defer.
	wsAcquire := func(c *fiber.Ctx) (bool, error) {
		if !websocket.IsWebSocketUpgrade(c) {
			return false, fiber.ErrUpgradeRequired
		}
		if !ipLimiter.acquire(c.IP()) {
			return false, fiber.NewError(fiber.StatusTooManyRequests, "too many WebSocket connections")
		}
		c.Locals("wsIPLimiter", ipLimiter)
		c.Locals("wsClientIP", c.IP())
		return true, nil
	}

	// wsRelease undoes an acquire when the middleware rejects before reaching
	// the WebSocket handler (so the handler's defer won't fire).
	wsRelease := func(c *fiber.Ctx) {
		ipLimiter.Release(c.IP())
	}

	resolveWSUserFromToken := func(c *fiber.Ctx, token string) error {
		if deps.JWTManager == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "ws auth unavailable")
		}
		claims, parseErr := deps.JWTManager.Parse(token)
		if parseErr != nil || claims == nil || claims.UserID <= 0 {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid ws token")
		}
		user, dbErr := userRepo.FindByID(c.Context(), claims.UserID)
		if dbErr != nil || !user.IsActive {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid ws token")
		}
		c.Locals("wsUserID", user.ID)
		c.Locals("wsUserIsAdmin", user.IsAdmin)
		return nil
	}

	v2.Use("/ws/realtime", func(c *fiber.Ctx) error {
		acquired, err := wsAcquire(c)
		if err != nil {
			return err
		}
		token := extractWSJWTToken(c)
		if token == "" {
			return c.Next()
		}
		if err := resolveWSUserFromToken(c, token); err != nil {
			if acquired {
				wsRelease(c)
			}
			return err
		}
		return c.Next()
	})
	v2.Get("/ws/realtime", websocket.New(wsHandler.HandleRealtime, websocket.Config{
		Subprotocols: []string{"grtblog.jwt"},
	}))

	v2.Use("/ws/presence", func(c *fiber.Ctx) error {
		_, err := wsAcquire(c)
		if err != nil {
			return err
		}
		return c.Next()
	})
	v2.Get("/ws/presence", websocket.New(wsHandler.HandlePresence))

	v2.Use("/ws", func(c *fiber.Ctx) error {
		path := c.Path()
		if strings.HasSuffix(path, "/ws/notifications") {
			return c.Next()
		}
		if !strings.HasSuffix(path, "/ws") {
			return c.Next()
		}
		_, err := wsAcquire(c)
		if err != nil {
			return err
		}

		roomKey, parseErr := parseWSRoomKey(c)
		if parseErr != nil {
			wsRelease(c)
			return fiber.NewError(fiber.StatusBadRequest, parseErr.Error())
		}
		c.Locals("wsRoomKey", roomKey)
		return c.Next()
	})

	v2.Get("/ws", websocket.New(wsHandler.Handle))

	v2.Use("/ws/notifications", func(c *fiber.Ctx) error {
		acquired, err := wsAcquire(c)
		if err != nil {
			return err
		}
		token := extractWSJWTToken(c)
		if token == "" {
			if acquired {
				wsRelease(c)
			}
			return fiber.NewError(fiber.StatusUnauthorized, "missing ws token")
		}
		if err := resolveWSUserFromToken(c, token); err != nil {
			if acquired {
				wsRelease(c)
			}
			return err
		}
		return c.Next()
	})
	v2.Get("/ws/notifications", websocket.New(wsHandler.HandleNotification, websocket.Config{
		Subprotocols: []string{"grtblog.jwt"},
	}))
}

func parseWSRoomKey(c *fiber.Ctx) (string, error) {
	roomType := strings.TrimSpace(c.Query("type"))
	if roomType == "" {
		return "", fmt.Errorf("missing room type")
	}
	switch roomType {
	case "article", "moment", "page":
	default:
		return "", fmt.Errorf("invalid room type")
	}

	idStr := strings.TrimSpace(c.Query("id"))
	if idStr == "" {
		return "", fmt.Errorf("missing room id")
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return "", fmt.Errorf("invalid room id")
	}

	return fmt.Sprintf("%s:%d", roomType, id), nil
}

func extractWSJWTToken(c *fiber.Ctx) string {
	if token := extractBearerToken(c.Get("Authorization")); token != "" {
		return token
	}

	protocols := splitHeaderTokens(c.Get("Sec-WebSocket-Protocol"))
	if len(protocols) >= 2 && strings.EqualFold(protocols[0], "grtblog.jwt") {
		return protocols[1]
	}

	for _, protocol := range protocols {
		const bearerPrefix = "bearer."
		if strings.HasPrefix(strings.ToLower(protocol), bearerPrefix) && len(protocol) > len(bearerPrefix) {
			return protocol[len(bearerPrefix):]
		}
	}
	return ""
}

func splitHeaderTokens(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	items := strings.Split(value, ",")
	out := make([]string, 0, len(items))
	for _, item := range items {
		token := strings.TrimSpace(item)
		if token != "" {
			out = append(out, token)
		}
	}
	return out
}

func extractBearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

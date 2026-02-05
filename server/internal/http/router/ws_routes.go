package router

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/ws"
)

func registerWSRoutes(v2 fiber.Router, manager *ws.Manager, deps Dependencies) {
	wsHandler := handler.NewWSHandler(manager, deps.Analytics)

	v2.Use("/ws", func(c *fiber.Ctx) error {
		path := c.Path()
		if strings.HasSuffix(path, "/ws/notifications") {
			return c.Next()
		}
		if !strings.HasSuffix(path, "/ws") {
			return c.Next()
		}
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		roomKey, err := parseWSRoomKey(c)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		c.Locals("wsRoomKey", roomKey)
		return c.Next()
	})

	v2.Get("/ws", websocket.New(wsHandler.Handle))

	v2.Use("/ws/notifications", func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}
		token := extractWSJWTToken(c)
		if token != "" && deps.JWTManager != nil {
			claims, err := deps.JWTManager.Parse(token)
			if err == nil && claims != nil && claims.UserID > 0 {
				c.Locals("wsUserID", claims.UserID)
				return c.Next()
			}
			return fiber.NewError(fiber.StatusUnauthorized, "invalid ws token")
		}
		return fiber.NewError(fiber.StatusUnauthorized, "missing ws token")
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

	// Fallback for non-browser debugging clients.
	return strings.TrimSpace(c.Query("access_token"))
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

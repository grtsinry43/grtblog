package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

// RegisterGrafanaProxy sets up /g/* as a reverse proxy to Grafana,
// gated behind Passkey session authentication.
// Grafana is configured with GF_SERVER_SERVE_FROM_SUB_PATH=true and
// GF_SERVER_ROOT_URL=.../g/, so it expects requests at /g/* and all
// its redirects/resources already include the /g/ prefix.
//
// Known limitation: WebSocket connections (Grafana Live /api/live/*)
// are not proxied — fasthttp does not support HTTP Upgrade.
// Static dashboards and queries work normally.
func RegisterGrafanaProxy(app *fiber.App, store *Store, grafanaURL string) {
	// /g without slash: check auth first, then redirect to /g/ only if valid.
	app.Get("/g", func(c *fiber.Ctx) error {
		token := c.Cookies(sessionCookieName)
		if token == "" {
			return c.Status(fiber.StatusNotFound).SendString("404 page not found")
		}
		valid, _ := store.ValidateAdminSession(c.UserContext(), token)
		if !valid {
			return c.Status(fiber.StatusNotFound).SendString("404 page not found")
		}
		return c.Redirect("/g/", fiber.StatusMovedPermanently)
	})

	app.Use("/g/", func(c *fiber.Ctx) error {
		token := c.Cookies(sessionCookieName)
		if token == "" {
			return c.Status(fiber.StatusNotFound).SendString("404 page not found")
		}
		valid, err := store.ValidateAdminSession(c.UserContext(), token)
		if err != nil || !valid {
			return c.Status(fiber.StatusNotFound).SendString("404 page not found")
		}

		// Forward the full path as-is to Grafana (it expects /g/* with sub-path mode).
		target := grafanaURL + c.OriginalURL()

		if err := proxy.Do(c, target); err != nil {
			return c.Status(fiber.StatusBadGateway).SendString("grafana unreachable")
		}

		c.Response().Header.Del("X-Frame-Options")
		return nil
	})
}

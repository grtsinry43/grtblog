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
	grafanaProxy := proxy.Balancer(proxy.Config{
		Servers: []string{grafanaURL},
	})

	// /g without slash: check auth first, then redirect to /g/ only if valid.
	app.Get("/g", func(c *fiber.Ctx) error {
		// Fiber's default non-strict routing also matches /g/. Let that exact
		// path continue into the Grafana proxy instead of redirecting to itself.
		if c.Path() != "/g" {
			return c.Next()
		}
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

		// Keep the original /g/* request URI. proxy.Do with an absolute target
		// mutates the active fasthttp request and can make Grafana canonicalise
		// /g/ back to itself indefinitely behind an HTTPS reverse proxy.
		if err := grafanaProxy(c); err != nil {
			return c.Status(fiber.StatusBadGateway).SendString("grafana unreachable")
		}

		c.Response().Header.Del("X-Frame-Options")
		return nil
	})
}

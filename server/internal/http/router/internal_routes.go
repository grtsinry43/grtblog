package router

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// registerInternalRoutes exposes endpoints intended for container-internal
// callers only (currently the renderer). nginx only proxies /api/, /uploads/
// and a few fixed paths, so /internal/ is unreachable from the public network
// in the standard deployment; inside the compose network the renderer reaches
// it via http://server:8080 directly.
func registerInternalRoutes(app *fiber.App, deps Dependencies) {
	// Backfill hook for the static-first architecture: when nginx misses a
	// static snapshot and falls back to SSR, the renderer reports the path
	// here so the ISR queue regenerates the snapshot in the background.
	app.Post("/internal/isr/revalidate", func(c *fiber.Ctx) error {
		if deps.ISR == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "isr not configured")
		}
		var req struct {
			URLPath string `json:"urlPath"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}
		urlPath := strings.TrimSpace(req.URLPath)
		if !strings.HasPrefix(urlPath, "/") {
			return fiber.NewError(fiber.StatusBadRequest, "urlPath must be an absolute path")
		}
		enqueued, err := deps.ISR.EnqueueURL(c.UserContext(), urlPath)
		if err != nil {
			log.Printf("[isr] revalidate enqueue failed url=%s err=%v", urlPath, err)
			return fiber.NewError(fiber.StatusInternalServerError, "enqueue failed")
		}
		return c.JSON(fiber.Map{"enqueued": enqueued, "urlPath": urlPath})
	})
}

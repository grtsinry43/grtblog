package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/webhook"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerWebhookAdminRoutes(v2 fiber.Router, deps Dependencies, webhookSvc *webhook.Service) {
	if webhookSvc == nil {
		return
	}
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminTokenRepo := persistence.NewAdminTokenRepository(deps.DB)
	adminGroup := v2.Group("", middleware.RequireAuth(deps.JWTManager, adminTokenRepo), middleware.RequireAdmin(identityRepo))
	webhookHandler := handler.NewWebhookHandler(webhookSvc)

	admin := adminGroup.Group("/admin")
	admin.Get("/webhooks", webhookHandler.ListWebhooks)
	admin.Get("/webhooks/events", webhookHandler.ListEvents)
	admin.Post("/webhooks", webhookHandler.CreateWebhook)
	admin.Put("/webhooks/:id", webhookHandler.UpdateWebhook)
	admin.Delete("/webhooks/:id", webhookHandler.DeleteWebhook)
	admin.Post("/webhooks/:id/test", webhookHandler.TestWebhook)

	admin.Get("/webhooks/deliveries", webhookHandler.ListHistory)
	admin.Post("/webhooks/deliveries/:id/replay", webhookHandler.ReplayHistory)
}

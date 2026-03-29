package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/thinking"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerThinkingPublicRoutes(v2 fiber.Router, deps Dependencies) {
	thinkingHandler := newThinkingHandler(deps)

	publicGroup := v2.Group("/thinkings")
	publicGroup.Get("/", thinkingHandler.ListThinkings)
	publicGroup.Post("/metrics", thinkingHandler.BatchGetThinkingMetrics) // POST /api/v2/thinkings/metrics
	publicGroup.Get("/:id/metrics", thinkingHandler.GetThinkingMetrics) // GET /api/v2/thinkings/123/metrics
	publicGroup.Get("/:id", thinkingHandler.GetThinking)
}

func registerThinkingAuthRoutes(v2 fiber.Router, deps Dependencies) {
	thinkingHandler := newThinkingHandler(deps)
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminTokenRepo := persistence.NewAdminTokenRepository(deps.DB)
	authGroup := v2.Group("/thinkings", middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo), middleware.RequireAdmin(identityRepo))
	authGroup.Post("/", thinkingHandler.CreateThinking)
	authGroup.Put("/:id", thinkingHandler.UpdateThinking)
	authGroup.Delete("/:id", thinkingHandler.DeleteThinking)
	adminGroup := v2.Group("/admin", middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo), middleware.RequireAdmin(identityRepo))
	adminGroup.Post("/thinkings/batch-delete", thinkingHandler.BatchDeleteThinkings)
}

func newThinkingHandler(deps Dependencies) *handler.ThinkingHandler {
	thinkingRepo := persistence.NewThinkingRepository(deps.DB)
	commentRepo := persistence.NewCommentRepository(deps.DB)
	thinkingSvc := thinking.NewService(thinkingRepo, commentRepo, deps.EventBus)
	userRepo := persistence.NewIdentityRepository(deps.DB)
	return handler.NewThinkingHandler(thinkingSvc, commentRepo, userRepo)
}

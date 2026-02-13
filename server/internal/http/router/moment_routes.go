package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/moment"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerMomentPublicRoutes(v2 fiber.Router, deps Dependencies) {
	momentHandler := newMomentHandler(deps)

	publicGroup := v2.Group("/moments")
	publicGroup.Get("/", momentHandler.ListMoments)  // GET /api/v2/moments
	publicGroup.Get("/:id", momentHandler.GetMoment) // GET /api/v2/moments/123
	publicGroup.Get("/:id/same-period-articles", momentHandler.ListSamePeriodArticles)
	publicGroup.Get("/short/:shortUrl", momentHandler.GetMomentByShortURL) // GET /api/v2/moments/short/abc123
	publicGroup.Post("/:id/latest", momentHandler.CheckMomentLatest)       // POST /api/v2/moments/123/latest

	v2.Get("/columns/short/:shortUrl/moments", momentHandler.ListMomentsByColumnShortURL)
}

func registerMomentAuthRoutes(v2 fiber.Router, deps Dependencies) {
	momentHandler := newMomentHandler(deps)
	adminTokenRepo := persistence.NewAdminTokenRepository(deps.DB)

	authGroup := v2.Group("/moments", middleware.RequireAuth(deps.JWTManager, adminTokenRepo))
	authGroup.Post("/", momentHandler.CreateMoment)      // POST /api/v2/moments
	authGroup.Put("/:id", momentHandler.UpdateMoment)    // PUT /api/v2/moments/123
	authGroup.Delete("/:id", momentHandler.DeleteMoment) // DELETE /api/v2/moments/123

	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminGroup := v2.Group("", middleware.RequireAuth(deps.JWTManager, adminTokenRepo), middleware.RequireAdmin(identityRepo))
	adminGroup.Get("/admin/moments", momentHandler.ListMomentsAdmin)                  // GET /api/v2/admin/moments
	adminGroup.Put("/admin/moments/published", momentHandler.BatchSetMomentPublished) // PUT /api/v2/admin/moments/published
	adminGroup.Put("/admin/moments/top", momentHandler.BatchSetMomentTop)             // PUT /api/v2/admin/moments/top
	adminGroup.Post("/admin/moments/batch-delete", momentHandler.BatchDeleteMoments)  // POST /api/v2/admin/moments/batch-delete
}

func newMomentHandler(deps Dependencies) *handler.MomentHandler {
	contentRepo := persistence.NewContentRepository(deps.DB)
	commentRepo := persistence.NewCommentRepository(deps.DB)
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	momentSvc := moment.NewService(contentRepo, commentRepo, deps.EventBus)
	return handler.NewMomentHandler(momentSvc, contentRepo, commentRepo, identityRepo)
}

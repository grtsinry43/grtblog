package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerBackupRoutes(v2 fiber.Router, deps Dependencies) {
	if deps.Backup == nil {
		return
	}
	h := handler.NewBackupHandler(deps.Backup)
	v2.Get("/backups/download", h.Download)

	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminTokenRepo := persistence.NewAdminTokenRepository(deps.DB)
	authMiddleware := middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo)
	adminMiddleware := middleware.RequireAdmin(identityRepo)
	backups := v2.Group("/admin/backups", authMiddleware, adminMiddleware)
	backups.Get("", h.List)
	backups.Post("", h.Create)
	backups.Get("/:id", h.Get)
	backups.Delete("/:id", h.Delete)
	backups.Post("/:id/download-ticket", h.IssueDownloadTicket)
}

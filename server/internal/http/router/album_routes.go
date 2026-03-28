package router

import (
	"github.com/gofiber/fiber/v2"

	appalbum "github.com/grtsinry43/grtblog-v2/server/internal/app/album"
	mediaapp "github.com/grtsinry43/grtblog-v2/server/internal/app/media"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerAlbumPublicRoutes(v2 fiber.Router, deps Dependencies) {
	albumHandler := newAlbumHandler(deps)

	publicGroup := v2.Group("/albums")
	publicGroup.Get("/", albumHandler.ListAlbums)                          // GET /api/v2/albums
	publicGroup.Get("/:id", albumHandler.GetAlbum)                         // GET /api/v2/albums/:id
	publicGroup.Get("/short/:shortUrl", albumHandler.GetAlbumByShortURL)   // GET /api/v2/albums/short/:shortUrl
}

func registerAlbumAuthRoutes(v2 fiber.Router, deps Dependencies) {
	albumHandler := newAlbumHandler(deps)
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminTokenRepo := persistence.NewAdminTokenRepository(deps.DB)

	authGroup := v2.Group("/albums", middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo), middleware.RequireAdmin(identityRepo))
	authGroup.Post("/", albumHandler.CreateAlbum)                              // POST /api/v2/albums
	authGroup.Put("/:id", albumHandler.UpdateAlbum)                            // PUT /api/v2/albums/:id
	authGroup.Delete("/:id", albumHandler.DeleteAlbum)                         // DELETE /api/v2/albums/:id
	authGroup.Post("/:id/photos", albumHandler.AddPhotos)                      // POST /api/v2/albums/:id/photos
	authGroup.Put("/:id/photos/reorder", albumHandler.ReorderPhotos)           // PUT /api/v2/albums/:id/photos/reorder
	authGroup.Put("/:id/photos/:photoId", albumHandler.UpdatePhoto)            // PUT /api/v2/albums/:id/photos/:photoId
	authGroup.Delete("/:id/photos/:photoId", albumHandler.DeletePhoto)         // DELETE /api/v2/albums/:id/photos/:photoId

	adminGroup := v2.Group("/admin", middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo), middleware.RequireAdmin(identityRepo))
	adminGroup.Get("/albums/:id", albumHandler.GetAlbumAdmin)                  // GET /api/v2/admin/albums/:id
	adminGroup.Get("/albums", albumHandler.ListAlbumsAdmin)                    // GET /api/v2/admin/albums
	adminGroup.Put("/albums/published", albumHandler.BatchSetAlbumPublished)   // PUT /api/v2/admin/albums/published
	adminGroup.Post("/albums/batch-delete", albumHandler.BatchDeleteAlbums)    // POST /api/v2/admin/albums/batch-delete
}

func newAlbumHandler(deps Dependencies) *handler.AlbumHandler {
	albumRepo := persistence.NewAlbumRepository(deps.DB)
	commentRepo := persistence.NewCommentRepository(deps.DB)
	uploadRepo := persistence.NewUploadFileRepository(deps.DB)
	albumSvc := appalbum.NewService(albumRepo, commentRepo, deps.EventBus)
	mediaSvc := mediaapp.NewService(uploadRepo, "", deps.EventBus)
	return handler.NewAlbumHandler(albumSvc, albumRepo, commentRepo, mediaSvc)
}

package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminnotification"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/auth"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/friendlink"
	mediaapp "github.com/grtsinry43/grtblog-v2/server/internal/app/media"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerUserRoutes(v2 fiber.Router, deps Dependencies, websiteInfoHandler *handler.WebsiteInfoHandler) {
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminTokenRepo := persistence.NewAdminTokenRepository(deps.DB)
	authMiddleware := middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo)
	adminMiddleware := middleware.RequireAdmin(identityRepo)

	oauthRepo := persistence.NewOAuthProviderRepository(deps.DB)
	var stateStore auth.StateStore
	if deps.Redis != nil {
		stateStore = auth.NewRedisStateStore(deps.Redis, deps.Config.Redis.Prefix)
	}
	authSvc := auth.NewService(identityRepo, oauthRepo, deps.JWTManager, stateStore, deps.Config.Auth)
	authHandler := handler.NewAuthHandler(authSvc, nil, nil, nil)
	authenticatedAuth := v2.Group("/auth", authMiddleware)
	authenticatedAuth.Get("/access-info", authHandler.AccessInfo)
	authenticatedAuth.Get("/profile", authHandler.Profile)
	authenticatedAuth.Put("/profile", authHandler.UpdateProfile)
	authenticatedAuth.Put("/password", authHandler.ChangePassword)
	authenticatedAuth.Get("/oauth-bindings", authHandler.ListOAuthBindings)
	authenticatedAuth.Post("/oauth-bindings/:provider/callback", authHandler.BindOAuth)
	authenticatedAuth.Delete("/oauth-bindings/:provider", authHandler.UnbindOAuth)

	friendLinkRepo := persistence.NewFriendLinkApplicationRepository(deps.DB)
	friendLinkSvc := friendlink.NewService(friendLinkRepo, deps.EventBus)
	friendLinkHandler := handler.NewFriendLinkHandler(friendLinkSvc, deps.SysConfig)
	friendLinks := v2.Group("/friend-links", authMiddleware)
	friendLinks.Post("/applications", friendLinkHandler.SubmitApplication)

	uploadRepo := persistence.NewUploadFileRepository(deps.DB)
	uploadSvc := mediaapp.NewService(uploadRepo, "", deps.EventBus)
	uploadHandler := handler.NewUploadHandler(uploadSvc)
	v2.Post("/upload", authMiddleware, adminMiddleware, uploadHandler.UploadFile)
	v2.Get("/uploads", authMiddleware, adminMiddleware, uploadHandler.ListUploads)
	v2.Post("/uploads/sync", authMiddleware, adminMiddleware, uploadHandler.SyncUploads)
	v2.Put("/upload/:id", authMiddleware, adminMiddleware, uploadHandler.RenameUpload)
	v2.Delete("/upload/:id", authMiddleware, adminMiddleware, uploadHandler.DeleteUpload)
	v2.Get("/upload/:id/download", authMiddleware, adminMiddleware, uploadHandler.DownloadUpload)

	adminNotificationRepo := persistence.NewAdminNotificationRepository(deps.DB)
	adminNotificationSvc := adminnotification.NewService(adminNotificationRepo, deps.EventBus)
	adminNotificationHandler := handler.NewAdminNotificationHandler(adminNotificationSvc)
	authenticatedNotifications := v2.Group("/notifications", authMiddleware)
	authenticatedNotifications.Get("", adminNotificationHandler.ListMine)
	authenticatedNotifications.Post("/:id/read", adminNotificationHandler.MarkRead)
	authenticatedNotifications.Post("/read-all", adminNotificationHandler.MarkAllRead)
}

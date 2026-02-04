package router

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminstats"
	appfed "github.com/grtsinry43/grtblog-v2/server/internal/app/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/federationconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/friendlink"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/hitokoto"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	fedinfra "github.com/grtsinry43/grtblog-v2/server/internal/infra/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
	"github.com/grtsinry43/grtblog-v2/server/internal/ws"
)

func registerAdminRoutes(v2 fiber.Router, deps Dependencies, websiteInfoHandler *handler.WebsiteInfoHandler, navMenuHandler *handler.NavMenuHandler, sysCfgSvc *sysconfig.Service, wsManager *ws.Manager) {
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminGroup := v2.Group("", middleware.RequireAuth(deps.JWTManager), middleware.RequireAdmin(identityRepo))

	websiteInfo := adminGroup.Group("/website-info")
	websiteInfo.Get("", websiteInfoHandler.List)
	websiteInfo.Put("/:key", websiteInfoHandler.Update)

	navMenus := adminGroup.Group("/admin/nav-menus")
	navMenus.Get("", navMenuHandler.ListAdmin)
	navMenus.Post("", navMenuHandler.Create)
	navMenus.Put("/reorder", navMenuHandler.Reorder)
	navMenus.Put("/:id", navMenuHandler.Update)
	navMenus.Delete("/:id", navMenuHandler.Delete)

	oauthRepo := persistence.NewOAuthProviderRepository(deps.DB)
	adminOAuth := handler.NewAdminOAuthHandler(oauthRepo)
	admin := adminGroup.Group("/admin")
	admin.Get("/oauth-providers", adminOAuth.List)
	admin.Post("/oauth-providers", adminOAuth.Create)
	admin.Put("/oauth-providers/:key", adminOAuth.Update)
	admin.Delete("/oauth-providers/:key", adminOAuth.Delete)

	commentHandler := newCommentHandler(deps)
	admin.Get("/comments", commentHandler.ListAdminComments)
	admin.Put("/comments/viewed", commentHandler.MarkCommentsViewed)
	admin.Post("/comments/:id/reply", commentHandler.ReplyComment)
	admin.Put("/comments/:id/status", commentHandler.UpdateCommentStatus)
	admin.Put("/comments/:id/author", commentHandler.SetCommentAuthor)
	admin.Put("/comments/:id/top", commentHandler.SetCommentTop)
	admin.Delete("/comments/:id", commentHandler.DeleteComment)
	admin.Put("/comments/areas/:areaId/close", commentHandler.SetCommentAreaClose)

	if sysCfgSvc != nil {
		sysConfigHandler := handler.NewSysConfigHandler(sysCfgSvc)
		admin.Get("/sysconfig", sysConfigHandler.ListSysConfig)
		admin.Put("/sysconfig", sysConfigHandler.UpdateSysConfig)
	}

	fedCfgRepo := persistence.NewFederationConfigRepository(deps.DB)
	fedCfgSvc := federationconfig.NewService(fedCfgRepo)
	fedCfgHandler := handler.NewFederationConfigHandler(fedCfgSvc)
	admin.Get("/federation/config", fedCfgHandler.ListFederationConfig)
	admin.Put("/federation/config", fedCfgHandler.UpdateFederationConfig)

	contentRepo := persistence.NewContentRepository(deps.DB)
	instanceRepo := persistence.NewFederationInstanceRepository(deps.DB)
	var cache fedinfra.Cache
	if deps.Redis != nil {
		cache = fedinfra.NewRedisCache(deps.Redis, deps.Config.Redis.Prefix)
	}
	resolver := fedinfra.NewResolver(&http.Client{Timeout: 10 * time.Second}, cache)
	outbound := appfed.NewOutboundService(fedCfgSvc, resolver, instanceRepo)
	federationAdminHandler := handler.NewFederationAdminHandler(fedCfgSvc, contentRepo, outbound, resolver)
	admin.Post("/federation/friendlinks/request", federationAdminHandler.RequestFriendLink)
	admin.Post("/federation/citations/request", federationAdminHandler.SendCitation)
	admin.Post("/federation/mentions/notify", federationAdminHandler.SendMention)
	admin.Get("/federation/remote/check", federationAdminHandler.CheckRemote)

	hitokotoSvc := hitokoto.NewService(deps.Redis, deps.Config.Redis.Prefix)
	hitokotoHandler := handler.NewAdminHitokotoHandler(hitokotoSvc)
	admin.Get("/hitokoto", hitokotoHandler.GetSentence)

	friendLinkAppRepo := persistence.NewFriendLinkApplicationRepository(deps.DB)
	friendLinkRepo := persistence.NewFriendLinkRepository(deps.DB)
	friendLinkAdminSvc := friendlink.NewAdminService(friendLinkAppRepo, friendLinkRepo, instanceRepo)
	friendLinkAdminHandler := handler.NewFriendLinkAdminHandler(friendLinkAdminSvc)
	admin.Get("/friend-links/applications", friendLinkAdminHandler.ListApplications)
	admin.Put("/friend-links/applications/:id/approve", friendLinkAdminHandler.ApproveApplication)
	admin.Put("/friend-links/applications/:id/reject", friendLinkAdminHandler.RejectApplication)
	admin.Put("/friend-links/applications/:id/block", friendLinkAdminHandler.BlockApplication)
	admin.Put("/friend-links/applications/:id/status", friendLinkAdminHandler.UpdateApplicationStatus)
	admin.Get("/friend-links", friendLinkAdminHandler.ListFriendLinks)
	admin.Post("/friend-links", friendLinkAdminHandler.CreateFriendLink)
	admin.Put("/friend-links/:id", friendLinkAdminHandler.UpdateFriendLink)
	admin.Put("/friend-links/:id/block", friendLinkAdminHandler.BlockFriendLink)
	admin.Delete("/friend-links/:id", friendLinkAdminHandler.DeleteFriendLink)

	logHandler := handler.NewAdminLogHandler("storage/logs/app.log", 200)
	systemHandler := handler.NewSystemHandler(deps.DB, deps.Redis)
	adminStatsSvc := adminstats.NewService(deps.DB, deps.Redis, deps.Config.Redis.Prefix, wsManager)
	adminStatsHandler := handler.NewAdminStatsHandler(adminStatsSvc)
	adminLogs := adminGroup.Group("/admin")
	adminLogs.Get("/logs", logHandler.List)
	adminLogs.Get("/system/status", systemHandler.GetStatus)
	adminLogs.Get("/stats/dashboard", adminStatsHandler.GetDashboard)
}

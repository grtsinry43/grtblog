package router

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	appap "github.com/grtsinry43/grtblog-v2/server/internal/app/activitypub"
	appapcfg "github.com/grtsinry43/grtblog-v2/server/internal/app/activitypubconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminnotification"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminstats"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminuser"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/email"
	appfed "github.com/grtsinry43/grtblog-v2/server/internal/app/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/federationconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/friendlink"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/globalnotification"
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
	adminTokenRepo := persistence.NewAdminTokenRepository(deps.DB)
	adminGroup := v2.Group("", middleware.RequireAuth(deps.JWTManager, adminTokenRepo), middleware.RequireAdmin(identityRepo))
	ownerStatusHandler := handler.NewOwnerStatusHandler(deps.OwnerStatus)
	adminGroup.Post("/onlineStatus", ownerStatusHandler.UpdateStatus)
	adminGroup.Post("/admin/owner-status/panel-heartbeat", ownerStatusHandler.PanelHeartbeat)

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
	adminTokenHandler := handler.NewAdminTokenHandler(adminTokenRepo, identityRepo)
	admin := adminGroup.Group("/admin")
	eventHandler := handler.NewEventHandler()
	admin.Get("/events", eventHandler.ListEvents)
	admin.Get("/events/catalog", eventHandler.ListEventCatalog)
	admin.Get("/events/catalog/:name", eventHandler.GetEventCatalogItem)
	admin.Get("/oauth-providers", adminOAuth.List)
	admin.Post("/oauth-providers", adminOAuth.Create)
	admin.Put("/oauth-providers/:key", adminOAuth.Update)
	admin.Delete("/oauth-providers/:key", adminOAuth.Delete)
	admin.Get("/tokens", adminTokenHandler.List)
	admin.Post("/tokens", adminTokenHandler.Create)
	admin.Delete("/tokens/:id", adminTokenHandler.Delete)

	commentHandler := newCommentHandler(deps)
	admin.Get("/comments", commentHandler.ListAdminComments)
	admin.Get("/visitors", commentHandler.ListAdminVisitors)
	admin.Get("/visitors/insights", commentHandler.GetAdminVisitorInsights)
	admin.Get("/visitors/:visitorId", commentHandler.GetAdminVisitorProfile)
	admin.Put("/comments/viewed", commentHandler.MarkCommentsViewed)
	admin.Post("/comments/:id/reply", commentHandler.ReplyComment)
	admin.Put("/comments/:id/status", commentHandler.UpdateCommentStatus)
	admin.Put("/comments/:id/author", commentHandler.SetCommentAuthor)
	admin.Put("/comments/:id/top", commentHandler.SetCommentTop)
	admin.Delete("/comments/:id", commentHandler.DeleteComment)
	admin.Put("/comments/areas/:areaId/close", commentHandler.SetCommentAreaClose)

	adminUserSvc := adminuser.NewService(identityRepo)
	adminUserHandler := handler.NewAdminUserHandler(adminUserSvc)
	admin.Get("/users", adminUserHandler.ListUsers)
	admin.Put("/users/:id", adminUserHandler.UpdateUser)

	rssAccessSvc := newRSSAccessAnalyticsService(deps)
	rssAdminHandler := handler.NewRSSAdminHandler(rssAccessSvc)
	admin.Get("/rss/access-stats", rssAdminHandler.GetAccessStats)

	if sysCfgSvc != nil {
		sysConfigHandler := handler.NewSysConfigHandler(sysCfgSvc)
		admin.Get("/sysconfig", sysConfigHandler.ListSysConfig)
		admin.Put("/sysconfig", sysConfigHandler.UpdateSysConfig)

		emailRepo := persistence.NewEmailRepository(deps.DB)
		emailSender := email.NewSender(sysCfgSvc)
		websiteInfoRepo := persistence.NewWebsiteInfoRepository(deps.DB)
		emailSvc := email.NewService(emailRepo, emailSender, websiteInfoRepo)
		emailHandler := handler.NewEmailTemplateHandler(emailSvc)
		admin.Get("/email/templates", emailHandler.ListEmailTemplates)
		admin.Post("/email/templates", emailHandler.CreateEmailTemplate)
		admin.Put("/email/templates/:code", emailHandler.UpdateEmailTemplate)
		admin.Delete("/email/templates/:code", emailHandler.DeleteEmailTemplate)
		admin.Post("/email/templates/:code/preview", emailHandler.PreviewEmailTemplate)
		admin.Post("/email/templates/:code/test", emailHandler.TestEmailTemplate)
		admin.Get("/email/subscriptions", emailHandler.ListEmailSubscriptions)
		admin.Put("/email/subscriptions/status", emailHandler.BatchUpdateEmailSubscriptionStatus)
	}

	fedCfgRepo := persistence.NewFederationConfigRepository(deps.DB)
	fedCfgSvc := federationconfig.NewService(fedCfgRepo)
	apCfgSvc := appapcfg.NewService(fedCfgRepo)
	fedCfgHandler := handler.NewFederationConfigHandler(fedCfgSvc)
	activityPubCfgHandler := handler.NewActivityPubConfigHandler(apCfgSvc)
	admin.Get("/federation/config", fedCfgHandler.ListFederationConfig)
	admin.Put("/federation/config", fedCfgHandler.UpdateFederationConfig)
	admin.Get("/activitypub/config", activityPubCfgHandler.ListActivityPubConfig)
	admin.Put("/activitypub/config", activityPubCfgHandler.UpdateActivityPubConfig)

	contentRepo := persistence.NewContentRepository(deps.DB)
	instanceRepo := persistence.NewFederationInstanceRepository(deps.DB)
	var cache fedinfra.Cache
	if deps.Redis != nil {
		cache = fedinfra.NewRedisCache(deps.Redis, deps.Config.Redis.Prefix)
	}
	resolver := fedinfra.NewResolver(&http.Client{Timeout: 10 * time.Second}, cache)
	outbound := appfed.NewOutboundService(fedCfgSvc, resolver, instanceRepo)
	outboundRepo := persistence.NewOutboundDeliveryRepository(deps.DB)
	deliverySvc := appfed.NewDeliveryService(outboundRepo, outbound, deps.EventBus)
	federationAdminHandler := handler.NewFederationAdminHandler(fedCfgSvc, contentRepo, deliverySvc, instanceRepo, resolver, deps.EventBus)
	federationReviewHandler := handler.NewFederationReviewHandler(
		persistence.NewFederatedCitationRepository(deps.DB),
		persistence.NewFederatedMentionRepository(deps.DB),
		instanceRepo,
		outbound,
	)
	activityPubSvc := appap.NewService(
		apCfgSvc,
		persistence.NewActivityPubFollowerRepository(deps.DB),
		persistence.NewActivityPubOutboxRepository(deps.DB),
		contentRepo,
		persistence.NewThinkingRepository(deps.DB),
		persistence.NewCommentRepository(deps.DB),
		identityRepo,
		adminnotification.NewService(persistence.NewAdminNotificationRepository(deps.DB), deps.EventBus),
	)
	activityPubAdminHandler := handler.NewActivityPubAdminHandler(activityPubSvc)
	admin.Post("/federation/friendlinks/request", federationAdminHandler.RequestFriendLink)
	admin.Post("/federation/citations/request", federationAdminHandler.SendCitation)
	admin.Post("/federation/mentions/notify", federationAdminHandler.SendMention)
	admin.Post("/activitypub/publish", activityPubAdminHandler.Publish)
	admin.Get("/activitypub/followers", activityPubAdminHandler.ListFollowers)
	admin.Post("/federation/activitypub/publish", activityPubAdminHandler.Publish)
	admin.Get("/federation/activitypub/followers", activityPubAdminHandler.ListFollowers)
	admin.Get("/federation/remote/check", federationAdminHandler.CheckRemote)
	admin.Get("/federation/instances", federationAdminHandler.ListInstances)
	admin.Get("/federation/instances/:id", federationAdminHandler.GetInstance)
	admin.Put("/federation/instances/:id/status", federationAdminHandler.UpdateInstanceStatus)
	admin.Get("/federation/outbound", federationAdminHandler.ListOutbound)
	admin.Get("/federation/outbound/:id", federationAdminHandler.GetOutbound)
	admin.Get("/federation/outbound/request/:requestId", federationAdminHandler.GetOutboundByRequestID)
	admin.Post("/federation/outbound/:id/retry", federationAdminHandler.RetryOutbound)
	admin.Get("/federation/reviews/pending", federationReviewHandler.ListPendingReviews)
	admin.Put("/federation/citations/:id/review", federationReviewHandler.ReviewCitation)
	admin.Put("/federation/mentions/:id/review", federationReviewHandler.ReviewMention)

	hitokotoSvc := hitokoto.NewService(deps.Redis, deps.Config.Redis.Prefix)
	hitokotoHandler := handler.NewAdminHitokotoHandler(hitokotoSvc)
	admin.Get("/hitokoto", hitokotoHandler.GetSentence)

	friendLinkAppRepo := persistence.NewFriendLinkApplicationRepository(deps.DB)
	friendLinkRepo := persistence.NewFriendLinkRepository(deps.DB)
	friendLinkAdminSvc := friendlink.NewAdminService(friendLinkAppRepo, friendLinkRepo, instanceRepo, identityRepo, deps.EventBus)
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

	globalNotificationRepo := persistence.NewGlobalNotificationRepository(deps.DB)
	globalNotificationSvc := globalnotification.NewService(globalNotificationRepo, deps.EventBus)
	globalNotificationHandler := handler.NewGlobalNotificationHandler(globalNotificationSvc)
	admin.Get("/global-notifications", globalNotificationHandler.ListAdmin)
	admin.Get("/global-notifications/:id", globalNotificationHandler.GetAdmin)
	admin.Post("/global-notifications", globalNotificationHandler.Create)
	admin.Put("/global-notifications/:id", globalNotificationHandler.Update)
	admin.Delete("/global-notifications/:id", globalNotificationHandler.Delete)

	logHandler := handler.NewAdminLogHandler("storage/logs/app.log", 200)
	systemHandler := handler.NewSystemHandler(deps.DB, deps.Redis, deps.EventBus)
	adminStatsSvc := adminstats.NewService(deps.DB, deps.Redis, deps.Config.Redis.Prefix, wsManager)
	adminStatsHandler := handler.NewAdminStatsHandler(adminStatsSvc)
	observabilityHandler := handler.NewAdminObservabilityHandler(deps.Observability)
	adminLogs := adminGroup.Group("/admin")
	adminLogs.Get("/logs", logHandler.List)
	adminLogs.Get("/system/status", systemHandler.GetStatus)
	adminLogs.Get("/stats/dashboard", adminStatsHandler.GetDashboard)
	adminLogs.Get("/observability/overview", observabilityHandler.GetOverview)
	adminLogs.Get("/observability/control-plane", observabilityHandler.GetControlPlane)
	adminLogs.Get("/observability/render-plane", observabilityHandler.GetRenderPlane)
	adminLogs.Get("/observability/realtime", observabilityHandler.GetRealtime)
	adminLogs.Get("/observability/federation", observabilityHandler.GetFederation)
	adminLogs.Get("/observability/storage", observabilityHandler.GetStorage)
	adminLogs.Get("/observability/timeline", observabilityHandler.GetTimeline)
	adminLogs.Get("/observability/alerts", observabilityHandler.GetAlerts)
	adminLogs.Get("/observability/pages", observabilityHandler.GetPageState)
	adminLogs.Post("/observability/pages/bootstrap", observabilityHandler.BootstrapPages)
	adminLogs.Post("/observability/pages/invalidate", observabilityHandler.InvalidatePages)
}

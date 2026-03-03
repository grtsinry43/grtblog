package router

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	appap "github.com/grtsinry43/grtblog-v2/server/internal/app/activitypub"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminnotification"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminstats"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminuser"
	appai "github.com/grtsinry43/grtblog-v2/server/internal/app/ai"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/email"
	appfed "github.com/grtsinry43/grtblog-v2/server/internal/app/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/friendlink"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/globalnotification"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/hitokoto"
	applike "github.com/grtsinry43/grtblog-v2/server/internal/app/like"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	fedinfra "github.com/grtsinry43/grtblog-v2/server/internal/infra/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
	"github.com/grtsinry43/grtblog-v2/server/internal/ws"
)

func registerAdminRoutes(v2 fiber.Router, deps Dependencies, websiteInfoHandler *handler.WebsiteInfoHandler, navMenuHandler *handler.NavMenuHandler, sysCfgSvc *sysconfig.Service, wsManager *ws.Manager, aiSvc *appai.Service) {
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminTokenRepo := persistence.NewAdminTokenRepository(deps.DB)
	authMiddleware := middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo)
	adminMiddleware := middleware.RequireAdmin(identityRepo)
	admin := v2.Group("/admin", authMiddleware, adminMiddleware)
	ownerStatusHandler := handler.NewOwnerStatusHandler(deps.OwnerStatus)
	v2.Post("/onlineStatus", authMiddleware, adminMiddleware, ownerStatusHandler.UpdateStatus)
	admin.Post("/owner-status/panel-heartbeat", ownerStatusHandler.PanelHeartbeat)

	websiteInfo := v2.Group("/website-info", authMiddleware, adminMiddleware)
	websiteInfo.Get("", websiteInfoHandler.List)
	websiteInfo.Put("/:key", websiteInfoHandler.Update)

	navMenus := admin.Group("/nav-menus")
	navMenus.Get("", navMenuHandler.ListAdmin)
	navMenus.Post("", navMenuHandler.Create)
	navMenus.Put("/reorder", navMenuHandler.Reorder)
	navMenus.Put("/:id", navMenuHandler.Update)
	navMenus.Delete("/:id", navMenuHandler.Delete)

	oauthRepo := persistence.NewOAuthProviderRepository(deps.DB)
	adminOAuth := handler.NewAdminOAuthHandler(oauthRepo)
	adminTokenHandler := handler.NewAdminTokenHandler(adminTokenRepo, identityRepo)
	htmlSnapshotHandler := handler.NewHTMLSnapshotHandler(deps.HTMLSnapshot, deps.ISR)
	admin.Post("/html/posts/refresh", htmlSnapshotHandler.RefreshPostsHTML)
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
	admin.Post("/comments/import", commentHandler.ImportComment)
	admin.Post("/comments/:id/reply", commentHandler.ReplyComment)
	admin.Put("/comments/:id/status", commentHandler.UpdateCommentStatus)
	admin.Put("/comments/:id/author", commentHandler.SetCommentAuthor)
	admin.Put("/comments/:id/top", commentHandler.SetCommentTop)
	admin.Delete("/comments/:id", commentHandler.DeleteComment)
	admin.Put("/comments/areas/:areaId/close", commentHandler.SetCommentAreaClose)

	likeRepo := persistence.NewLikeRepository(deps.DB)
	likeHandler := handler.NewLikeHandler(applike.NewService(likeRepo))
	admin.Post("/likes/import", likeHandler.ImportLikeBatch)

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
		emailSvc := email.NewService(emailRepo, emailSender, sysCfgSvc)
		emailHandler := handler.NewEmailTemplateHandler(emailSvc)
		admin.Get("/email/templates", emailHandler.ListEmailTemplates)
		admin.Post("/email/templates", emailHandler.CreateEmailTemplate)
		admin.Put("/email/templates/:code", emailHandler.UpdateEmailTemplate)
		admin.Delete("/email/templates/:code", emailHandler.DeleteEmailTemplate)
		admin.Post("/email/templates/:code/preview", emailHandler.PreviewEmailTemplate)
		admin.Post("/email/templates/:code/test", emailHandler.TestEmailTemplate)
		admin.Get("/email/subscriptions", emailHandler.ListEmailSubscriptions)
		admin.Put("/email/subscriptions/status", emailHandler.BatchUpdateEmailSubscriptionStatus)
		admin.Get("/email/outbox", emailHandler.ListEmailOutbox)
		admin.Get("/email/outbox/:id", emailHandler.GetEmailOutbox)
	}

	fedCfgHandler := handler.NewFederationConfigHandler(sysCfgSvc)
	activityPubCfgHandler := handler.NewActivityPubConfigHandler(sysCfgSvc)
	admin.Get("/federation/config", fedCfgHandler.ListFederationConfig)
	admin.Put("/federation/config", fedCfgHandler.UpdateFederationConfig)
	admin.Get("/activitypub/config", activityPubCfgHandler.ListActivityPubConfig)
	admin.Put("/activitypub/config", activityPubCfgHandler.UpdateActivityPubConfig)
	admin.Get("/federation/export", fedCfgHandler.ExportFederationConfigs)
	admin.Post("/federation/import", fedCfgHandler.ImportFederationConfigs)

	contentRepo := persistence.NewContentRepository(deps.DB)
	instanceRepo := persistence.NewFederationInstanceRepository(deps.DB)
	var cache fedinfra.Cache
	if deps.Redis != nil {
		cache = fedinfra.NewRedisCache(deps.Redis, deps.Config.Redis.Prefix)
	}
	resolver := fedinfra.NewResolver(&http.Client{Timeout: 10 * time.Second}, cache)
	outbound := appfed.NewOutboundService(sysCfgSvc, resolver, instanceRepo)
	outboundRepo := persistence.NewOutboundDeliveryRepository(deps.DB)
	deliverySvc := appfed.NewDeliveryService(outboundRepo, outbound, deps.EventBus)
	postCacheRepo := persistence.NewFederatedPostCacheRepository(deps.DB)
	federationAdminHandler := handler.NewFederationAdminHandler(sysCfgSvc, contentRepo, deliverySvc, instanceRepo, postCacheRepo, resolver, deps.EventBus)
	federationReviewHandler := handler.NewFederationReviewHandler(
		persistence.NewFederatedCitationRepository(deps.DB),
		persistence.NewFederatedMentionRepository(deps.DB),
		instanceRepo,
		outbound,
	)
	activityPubSvc := appap.NewService(
		sysCfgSvc,
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
	admin.Get("/federation/instances/:id/posts", federationAdminHandler.ListInstancePosts)
	admin.Put("/federation/instances/:id/status", federationAdminHandler.UpdateInstanceStatus)
	admin.Get("/federation/authors/search", federationAdminHandler.SearchAuthors)
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
	systemHandler := handler.NewSystemHandler(deps.Config.App, deps.DB, deps.Redis, deps.EventBus, deps.HealthState)
	adminStatsSvc := adminstats.NewService(deps.DB, deps.Redis, deps.Config.Redis.Prefix, wsManager)
	adminStatsHandler := handler.NewAdminStatsHandler(adminStatsSvc)
	observabilityHandler := handler.NewAdminObservabilityHandler(deps.Observability)
	admin.Get("/logs", logHandler.List)
	admin.Get("/system/status", systemHandler.GetStatus)
	admin.Get("/system/update-check", systemHandler.GetUpdateCheck)
	admin.Get("/stats/dashboard", adminStatsHandler.GetDashboard)
	admin.Get("/observability/overview", observabilityHandler.GetOverview)
	admin.Get("/observability/control-plane", observabilityHandler.GetControlPlane)
	admin.Get("/observability/render-plane", observabilityHandler.GetRenderPlane)
	admin.Get("/observability/realtime", observabilityHandler.GetRealtime)
	admin.Get("/observability/federation", observabilityHandler.GetFederation)
	admin.Get("/observability/storage", observabilityHandler.GetStorage)
	admin.Get("/observability/timeline", observabilityHandler.GetTimeline)
	admin.Get("/observability/alerts", observabilityHandler.GetAlerts)
	admin.Get("/observability/pages", observabilityHandler.GetPageState)
	admin.Post("/observability/pages/bootstrap", observabilityHandler.BootstrapPages)
	admin.Post("/observability/pages/invalidate", observabilityHandler.InvalidatePages)

	// AI 功能
	aiHandler := handler.NewAIHandler(aiSvc)
	admin.Get("/ai/providers", aiHandler.ListProviders)
	admin.Post("/ai/providers", aiHandler.CreateProvider)
	admin.Put("/ai/providers/:id", aiHandler.UpdateProvider)
	admin.Delete("/ai/providers/:id", aiHandler.DeleteProvider)
	admin.Get("/ai/models", aiHandler.ListModels)
	admin.Post("/ai/models", aiHandler.CreateModel)
	admin.Put("/ai/models/:id", aiHandler.UpdateModel)
	admin.Delete("/ai/models/:id", aiHandler.DeleteModel)
	admin.Post("/ai/moderate-comment", aiHandler.ModerateComment)
	admin.Post("/ai/generate-title", aiHandler.GenerateTitle)
	admin.Post("/ai/rewrite-content", aiHandler.RewriteContent)
	admin.Post("/ai/rewrite-content/stream", aiHandler.RewriteContentStream)
	admin.Post("/ai/generate-summary/stream", aiHandler.GenerateSummaryStream)
	admin.Get("/ai/task-logs", aiHandler.ListTaskLogs)
	admin.Get("/ai/task-logs/:id", aiHandler.GetTaskLog)
}

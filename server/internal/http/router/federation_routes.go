package router

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	appap "github.com/grtsinry43/grtblog-v2/server/internal/app/activitypub"
	appapcfg "github.com/grtsinry43/grtblog-v2/server/internal/app/activitypubconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminnotification"
	appfed "github.com/grtsinry43/grtblog-v2/server/internal/app/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/federationconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerFederationRoutes(app *fiber.App, deps Dependencies) {
	cfgRepo := persistence.NewFederationConfigRepository(deps.DB)
	cfgSvc := federationconfig.NewService(cfgRepo)
	apCfgSvc := appapcfg.NewService(cfgRepo)
	instanceRepo := persistence.NewFederationInstanceRepository(deps.DB)
	linkRepo := persistence.NewFriendLinkRepository(deps.DB)
	appRepo := persistence.NewFriendLinkApplicationRepository(deps.DB)
	contentRepo := persistence.NewContentRepository(deps.DB)
	commentRepo := persistence.NewCommentRepository(deps.DB)
	thinkingRepo := persistence.NewThinkingRepository(deps.DB)
	userRepo := persistence.NewIdentityRepository(deps.DB)
	citationRepo := persistence.NewFederatedCitationRepository(deps.DB)
	mentionRepo := persistence.NewFederatedMentionRepository(deps.DB)
	postCacheRepo := persistence.NewFederatedPostCacheRepository(deps.DB)
	outboundRepo := persistence.NewOutboundDeliveryRepository(deps.DB)
	apFollowerRepo := persistence.NewActivityPubFollowerRepository(deps.DB)
	apOutboxRepo := persistence.NewActivityPubOutboxRepository(deps.DB)
	adminNotifRepo := persistence.NewAdminNotificationRepository(deps.DB)
	adminNotifSvc := adminnotification.NewService(adminNotifRepo, deps.EventBus)

	var cache federation.Cache
	var rateLimiter federation.RateLimiter
	if deps.Redis != nil {
		cache = federation.NewRedisCache(deps.Redis, deps.Config.Redis.Prefix)
		rateLimiter = federation.NewRedisRateLimiter(deps.Redis, deps.Config.Redis.Prefix)
	}
	resolver := federation.NewResolver(&http.Client{Timeout: 10 * time.Second}, cache)
	verifier := federation.NewVerifier(resolver, 5*time.Minute)
	outbound := appfed.NewOutboundService(cfgSvc, resolver, instanceRepo)
	deliverySvc := appfed.NewDeliveryService(outboundRepo, outbound, deps.EventBus)

	wellKnownHandler := handler.NewFederationWellKnownHandler(cfgSvc, deps.Config.App)
	app.Get("/.well-known/blog-federation/manifest.json", wellKnownHandler.Manifest)
	app.Get("/.well-known/blog-federation/public-key.json", wellKnownHandler.PublicKey)
	app.Get("/.well-known/blog-federation/endpoints.json", wellKnownHandler.Endpoints)
	apSvc := appap.NewService(apCfgSvc, apFollowerRepo, apOutboxRepo, contentRepo, thinkingRepo, commentRepo, userRepo, adminNotifSvc)
	appap.RegisterSubscribers(deps.EventBus, apSvc)
	apHandler := handler.NewActivityPubHandler(apSvc)
	app.Get("/.well-known/nodeinfo", apHandler.NodeInfoDiscovery)
	app.Get("/nodeinfo/2.0", apHandler.NodeInfo20)
	app.Get("/.well-known/webfinger", apHandler.WebFinger)
	app.Get("/ap/actor", apHandler.Actor)
	app.Get("/ap/followers", apHandler.Followers)
	app.Get("/ap/outbox", apHandler.Outbox)
	app.Get("/ap/objects/:id", apHandler.Object)
	app.Post("/ap/inbox", apHandler.Inbox)

	federationGroup := app.Group("/api/federation")
	friendLinkHandler := handler.NewFederationFriendLinkHandler(cfgSvc, instanceRepo, linkRepo, appRepo, resolver, verifier, rateLimiter, deps.EventBus)
	federationGroup.Post("/friendlinks/request", friendLinkHandler.RequestFriendLink)

	timelineHandler := handler.NewFederationTimelineHandler(contentRepo, userRepo, cfgSvc)
	federationGroup.Get("/timeline/posts", timelineHandler.ListTimelinePosts)

	postHandler := handler.NewFederationPostHandler(contentRepo, userRepo, postCacheRepo, cfgSvc)
	federationGroup.Get("/posts/:id", postHandler.GetPostDetail)

	citationHandler := handler.NewFederationCitationHandler(cfgSvc, contentRepo, instanceRepo, citationRepo, linkRepo, resolver, verifier, rateLimiter, deps.EventBus)
	federationGroup.Post("/citations/request", citationHandler.RequestCitation)

	mentionHandler := handler.NewFederationMentionHandler(cfgSvc, instanceRepo, mentionRepo, userRepo, resolver, verifier, rateLimiter, deps.EventBus)
	federationGroup.Post("/mentions/notify", mentionHandler.NotifyMention)

	outboundResultHandler := handler.NewFederationOutboundResultHandler(deliverySvc, verifier)
	federationGroup.Post("/outbound/result", outboundResultHandler.ResultCallback)
}

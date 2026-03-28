package router

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	appap "github.com/grtsinry43/grtblog-v2/server/internal/app/activitypub"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminnotification"
	appfed "github.com/grtsinry43/grtblog-v2/server/internal/app/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerFederationRoutes(app *fiber.App, deps Dependencies) {
	sysCfgSvc := deps.SysConfig
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
	} else {
		rateLimiter = federation.NewInMemoryRateLimiter()
	}
	resolver := federation.NewResolver(&http.Client{Timeout: 10 * time.Second}, cache)
	verifier := federation.NewVerifier(resolver, 5*time.Minute)
	outbound := appfed.NewOutboundService(sysCfgSvc, resolver, instanceRepo)
	deliverySvc := appfed.NewDeliveryService(outboundRepo, outbound, linkRepo, deps.EventBus)

	wellKnownLimiter := limiter.New(limiter.Config{
		Max:        120,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})
	wellKnownHandler := handler.NewFederationWellKnownHandler(sysCfgSvc, deps.Config.App)
	app.Get("/.well-known/blog-federation/manifest.json", wellKnownLimiter, wellKnownHandler.Manifest)
	app.Get("/.well-known/blog-federation/public-key.json", wellKnownLimiter, wellKnownHandler.PublicKey)
	app.Get("/.well-known/blog-federation/endpoints.json", wellKnownLimiter, wellKnownHandler.Endpoints)
	apSvc := appap.NewService(sysCfgSvc, apFollowerRepo, apOutboxRepo, contentRepo, thinkingRepo, commentRepo, userRepo, adminNotifSvc)
	appap.RegisterSubscribers(deps.EventBus, apSvc)
	apHandler := handler.NewActivityPubHandler(apSvc)
	app.Get("/.well-known/nodeinfo", wellKnownLimiter, apHandler.NodeInfoDiscovery)
	app.Get("/nodeinfo/2.0", wellKnownLimiter, apHandler.NodeInfo20)
	app.Get("/.well-known/webfinger", wellKnownLimiter, apHandler.WebFinger)
	apLimiter := limiter.New(limiter.Config{
		Max:        60,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})
	app.Get("/ap/actor", apLimiter, apHandler.Actor)
	app.Get("/ap/followers", apLimiter, apHandler.Followers)
	app.Get("/ap/outbox", apLimiter, apHandler.Outbox)
	app.Get("/ap/objects/:id", apLimiter, apHandler.Object)
	app.Post("/ap/inbox", apLimiter, apHandler.Inbox)

	federationGroup := app.Group("/api/federation", limiter.New(limiter.Config{
		Max:        60,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}))
	friendLinkHandler := handler.NewFederationFriendLinkHandler(sysCfgSvc, instanceRepo, linkRepo, appRepo, resolver, verifier, rateLimiter, deps.EventBus)
	federationGroup.Post("/friendlinks/request", friendLinkHandler.RequestFriendLink)

	timelineHandler := handler.NewFederationTimelineHandler(contentRepo, userRepo, sysCfgSvc)
	federationGroup.Get("/timeline/posts", timelineHandler.ListTimelinePosts)

	postHandler := handler.NewFederationPostHandler(contentRepo, userRepo, postCacheRepo, sysCfgSvc)
	federationGroup.Get("/posts/:id", postHandler.GetPostDetail)

	citationHandler := handler.NewFederationCitationHandler(sysCfgSvc, contentRepo, instanceRepo, citationRepo, linkRepo, resolver, verifier, rateLimiter, deps.EventBus)
	federationGroup.Post("/citations/request", citationHandler.RequestCitation)

	mentionHandler := handler.NewFederationMentionHandler(sysCfgSvc, instanceRepo, mentionRepo, linkRepo, userRepo, resolver, verifier, rateLimiter, deps.EventBus)
	federationGroup.Post("/mentions/notify", mentionHandler.NotifyMention)

	outboundResultHandler := handler.NewFederationOutboundResultHandler(deliverySvc, verifier)
	federationGroup.Post("/outbound/result", outboundResultHandler.ResultCallback)
}

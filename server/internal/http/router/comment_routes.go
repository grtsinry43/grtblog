package router

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/clientinfo"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/geoip"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerCommentPublicRoutes(v2 fiber.Router, deps Dependencies) {
	commentHandler := newCommentHandler(deps)

	publicGroup := v2.Group("/comments")
	publicGroup.Get("/areas/:areaId", commentHandler.ListCommentTree)
	publicGroup.Post("/areas/:areaId/visitor", limiter.New(limiter.Config{
		Max:        20,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			handler.Audit(c, "comments.visitor.rate_limited", map[string]any{"ip": c.IP(), "areaId": c.Params("areaId")})
			return response.NewBizErrorWithMsg(response.TooManyRequests, "")
		},
	}), commentHandler.CreateCommentVisitor)
}

func registerCommentAuthRoutes(v2 fiber.Router, deps Dependencies) {
	commentHandler := newCommentHandler(deps)
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminTokenRepo := persistence.NewAdminTokenRepository(deps.DB)

	authGroup := v2.Group("/comments", middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo))
	authGroup.Post("/areas/:areaId", commentHandler.CreateCommentLogin)
}

func newCommentHandler(deps Dependencies) *handler.CommentHandler {
	commentRepo := persistence.NewCommentRepository(deps.DB)
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	friendLinkRepo := persistence.NewFriendLinkRepository(deps.DB)
	clientInfoResolver := clientinfo.NewUAParser()

	var geoResolver comment.GeoIPResolver
	if deps.Config.GeoIP.DBPath != "" {
		geoip.EnsureDatabasesAsync(
			context.Background(),
			deps.Config.GeoIP.DBPath,
			deps.Config.GeoIP.DownloadURL,
			deps.Config.GeoIP.ASNPath,
			deps.Config.GeoIP.ASNURL,
			log.Printf,
		)
		resolver, err := geoip.NewResolver(deps.Config.GeoIP.DBPath, deps.Config.GeoIP.ASNPath)
		if err != nil {
			log.Printf("geoip resolver init failed: %v", err)
			geoResolver = geoip.NewLazyResolver(deps.Config.GeoIP.DBPath, deps.Config.GeoIP.ASNPath)
		} else {
			geoResolver = resolver
		}
	}

	commentSvc := comment.NewService(commentRepo, identityRepo, friendLinkRepo, deps.SysConfig, clientInfoResolver, geoResolver, deps.EventBus)
	return handler.NewCommentHandler(commentSvc, deps.JWTManager)
}

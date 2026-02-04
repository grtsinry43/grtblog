package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/friendlink"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/htmlsnapshot"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerPublicRoutes(v2 fiber.Router, deps Dependencies, websiteInfoHandler *handler.WebsiteInfoHandler, htmlSnapshotSvc *htmlsnapshot.Service, navMenuHandler *handler.NavMenuHandler) {
	public := v2.Group("/public")
	public.Get("/website-info", websiteInfoHandler.PublicList)
	public.Get("/nav-menus", navMenuHandler.ListPublic)

	htmlSnapshotHandler := handler.NewHTMLSnapshotHandler(htmlSnapshotSvc)
	public.Post("/html/posts/refresh", htmlSnapshotHandler.RefreshPostsHTML)

	articleHandler := newArticleHandler(deps)
	public.Get("/articles/recent", articleHandler.ListRecentPublicArticles)

	momentHandler := newMomentHandler(deps)
	public.Get("/moments/recent", momentHandler.ListRecentPublicMoments)

	friendLinkRepo := persistence.NewFriendLinkRepository(deps.DB)
	friendLinkSvc := friendlink.NewLinkService(friendLinkRepo)
	friendLinkHandler := handler.NewFriendLinkPublicHandler(friendLinkSvc)
	public.Get("/friend-links", friendLinkHandler.ListPublic)

	if deps.Analytics != nil {
		analyticsHandler := handler.NewAnalyticsHandler(deps.Analytics)
		public.Post("/analytics/view", analyticsHandler.TrackView)
	}
}

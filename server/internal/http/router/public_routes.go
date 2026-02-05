package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/friendlink"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/friendtimeline"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/globalnotification"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/htmlsnapshot"
	appsearch "github.com/grtsinry43/grtblog-v2/server/internal/app/search"
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
	friendTimelineSvc := friendtimeline.NewService(persistence.NewFederatedPostCacheRepository(deps.DB), deps.Redis, deps.Config.Redis.Prefix)
	friendTimelineHandler := handler.NewFriendTimelineHandler(friendTimelineSvc)
	public.Get("/friend-timeline", friendTimelineHandler.ListPublic)

	globalNotificationRepo := persistence.NewGlobalNotificationRepository(deps.DB)
	globalNotificationSvc := globalnotification.NewService(globalNotificationRepo, deps.EventBus)
	globalNotificationHandler := handler.NewGlobalNotificationHandler(globalNotificationSvc)
	public.Get("/global-notifications", globalNotificationHandler.ListPublicActive)

	searchRepo := persistence.NewSearchRepository(deps.DB)
	searchSvc := appsearch.NewService(searchRepo, deps.Redis, deps.Config.Redis.Prefix)
	searchHandler := handler.NewSearchHandler(searchSvc)
	public.Get("/search", searchHandler.SiteSearch)

	if deps.Analytics != nil {
		analyticsHandler := handler.NewAnalyticsHandler(deps.Analytics)
		public.Post("/analytics/view", analyticsHandler.TrackView)
	}
}

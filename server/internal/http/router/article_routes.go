package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/article"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerArticlePublicRoutes(v2 fiber.Router, deps Dependencies) {
	articleHandler := newArticleHandler(deps)

	publicGroup := v2.Group("/articles")
	publicGroup.Get("/", articleHandler.ListArticles)  // GET /api/v2/articles
	publicGroup.Get("/:id", articleHandler.GetArticle) // GET /api/v2/articles/123
	publicGroup.Get("/:id/same-period-moments", articleHandler.ListSamePeriodMoments)
	publicGroup.Get("/short/:shortUrl", articleHandler.GetArticleByShortURL) // GET /api/v2/articles/short/abc123
	publicGroup.Post("/:id/latest", articleHandler.CheckArticleLatest)       // POST /api/v2/articles/123/latest
	publicGroup.Get("/:id/metrics", articleHandler.GetArticleMetrics)       // GET /api/v2/articles/123/metrics

	v2.Get("/categories/short/:shortUrl/articles", articleHandler.ListArticlesByCategoryShortURL)
}

func registerArticleAuthRoutes(v2 fiber.Router, deps Dependencies) {
	articleHandler := newArticleHandler(deps)
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminTokenRepo := persistence.NewAdminTokenRepository(deps.DB)

	authGroup := v2.Group("/articles", middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo), middleware.RequireAdmin(identityRepo))
	authGroup.Post("/", articleHandler.CreateArticle)      // POST /api/v2/articles
	authGroup.Put("/:id", articleHandler.UpdateArticle)    // PUT /api/v2/articles/123
	authGroup.Delete("/:id", articleHandler.DeleteArticle) // DELETE /api/v2/articles/123
	federationInteractionHandler := handler.NewFederationInteractionHandler(
		persistence.NewContentRepository(deps.DB),
		persistence.NewFederatedCitationRepository(deps.DB),
		persistence.NewOutboundDeliveryRepository(deps.DB),
	)
	authGroup.Get("/:id/federation/interactions", federationInteractionHandler.GetArticleInteractions)

	adminGroup := v2.Group("/admin", middleware.RequireAuth(deps.JWTManager, identityRepo, adminTokenRepo), middleware.RequireAdmin(identityRepo))
	adminGroup.Get("/articles/:id", articleHandler.GetArticleAdmin)
	adminGroup.Get("/articles", articleHandler.ListArticlesAdmin)                  // GET /api/v2/admin/articles
	adminGroup.Put("/articles/published", articleHandler.BatchSetArticlePublished) // PUT /api/v2/admin/articles/published
	adminGroup.Put("/articles/top", articleHandler.BatchSetArticleTop)             // PUT /api/v2/admin/articles/top
	adminGroup.Post("/articles/batch-delete", articleHandler.BatchDeleteArticles)  // POST /api/v2/admin/articles/batch-delete
	adminGroup.Post("/articles/:id/federation/signals/reset", articleHandler.ResetArticleFederationSignals)
}

func newArticleHandler(deps Dependencies) *handler.ArticleHandler {
	contentRepo := persistence.NewContentRepository(deps.DB)
	commentRepo := persistence.NewCommentRepository(deps.DB)
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	articleSvc := article.NewService(contentRepo, commentRepo, deps.EventBus)
	return handler.NewArticleHandler(articleSvc, contentRepo, commentRepo, identityRepo, deps.SysConfig,
		handler.WithFederationRepos(
			persistence.NewOutboundDeliveryRepository(deps.DB),
			persistence.NewFederatedPostCacheRepository(deps.DB),
			persistence.NewFederationInstanceRepository(deps.DB),
		),
	)
}

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
	publicGroup.Get("/", articleHandler.ListArticles)                        // GET /api/v2/articles
	publicGroup.Get("/:id", articleHandler.GetArticle)                       // GET /api/v2/articles/123
	publicGroup.Get("/short/:shortUrl", articleHandler.GetArticleByShortURL) // GET /api/v2/articles/short/abc123
	publicGroup.Post("/:id/latest", articleHandler.CheckArticleLatest)       // POST /api/v2/articles/123/latest
}

func registerArticleAuthRoutes(v2 fiber.Router, deps Dependencies) {
	articleHandler := newArticleHandler(deps)

	authGroup := v2.Group("/articles", middleware.RequireAuth(deps.JWTManager))
	authGroup.Post("/", articleHandler.CreateArticle)      // POST /api/v2/articles
	authGroup.Put("/:id", articleHandler.UpdateArticle)    // PUT /api/v2/articles/123
	authGroup.Delete("/:id", articleHandler.DeleteArticle) // DELETE /api/v2/articles/123

	identityRepo := persistence.NewIdentityRepository(deps.DB)
	adminGroup := v2.Group("", middleware.RequireAuth(deps.JWTManager), middleware.RequireAdmin(identityRepo))
	adminGroup.Get("/admin/articles", articleHandler.ListArticlesAdmin) // GET /api/v2/admin/articles
}

func newArticleHandler(deps Dependencies) *handler.ArticleHandler {
	contentRepo := persistence.NewContentRepository(deps.DB)
	commentRepo := persistence.NewCommentRepository(deps.DB)
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	articleSvc := article.NewService(contentRepo, commentRepo, deps.EventBus)
	return handler.NewArticleHandler(articleSvc, contentRepo, commentRepo, identityRepo)
}

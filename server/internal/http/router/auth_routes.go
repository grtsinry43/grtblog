package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/auth"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/setupstate"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

const (
	authCredentialRateMax = 10
	authProbeRateMax      = 60
	authRateWindow        = time.Minute
)

func registerAuthRoutes(v2 fiber.Router, deps Dependencies, sysCfgSvc *sysconfig.Service) {
	identityRepo := persistence.NewIdentityRepository(deps.DB)
	oauthRepo := persistence.NewOAuthProviderRepository(deps.DB)
	var stateStore auth.StateStore
	if deps.Redis != nil {
		stateStore = auth.NewRedisStateStore(deps.Redis, deps.Config.Redis.Prefix)
	}
	authSvc := auth.NewService(identityRepo, oauthRepo, deps.JWTManager, stateStore, deps.Config.Auth)
	setupStateSvc := setupstate.NewService(identityRepo, sysCfgSvc, deps.DB)
	authHandler := handler.NewAuthHandler(authSvc, setupStateSvc, sysCfgSvc, deps.Turnstile)
	oauthHandler := handler.NewOAuthHandler(authSvc, deps.Config.Auth.OAuthStateTTL)

	authGroup := v2.Group("/auth")

	// Strict: login / register (credential stuffing / brute-force).
	credentialLimiter := newRateLimiter(deps, limiter.Config{
		Max:        authCredentialRateMax,
		Expiration: authRateWindow,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "auth:credential:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			handler.Audit(c, "auth.credential.rate_limited", map[string]any{"ip": c.IP()})
			return response.NewBizErrorWithMsg(response.TooManyRequests, "")
		},
	})

	// Loose: setup probes / OAuth discovery (frontend may hit these often).
	probeLimiter := newRateLimiter(deps, limiter.Config{
		Max:        authProbeRateMax,
		Expiration: authRateWindow,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "auth:probe:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			handler.Audit(c, "auth.probe.rate_limited", map[string]any{"ip": c.IP()})
			return response.NewBizErrorWithMsg(response.TooManyRequests, "")
		},
	})

	authGroup.Post("/register", credentialLimiter, authHandler.Register)
	authGroup.Post("/login", credentialLimiter, authHandler.Login)

	authGroup.Get("/init-state", probeLimiter, authHandler.InitState)
	authGroup.Get("/setup-state", probeLimiter, authHandler.SetupState)
	authGroup.Get("/turnstile", probeLimiter, authHandler.TurnstileState)
	authGroup.Get("/providers", probeLimiter, oauthHandler.ListProviders)
	authGroup.Get("/providers/:provider/authorize", probeLimiter, oauthHandler.Authorize)
	authGroup.Post("/providers/:provider/callback", probeLimiter, oauthHandler.Callback)
}

package router

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/email"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/handler"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

func registerEmailPublicRoutes(v2 fiber.Router, deps Dependencies, sysCfgSvc *sysconfig.Service) {
	if sysCfgSvc == nil {
		return
	}
	emailRepo := persistence.NewEmailRepository(deps.DB)
	emailSender := email.NewSender(sysCfgSvc)
	websiteInfoRepo := persistence.NewWebsiteInfoRepository(deps.DB)
	emailSvc := email.NewService(emailRepo, emailSender, websiteInfoRepo)
	emailHandler := handler.NewEmailTemplateHandler(emailSvc)

	public := v2.Group("/public/email")
	public.Get("/events", emailHandler.ListPublicEmailEvents)

	guarded := public.Group("", limiter.New(limiter.Config{
		Max:        20,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			handler.Audit(c, "email.subscription.rate_limited", map[string]any{"ip": c.IP()})
			return response.NewBizErrorWithMsg(response.TooManyRequests, "")
		},
	}))
	guarded.Use(func(c *fiber.Ctx) error {
		blocked, err := sysCfgSvc.EmailSubscriptionBlockedIPs(c.Context())
		if err != nil {
			return err
		}
		ip := strings.TrimSpace(c.IP())
		for _, item := range blocked {
			if ip == strings.TrimSpace(item) {
				handler.Audit(c, "email.subscription.blocked_ip", map[string]any{"ip": ip})
				return response.NewBizErrorWithMsg(response.Unauthorized, "IP 已被限制访问")
			}
		}
		return c.Next()
	})
	guarded.Post("/subscriptions", emailHandler.SubscribeEmail)
	guarded.Post("/subscriptions/unsubscribe", emailHandler.UnsubscribeEmail)
}

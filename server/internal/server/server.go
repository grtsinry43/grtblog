package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/analytics"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/article"
	appfed "github.com/grtsinry43/grtblog-v2/server/internal/app/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/health"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/htmlsnapshot"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/isr"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/buildinfo"
	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/router"
	infraevent "github.com/grtsinry43/grtblog-v2/server/internal/infra/event"
	fedinfra "github.com/grtsinry43/grtblog-v2/server/internal/infra/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/metrics"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
	"github.com/grtsinry43/grtblog-v2/server/internal/security/jwt"
	"github.com/grtsinry43/grtblog-v2/server/internal/security/turnstile"
)

// Server wraps Fiber with configuration and dependencies.
type Server struct {
	cfg           config.Config
	db            *gorm.DB
	app           *fiber.App
	logFile       *os.File
	ctx           context.Context
	cancel        context.CancelFunc
	articleSvc    *article.Service
	sysCfgSvc     *sysconfig.Service
	analytics     *analytics.Service
	isrSvc        *isr.Service
	fedSync       *appfed.SyncWorker
	fedDeliver    *appfed.DeliveryService
	healthChecker *health.Checker
	version       string
}

// New builds a Fiber server with registered routes and middlewares.
func New(cfg config.Config, db *gorm.DB) *Server {
	if err := validateSecurityConfig(cfg); err != nil {
		log.Fatalf("invalid security configuration: %v", err)
	}

	logFile := initLogging()
	sysCfgRepo := persistence.NewSysConfigRepository(db)
	eventBus := infraevent.NewInMemoryBus()
	sysCfgSvc := sysconfig.NewService(sysCfgRepo, cfg.Turnstile, eventBus)
	contentRepo := persistence.NewContentRepository(db)
	commentRepo := persistence.NewCommentRepository(db)
	articleSvc := article.NewService(contentRepo, commentRepo, eventBus)

	ctx, cancel := context.WithCancel(context.Background())
	bodyLimit := sysCfgSvc.UploadMaxSizeBytes(ctx)

	app := fiber.New(fiber.Config{
		AppName:           cfg.App.Name,
		EnablePrintRoutes: cfg.App.Env == "development",
		BodyLimit:         bodyLimit,
		// Resolve client IP behind reverse proxies safely.
		ProxyHeader:             cfg.App.ProxyHeader,
		EnableIPValidation:      cfg.App.IPValidation,
		EnableTrustedProxyCheck: cfg.App.TrustedProxyCheck,
		TrustedProxies:          cfg.App.TrustedProxies,

		// 核心：全局错误处理，自动把业务错误包装成统一响应
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// 1. 我们自己抛出的业务错误：*response.AppError
			if ae, ok := err.(*response.AppError); ok {
				detail := fmt.Sprintf("biz=%s code=%d msg=%s", ae.Biz.BizErr, ae.Biz.Code, ae.Error())
				if ae.Cause != nil {
					detail = fmt.Sprintf("%s cause=%v", detail, ae.Cause)
				}
				logRequestError(c, "biz", detail)
				return response.ErrorWithMsg[any](c, ae.Biz, ae.Message)
			}

			// 2. Fiber 内置错误（比如 fiber.ErrNotFound / ErrMethodNotAllowed）
			if fe, ok := err.(*fiber.Error); ok {
				logRequestError(c, "http", fmt.Sprintf("status=%d msg=%s", fe.Code, fe.Message))
				// 这里可以按需映射到你的 BizError
				switch fe.Code {
				case fiber.StatusNotFound:
					return response.ErrorFromBiz[any](c, response.NotFound)
				case fiber.StatusMethodNotAllowed:
					return response.ErrorFromBiz[any](c, response.MethodNotAllowed)
				default:
					// 其他 HTTP 错误，统一当作 SERVER_ERROR 或自定义映射
					return response.ErrorFromBiz[any](c, response.ServerError)
				}
			}

			// 3. 其他未识别错误，统一视为服务器内部错误
			logRequestError(c, "unhandled", fmt.Sprintf("err=%v", err))
			return response.ErrorFromBiz[any](c, response.ServerError)
		},
	})

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			reqID, _ := c.Locals("requestId").(string)
			stack := debug.Stack()
			if reqID != "" {
				log.Printf("[panic] req=%s %s %s: %v\n%s", reqID, c.Method(), c.Path(), e, stack)
			} else {
				log.Printf("[panic] %s %s: %v\n%s", c.Method(), c.Path(), e, stack)
			}
		},
	}))

	// CORS: read allowed origins from sysconfig (site.public_url, site.api_url).
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			if origin == "" {
				return false
			}
			info, err := sysCfgSvc.WebsiteInfo(ctx)
			if err != nil {
				return false
			}
			for _, key := range []string{"public_url", "api_url"} {
				allowed := strings.TrimRight(strings.TrimSpace(info[key]), "/")
				if allowed != "" && strings.EqualFold(origin, allowed) {
					return true
				}
			}
			// In development mode, allow localhost origins.
			if strings.ToLower(cfg.App.Env) == "development" {
				return strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1")
			}
			return false
		},
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Request-ID",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	jwtManager := jwt.NewManager(cfg.Auth)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	turnstileClient := turnstile.NewClient(cfg.Turnstile)
	analyticsSvc := analytics.NewService(cfg, db, redisClient)
	htmlSnapshotSvc := htmlsnapshot.NewService(contentRepo, cfg.App.HTMLSnapshotBaseURL, redisClient, cfg.Redis.Prefix)
	isrSvc := isr.NewService(redisClient, cfg.Redis.Prefix, htmlSnapshotSvc, contentRepo)
	httpStats := metrics.NewHTTPStats(6 * time.Hour)
	fedResolver := fedinfra.NewResolver(&http.Client{Timeout: 10 * time.Second}, fedinfra.NewRedisCache(redisClient, cfg.Redis.Prefix))
	fedOutbound := appfed.NewOutboundService(sysCfgSvc, fedResolver, persistence.NewFederationInstanceRepository(db))
	fedDeliver := appfed.NewDeliveryService(persistence.NewOutboundDeliveryRepository(db), fedOutbound, eventBus)
	fedSync := appfed.NewSyncWorker(
		persistence.NewFederationInstanceRepository(db),
		persistence.NewFederatedPostCacheRepository(db),
		persistence.NewFriendLinkRepository(db),
		persistence.NewFriendLinkSyncJobRepository(db),
		fedResolver,
	)

	// Health state machine.
	isDev := strings.ToLower(cfg.App.Env) == "development"
	healthState := health.NewState(isDev)
	healthChecker := health.NewChecker(healthState, db, redisClient, sysCfgSvc, eventBus, 0, cfg.App.HTMLSnapshotBaseURL)

	app.Use(func(c *fiber.Ctx) error {
		if c.Locals("requestId") == nil {
			reqID := c.Get("X-Request-ID")
			if reqID == "" {
				reqID = uuid.NewString()
			}
			c.Locals("requestId", reqID)
		}
		return c.Next()
	})
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		status := c.Response().StatusCode()
		httpStats.Record(status, time.Since(start))
		return err
	})

	// 注册路由
	router.Register(app, router.Dependencies{
		DB:            db,
		Config:        cfg,
		JWTManager:    jwtManager,
		Turnstile:     turnstileClient,
		SysConfig:     sysCfgSvc,
		EventBus:      eventBus,
		Redis:         redisClient,
		Analytics:     analyticsSvc,
		HTTPStats:     httpStats,
		HTMLSnapshot:  htmlSnapshotSvc,
		ISR:           isrSvc,
		HealthState:   healthState,
		HealthChecker: healthChecker,
	})

	return &Server{
		cfg:           cfg,
		db:            db,
		app:           app,
		logFile:       logFile,
		ctx:           ctx,
		cancel:        cancel,
		articleSvc:    articleSvc,
		sysCfgSvc:     sysCfgSvc,
		analytics:     analyticsSvc,
		isrSvc:        isrSvc,
		fedSync:       fedSync,
		fedDeliver:    fedDeliver,
		healthChecker: healthChecker,
		version:       buildinfo.Version(),
	}
}

func validateSecurityConfig(cfg config.Config) error {
	env := strings.ToLower(strings.TrimSpace(cfg.App.Env))
	if env == "development" || env == "test" {
		return nil
	}

	secret := strings.TrimSpace(cfg.Auth.Secret)
	if secret == "" || secret == "change-me" {
		return fmt.Errorf("AUTH_SECRET must be explicitly configured outside development/test")
	}
	if len(secret) < 32 {
		return fmt.Errorf("AUTH_SECRET must be at least 32 characters outside development/test")
	}
	return nil
}

// Start launches the Fiber HTTP server and background workers.
func (s *Server) Start() error {
	// 启动健康状态检查
	if s.healthChecker != nil {
		go s.healthChecker.Run(s.ctx)
	}
	// 启动热门文章同步任务
	go s.runHotArticleSyncWorker()
	if s.analytics != nil {
		go s.analytics.RunViewEventWorker(s.ctx)
	}
	if s.isrSvc != nil {
		go s.runISRBootstrapIfNeeded()
		go s.isrSvc.RunWorker(s.ctx, 20, time.Second)
	}
	if s.fedSync != nil {
		go s.fedSync.Run(s.ctx, 30*time.Minute)
	}
	if s.fedDeliver != nil {
		go s.runFederationRetryWorker()
	}

	addr := fmt.Sprintf(":%s", s.cfg.App.Port)
	return s.app.Listen(addr)
}

func (s *Server) runFederationRetryWorker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			_ = s.fedDeliver.ProcessRetryQueue(s.ctx, 20)
		}
	}
}

func (s *Server) runISRBootstrapIfNeeded() {
	ctx, cancel := context.WithTimeout(s.ctx, 15*time.Minute)
	defer cancel()

	log.Printf("[isr] bootstrap check start")
	need, reason, err := s.isrSvc.NeedsBootstrapForVersion(ctx, s.version)
	if err != nil {
		log.Printf("[isr] bootstrap check failed: %v", err)
		return
	}
	if !need {
		snapshot, snapErr := s.isrSvc.Snapshot(ctx, 5, 5)
		if snapErr != nil {
			log.Printf("[isr] bootstrap skipped reason=%s", reason)
			return
		}
		log.Printf("[isr] bootstrap skipped reason=%s urlKeys=%d depKeys=%d queueDepth=%d", reason, snapshot.URLKeyCount, snapshot.DepKeyCount, snapshot.QueueDepth)
		return
	}

	log.Printf("[isr] bootstrap start reason=%s version=%s", reason, s.version)
	report, err := s.isrSvc.Bootstrap(ctx)
	if err != nil {
		log.Printf("[isr] bootstrap failed: %v", err)
		return
	}
	if err := s.isrSvc.MarkBootstrapVersion(ctx, s.version); err != nil {
		log.Printf("[isr] bootstrap version mark failed: %v", err)
	}
	log.Printf("[isr] bootstrap done routes=%d rendered=%d failed=%d durationMs=%d", report.TotalRoutes, report.RenderedCount, len(report.Failed), report.DurationMS)
}

// Shutdown gracefully stops Fiber and background workers.
func (s *Server) Shutdown(ctx context.Context) error {
	s.cancel() // 停止所有后台任务
	if s.logFile != nil {
		_ = s.logFile.Close()
	}
	return s.app.ShutdownWithContext(ctx)
}

// runHotArticleSyncWorker 定期同步热门文章状态
func (s *Server) runHotArticleSyncWorker() {
	log.Println("[worker] hot article sync worker started")
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	// 启动时立即执行一次
	s.syncHotArticles()

	for {
		select {
		case <-s.ctx.Done():
			log.Println("[worker] hot article sync worker stopped")
			return
		case <-ticker.C:
			s.syncHotArticles()
		}
	}
}

func (s *Server) syncHotArticles() {
	// 增加超时控制，防止单次同步阻塞整个 worker
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	thresholds := s.sysCfgSvc.HotArticleThresholds(ctx)
	err := s.articleSvc.UpdateHotArticles(ctx, thresholds.Views, thresholds.Likes, thresholds.Comments)
	if err != nil {
		log.Printf("[worker] failed to sync hot articles: %v", err)
	}
}

// App exposes the underlying Fiber instance for testing.
func (s *Server) App() *fiber.App {
	return s.app
}

func logRequestError(c *fiber.Ctx, kind string, detail string) {
	reqID, _ := c.Locals("requestId").(string)
	if reqID == "" {
		reqID = "-"
	}
	log.Printf("[error] req=%s %s %s kind=%s %s", reqID, c.Method(), c.Path(), kind, detail)
}

func initLogging() *os.File {
	logDir := filepath.Join("storage", "logs")
	_ = os.MkdirAll(logDir, 0o755)
	logPath := filepath.Join(logDir, "app.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Printf("failed to open log file: %v", err)
		return nil
	}
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC)
	return f
}

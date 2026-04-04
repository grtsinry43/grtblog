package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}

	if err := Migrate(ctx, pool); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	store := NewStore(pool)
	limiter := NewRateLimiter(ctx)
	passkeys := NewPasskeyService(store, cfg)

	app := fiber.New(fiber.Config{
		AppName:   "grtblog-telemetry",
		BodyLimit: 5 * 1024 * 1024,
	})

	app.Use(recover.New())

	// Public: literary page (always visible, permanent disguise).
	RegisterAdminPage(app)

	// Public: health check.
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// /collect needs CORS (called cross-origin by blog instances).
	collectCors := cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, User-Agent",
		AllowMethods: "POST, OPTIONS",
	})
	app.Post("/collect", collectCors, CollectHandler(store, limiter))

	// Passkey auth ceremony endpoints (same-origin, no CORS needed).
	app.Get("/auth/passkey/status", func(c *fiber.Ctx) error {
		has, _ := passkeys.store.HasAnyCredential(c.UserContext())
		return c.JSON(fiber.Map{"hasCredential": has})
	})
	auth := app.Group("/auth/passkey")
	auth.Post("/register/begin", passkeys.RegisterBeginHandler())
	auth.Post("/register/finish", passkeys.RegisterFinishHandler())
	auth.Post("/login/begin", passkeys.LoginBeginHandler())
	auth.Post("/login/finish", passkeys.LoginFinishHandler())

	// Grafana reverse proxy (gated behind Passkey session).
	RegisterGrafanaProxy(app, store, cfg.GrafanaURL)

	// Background: clean up expired sessions.
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := store.DeleteExpiredSessions(ctx); err != nil {
					log.Printf("[cleanup] expired sessions: %v", err)
				}
			}
		}
	}()

	// Graceful shutdown.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		cancel()
		_ = app.Shutdown()
	}()

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("telemetry collector listening on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/grtsinry43/grtblog-v2/server/internal/buildinfo"
	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/database"
	appserver "github.com/grtsinry43/grtblog-v2/server/internal/server"
)

// @title grtblog API v2
// @version 2.0.0
// @description grtblog 后端接口
// @BasePath /api/v2
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, falling back to system env vars")
	}

	// 现在 config.Load() 才能读到 .env 里的值
	cfg := config.Load()

	log.Println(startupBanner(cfg))

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	srv := appserver.New(cfg, db)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("server exiting: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func startupBanner(cfg config.Config) string {
	const apiBasePath = "/api/v2"
	version := buildinfo.Version()
	return "\n" +
		"================================================================\n" +
		"> " + cfg.App.Name + " " + version + "\n" +
		"> " + "不仅是博客，也是全新的内容基础设施。\n" +
		"\n" +
		"by @grtsinry43 · github.com/grtsinry43\n" +
		"“代码是写给人看的，顺便在机器上运行的。”\n" +
		"\n" +
		"- API 前缀: " + apiBasePath + "\n" +
		"- 监听端口: :" + cfg.App.Port + "\n" +
		"- 运行环境: " + cfg.App.Env + "\n" +
		"================================================================"
}

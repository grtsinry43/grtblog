package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"

	backupapp "github.com/grtsinry43/grtblog-v2/server/internal/app/backup"
	"github.com/grtsinry43/grtblog-v2/server/internal/config"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	timeout := cfg.Backup.CommandTimeout
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	log.Println("[restore] pending full-site restore detected")
	if err := backupapp.ExecutePendingRestore(ctx, cfg.Backup, cfg.Database.DSN, cfg.Redis); err != nil {
		log.Fatalf("[restore] failed: %v", err)
	}
	log.Println("[restore] full-site restore completed")
}

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"qr-parking/bot"
	"qr-parking/config"
	"qr-parking/db"
	"qr-parking/db/repositories"
	"qr-parking/pkg/logger"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	zapLogger, err := logger.New(cfg.App.Environment)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer zapLogger.Sync()

	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DB.URL, cfg.DB.MaxConns, cfg.DB.MinConns)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.URL,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close()

	tgRepo := repositories.NewTelegramRepo(pool)
	userRepo := repositories.NewUserRepo(pool)

	b, err := bot.New(cfg.Telegram.BotToken, tgRepo, userRepo, rdb, zapLogger)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go b.Start()

	<-quit
	zapLogger.Info("Shutting down bot...")
	b.Stop()
}

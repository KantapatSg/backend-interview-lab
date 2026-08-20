package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/KantapatSg/backend-interview-lab/platform/internal/adapter/postgres"
	cache "github.com/KantapatSg/backend-interview-lab/platform/internal/adapter/redis"
	"github.com/KantapatSg/backend-interview-lab/platform/internal/app"
	"github.com/KantapatSg/backend-interview-lab/platform/internal/httpapi"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	databaseURL := env("DATABASE_URL", "postgres://app:app@localhost:5432/interview_lab?sslmode=disable")
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		slog.Error("database pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		slog.Error("database ping", "error", err)
		os.Exit(1)
	}

	var redisCache *cache.Cache
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		redisCache = cache.New(addr)
		defer redisCache.Close()
	}
	service := app.NewJobService(postgres.New(pool), redisCache, time.Now)
	server := &http.Server{Addr: env("HTTP_ADDR", ":8080"), Handler: httpapi.NewServer(service, pool.Ping), ReadHeaderTimeout: 5 * time.Second}
	slog.Info("api listening", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("http server", "error", err)
		os.Exit(1)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

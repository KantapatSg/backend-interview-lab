package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	kafkatransport "github.com/KantapatSg/backend-interview-lab/platform/internal/adapter/kafka"
	"github.com/KantapatSg/backend-interview-lab/platform/internal/adapter/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, err := pgxpool.New(ctx, env("DATABASE_URL", "postgres://app:app@localhost:5432/interview_lab?sslmode=disable"))
	if err != nil {
		slog.Error("database pool", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		slog.Error("database ping", "error", err)
		os.Exit(1)
	}
	repo := postgres.New(db)
	producer := kafkatransport.NewProducer(env("KAFKA_BROKER", "localhost:29092"), env("KAFKA_TOPIC", "jobs.events.v1"))
	defer producer.Close()
	consumer := kafkatransport.NewConsumer(env("KAFKA_BROKER", "localhost:29092"), env("KAFKA_TOPIC", "jobs.events.v1"), env("KAFKA_GROUP", "job-worker"))
	defer consumer.Close()

	go publishOutbox(ctx, repo, producer)
	if err := consumer.Run(ctx, repo.ProcessJobEvent); err != nil && ctx.Err() == nil {
		slog.Error("consumer stopped", "error", err)
		os.Exit(1)
	}
}

func publishOutbox(ctx context.Context, repo *postgres.Repository, producer *kafkatransport.Producer) {
	interval, _ := strconv.Atoi(env("OUTBOX_POLL_MS", "500"))
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			events, err := repo.ClaimOutbox(ctx, 50)
			if err != nil {
				slog.Error("claim outbox", "error", err)
				continue
			}
			for _, event := range events {
				if err := producer.Publish(ctx, event); err != nil {
					slog.Error("publish outbox", "event_id", event.ID, "error", err)
					continue
				}
				if err := repo.MarkOutboxPublished(ctx, event.ID); err != nil {
					slog.Error("mark outbox", "event_id", event.ID, "error", err)
				}
			}
		}
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	neo4jadapter "github.com/trueconnect/backend/internal/adapter/neo4j"
	"github.com/trueconnect/backend/internal/adapter/postgres"
	redisadapter "github.com/trueconnect/backend/internal/adapter/redis"
	"github.com/trueconnect/backend/internal/config"
	"github.com/trueconnect/backend/internal/pkg/logger"
	"github.com/trueconnect/backend/internal/service"
	"github.com/trueconnect/backend/internal/worker"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	log := logger.New(cfg.Server.Env)
	slog.SetDefault(log)

	log.Info("starting TrueConnect worker",
		slog.String("env", cfg.Server.Env),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ── Data connections ──────────────────────────────────────────────

	pgPool, err := postgres.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		return fmt.Errorf("connecting to postgres: %w", err)
	}
	defer pgPool.Close()
	log.Info("connected to PostgreSQL")

	neo4jDriver, err := neo4jadapter.New(ctx, cfg.Neo4j.URI, cfg.Neo4j.User, cfg.Neo4j.Password)
	if err != nil {
		return fmt.Errorf("connecting to neo4j: %w", err)
	}
	defer neo4jDriver.Close(ctx)
	log.Info("connected to Neo4j")

	redisClient, err := redisadapter.New(ctx, cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		return fmt.Errorf("connecting to redis: %w", err)
	}
	defer redisClient.Close()
	log.Info("connected to Redis")

	// ── Repositories ─────────────────────────────────────────────────

	userRepo := postgres.NewUserRepo(pgPool)
	graphRepo := neo4jadapter.NewTrustGraphRepo(neo4jDriver)
	tokenRepo := postgres.NewRefreshTokenRepo(pgPool)
	sessionStore := redisadapter.NewSessionStore(redisClient)

	// Decode encryption key
	encryptionKey, err := hex.DecodeString(cfg.Auth.EncryptionKey)
	if err != nil {
		return fmt.Errorf("decoding encryption key: %w", err)
	}

	userSvc := service.NewUserService(userRepo, tokenRepo, sessionStore, graphRepo, encryptionKey)

	// ─── Workers ───────────────────────────────────────────────────────────────
	// NOTE: TrustEngine runs inside the API process (cmd/api), not here.
	// It consumes events from InteractionService which only exist in the API.
	// This worker binary runs scheduled background jobs only.

	sybilDetector := worker.NewSybilDetector(graphRepo, userRepo, adminRepo, userSvc, log, 6*time.Hour)
	go sybilDetector.Run(ctx)

	log.Info("worker running, waiting for shutdown signal")
	<-ctx.Done()

	log.Info("worker stopped")
	return nil
}

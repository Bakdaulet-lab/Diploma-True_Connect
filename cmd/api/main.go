package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/adapter/kyc"
	minioadapter "github.com/trueconnect/backend/internal/adapter/minio"
	neo4jadapter "github.com/trueconnect/backend/internal/adapter/neo4j"
	"github.com/trueconnect/backend/internal/adapter/postgres"
	redisadapter "github.com/trueconnect/backend/internal/adapter/redis"
	"github.com/trueconnect/backend/internal/config"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/handler"
	tcjwt "github.com/trueconnect/backend/internal/pkg/jwt"
	"github.com/trueconnect/backend/internal/pkg/logger"
	"github.com/trueconnect/backend/internal/provider"
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

	log.Info("starting TrueConnect API",
		slog.String("env", cfg.Server.Env),
		slog.Int("port", cfg.Server.Port),
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

	if err := neo4jadapter.SetupConstraints(ctx, neo4jDriver); err != nil {
		return fmt.Errorf("setting up neo4j constraints: %w", err)
	}
	log.Info("connected to Neo4j")

	redisClient, err := redisadapter.New(ctx, cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		return fmt.Errorf("connecting to redis: %w", err)
	}
	defer redisClient.Close()
	log.Info("connected to Redis")

	minioClient, err := minioadapter.New(ctx, cfg.MinIO.Endpoint, cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, cfg.MinIO.UseSSL, cfg.MinIO.Bucket)
	if err != nil {
		return fmt.Errorf("connecting to minio: %w", err)
	}
	log.Info("connected to MinIO")

	// ── Decode encryption key ─────────────────────────────────────────

	encryptionKey, err := hex.DecodeString(cfg.Auth.EncryptionKey)
	if err != nil {
		return fmt.Errorf("decoding encryption key: %w", err)
	}

	// ── Repositories ─────────────────────────────────────────────────

	userRepo := postgres.NewUserRepo(pgPool)
	tokenRepo := postgres.NewRefreshTokenRepo(pgPool)
	sessionStore := redisadapter.NewSessionStore(redisClient)
	graphRepo := neo4jadapter.NewTrustGraphRepo(neo4jDriver)

	// Sprint 2 repos
	profileRepo := postgres.NewProfileRepo(pgPool)
	mediaRepo := postgres.NewMediaRepo(pgPool)
	matchRepo := postgres.NewMatchRepo(pgPool)
	settingsRepo := postgres.NewSettingsRepo(pgPool)
	matchingCache := redisadapter.NewMatchingCache(redisClient)
	mediaStore := minioadapter.NewMediaStore(minioClient, cfg.MinIO.Bucket)

	// ── JWT manager ───────────────────────────────────────────────────

	jwtManager := tcjwt.NewManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenExpiry)

	// ── Services ─────────────────────────────────────────────────────

	authSvc := service.NewAuthService(
		userRepo,
		tokenRepo,
		sessionStore,
		graphRepo,
		jwtManager,
		encryptionKey,
		cfg.Auth.RefreshTokenExpiry,
		log,
	)

	// Sprint 2 services
	profileSvc := service.NewProfileService(profileRepo, mediaRepo, userRepo, mediaStore, matchingCache)
	matchingSvc := service.NewMatchingService(profileRepo, matchRepo, settingsRepo, matchingCache)
	settingsSvc := service.NewSettingsService(settingsRepo)

	// Sprint 3 repos, services, and trust engine
	interactionRepo := postgres.NewInteractionRepo(pgPool)
	eventCh := make(chan uuid.UUID, 100)
	uow := postgres.NewUoW(pgPool)
	interactionSvc := service.NewInteractionService(
		interactionRepo,
		matchRepo,
		graphRepo,
		uow,
		eventCh,
	)
	reputeSvc := service.NewReputationService(graphRepo, userRepo, matchingCache)

	trustEngine := worker.NewTrustEngine(eventCh, reputeSvc, log)
	go trustEngine.Run(ctx)

	// Push Notifications
	pushCh := make(chan domain.PushEvent, 100)
	var pushProvider provider.PushProvider
	if cfg.Firebase.CredentialsFile != "" {
		pushProvider, err = provider.NewFCMPushProvider(ctx, cfg.Firebase.CredentialsFile)
		if err != nil {
			return fmt.Errorf("initializing fcm: %w", err)
		}
		log.Info("initialized FCM push provider")
	} else {
		pushProvider = provider.NewMockPushProvider()
		log.Warn("using Mock push provider (no FIREBASE_CREDENTIALS set)")
	}

	pushWorker := worker.NewPushWorker(pushProvider, userRepo, pushCh, log)
	go pushWorker.Run(ctx)

	// Sprint 4 repos and services
	postRepo := postgres.NewPostRepo(pgPool)
	messageRepo := postgres.NewMessageRepo(pgPool)

	postSvc := service.NewPostService(postRepo)
	chatSvc := service.NewChatService(messageRepo, matchRepo, encryptionKey, pushCh)

	// ── Handlers ────────────────────────────────────────────────────────────────

	authHandler := handler.NewAuthHandler(authSvc, log, cfg.Server.Env == "development")

	// Sprint 2 handlers
	profileHandler := handler.NewProfileHandler(profileSvc, log)
	matchingHandler := handler.NewMatchingHandler(matchingSvc, log)
	settingsHandler := handler.NewSettingsHandler(settingsSvc, log)

	// Sprint 3 handler
	interactionHandler := handler.NewInteractionHandler(interactionSvc, reputeSvc, log)

	// Sprint 4 handlers
	postHandler := handler.NewPostHandler(postSvc, log)
	chatHub := handler.NewHub(chatSvc, matchingSvc, redisClient, jwtManager, log,
		cfg.Server.CORSOrigins, cfg.Server.Env == "development")

	kycProvider := kyc.NewSumsubProvider("dummy-token", "dummy-secret", log)
	kycHandler := handler.NewKYCHandler(mediaStore, userRepo, kycProvider, log)

	// Sprint 5 service and handler
	userSvc := service.NewUserService(userRepo, tokenRepo, sessionStore, graphRepo, encryptionKey)
	userHandler := handler.NewUserHandler(userSvc, log)

	reportRepo := postgres.NewReportRepo(pgPool)
	reportSvc := service.NewReportService(reportRepo, graphRepo)
	reportHandler := handler.NewReportHandler(reportSvc, log)

	adminHandler := handler.NewAdminHandler(userSvc, log)

	auditRepo := postgres.NewAuditRepo(pgPool)

	healthDeps := &handler.HealthDeps{
		PG:    pgPool,
		Neo4j: neo4jDriver,
		Redis: redisClient,
	}

	// ── Router ───────────────────────────────────────────────────────

	router := handler.NewRouter(&handler.RouterDeps{
		Health:      healthDeps,
		Auth:        authHandler,
		Profile:     profileHandler,
		Matching:    matchingHandler,
		Settings:    settingsHandler,
		Interaction: interactionHandler,
		Post:        postHandler,
		Chat:        chatHub,
		KYC:         kycHandler,
		User:        userHandler,
		Report:      reportHandler,
		Admin:       adminHandler,
		AuditRepo:   auditRepo,
		JWT:         jwtManager,
		Redis:       redisClient,
		CORSOrigins: cfg.Server.CORSOrigins,
		Log:         log,
	})

	// ── HTTP server ───────────────────────────────────────────────────

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("HTTP server listening", slog.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutting down server: %w", err)
	}

	log.Info("server stopped gracefully")
	return nil
}

package handler

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/trueconnect/backend/internal/handler/middleware"
	tcjwt "github.com/trueconnect/backend/internal/pkg/jwt"
	"github.com/trueconnect/backend/internal/repository"
)

// RouterDeps holds all handler and middleware dependencies needed to build the router.
type RouterDeps struct {
	Health      *HealthDeps
	Auth        *AuthHandler
	Profile     *ProfileHandler
	Matching    *MatchingHandler
	Settings    *SettingsHandler
	Interaction *InteractionHandler
	Post        *PostHandler
	Chat        *Hub
	KYC         *KYCHandler
	User        *UserHandler
	Report      *ReportHandler
	Admin       *AdminHandler
	AuditRepo   repository.AuditRepository
	JWT         *tcjwt.Manager
	Redis       *redis.Client
	CORSOrigins []string
	Log         *slog.Logger
}

// NewRouter creates the Gin engine with all routes and middleware registered.
func NewRouter(deps *RouterDeps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// Global middleware applied to every request.
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(deps.Log))
	r.Use(middleware.Recovery(deps.Log))
	r.Use(middleware.CORS(deps.CORSOrigins))

	// Global IP rate limit: 120 req/min — basic flood protection.
	r.Use(middleware.RateLimit(deps.Redis, middleware.RateLimitConfig{
		Requests: 120,
		Window:   time.Minute,
	}))

	v1 := r.Group("/v1")

	// ── Health check (public) ─────────────────────────────────────────
	v1.GET("/health", HealthHandler(deps.Health))

	// ── Auth (public, tighter rate limit: 20 req/min per IP) ─────────
	auth := v1.Group("/auth")
	auth.Use(middleware.RateLimit(deps.Redis, middleware.RateLimitConfig{
		Requests: 20,
		Window:   time.Minute,
	}))
	{
		auth.POST("/register", deps.Auth.Register)
		auth.POST("/login", deps.Auth.Login)
		auth.POST("/refresh", deps.Auth.Refresh)
		auth.POST("/logout", deps.Auth.Logout)
		auth.POST("/verify-phone", deps.Auth.VerifyPhone)
	}

	// ── Protected routes (JWT required) ──────────────────────────────
	protected := v1.Group("")
	protected.Use(middleware.Auth(deps.JWT))
	protected.Use(middleware.AuditLogMiddleware(deps.Log, deps.AuditRepo))

	// Per-user rate limit for write-heavy endpoints: 30 req/min.
	userRL := middleware.RateLimitByUser(deps.Redis, middleware.RateLimitConfig{
		Requests: 30,
		Window:   time.Minute,
	})

	{
		// ── Profiles ─────────────────────────────────────────────────
		protected.GET("/profiles/:id", deps.Profile.GetProfile)
		protected.PUT("/profiles/me", deps.Profile.UpsertProfile)
		protected.GET("/profiles/me/photos", deps.Profile.ListPhotos)
		protected.POST("/profiles/me/photos", deps.Profile.UploadPhoto)
		protected.DELETE("/profiles/me/photos/:photoID", deps.Profile.DeletePhoto)

		// ── Matching ─────────────────────────────────────────────────
		protected.GET("/matching/candidates", deps.Matching.GetCandidates)
		protected.POST("/matching/like", userRL, deps.Matching.Like)
		protected.POST("/matching/pass", userRL, deps.Matching.Pass)
		protected.GET("/matches", deps.Matching.ListMatches)

		// ── Settings ─────────────────────────────────────────────────
		protected.GET("/settings", deps.Settings.GetSettings)
		protected.PATCH("/settings", deps.Settings.UpdateSettings)

		// ── User Account ─────────────────────────────────────────────
		protected.GET("/users/me", deps.User.GetMe)
		protected.POST("/users/me/fcm-token", deps.User.UpdateFCMToken)
		protected.DELETE("/users/me", deps.User.DeleteMe)

		// ── Interactions & Reputation ────────────────────────────────
		protected.POST("/interactions", userRL, deps.Interaction.SubmitRating)
		protected.POST("/interactions/:id/confirm", deps.Interaction.ConfirmInteraction)
		protected.GET("/users/:id/reputation", deps.Interaction.GetReputation)

		// ── Social Feed (Posts) ──────────────────────────────────────
		protected.GET("/posts", deps.Post.ListFeed)
		protected.POST("/posts", userRL, deps.Post.CreatePost)
		protected.GET("/posts/:id", deps.Post.GetPost)
		protected.DELETE("/posts/:id", deps.Post.DeletePost)
		protected.POST("/posts/:id/like", userRL, deps.Post.LikePost)
		protected.DELETE("/posts/:id/like", deps.Post.UnlikePost)
		protected.POST("/posts/:id/comments", userRL, deps.Post.CreateComment)
		protected.GET("/posts/:id/comments", deps.Post.ListComments)

		// ── Chat (message history) ───────────────────────────────────
		protected.GET("/matches/:id/messages", deps.Chat.GetMessages)

		// ── KYC ──────────────────────────────────────────────────────
		protected.POST("/kyc/submit", userRL, deps.KYC.SubmitKYC)
		protected.GET("/kyc/status", deps.KYC.GetStatus)

		// ── Reports ──────────────────────────────────────────────────
		protected.POST("/reports", userRL, deps.Report.CreateReport)

		// ────── Admin / Review Workflow ──────────────────────────────────────────
		if deps.Admin != nil {
			adminGroup := protected.Group("/admin")
			adminGroup.GET("/users/under-review", deps.Admin.ListUnderReview)
			adminGroup.POST("/users/:id/review", deps.Admin.ReviewVerdict)
		}
	}

	// ── WebSocket (public route, auth via first message) ────────────
	v1.GET("/ws", deps.Chat.HandleWS)

	return r
}

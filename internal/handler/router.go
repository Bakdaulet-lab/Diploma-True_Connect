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
	Health       *HealthDeps
	Auth         *AuthHandler
	Profile      *ProfileHandler
	Matching     *MatchingHandler
	Settings     *SettingsHandler
	Interaction  *InteractionHandler
	Post         *PostHandler
	Chat         *Hub
	KYC          *KYCHandler
	Notification *NotificationHandler
	User         *UserHandler
	Report       *ReportHandler
	Admin        *AdminHandler
	AuditRepo    repository.AuditRepository
	JWT          *tcjwt.Manager
	Redis        *redis.Client
	CORSOrigins  []string
	Log          *slog.Logger
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

	// Global IP rate limit: 120 req/min вЂ” basic flood protection.
	r.Use(middleware.RateLimit(deps.Redis, middleware.RateLimitConfig{
		Requests: 120,
		Window:   time.Minute,
	}))

	v1 := r.Group("/v1")

	// в”Ђв”Ђ Health check (public) в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
	v1.GET("/health", HealthHandler(deps.Health))

	// в”Ђв”Ђ KYC Webhook (public) в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
	if deps.KYC != nil {
		v1.POST("/kyc/webhook", deps.KYC.HandleWebhook)
	}

	// в”Ђв”Ђ Auth (public, tighter rate limit: 20 req/min per IP) в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
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

	// в”Ђв”Ђ Protected routes (JWT required) в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
	protected := v1.Group("")
	protected.Use(middleware.Auth(deps.JWT))
	protected.Use(middleware.AuditLogMiddleware(deps.Log, deps.AuditRepo))

	// Per-user rate limit for write-heavy endpoints: 30 req/min.
	userRL := middleware.RateLimitByUser(deps.Redis, middleware.RateLimitConfig{
		Requests: 30,
		Window:   time.Minute,
	})

	{
		// в”Ђв”Ђ Profiles в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
		// РЎРќРђР§РђР›Рђ СЃС‚Р°С‚РёС‡РµСЃРєРёРµ РјР°СЂС€СЂСѓС‚С‹ (me)
		protected.GET("/profiles/me", deps.Profile.GetMyProfile) // <-- Р”РћР‘РђР’РРўР¬ Р­РўРЈ РЎРўР РћРљРЈ
		protected.PUT("/profiles/me", deps.Profile.UpsertProfile)
		protected.GET("/profiles/me/photos", deps.Profile.ListPhotos)
		protected.POST("/profiles/me/photos", deps.Profile.UploadPhoto)
		protected.DELETE("/profiles/me/photos/:photoID", deps.Profile.DeletePhoto)

		// РџРћРўРћРњ РґРёРЅР°РјРёС‡РµСЃРєРёРµ РјР°СЂС€СЂСѓС‚С‹ СЃ ID
		protected.GET("/profiles/:id", deps.Profile.GetProfile)

		// в”Ђв”Ђ Matching в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
		protected.GET("/matching/candidates", deps.Matching.GetCandidates)
		protected.GET("/matching/graph-candidates", deps.Matching.GetGraphCandidates)
		protected.POST("/matching/like", userRL, deps.Matching.Like)
		protected.POST("/matching/pass", userRL, deps.Matching.Pass)
		protected.GET("/matches", deps.Matching.ListMatches)

		// в”Ђв”Ђ Settings в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
		protected.GET("/settings", deps.Settings.GetSettings)
		protected.PATCH("/settings", deps.Settings.UpdateSettings)

		// в”Ђв”Ђ User Account в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
		protected.GET("/users/me", deps.User.GetMe)
		protected.POST("/users/me/fcm-token", deps.User.UpdateFCMToken)
		protected.DELETE("/users/me", deps.User.DeleteMe)

		// в”Ђв”Ђ Interactions & Reputation в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
		protected.POST("/interactions", userRL, deps.Interaction.SubmitRating)
		protected.POST("/interactions/:id/confirm", deps.Interaction.ConfirmInteraction)
		protected.GET("/users/:id/reputation", deps.Interaction.GetReputation)
		protected.GET("/reputation/leaderboard", deps.Interaction.GetLeaderboard)
		// в”Ђв”Ђ Social Feed (Posts) в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
		protected.GET("/posts", deps.Post.ListFeed)
		protected.POST("/posts", userRL, deps.Post.CreatePost)
		protected.GET("/posts/:id", deps.Post.GetPost)
		protected.DELETE("/posts/:id", deps.Post.DeletePost)
		protected.POST("/posts/:id/like", userRL, deps.Post.LikePost)
		protected.DELETE("/posts/:id/like", deps.Post.UnlikePost)
		protected.POST("/posts/:id/comments", userRL, deps.Post.CreateComment)
		protected.GET("/posts/:id/comments", deps.Post.ListComments)

		// в”Ђв”Ђ Chat (message history) в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
		protected.GET("/matches/:id/messages", deps.Chat.GetMessages)

		// в”Ђв”Ђ Notifications в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
		protected.GET("/notifications", deps.Notification.List)
		protected.PATCH("/notifications/:id/read", deps.Notification.MarkAsRead)
		protected.POST("/notifications/read-all", deps.Notification.MarkAllAsRead)
		protected.GET("/notifications/unread-count", deps.Notification.GetUnreadCount)

		// в”Ђв”Ђ KYC в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
		protected.POST("/kyc/submit", userRL, deps.KYC.SubmitKYC)
		protected.GET("/kyc/status", deps.KYC.GetStatus)

		// в”Ђв”Ђ Reports в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
		protected.POST("/reports", userRL, deps.Report.CreateReport)

		// в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ Admin / Review Workflow в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
		if deps.Admin != nil {
			adminGroup := protected.Group("/admin")
			adminGroup.GET("/users/under-review", deps.Admin.ListUnderReview)
			adminGroup.POST("/users/:id/review", deps.Admin.ReviewVerdict)
			adminGroup.GET("/sybil-clusters", deps.Admin.GetSybilClusters)
			adminGroup.GET("/analytics", deps.Admin.GetAnalytics)
			adminGroup.GET("/users", deps.Admin.SearchUsers)
			adminGroup.GET("/kyc/pending", deps.Admin.GetPendingKYC)
			adminGroup.GET("/kyc/:id/document", deps.Admin.GetKYCDocument)
			adminGroup.POST("/kyc/:id/review", deps.Admin.ReviewKYC)
		}
	}

	// ── WebSocket (public route, auth via first message) ──────────
	v1.GET("/ws", deps.Chat.HandleWS)

	return r
}

package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	redisadapter "github.com/trueconnect/backend/internal/adapter/redis"
	"github.com/trueconnect/backend/internal/domain"
	tcjwt "github.com/trueconnect/backend/internal/pkg/jwt"
	"github.com/trueconnect/backend/internal/repository"
)

const (
	// ContextKeyUserID is the gin context key for the authenticated user's UUID.
	ContextKeyUserID = "user_id"
	// ContextKeyVerificationLevel is the gin context key for the user's verification level.
	ContextKeyVerificationLevel = "verification_level"
	// ContextKeyTrustStatus is the gin context key for the user's trust status.
	ContextKeyTrustStatus = "trust_status"
	// ContextKeyIsAdmin is the gin context key for the admin flag.
	ContextKeyIsAdmin = "is_admin"
)

// AuthMiddleware validates JWT tokens and enforces trust status on every protected request.
type AuthMiddleware struct {
	jwt      *tcjwt.Manager
	userRepo repository.UserRepository
	tsCache  *redisadapter.TrustStatusCache
	log      *slog.Logger
}

// NewAuthMiddleware constructs an AuthMiddleware. tsCache may be nil (disables caching).
func NewAuthMiddleware(jwt *tcjwt.Manager, userRepo repository.UserRepository, tsCache *redisadapter.TrustStatusCache, log *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt, userRepo: userRepo, tsCache: tsCache, log: log}
}

// Authenticate returns a Gin handler that verifies the Bearer token, then checks
// trust_status from Redis (cache) → PostgreSQL (source of truth). Banned, suspended,
// and under_review accounts receive 403. DB errors are fail-closed (500).
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_token"})
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")

		claims, err := m.jwt.Verify(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_user_id"})
			return
		}

		status, err := m.fetchTrustStatus(c.Request.Context(), userID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user_not_found"})
				return
			}
			m.log.Error("trust_status lookup failed",
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()),
			)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}

		switch status {
		case string(domain.TrustStatusBanned),
			string(domain.TrustStatusSuspended),
			string(domain.TrustStatusUnderReview):
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":  "account_restricted",
				"status": status,
			})
			return
		}

		c.Set(ContextKeyUserID, userID)
		c.Set(ContextKeyVerificationLevel, claims.VerificationLevel)
		c.Set(ContextKeyTrustStatus, status)
		c.Set(ContextKeyIsAdmin, claims.IsAdmin)
		c.Next()
	}
}

// fetchTrustStatus resolves the user's trust_status: cache hit → return; cache miss or
// Redis error → DB read; DB error → fail-closed (return error).
func (m *AuthMiddleware) fetchTrustStatus(ctx context.Context, userID uuid.UUID) (string, error) {
	if m.tsCache != nil {
		cached, err := m.tsCache.Get(ctx, userID)
		if err == nil {
			return cached, nil
		}
		if !errors.Is(err, redisadapter.ErrTrustStatusCacheMiss) {
			m.log.Warn("trust_status cache get failed, falling back to DB",
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()),
			)
		}
	}

	user, err := m.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	status := string(user.TrustStatus)

	if m.tsCache != nil {
		if cacheErr := m.tsCache.Set(ctx, userID, status); cacheErr != nil {
			m.log.Warn("trust_status cache set failed",
				slog.String("user_id", userID.String()),
				slog.String("error", cacheErr.Error()),
			)
		}
	}

	return status, nil
}

// GetUserID extracts the authenticated user's UUID from the gin context.
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextKeyUserID)
	if !exists {
		return uuid.Nil, false
	}
	uid, ok := val.(uuid.UUID)
	return uid, ok
}

// RequireAdmin aborts with 403 if the authenticated user does not have the admin flag set.
// Must be placed after Authenticate() in the middleware chain.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, _ := c.Get(ContextKeyIsAdmin)
		if admin, ok := isAdmin.(bool); !ok || !admin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin_required"})
			return
		}
		c.Next()
	}
}

// Auth is deprecated — use NewAuthMiddleware(...).Authenticate() instead.
func Auth(jwtManager *tcjwt.Manager) gin.HandlerFunc {
	panic("middleware.Auth(jwt) is deprecated — use NewAuthMiddleware(...).Authenticate()")
}

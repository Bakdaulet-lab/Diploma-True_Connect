package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	tcjwt "github.com/trueconnect/backend/internal/pkg/jwt"
)

const (
	// ContextKeyUserID is the gin context key for the authenticated user's UUID.
	ContextKeyUserID = "user_id"
	// ContextKeyVerificationLevel is the gin context key for the user's verification level.
	ContextKeyVerificationLevel = "verification_level"
	// ContextKeyTrustStatus is the gin context key for the user's trust status.
	ContextKeyTrustStatus = "trust_status"
)

// Auth returns a middleware that validates JWT tokens from the Authorization header.
func Auth(jwtManager *tcjwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			abortUnauthorized(c, "missing authorization header")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			abortUnauthorized(c, "invalid authorization header format")
			return
		}

		claims, err := jwtManager.Verify(parts[1])
		if err != nil {
			abortUnauthorized(c, "invalid or expired token")
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			abortUnauthorized(c, "invalid token subject")
			return
		}

		// Check if the user is suspended/banned at the token level.
		// This avoids a DB hit for every request; the JWT is refreshed often enough
		// that a status change propagates within 15 minutes.
		if claims.TrustStatus == "suspended" || claims.TrustStatus == "banned" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "ACCOUNT_SUSPENDED",
					"message": "your account has been suspended",
				},
			})
			return
		}

		// Set authenticated context values
		c.Set(ContextKeyUserID, userID)
		c.Set(ContextKeyVerificationLevel, claims.VerificationLevel)
		c.Set(ContextKeyTrustStatus, claims.TrustStatus)

		c.Next()
	}
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

func abortUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{
			"code":    "UNAUTHORIZED",
			"message": message,
		},
	})
}

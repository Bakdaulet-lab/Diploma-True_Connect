package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/repository"
)

// AuditLogMiddleware creates an audit trail for important requests.
// Persists the access logs into identity_vault.access_log schema.
func AuditLogMiddleware(log *slog.Logger, auditRepo repository.AuditRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		method := c.Request.Method
		path := c.Request.URL.Path
		status := c.Writer.Status()

		if status >= 200 && status < 300 {
			uidStr := "unknown"
			var uid uuid.UUID
			if val, exists := c.Get(ContextKeyUserID); exists {
				if id, ok := val.(uuid.UUID); ok {
					uidStr = id.String()
					uid = id
				}
			}

			if method != "GET" {
				log.Info("AUDIT_TRAIL",
					slog.String("user_id", uidStr),
					slog.String("method", method),
					slog.String("path", path),
					slog.Int("status", status),
					slog.Duration("duration", duration),
					slog.String("ip", c.ClientIP()),
				)

				// Async save to database
				if uid != uuid.Nil {
					go func() {
						_ = auditRepo.LogAction(c.Copy(), &repository.AuditLog{
							UserID:    uid,
							Path:      path,
							Method:    method,
							IPAddress: c.ClientIP(),
						})
					}()
				}
			}
		}
	}
}

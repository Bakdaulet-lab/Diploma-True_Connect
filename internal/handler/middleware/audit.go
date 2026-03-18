package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuditLogMiddleware creates an audit trail for important requests.
// In a full system, this would write to a specialized database or log stream.
func AuditLogMiddleware(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		method := c.Request.Method
		path := c.Request.URL.Path
		status := c.Writer.Status()

		if status >= 200 && status < 300 {
			uidStr := "unknown"
			if val, exists := c.Get(ContextKeyUserID); exists {
				if id, ok := val.(uuid.UUID); ok {
					uidStr = id.String()
				}
			}

			// Do not flood audit logs with simple GETs unless required
			if method != "GET" {
				log.Info("AUDIT_TRAIL",
					slog.String("user_id", uidStr),
					slog.String("method", method),
					slog.String("path", path),
					slog.Int("status", status),
					slog.Duration("duration", duration),
					slog.String("ip", c.ClientIP()),
				)
			}
		}
	}
}

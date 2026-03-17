package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a middleware that logs each request using slog.
func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("ip", c.ClientIP()),
			slog.String("request_id", c.GetString("request_id")),
		}

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		log.LogAttrs(c.Request.Context(), level, "request", attrs...)
	}
}

// Recovery returns a middleware that recovers from panics and logs them.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic recovered",
					slog.Any("error", err),
					slog.String("path", c.Request.URL.Path),
					slog.String("request_id", c.GetString("request_id")),
				)
				c.AbortWithStatusJSON(500, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "an unexpected error occurred",
					},
				})
			}
		}()
		c.Next()
	}
}

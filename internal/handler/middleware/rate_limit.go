package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig holds the configuration for rate limiting.
type RateLimitConfig struct {
	// Requests is the maximum number of requests allowed in the window.
	Requests int64
	// Window is the sliding window duration.
	Window time.Duration
}

// RateLimit returns a middleware that enforces per-IP rate limiting using Redis.
func RateLimit(client *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		windowKey := time.Now().Truncate(cfg.Window).Unix()
		key := fmt.Sprintf("rl:ip:%s:%d", ip, windowKey)

		pipe := client.Pipeline()
		incrCmd := pipe.Incr(c.Request.Context(), key)
		pipe.Expire(c.Request.Context(), key, cfg.Window*2)
		_, err := pipe.Exec(c.Request.Context())
		if err != nil {
			// If Redis is down, fail open — don't block requests.
			c.Next()
			return
		}

		count := incrCmd.Val()
		if count > cfg.Requests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "too many requests, please try again later",
				},
			})
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Requests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", cfg.Requests-count))

		c.Next()
	}
}

// RateLimitByUser returns a middleware that enforces per-user rate limiting.
// Must be placed AFTER the Auth middleware so that user_id is available.
func RateLimitByUser(client *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.Next()
			return
		}

		windowKey := time.Now().Truncate(cfg.Window).Unix()
		key := fmt.Sprintf("rl:user:%s:%d", userID.String(), windowKey)

		pipe := client.Pipeline()
		incrCmd := pipe.Incr(c.Request.Context(), key)
		pipe.Expire(c.Request.Context(), key, cfg.Window*2)
		_, err := pipe.Exec(c.Request.Context())
		if err != nil {
			c.Next()
			return
		}

		count := incrCmd.Val()
		if count > cfg.Requests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "too many requests, please try again later",
				},
			})
			return
		}

		c.Next()
	}
}

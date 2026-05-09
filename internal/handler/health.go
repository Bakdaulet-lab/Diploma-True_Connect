package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"
)

// HealthDeps holds the dependencies needed for health checks.
type HealthDeps struct {
	PG    *pgxpool.Pool
	Neo4j neo4j.DriverWithContext
	Redis *redis.Client
}

// HealthHandler returns a handler that checks all service dependencies.
func HealthHandler(deps *HealthDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Check PostgreSQL
		if err := deps.PG.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "postgres: " + err.Error(),
			})
			return
		}

		// Check Neo4j
		if err := deps.Neo4j.VerifyConnectivity(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "neo4j: " + err.Error(),
			})
			return
		}

		// Check Redis
		if err := deps.Redis.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "redis: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}

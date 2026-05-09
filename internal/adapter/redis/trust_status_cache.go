package redisadapter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// TrustStatusCache caches the user's trust_status field in Redis to avoid
// hitting PostgreSQL on every authenticated request. Cache miss falls through
// to the database (the cache is an optimization, not a source of truth).
type TrustStatusCache struct {
	client *redis.Client
}

const (
	trustStatusKeyPrefix = "trust_status:"
	trustStatusTTL       = 60 * time.Second
)

// ErrTrustStatusCacheMiss indicates the requested key is not in the cache.
// Callers must handle it by reading from the source of truth (DB).
var ErrTrustStatusCacheMiss = errors.New("trust status cache miss")

// NewTrustStatusCache creates a new Redis-backed trust status cache.
func NewTrustStatusCache(client *redis.Client) *TrustStatusCache {
	return &TrustStatusCache{client: client}
}

// Get returns the cached trust_status string for a user, or ErrTrustStatusCacheMiss.
func (c *TrustStatusCache) Get(ctx context.Context, userID uuid.UUID) (string, error) {
	key := trustStatusKeyPrefix + userID.String()
	val, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrTrustStatusCacheMiss
	}
	if err != nil {
		return "", fmt.Errorf("trust status cache get: %w", err)
	}
	return val, nil
}

// Set caches the user's current trust_status with a short TTL.
func (c *TrustStatusCache) Set(ctx context.Context, userID uuid.UUID, status string) error {
	key := trustStatusKeyPrefix + userID.String()
	if err := c.client.Set(ctx, key, status, trustStatusTTL).Err(); err != nil {
		return fmt.Errorf("trust status cache set: %w", err)
	}
	return nil
}

// Delete invalidates the cached entry. Used after any change to user.trust_status
// so the next request reads the fresh value from the database.
func (c *TrustStatusCache) Delete(ctx context.Context, userID uuid.UUID) error {
	key := trustStatusKeyPrefix + userID.String()
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("trust status cache delete: %w", err)
	}
	return nil
}

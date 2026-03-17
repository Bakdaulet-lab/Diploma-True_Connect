package redisadapter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trueconnect/backend/internal/repository"
)

var _ repository.SessionStore = (*SessionStore)(nil)

// SessionStore manages refresh token tracking and rate limiting in Redis.
type SessionStore struct {
	client *redis.Client
}

// NewSessionStore creates a new Redis-backed session store.
func NewSessionStore(client *redis.Client) *SessionStore {
	return &SessionStore{client: client}
}

// StoreRefreshToken adds a refresh token hash to the user's active set.
func (s *SessionStore) StoreRefreshToken(ctx context.Context, userID string, tokenHash string, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s:refresh_tokens", userID)

	pipe := s.client.Pipeline()
	pipe.SAdd(ctx, key, tokenHash)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("storing refresh token in redis: %w", err)
	}

	return nil
}

// ValidateRefreshToken checks if a refresh token hash exists in the user's active set.
func (s *SessionStore) ValidateRefreshToken(ctx context.Context, userID string, tokenHash string) (bool, error) {
	key := fmt.Sprintf("session:%s:refresh_tokens", userID)

	exists, err := s.client.SIsMember(ctx, key, tokenHash).Result()
	if err != nil {
		return false, fmt.Errorf("validating refresh token in redis: %w", err)
	}

	return exists, nil
}

// RemoveRefreshToken removes a specific refresh token hash from the user's active set.
func (s *SessionStore) RemoveRefreshToken(ctx context.Context, userID string, tokenHash string) error {
	key := fmt.Sprintf("session:%s:refresh_tokens", userID)

	err := s.client.SRem(ctx, key, tokenHash).Err()
	if err != nil {
		return fmt.Errorf("removing refresh token from redis: %w", err)
	}

	return nil
}

// RemoveAllRefreshTokens clears all refresh tokens for a user (logout-all / compromise).
func (s *SessionStore) RemoveAllRefreshTokens(ctx context.Context, userID string) error {
	key := fmt.Sprintf("session:%s:refresh_tokens", userID)

	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("removing all refresh tokens from redis: %w", err)
	}

	return nil
}

// IncrementAuthFailure increments the auth failure counter for brute-force protection.
// Returns the new count. Key auto-expires after the window.
func (s *SessionStore) IncrementAuthFailure(ctx context.Context, phoneHash string, window time.Duration) (int64, error) {
	key := fmt.Sprintf("rl:auth:fail:%s", phoneHash)

	pipe := s.client.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("incrementing auth failure: %w", err)
	}

	return incrCmd.Val(), nil
}

// GetAuthFailureCount returns the current auth failure count for a phone hash.
func (s *SessionStore) GetAuthFailureCount(ctx context.Context, phoneHash string) (int64, error) {
	key := fmt.Sprintf("rl:auth:fail:%s", phoneHash)

	count, err := s.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("getting auth failure count: %w", err)
	}

	return count, nil
}

// ClearAuthFailures resets the failure counter after successful login.
func (s *SessionStore) ClearAuthFailures(ctx context.Context, phoneHash string) error {
	key := fmt.Sprintf("rl:auth:fail:%s", phoneHash)

	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("clearing auth failures: %w", err)
	}

	return nil
}

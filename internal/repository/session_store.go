package repository

import (
	"context"
	"time"
)

// SessionStore abstracts session and brute-force protection operations backed by Redis.
type SessionStore interface {
	StoreRefreshToken(ctx context.Context, userID string, tokenHash string, ttl time.Duration) error
	ValidateRefreshToken(ctx context.Context, userID string, tokenHash string) (bool, error)
	RemoveRefreshToken(ctx context.Context, userID string, tokenHash string) error
	RemoveAllRefreshTokens(ctx context.Context, userID string) error
	PublishUserBanned(ctx context.Context, userID string) error

	IncrementAuthFailure(ctx context.Context, phoneHash string, window time.Duration) (int64, error)
	GetAuthFailureCount(ctx context.Context, phoneHash string) (int64, error)
	ClearAuthFailures(ctx context.Context, phoneHash string) error
}

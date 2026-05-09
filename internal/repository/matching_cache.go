package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MatchingCache abstracts Redis operations for the matching feed.
type MatchingCache interface {
	// AddSeen marks one or more user IDs as already shown to the requester.
	AddSeen(ctx context.Context, requesterID uuid.UUID, seenIDs []uuid.UUID, ttl time.Duration) error

	// GetSeenIDs returns the set of user IDs already shown to the requester.
	GetSeenIDs(ctx context.Context, requesterID uuid.UUID) ([]uuid.UUID, error)

	// CacheTrustScore stores a user's trust score with a short TTL.
	CacheTrustScore(ctx context.Context, userID uuid.UUID, score int, ttl time.Duration) error

	// GetCachedTrustScore retrieves the cached trust score. Returns -1 if not cached.
	GetCachedTrustScore(ctx context.Context, userID uuid.UUID) (int, error)
}

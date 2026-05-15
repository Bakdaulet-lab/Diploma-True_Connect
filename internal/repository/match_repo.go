package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// MatchRepository manages swipe likes and mutual matches.
type MatchRepository interface {
	// RecordLike records that userID liked targetID.
	// Returns (matched=true, matchID) when both parties have now liked each other.
	RecordLike(ctx context.Context, userID, targetID uuid.UUID) (matched bool, matchID uuid.UUID, err error)

	// RecordPass persists that userID swiped left on targetID so they never
	// reappear in discovery even after Redis TTL expiry.
	RecordPass(ctx context.Context, userID, targetID uuid.UUID) error

	// GetRejectedIDs returns all target IDs that userID has ever passed on.
	GetRejectedIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)

	// ListMatches returns all mutual matches for a user, newest first.
	// Uses cursor-based pagination (matched_at or created_at).
	ListMatches(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]*domain.Match, string, error)

	// GetMatch returns a specific match only if userID is a participant.
	GetMatch(ctx context.Context, matchID uuid.UUID, userID uuid.UUID) (*domain.Match, error)

	// IsMatched reports whether two users have a mutual match.
	IsMatched(ctx context.Context, userA, userB uuid.UUID) (bool, error)

	// FindExpiredNiyyahMatches returns mutual matches whose 90-day niyyah timer
	// expired within the last 25 hours (so the daily worker processes each once).
	FindExpiredNiyyahMatches(ctx context.Context) ([]*domain.Match, error)

	// MarkFamilyIntroDone sets family_intro_done = true on the given match.
	MarkFamilyIntroDone(ctx context.Context, matchID uuid.UUID) error

	// MarkImamConfirmed sets imam_confirmed = true on the given match.
	MarkImamConfirmed(ctx context.Context, matchID uuid.UUID) error

	// BlockUser records blockerID blocking blockedID.
	BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error

	// GetBlockedIDs returns all user IDs that blockerID has blocked.
	GetBlockedIDs(ctx context.Context, blockerID uuid.UUID) ([]uuid.UUID, error)

	// Unmatch deletes a mutual match row (both parties lose the match).
	Unmatch(ctx context.Context, matchID, callerID uuid.UUID) error

	// GetPendingLikes returns user IDs who liked userID but haven't been liked back yet.
	GetPendingLikes(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

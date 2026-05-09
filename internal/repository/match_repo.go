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

	// RecordPass records that userID passed on targetID (no DB persistence needed,
	// handled via Redis seen-set; this method is a no-op stub for audit logging).
	RecordPass(ctx context.Context, userID, targetID uuid.UUID) error

	// ListMatches returns all mutual matches for a user, newest first.
	// Uses cursor-based pagination (matched_at or created_at).
	ListMatches(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]*domain.Match, string, error)

	// GetMatch returns a specific match only if userID is a participant.
	GetMatch(ctx context.Context, matchID uuid.UUID, userID uuid.UUID) (*domain.Match, error)

	// IsMatched reports whether two users have a mutual match.
	IsMatched(ctx context.Context, userA, userB uuid.UUID) (bool, error)
}

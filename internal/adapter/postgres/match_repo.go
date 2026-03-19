package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// MatchRepo implements repository.MatchRepository using PostgreSQL.
type MatchRepo struct {
	pool *pgxpool.Pool
}

var _ repository.MatchRepository = (*MatchRepo)(nil)

// NewMatchRepo creates a new PostgreSQL-backed match repository.
func NewMatchRepo(pool *pgxpool.Pool) *MatchRepo {
	return &MatchRepo{pool: pool}
}

// RecordLike records a like from userID → targetID.
// The matches table enforces user_a_id < user_b_id so we always sort the pair.
// Returns matched=true and the match ID when both users have now liked each other.
func (r *MatchRepo) RecordLike(ctx context.Context, userID, targetID uuid.UUID) (bool, uuid.UUID, error) {
	// Sort the pair so user_a_id < user_b_id (UUID lexicographic order).
	userA, userB := orderPair(userID, targetID)
	isUserA := userID == userA

	var setCol, checkCol string
	if isUserA {
		setCol, checkCol = "user_a_liked", "user_b_liked"
	} else {
		setCol, checkCol = "user_b_liked", "user_a_liked"
	}

	query := fmt.Sprintf(`
                INSERT INTO social.matches (user_a_id, user_b_id, %[1]s)
                VALUES ($1, $2, true)
                ON CONFLICT (user_a_id, user_b_id) DO UPDATE
                        SET %[1]s = true,
                            matched_at = CASE
                                WHEN social.matches.%[2]s = true THEN COALESCE(social.matches.matched_at, NOW())
                                ELSE social.matches.matched_at
                            END
                RETURNING id, %[2]s, matched_at`, setCol, checkCol)

        var matchID uuid.UUID
        var otherLiked bool
        var matchedAt interface{} // may be null

        err := runner(ctx, r.pool).QueryRow(ctx, query, userA, userB).Scan(&matchID, &otherLiked, &matchedAt)
        if err != nil {
                return false, uuid.Nil, fmt.Errorf("recording like: %w", err)
        }

        // The query atomically updates matched_at if both have liked.
        // We consider it a "new mutual match" if both have liked and matchedAt is NOT nil.
        isMatch := otherLiked && matchedAt != nil
        return isMatch, matchID, nil
}

// RecordPass is a no-op in PostgreSQL (the seen-set lives in Redis).
func (r *MatchRepo) RecordPass(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (r *MatchRepo) ListMatches(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Match, error) {
	query := `
		SELECT id, user_a_id, user_b_id, user_a_liked, user_b_liked, matched_at, created_at
		FROM social.matches
		WHERE (user_a_id = $1 OR user_b_id = $1)
		  AND matched_at IS NOT NULL
		ORDER BY matched_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := runner(ctx, r.pool).Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing matches: %w", err)
	}
	defer rows.Close()

	var matches []*domain.Match
	for rows.Next() {
		m := &domain.Match{}
		if err := rows.Scan(
			&m.ID, &m.UserAID, &m.UserBID,
			&m.UserALiked, &m.UserBLiked,
			&m.MatchedAt, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning match row: %w", err)
		}
		matches = append(matches, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating matches: %w", err)
	}

	return matches, nil
}

func (r *MatchRepo) GetMatch(ctx context.Context, matchID uuid.UUID, userID uuid.UUID) (*domain.Match, error) {
	query := `
		SELECT id, user_a_id, user_b_id, user_a_liked, user_b_liked, matched_at, created_at
		FROM social.matches
		WHERE id = $1 AND (user_a_id = $2 OR user_b_id = $2)`

	m := &domain.Match{}
	err := runner(ctx, r.pool).QueryRow(ctx, query, matchID, userID).Scan(
		&m.ID, &m.UserAID, &m.UserBID,
		&m.UserALiked, &m.UserBLiked,
		&m.MatchedAt, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("getting match: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("getting match: %w", err)
	}

	return m, nil
}

func (r *MatchRepo) IsMatched(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	a, b := orderPair(userA, userB)
	query := `
		SELECT 1 FROM social.matches
		WHERE user_a_id = $1 AND user_b_id = $2 AND matched_at IS NOT NULL
		LIMIT 1`

	var dummy int
	err := runner(ctx, r.pool).QueryRow(ctx, query, a, b).Scan(&dummy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("checking match: %w", err)
	}

	return true, nil
}

// orderPair returns (smaller, larger) UUID so the pair is always consistently ordered.
func orderPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if a.String() < b.String() {
		return a, b
	}
	return b, a
}

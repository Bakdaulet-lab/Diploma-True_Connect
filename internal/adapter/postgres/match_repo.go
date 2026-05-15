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

// RecordLike records a like from userID → targetID inside an explicit transaction
// with a row-level lock to prevent the race where two concurrent likes both see
// matched_at IS NULL before either has committed, causing duplicate match events.
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

	// Use an advisory lock keyed on the sorted UUID pair to serialize concurrent
	// likes for the same user pair without blocking unrelated pairs.
	lockKey := int64(userA[0])<<56 | int64(userA[1])<<48 | int64(userB[0])<<8 | int64(userB[1])

	upsertQuery := fmt.Sprintf(`
		INSERT INTO social.matches (user_a_id, user_b_id, %[1]s)
		VALUES ($1, $2, true)
		ON CONFLICT (user_a_id, user_b_id) DO UPDATE
			SET %[1]s = true,
			    matched_at = CASE
			        WHEN social.matches.%[2]s = true THEN COALESCE(social.matches.matched_at, NOW())
			        ELSE social.matches.matched_at
			    END,
			    niyyah_timer_ends_at = CASE
			        WHEN social.matches.%[2]s = true THEN COALESCE(social.matches.niyyah_timer_ends_at, NOW() + INTERVAL '90 days')
			        ELSE social.matches.niyyah_timer_ends_at
			    END
		RETURNING id, %[2]s, matched_at`, setCol, checkCol)

	var matchID uuid.UUID
	var otherLiked bool
	var matchedAt interface{}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, uuid.Nil, fmt.Errorf("recording like: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, lockKey); err != nil {
		return false, uuid.Nil, fmt.Errorf("recording like: advisory lock: %w", err)
	}

	if err := tx.QueryRow(ctx, upsertQuery, userA, userB).Scan(&matchID, &otherLiked, &matchedAt); err != nil {
		return false, uuid.Nil, fmt.Errorf("recording like: upsert: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, uuid.Nil, fmt.Errorf("recording like: commit: %w", err)
	}

	isMatch := otherLiked && matchedAt != nil
	return isMatch, matchID, nil
}

// RecordPass persists the pass so the candidate never reappears after Redis TTL.
func (r *MatchRepo) RecordPass(ctx context.Context, userID, targetID uuid.UUID) error {
	const q = `
		INSERT INTO social.swipe_rejections (user_id, target_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`
	if _, err := runner(ctx, r.pool).Exec(ctx, q, userID, targetID); err != nil {
		return fmt.Errorf("recording pass: %w", err)
	}
	return nil
}

// GetRejectedIDs returns all target IDs the user has ever passed on.
func (r *MatchRepo) GetRejectedIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	const q = `SELECT target_id FROM social.swipe_rejections WHERE user_id = $1`
	rows, err := runner(ctx, r.pool).Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("getting rejected IDs: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning rejected ID: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *MatchRepo) ListMatches(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]*domain.Match, string, error) {
	query := `
		SELECT id, user_a_id, user_b_id, user_a_liked, user_b_liked, matched_at, niyyah_timer_ends_at, created_at
		FROM social.matches
		WHERE (user_a_id = $1 OR user_b_id = $1)
		  AND matched_at IS NOT NULL
		  AND ($3::timestamptz IS NULL OR matched_at < $3::timestamptz)
		ORDER BY matched_at DESC
		LIMIT $2`

	var cursorArg interface{}
	if cursor == "" {
		cursorArg = nil
	} else {
		cursorArg = cursor
	}

	rows, err := runner(ctx, r.pool).Query(ctx, query, userID, limit, cursorArg)
	if err != nil {
		return nil, "", fmt.Errorf("listing matches: %w", err)
	}
	defer rows.Close()

	var matches []*domain.Match
	for rows.Next() {
		m := &domain.Match{}
		if err := rows.Scan(
			&m.ID, &m.UserAID, &m.UserBID,
			&m.UserALiked, &m.UserBLiked,
			&m.MatchedAt, &m.NiyyahTimerEndsAt, &m.CreatedAt,
		); err != nil {
			return nil, "", fmt.Errorf("scanning match row: %w", err)
		}
		matches = append(matches, m)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterating matches: %w", err)
	}

	nextCursor := ""
	if len(matches) == limit {
		nextCursor = matches[len(matches)-1].MatchedAt.Format("2006-01-02T15:04:05.999999Z07:00")
	}

	return matches, nextCursor, nil
}

// ListMatchViews returns matches joined with the other user's profile in one query.
func (r *MatchRepo) ListMatchViews(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]*repository.MatchViewRow, string, error) {
	query := `
		SELECT
			m.id                              AS match_id,
			CASE WHEN m.user_a_id = $1 THEN m.user_b_id ELSE m.user_a_id END AS other_user_id,
			COALESCE(p.display_name, '')      AS display_name,
			COALESCE(p.avatar_url, '')        AS avatar_url,
			COALESCE(u.trust_score, 0)        AS trust_score,
			u.public_key,
			m.niyyah_timer_ends_at,
			m.matched_at
		FROM social.matches m
		JOIN social.users u
			ON u.id = CASE WHEN m.user_a_id = $1 THEN m.user_b_id ELSE m.user_a_id END
		LEFT JOIN social.profiles p
			ON p.user_id = u.id
		WHERE (m.user_a_id = $1 OR m.user_b_id = $1)
		  AND m.matched_at IS NOT NULL
		  AND ($3::timestamptz IS NULL OR m.matched_at < $3::timestamptz)
		ORDER BY m.matched_at DESC
		LIMIT $2`

	var cursorArg interface{}
	if cursor != "" {
		cursorArg = cursor
	}

	rows, err := runner(ctx, r.pool).Query(ctx, query, userID, limit, cursorArg)
	if err != nil {
		return nil, "", fmt.Errorf("listing match views: %w", err)
	}
	defer rows.Close()

	var result []*repository.MatchViewRow
	for rows.Next() {
		row := &repository.MatchViewRow{}
		if err := rows.Scan(
			&row.MatchID, &row.OtherUserID,
			&row.DisplayName, &row.AvatarURL,
			&row.TrustScore, &row.PublicKey,
			&row.NiyyahTimerEndsAt, &row.MatchedAt,
		); err != nil {
			return nil, "", fmt.Errorf("scanning match view: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterating match views: %w", err)
	}

	nextCursor := ""
	if len(result) == limit && result[len(result)-1].MatchedAt != nil {
		nextCursor = result[len(result)-1].MatchedAt.Format("2006-01-02T15:04:05.999999Z07:00")
	}
	return result, nextCursor, nil
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

// FindExpiredNiyyahMatches returns mutual matches whose 90-day niyyah timer
// expired within the last 25 hours so the daily worker processes each once.
func (r *MatchRepo) FindExpiredNiyyahMatches(ctx context.Context) ([]*domain.Match, error) {
	query := `
		SELECT id, user_a_id, user_b_id, user_a_liked, user_b_liked, matched_at, niyyah_timer_ends_at, created_at
		FROM social.matches
		WHERE niyyah_timer_ends_at IS NOT NULL
		  AND niyyah_timer_ends_at < NOW()
		  AND niyyah_timer_ends_at > NOW() - INTERVAL '25 hours'
		  AND matched_at IS NOT NULL`

	rows, err := runner(ctx, r.pool).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("finding expired niyyah matches: %w", err)
	}
	defer rows.Close()

	var matches []*domain.Match
	for rows.Next() {
		m := &domain.Match{}
		if err := rows.Scan(
			&m.ID, &m.UserAID, &m.UserBID,
			&m.UserALiked, &m.UserBLiked,
			&m.MatchedAt, &m.NiyyahTimerEndsAt, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning expired match: %w", err)
		}
		matches = append(matches, m)
	}
	return matches, rows.Err()
}

// MarkFamilyIntroDone sets family_intro_done = true on the match.
func (r *MatchRepo) MarkFamilyIntroDone(ctx context.Context, matchID uuid.UUID) error {
	const q = `UPDATE social.matches SET family_intro_done = true WHERE id = $1`
	if _, err := runner(ctx, r.pool).Exec(ctx, q, matchID); err != nil {
		return fmt.Errorf("marking family intro done: %w", err)
	}
	return nil
}

// MarkImamConfirmed sets imam_confirmed = true on the match.
func (r *MatchRepo) MarkImamConfirmed(ctx context.Context, matchID uuid.UUID) error {
	const q = `UPDATE social.matches SET imam_confirmed = true WHERE id = $1`
	if _, err := runner(ctx, r.pool).Exec(ctx, q, matchID); err != nil {
		return fmt.Errorf("marking imam confirmed: %w", err)
	}
	return nil
}

// BlockUser records that blockerID has blocked blockedID.
func (r *MatchRepo) BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	const q = `
		INSERT INTO social.blocked_users (blocker_id, blocked_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`
	if _, err := runner(ctx, r.pool).Exec(ctx, q, blockerID, blockedID); err != nil {
		return fmt.Errorf("blocking user: %w", err)
	}
	return nil
}

// GetBlockedIDs returns all user IDs in any block relationship with userID:
// users that userID blocked AND users who blocked userID.
// Both directions are excluded from the discovery feed.
func (r *MatchRepo) GetBlockedIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	const q = `
		SELECT blocked_id FROM social.blocked_users WHERE blocker_id = $1
		UNION
		SELECT blocker_id FROM social.blocked_users WHERE blocked_id = $1`
	rows, err := runner(ctx, r.pool).Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("getting blocked IDs: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning blocked ID: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// UnmatchByUsers removes any match row between the two users regardless of order.
func (r *MatchRepo) UnmatchByUsers(ctx context.Context, userA, userB uuid.UUID) error {
	a, b := orderPair(userA, userB)
	const q = `DELETE FROM social.matches WHERE user_a_id = $1 AND user_b_id = $2`
	_, err := runner(ctx, r.pool).Exec(ctx, q, a, b)
	if err != nil {
		return fmt.Errorf("unmatching by users: %w", err)
	}
	return nil
}

// Unmatch removes a mutual match. callerID must be one of the participants.
func (r *MatchRepo) Unmatch(ctx context.Context, matchID, callerID uuid.UUID) error {
	const q = `
		DELETE FROM social.matches
		WHERE id = $1 AND (user_a_id = $2 OR user_b_id = $2)`
	tag, err := runner(ctx, r.pool).Exec(ctx, q, matchID, callerID)
	if err != nil {
		return fmt.Errorf("unmatching: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("unmatching: %w", domain.ErrNotFound)
	}
	return nil
}

// GetPendingLikes returns IDs of users who liked userID but haven't been liked back.
func (r *MatchRepo) GetPendingLikes(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	const q = `
		SELECT
			CASE WHEN user_b_id = $1 THEN user_a_id ELSE user_b_id END AS liker_id
		FROM social.matches
		WHERE
			(user_b_id = $1 AND user_a_liked = true AND user_b_liked = false)
			OR
			(user_a_id = $1 AND user_b_liked = true AND user_a_liked = false)
		ORDER BY created_at DESC
		LIMIT 50`

	rows, err := runner(ctx, r.pool).Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("getting pending likes: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning liker id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// orderPair returns (smaller, larger) UUID so the pair is always consistently ordered.
func orderPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if a.String() < b.String() {
		return a, b
	}
	return b, a
}

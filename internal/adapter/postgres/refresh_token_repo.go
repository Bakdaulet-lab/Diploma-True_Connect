package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// RefreshTokenRepo handles refresh token persistence in PostgreSQL.
type RefreshTokenRepo struct {
	pool *pgxpool.Pool
}

var _ repository.RefreshTokenRepository = (*RefreshTokenRepo)(nil)

// NewRefreshTokenRepo creates a new PostgreSQL-backed refresh token repository.
func NewRefreshTokenRepo(pool *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{pool: pool}
}

// Create stores a new refresh token hash.
func (r *RefreshTokenRepo) Create(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) (uuid.UUID, error) {
	query := `
		INSERT INTO social.refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id`

	var id uuid.UUID
	err := runner(ctx, r.pool).QueryRow(ctx, query, userID, tokenHash, expiresAt).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("creating refresh token: %w", err)
	}

	return id, nil
}

// GetByTokenHash retrieves a token record by its SHA-256 hash.
// Returns ErrNotFound if the token doesn't exist or has expired.
// Returns tokens regardless of revocation status for reuse detection at the service layer.
func (r *RefreshTokenRepo) GetByTokenHash(ctx context.Context, tokenHash []byte) (*repository.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked
		FROM social.refresh_tokens
		WHERE token_hash = $1 AND expires_at > NOW()`

	token := &repository.RefreshToken{}
	err := runner(ctx, r.pool).QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.Revoked,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("getting refresh token: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("getting refresh token: %w", err)
	}

	return token, nil
}

// Revoke marks a specific refresh token as revoked.
func (r *RefreshTokenRepo) Revoke(ctx context.Context, tokenHash []byte) error {
	query := `
		UPDATE social.refresh_tokens
		SET revoked = true
		WHERE token_hash = $1 AND revoked = false`

	_, err := runner(ctx, r.pool).Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("revoking refresh token: %w", err)
	}

	return nil
}

// RevokeAllForUser revokes all refresh tokens belonging to a user.
func (r *RefreshTokenRepo) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE social.refresh_tokens
		SET revoked = true
		WHERE user_id = $1 AND revoked = false`

	_, err := runner(ctx, r.pool).Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("revoking all tokens for user: %w", err)
	}

	return nil
}

// DeleteExpired removes expired and revoked token rows. Call periodically for cleanup.
func (r *RefreshTokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM social.refresh_tokens WHERE expires_at < NOW() OR revoked = true`

	tag, err := runner(ctx, r.pool).Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("deleting expired tokens: %w", err)
	}

	return tag.RowsAffected(), nil
}

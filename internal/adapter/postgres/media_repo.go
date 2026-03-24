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

// MediaRepo implements repository.MediaRepository using PostgreSQL.
type MediaRepo struct {
	pool *pgxpool.Pool
}

var _ repository.MediaRepository = (*MediaRepo)(nil)

// NewMediaRepo creates a new PostgreSQL-backed media repository.
func NewMediaRepo(pool *pgxpool.Pool) *MediaRepo {
	return &MediaRepo{pool: pool}
}

func (r *MediaRepo) Create(ctx context.Context, m *domain.Media) error {
	query := `
		INSERT INTO social.media (id, user_id, object_key, media_type, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`

	err := runner(ctx, r.pool).QueryRow(ctx, query,
		m.ID, m.UserID, m.ObjectKey, m.MediaType, m.SortOrder,
	).Scan(&m.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating media record: %w", err)
	}

	return nil
}

func (r *MediaRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Media, error) {
	query := `
		SELECT id, user_id, object_key, media_type, sort_order, is_verified, created_at
		FROM social.media
		WHERE user_id = $1
		ORDER BY sort_order ASC, created_at ASC`

	rows, err := runner(ctx, r.pool).Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("listing media: %w", err)
	}
	defer rows.Close()

	var items []*domain.Media
	for rows.Next() {
		m := &domain.Media{}
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.ObjectKey, &m.MediaType,
			&m.SortOrder, &m.IsVerified, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning media row: %w", err)
		}
		items = append(items, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating media: %w", err)
	}

	return items, nil
}

func (r *MediaRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	query := `
		SELECT id, user_id, object_key, media_type, sort_order, is_verified, created_at
		FROM social.media
		WHERE id = $1`

	m := &domain.Media{}
	err := runner(ctx, r.pool).QueryRow(ctx, query, id).Scan(
		&m.ID, &m.UserID, &m.ObjectKey, &m.MediaType,
		&m.SortOrder, &m.IsVerified, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("getting media: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("getting media: %w", err)
	}

	return m, nil
}

// Delete removes a media record only if the ownerID matches (prevents cross-user deletion).
func (r *MediaRepo) Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error {
	query := `DELETE FROM social.media WHERE id = $1 AND user_id = $2`

	tag, err := runner(ctx, r.pool).Exec(ctx, query, id, ownerID)
	if err != nil {
		return fmt.Errorf("deleting media: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("deleting media: %w", domain.ErrNotFound)
	}

	return nil
}

func (r *MediaRepo) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM social.media WHERE user_id = $1`

	var count int
	if err := runner(ctx, r.pool).QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting media: %w", err)
	}

	return count, nil
}

// UpdateAvatar updates the avatar_url column on the user's profile.
func (r *MediaRepo) UpdateAvatar(ctx context.Context, userID uuid.UUID, objectKey string) error {
	query := `
		UPDATE social.profiles
		SET avatar_url = $2, updated_at = NOW()
		WHERE user_id = $1`

	_, err := runner(ctx, r.pool).Exec(ctx, query, userID, objectKey)
	if err != nil {
		return fmt.Errorf("updating avatar: %w", err)
	}

	return nil
}

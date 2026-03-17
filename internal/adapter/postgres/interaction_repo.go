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

// InteractionRepo implements repository.InteractionRepository using PostgreSQL.
type InteractionRepo struct {
	pool *pgxpool.Pool
}

var _ repository.InteractionRepository = (*InteractionRepo)(nil)

// NewInteractionRepo creates a new PostgreSQL-backed interaction repository.
func NewInteractionRepo(pool *pgxpool.Pool) *InteractionRepo {
	return &InteractionRepo{pool: pool}
}

func (r *InteractionRepo) Create(ctx context.Context, interaction *domain.Interaction) error {
	query := `
		INSERT INTO social.interactions (
			rater_id, rated_id, rating, context, comment, is_verified
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		interaction.RaterID,
		interaction.RatedID,
		interaction.Rating,
		interaction.Context,
		interaction.Comment,
		interaction.IsVerified,
	).Scan(&interaction.ID, &interaction.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating interaction: %w", err)
	}

	return nil
}

func (r *InteractionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Interaction, error) {
	query := `
		SELECT id, rater_id, rated_id, rating, context, comment, is_verified, created_at
		FROM social.interactions
		WHERE id = $1`

	i := &domain.Interaction{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&i.ID, &i.RaterID, &i.RatedID,
		&i.Rating, &i.Context, &i.Comment,
		&i.IsVerified, &i.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("getting interaction: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("getting interaction: %w", err)
	}

	return i, nil
}

func (r *InteractionRepo) ConfirmInteraction(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE social.interactions
		SET is_verified = true
		WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("confirming interaction: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("confirming interaction: %w", domain.ErrNotFound)
	}

	return nil
}

func (r *InteractionRepo) GetByRatedUser(ctx context.Context, ratedID uuid.UUID, limit, offset int) ([]domain.Interaction, error) {
	query := `
		SELECT id, rater_id, rated_id, rating, context, comment, is_verified, created_at
		FROM social.interactions
		WHERE rated_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, ratedID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing interactions: %w", err)
	}
	defer rows.Close()

	var interactions []domain.Interaction
	for rows.Next() {
		var i domain.Interaction
		if err := rows.Scan(
			&i.ID, &i.RaterID, &i.RatedID,
			&i.Rating, &i.Context, &i.Comment,
			&i.IsVerified, &i.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning interaction row: %w", err)
		}
		interactions = append(interactions, i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating interactions: %w", err)
	}

	return interactions, nil
}

func (r *InteractionRepo) ExistsBetweenUsersAfter(ctx context.Context, raterID, ratedID uuid.UUID, after string) (bool, error) {
	afterTime, err := time.Parse(time.RFC3339, after)
	if err != nil {
		return false, fmt.Errorf("parsing after timestamp: %w", err)
	}

	query := `
		SELECT 1 FROM social.interactions
		WHERE rater_id = $1 AND rated_id = $2 AND created_at > $3
		LIMIT 1`

	var dummy int
	err = r.pool.QueryRow(ctx, query, raterID, ratedID, afterTime).Scan(&dummy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("checking recent interaction: %w", err)
	}

	return true, nil
}

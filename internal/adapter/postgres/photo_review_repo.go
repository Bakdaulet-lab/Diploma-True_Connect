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

// PhotoReviewRepo implements repository.PhotoReviewRepository.
type PhotoReviewRepo struct {
	pool *pgxpool.Pool
}

var _ repository.PhotoReviewRepository = (*PhotoReviewRepo)(nil)

func NewPhotoReviewRepo(pool *pgxpool.Pool) *PhotoReviewRepo {
	return &PhotoReviewRepo{pool: pool}
}

func (r *PhotoReviewRepo) Enqueue(ctx context.Context, pr *domain.PhotoReview) error {
	query := `
		INSERT INTO social.photo_review_queue (user_id, object_key, reasons, nsfw_score, face_count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, status, created_at`
	err := runner(ctx, r.pool).QueryRow(ctx, query,
		pr.UserID, pr.ObjectKey, pr.Reasons, pr.NSFWScore, pr.FaceCount,
	).Scan(&pr.ID, &pr.Status, &pr.CreatedAt)
	if err != nil {
		return fmt.Errorf("enqueue photo review: %w", err)
	}
	return nil
}

func (r *PhotoReviewRepo) ListPending(ctx context.Context, limit int) ([]*domain.PhotoReview, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	query := `
		SELECT id, user_id, object_key, reasons, nsfw_score, face_count, status, created_at
		FROM social.photo_review_queue
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1`
	rows, err := runner(ctx, r.pool).Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("listing photo reviews: %w", err)
	}
	defer rows.Close()

	var out []*domain.PhotoReview
	for rows.Next() {
		pr := &domain.PhotoReview{}
		if err := rows.Scan(&pr.ID, &pr.UserID, &pr.ObjectKey, &pr.Reasons,
			&pr.NSFWScore, &pr.FaceCount, &pr.Status, &pr.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning photo review: %w", err)
		}
		out = append(out, pr)
	}
	return out, rows.Err()
}

func (r *PhotoReviewRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PhotoReview, error) {
	query := `
		SELECT id, user_id, object_key, reasons, nsfw_score, face_count, status, created_at, reviewed_at, reviewed_by
		FROM social.photo_review_queue
		WHERE id = $1`
	pr := &domain.PhotoReview{}
	err := runner(ctx, r.pool).QueryRow(ctx, query, id).Scan(
		&pr.ID, &pr.UserID, &pr.ObjectKey, &pr.Reasons, &pr.NSFWScore,
		&pr.FaceCount, &pr.Status, &pr.CreatedAt, &pr.ReviewedAt, &pr.ReviewedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get photo review: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get photo review: %w", err)
	}
	return pr, nil
}

func (r *PhotoReviewRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, reviewerID uuid.UUID) error {
	query := `
		UPDATE social.photo_review_queue
		SET status = $2, reviewed_by = $3, reviewed_at = NOW()
		WHERE id = $1 AND status = 'pending'`
	tag, err := runner(ctx, r.pool).Exec(ctx, query, id, status, reviewerID)
	if err != nil {
		return fmt.Errorf("update photo review: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update photo review: %w", domain.ErrNotFound)
	}
	return nil
}

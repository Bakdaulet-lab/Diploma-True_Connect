package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// PhotoReviewRepository manages the admin re-review queue for rejected photos.
type PhotoReviewRepository interface {
	// Enqueue adds a held photo for manual review.
	Enqueue(ctx context.Context, r *domain.PhotoReview) error
	// ListPending returns pending reviews, oldest first.
	ListPending(ctx context.Context, limit int) ([]*domain.PhotoReview, error)
	// GetByID fetches a single review record.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PhotoReview, error)
	// UpdateStatus sets the verdict and records who/when.
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, reviewerID uuid.UUID) error
}

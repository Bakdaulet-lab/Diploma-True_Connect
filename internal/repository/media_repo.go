package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// MediaRepository manages media records (photo metadata) in PostgreSQL.
type MediaRepository interface {
	Create(ctx context.Context, media *domain.Media) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Media, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error)
	Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error
	CountByUser(ctx context.Context, userID uuid.UUID) (int, error)
	UpdateAvatar(ctx context.Context, userID uuid.UUID, objectKey string) error
}

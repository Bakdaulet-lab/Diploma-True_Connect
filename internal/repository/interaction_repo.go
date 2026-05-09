package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

type InteractionRepository interface {
	Create(ctx context.Context, interaction *domain.Interaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Interaction, error)
	ConfirmInteraction(ctx context.Context, id uuid.UUID) error
	GetByRatedUser(ctx context.Context, ratedID uuid.UUID, limit, offset int) ([]domain.Interaction, error)
	ExistsBetweenUsersAfter(ctx context.Context, raterID, ratedID uuid.UUID, after string) (bool, error)
}

package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByPhoneHash(ctx context.Context, phoneHash []byte) (*domain.User, error)
	UpdateVerificationLevel(ctx context.Context, id uuid.UUID, level domain.VerificationLevel) error
	UpdateTrustScore(ctx context.Context, id uuid.UUID, score int) error
	UpdateTrustStatus(ctx context.Context, id uuid.UUID, status domain.TrustStatus) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdateFCMToken(ctx context.Context, id uuid.UUID, token string) error
	ListByTrustStatus(ctx context.Context, status domain.TrustStatus, limit, offset int) ([]*domain.User, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

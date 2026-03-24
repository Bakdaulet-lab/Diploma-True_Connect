package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a stored refresh token record.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash []byte
	ExpiresAt time.Time
	Revoked   bool
}

// RefreshTokenRepository defines storage operations for refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) (uuid.UUID, error)
	GetByTokenHash(ctx context.Context, tokenHash []byte) (*RefreshToken, error)
	Revoke(ctx context.Context, tokenHash []byte) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

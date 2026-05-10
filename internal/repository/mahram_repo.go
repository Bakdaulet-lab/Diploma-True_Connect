package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// MahramRepository manages mahram (chaperone) records for female users.
type MahramRepository interface {
	// Create registers a new mahram for a woman. The phone is pre-encrypted and pre-hashed
	// by the service layer before being passed here.
	Create(ctx context.Context, womanUserID uuid.UUID, phoneEncrypted, phoneHash []byte) (*domain.Mahram, error)

	// GetByWomanID returns all mahrams registered by the given woman.
	GetByWomanID(ctx context.Context, womanUserID uuid.UUID) ([]*domain.Mahram, error)

	// GetByID returns a single mahram record.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Mahram, error)

	// UpdateStatus transitions a mahram's verification_status.
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, verifiedAt *interface{}) error

	// SetTelegramChatID stores the Telegram chat ID after the bot receives the first message.
	SetTelegramChatID(ctx context.Context, id uuid.UUID, chatID int64) error
}

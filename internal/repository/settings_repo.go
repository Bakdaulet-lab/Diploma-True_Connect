package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// UserSettingsRepository manages per-user notification and matching preferences.
type UserSettingsRepository interface {
	// Get returns the settings for a user, or defaults if none exist yet.
	Get(ctx context.Context, userID uuid.UUID) (*domain.UserSettings, error)

	// Upsert creates or updates settings for a user.
	Upsert(ctx context.Context, settings *domain.UserSettings) error
}

package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// UpdateSettingsInput contains all fields that can be patched.
// Pointer fields are optional — nil means "don't change".
type UpdateSettingsInput struct {
	PushNotifications *bool
	ShowOnlineStatus  *bool
	DistanceUnit      *string
	MaxDistanceKm     *int
	AgeRangeMin       *int
	AgeRangeMax       *int
	ModestyLevel      *int
	NiyyahFilter      *string // empty string clears the filter
	MadhabFilter      *string // empty string clears the filter
}

// SettingsService manages user notification and matching preferences.
type SettingsService struct {
	repo repository.UserSettingsRepository
}

// NewSettingsService creates a new settings service.
func NewSettingsService(repo repository.UserSettingsRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

// Get returns settings for the user, applying defaults if not yet configured.
func (s *SettingsService) Get(ctx context.Context, userID uuid.UUID) (*domain.UserSettings, error) {
	settings, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return settings, nil
}

// Update applies a partial update to the user's settings (PATCH semantics).
func (s *SettingsService) Update(ctx context.Context, userID uuid.UUID, input UpdateSettingsInput) (*domain.UserSettings, error) {
	// Load existing (or default) settings first.
	existing, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("update settings: loading existing: %w", err)
	}

	// Apply only the fields that were provided.
	if input.PushNotifications != nil {
		existing.PushNotifications = *input.PushNotifications
	}
	if input.ShowOnlineStatus != nil {
		existing.ShowOnlineStatus = *input.ShowOnlineStatus
	}
	if input.DistanceUnit != nil {
		if *input.DistanceUnit != "km" && *input.DistanceUnit != "mi" {
			return nil, fmt.Errorf("update settings: distance_unit must be 'km' or 'mi': %w", domain.ErrInvalidInput)
		}
		existing.DistanceUnit = *input.DistanceUnit
	}
	if input.MaxDistanceKm != nil {
		if *input.MaxDistanceKm < 1 || *input.MaxDistanceKm > 500 {
			return nil, fmt.Errorf("update settings: max_distance_km must be between 1 and 500: %w", domain.ErrInvalidInput)
		}
		existing.MaxDistanceKm = *input.MaxDistanceKm
	}
	if input.AgeRangeMin != nil {
		if *input.AgeRangeMin < 18 || *input.AgeRangeMin > 99 {
			return nil, fmt.Errorf("update settings: age_range_min must be between 18 and 99: %w", domain.ErrInvalidInput)
		}
		existing.AgeRangeMin = *input.AgeRangeMin
	}
	if input.AgeRangeMax != nil {
		if *input.AgeRangeMax < 18 || *input.AgeRangeMax > 99 {
			return nil, fmt.Errorf("update settings: age_range_max must be between 18 and 99: %w", domain.ErrInvalidInput)
		}
		existing.AgeRangeMax = *input.AgeRangeMax
	}

	// Validate range order after applying both ends.
	if existing.AgeRangeMin > existing.AgeRangeMax {
		return nil, fmt.Errorf("update settings: age_range_min cannot exceed age_range_max: %w", domain.ErrInvalidInput)
	}

	if input.ModestyLevel != nil {
		if *input.ModestyLevel < 0 || *input.ModestyLevel > 5 {
			return nil, fmt.Errorf("update settings: modesty_level must be between 0 and 5: %w", domain.ErrInvalidInput)
		}
		existing.ModestyLevel = *input.ModestyLevel
	}
	if input.NiyyahFilter != nil {
		if *input.NiyyahFilter == "" {
			existing.NiyyahFilter = nil
		} else {
			existing.NiyyahFilter = input.NiyyahFilter
		}
	}
	if input.MadhabFilter != nil {
		if *input.MadhabFilter == "" {
			existing.MadhabFilter = nil
		} else {
			existing.MadhabFilter = input.MadhabFilter
		}
	}

	if err := s.repo.Upsert(ctx, existing); err != nil {
		return nil, fmt.Errorf("update settings: persisting: %w", err)
	}

	return existing, nil
}

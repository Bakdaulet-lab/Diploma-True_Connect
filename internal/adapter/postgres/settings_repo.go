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

// SettingsRepo implements repository.UserSettingsRepository using PostgreSQL.
type SettingsRepo struct {
	pool *pgxpool.Pool
}

var _ repository.UserSettingsRepository = (*SettingsRepo)(nil)

// NewSettingsRepo creates a new PostgreSQL-backed settings repository.
func NewSettingsRepo(pool *pgxpool.Pool) *SettingsRepo {
	return &SettingsRepo{pool: pool}
}

func (r *SettingsRepo) Get(ctx context.Context, userID uuid.UUID) (*domain.UserSettings, error) {
	query := `
		SELECT
			user_id::text,
			push_notifications,
			show_online_status,
			distance_unit,
			max_distance_km,
			age_range_min,
			age_range_max,
			updated_at
		FROM social.user_settings
		WHERE user_id = $1`

	s := &domain.UserSettings{}
	err := runner(ctx, r.pool).QueryRow(ctx, query, userID).Scan(
		&s.UserID,
		&s.PushNotifications,
		&s.ShowOnlineStatus,
		&s.DistanceUnit,
		&s.MaxDistanceKm,
		&s.AgeRangeMin,
		&s.AgeRangeMax,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return defaults when no row exists yet.
			return domain.DefaultSettings(userID.String()), nil
		}
		return nil, fmt.Errorf("getting settings: %w", err)
	}

	return s, nil
}

func (r *SettingsRepo) Upsert(ctx context.Context, s *domain.UserSettings) error {
	query := `
		INSERT INTO social.user_settings (
			user_id, push_notifications, show_online_status,
			distance_unit, max_distance_km, age_range_min, age_range_max
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			push_notifications = EXCLUDED.push_notifications,
			show_online_status = EXCLUDED.show_online_status,
			distance_unit      = EXCLUDED.distance_unit,
			max_distance_km    = EXCLUDED.max_distance_km,
			age_range_min      = EXCLUDED.age_range_min,
			age_range_max      = EXCLUDED.age_range_max,
			updated_at         = NOW()
		RETURNING updated_at`

	userID, err := uuid.Parse(s.UserID)
	if err != nil {
		return fmt.Errorf("parsing user id: %w", err)
	}

	return runner(ctx, r.pool).QueryRow(ctx, query,
		userID,
		s.PushNotifications,
		s.ShowOnlineStatus,
		s.DistanceUnit,
		s.MaxDistanceKm,
		s.AgeRangeMin,
		s.AgeRangeMax,
	).Scan(&s.UpdatedAt)
}

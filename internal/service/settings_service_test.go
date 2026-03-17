package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/service"
)

func newTestSettingsService() (*service.SettingsService, *mockSettingsRepo) {
	repo := newMockSettingsRepo()
	return service.NewSettingsService(repo), repo
}

// ── Get ───────────────────────────────────────────────────────────────────────

func TestSettings_Get_ReturnsDefaults(t *testing.T) {
	t.Parallel()

	svc, _ := newTestSettingsService()
	userID := uuid.New()

	settings, err := svc.Get(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// DefaultSettings enables push notifications.
	if !settings.PushNotifications {
		t.Error("expected PushNotifications=true in defaults")
	}
	if settings.AgeRangeMin < 18 || settings.AgeRangeMax > 99 {
		t.Errorf("default age range out of bounds: %d–%d", settings.AgeRangeMin, settings.AgeRangeMax)
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestSettings_Update_PartialPatch(t *testing.T) {
	t.Parallel()

	svc, _ := newTestSettingsService()
	userID := uuid.New()

	// Only change distance unit; other fields stay at defaults.
	unit := "mi"
	updated, err := svc.Update(context.Background(), userID, service.UpdateSettingsInput{
		DistanceUnit: &unit,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.DistanceUnit != "mi" {
		t.Errorf("expected DistanceUnit=mi, got %q", updated.DistanceUnit)
	}
	// push notifications should still be the default (true).
	if !updated.PushNotifications {
		t.Error("PushNotifications should remain unchanged at default")
	}
}

func TestSettings_Update_InvalidDistanceUnit(t *testing.T) {
	t.Parallel()

	svc, _ := newTestSettingsService()
	userID := uuid.New()

	unit := "yards"
	_, err := svc.Update(context.Background(), userID, service.UpdateSettingsInput{
		DistanceUnit: &unit,
	})
	if err == nil {
		t.Fatal("expected error for invalid distance unit")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestSettings_Update_MaxDistanceOutOfRange(t *testing.T) {
	t.Parallel()

	svc, _ := newTestSettingsService()
	userID := uuid.New()

	dist := 9999
	_, err := svc.Update(context.Background(), userID, service.UpdateSettingsInput{
		MaxDistanceKm: &dist,
	})
	if err == nil {
		t.Fatal("expected error for max_distance_km > 500")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestSettings_Update_AgeRangeInverted(t *testing.T) {
	t.Parallel()

	svc, _ := newTestSettingsService()
	userID := uuid.New()

	min, max := 40, 25
	_, err := svc.Update(context.Background(), userID, service.UpdateSettingsInput{
		AgeRangeMin: &min,
		AgeRangeMax: &max,
	})
	if err == nil {
		t.Fatal("expected error when age_range_min > age_range_max")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestSettings_Update_TogglePushNotifications(t *testing.T) {
	t.Parallel()

	svc, _ := newTestSettingsService()
	userID := uuid.New()

	off := false
	updated, err := svc.Update(context.Background(), userID, service.UpdateSettingsInput{
		PushNotifications: &off,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.PushNotifications != false {
		t.Error("expected PushNotifications=false after update")
	}
}

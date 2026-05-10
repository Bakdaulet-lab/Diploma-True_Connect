package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/service"
)

// catalogImamID is one of the real IDs embedded in catalog.json.
var catalogImamID = uuid.MustParse("a0000001-a000-4001-a000-000000000001")

func newTestImamService() (*service.ImamService, *mockMatchRepo, *mockProfileRepo) {
	matchRepo := newMockMatchRepo()
	profileRepo := newMockProfileRepo()
	// notifSvc = nil: imam service guards nil before calling it.
	svc := service.NewImamService(matchRepo, profileRepo, nil, nil)
	return svc, matchRepo, profileRepo
}

// ── ListImams ─────────────────────────────────────────────────────────────────

func TestImamService_ListImams_KnownCity(t *testing.T) {
	t.Parallel()

	svc, _, _ := newTestImamService()
	imams, err := svc.ListImams(context.Background(), "Almaty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(imams) == 0 {
		t.Fatal("expected at least one imam for Almaty")
	}
	for _, im := range imams {
		if im.City != "Almaty" {
			t.Errorf("expected city=Almaty, got %s", im.City)
		}
	}
}

func TestImamService_ListImams_UnknownCity(t *testing.T) {
	t.Parallel()

	svc, _, _ := newTestImamService()
	imams, err := svc.ListImams(context.Background(), "Timbuktu")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(imams) != 0 {
		t.Errorf("expected empty list for unknown city, got %d", len(imams))
	}
}

func TestImamService_ListImams_CaseInsensitive(t *testing.T) {
	t.Parallel()

	svc, _, _ := newTestImamService()
	upper, _ := svc.ListImams(context.Background(), "ALMATY")
	lower, _ := svc.ListImams(context.Background(), "almaty")
	if len(upper) != len(lower) {
		t.Errorf("case sensitivity mismatch: upper=%d lower=%d", len(upper), len(lower))
	}
}

// ── ConfirmNikah ──────────────────────────────────────────────────────────────

func TestImamService_ConfirmNikah_Success(t *testing.T) {
	t.Parallel()

	svc, matchRepo, _ := newTestImamService()
	userA, userB := uuid.New(), uuid.New()
	matchID := uuid.New()
	now := time.Now()
	matchRepo.mu.Lock()
	matchRepo.matches[matchID] = &domain.Match{
		ID:        matchID,
		UserAID:   userA,
		UserBID:   userB,
		MatchedAt: &now,
	}
	matchRepo.mu.Unlock()

	err := svc.ConfirmNikah(context.Background(), matchID, catalogImamID, userA)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestImamService_ConfirmNikah_UnknownImam(t *testing.T) {
	t.Parallel()

	svc, _, _ := newTestImamService()
	err := svc.ConfirmNikah(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for unknown imam, got %v", err)
	}
}

func TestImamService_ConfirmNikah_CallerNotInMatch(t *testing.T) {
	t.Parallel()

	svc, matchRepo, _ := newTestImamService()
	userA, userB := uuid.New(), uuid.New()
	outsider := uuid.New()
	matchID := uuid.New()
	now := time.Now()
	matchRepo.mu.Lock()
	matchRepo.matches[matchID] = &domain.Match{
		ID:        matchID,
		UserAID:   userA,
		UserBID:   userB,
		MatchedAt: &now,
	}
	matchRepo.mu.Unlock()

	err := svc.ConfirmNikah(context.Background(), matchID, catalogImamID, outsider)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestImamService_ConfirmNikah_Idempotent(t *testing.T) {
	t.Parallel()

	svc, matchRepo, _ := newTestImamService()
	userA, userB := uuid.New(), uuid.New()
	matchID := uuid.New()
	now := time.Now()
	matchRepo.mu.Lock()
	matchRepo.matches[matchID] = &domain.Match{
		ID:            matchID,
		UserAID:       userA,
		UserBID:       userB,
		MatchedAt:     &now,
		ImamConfirmed: true, // already confirmed
	}
	matchRepo.mu.Unlock()

	// Should succeed without error (idempotent).
	err := svc.ConfirmNikah(context.Background(), matchID, catalogImamID, userA)
	if err != nil {
		t.Fatalf("expected no error for already-confirmed match, got: %v", err)
	}
}

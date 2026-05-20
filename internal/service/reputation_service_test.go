package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/service"
)

func setupReputationService(t *testing.T) (
	*service.ReputationService,
	*trackingGraphRepo,
	*mockUserRepoS3,
	*mockMatchingCache,
) {
	t.Helper()
	graphRepo := newTrackingGraphRepo()
	userRepo := newMockUserRepoS3()
	cache := newMockMatchingCache()
	profileRepo := newMockProfileRepo()
	svc := service.NewReputationService(graphRepo, userRepo, profileRepo, cache)
	return svc, graphRepo, userRepo, cache
}

// ── RecalculateScore tests ─────────────────────────────────────────────────

func TestRecalculateScore_WritesToAllStores(t *testing.T) {
	svc, graphRepo, userRepo, cache := setupReputationService(t)
	userID := uuid.New()
	graphRepo.scoreToReturn = 72

	score, err := svc.RecalculateScore(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if score != 72 {
		t.Fatalf("expected score 72, got %d", score)
	}

	// PG updated.
	userRepo.mu.Lock()
	pgScore, ok := userRepo.trustScores[userID]
	userRepo.mu.Unlock()
	if !ok || pgScore != 72 {
		t.Errorf("expected PG trust_score=72, got %d (exists=%v)", pgScore, ok)
	}

	// Redis cached.
	cache.mu.Lock()
	cachedScore, ok := cache.trustScores[userID]
	cache.mu.Unlock()
	if !ok || cachedScore != 72 {
		t.Errorf("expected cached trust_score=72, got %d (exists=%v)", cachedScore, ok)
	}
}

func TestRecalculateScore_DefaultScore(t *testing.T) {
	svc, graphRepo, _, _ := setupReputationService(t)
	graphRepo.scoreToReturn = 50

	score, err := svc.RecalculateScore(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if score != 50 {
		t.Fatalf("expected neutral score 50, got %d", score)
	}
}

// ── GetScore tests ─────────────────────────────────────────────────────────

func TestGetScore_ReturnsCached(t *testing.T) {
	svc, graphRepo, _, cache := setupReputationService(t)
	userID := uuid.New()

	// Pre-seed cache.
	cache.mu.Lock()
	cache.trustScores[userID] = 85
	cache.mu.Unlock()

	graphRepo.scoreToReturn = 60 // Should NOT be called.

	ts, err := svc.GetScore(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if ts.Score != 85 {
		t.Fatalf("expected cached score 85, got %d", ts.Score)
	}
}

func TestGetScore_FallsBackToCompute(t *testing.T) {
	svc, graphRepo, _, cache := setupReputationService(t)
	userID := uuid.New()
	graphRepo.scoreToReturn = 63

	ts, err := svc.GetScore(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if ts.Score != 63 {
		t.Fatalf("expected computed score 63, got %d", ts.Score)
	}

	// Verify it was then cached.
	cache.mu.Lock()
	cached, ok := cache.trustScores[userID]
	cache.mu.Unlock()
	if !ok || cached != 63 {
		t.Errorf("expected score to be cached as 63, got %d (exists=%v)", cached, ok)
	}
}

func TestGetScore_UserIDPreserved(t *testing.T) {
	svc, _, _, _ := setupReputationService(t)
	userID := uuid.New()

	ts, err := svc.GetScore(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if ts.UserID != userID {
		t.Fatalf("expected UserID %s, got %s", userID, ts.UserID)
	}
}

// ── GetScoreBreakdown tests ────────────────────────────────────────────────

func TestGetScoreBreakdown_PassthroughAndBadge(t *testing.T) {
	svc, graphRepo, _, _ := setupReputationService(t)
	userID := uuid.New()
	graphRepo.breakdown = &domain.TrustScoreBreakdown{
		Score:          72,
		RatingCount:    4,
		SmoothedRating: 3.6,
		BaseScore:      72.0,
		KYCBonus:       10.0,
		ReportCount:    1,
		ReportPenalty:  15.0,
		RawScore:       67.0,
	}

	b, err := svc.GetScoreBreakdown(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if b.UserID != userID {
		t.Fatalf("expected UserID %s, got %s", userID, b.UserID)
	}
	if b.Score != 72 || b.RatingCount != 4 || b.KYCBonus != 10.0 || b.ReportPenalty != 15.0 {
		t.Fatalf("breakdown components not passed through: %+v", b)
	}
	// Badge must be derived from Score (72 → Gold tier), not left empty.
	if b.Badge != "Gold / Member" {
		t.Fatalf("expected badge 'Gold / Member' for score 72, got %q", b.Badge)
	}
}

func TestGetScoreBreakdown_NeutralDefault(t *testing.T) {
	svc, graphRepo, _, _ := setupReputationService(t)
	graphRepo.scoreToReturn = 50 // mock returns a neutral breakdown by default

	b, err := svc.GetScoreBreakdown(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if b.Score != 50 || b.Badge != "Silver / Newbie" {
		t.Fatalf("expected neutral score 50 / Silver, got %d / %q", b.Score, b.Badge)
	}
}

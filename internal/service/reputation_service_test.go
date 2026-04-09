package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
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

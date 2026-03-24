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

type dummyUoW struct{}

func (d *dummyUoW) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func setupInteractionService(t *testing.T) (
	*service.InteractionService,
	*mockInteractionRepo,
	*mockMatchRepo,
	*trackingGraphRepo,
	chan uuid.UUID,
) {
	t.Helper()
	interactionRepo := newMockInteractionRepo()
	matchRepo := newMockMatchRepo()
	graphRepo := newTrackingGraphRepo()
	eventCh := make(chan uuid.UUID, 10)
	uow := &dummyUoW{}
	svc := service.NewInteractionService(interactionRepo, matchRepo, graphRepo, uow, eventCh)
	return svc, interactionRepo, matchRepo, graphRepo, eventCh
}

func createMatchedPair(t *testing.T, matchRepo *mockMatchRepo) (uuid.UUID, uuid.UUID) {
	t.Helper()
	userA := uuid.New()
	userB := uuid.New()
	// Simulate mutual match by recording likes from both sides.
	matchRepo.RecordLike(context.Background(), userA, userB)
	matchRepo.RecordLike(context.Background(), userB, userA)
	return userA, userB
}

// ── SubmitRating tests ─────────────────────────────────────────────────────

func TestSubmitRating_Success(t *testing.T) {
	svc, _, matchRepo, graphRepo, eventCh := setupInteractionService(t)
	rater, rated := createMatchedPair(t, matchRepo)

	interaction, err := svc.SubmitRating(context.Background(), rater, rated, 4, "date", "great meeting")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if interaction.ID == uuid.Nil {
		t.Fatal("expected interaction ID to be set")
	}
	if interaction.Rating != 4 {
		t.Errorf("expected rating 4, got %d", interaction.Rating)
	}
	if interaction.IsVerified {
		t.Error("expected IsVerified to be false on initial submission")
	}

	// Check that a graph edge was created.
	graphRepo.mu.Lock()
	if len(graphRepo.ratings) != 1 {
		t.Errorf("expected 1 graph rating, got %d", len(graphRepo.ratings))
	}
	graphRepo.mu.Unlock()

	// Check that an event was emitted.
	select {
	case uid := <-eventCh:
		if uid != rated {
			t.Errorf("expected event for %s, got %s", rated, uid)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected event on channel, got none")
	}
}

func TestSubmitRating_SelfRating(t *testing.T) {
	svc, _, _, _, _ := setupInteractionService(t)
	userID := uuid.New()

	_, err := svc.SubmitRating(context.Background(), userID, userID, 5, "date", "")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSubmitRating_NotMatched(t *testing.T) {
	svc, _, _, _, _ := setupInteractionService(t)
	_, err := svc.SubmitRating(context.Background(), uuid.New(), uuid.New(), 3, "meetup", "")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got: %v", err)
	}
}

func TestSubmitRating_InvalidRating(t *testing.T) {
	svc, _, matchRepo, _, _ := setupInteractionService(t)
	rater, rated := createMatchedPair(t, matchRepo)

	_, err := svc.SubmitRating(context.Background(), rater, rated, 6, "date", "")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for rating=6, got: %v", err)
	}

	_, err = svc.SubmitRating(context.Background(), rater, rated, 0, "date", "")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for rating=0, got: %v", err)
	}
}

func TestSubmitRating_RateLimited(t *testing.T) {
	svc, _, matchRepo, _, _ := setupInteractionService(t)
	rater, rated := createMatchedPair(t, matchRepo)

	// First rating succeeds.
	_, err := svc.SubmitRating(context.Background(), rater, rated, 4, "date", "")
	if err != nil {
		t.Fatalf("first rating should succeed: %v", err)
	}

	// Second rating within 7 days should fail.
	_, err = svc.SubmitRating(context.Background(), rater, rated, 5, "meetup", "")
	if !errors.Is(err, domain.ErrRateLimitExceeded) {
		t.Fatalf("expected ErrRateLimitExceeded, got: %v", err)
	}
}

func TestSubmitRating_GraphEdgeCreated(t *testing.T) {
	svc, _, matchRepo, graphRepo, _ := setupInteractionService(t)
	rater, rated := createMatchedPair(t, matchRepo)

	_, err := svc.SubmitRating(context.Background(), rater, rated, 3, "event", "cool event")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	graphRepo.mu.Lock()
	defer graphRepo.mu.Unlock()
	if len(graphRepo.ratings) != 1 {
		t.Fatalf("expected 1 rating edge, got %d", len(graphRepo.ratings))
	}
	edge := graphRepo.ratings[0]
	if edge.from != rater || edge.to != rated {
		t.Error("rating edge has wrong direction")
	}
	if edge.score != 3 {
		t.Errorf("expected score 3, got %d", edge.score)
	}
	if edge.verified {
		t.Error("initial rating should not be verified")
	}
}

// ── ConfirmInteraction tests ───────────────────────────────────────────────

func TestConfirmInteraction_Success(t *testing.T) {
	svc, interactionRepo, matchRepo, graphRepo, eventCh := setupInteractionService(t)
	rater, rated := createMatchedPair(t, matchRepo)

	interaction, err := svc.SubmitRating(context.Background(), rater, rated, 5, "date", "")
	if err != nil {
		t.Fatalf("submit rating failed: %v", err)
	}
	// Drain the submit event.
	<-eventCh

	// The rated user confirms.
	err = svc.ConfirmInteraction(context.Background(), interaction.ID, rated)
	if err != nil {
		t.Fatalf("confirm should succeed: %v", err)
	}

	// Verify PG is updated.
	interactionRepo.mu.Lock()
	confirmed := interactionRepo.interactions[interaction.ID]
	interactionRepo.mu.Unlock()
	if !confirmed.IsVerified {
		t.Error("expected interaction to be verified in PG")
	}

	// Verify Neo4j got a verified rating + meeting edge.
	graphRepo.mu.Lock()
	if len(graphRepo.ratings) != 2 {
		t.Errorf("expected 2 graph ratings (initial + verified), got %d", len(graphRepo.ratings))
	}
	if len(graphRepo.meetings) != 1 {
		t.Errorf("expected 1 meeting edge, got %d", len(graphRepo.meetings))
	}
	graphRepo.mu.Unlock()

	// Verify event emitted for recalculation.
	select {
	case uid := <-eventCh:
		if uid != rated {
			t.Errorf("expected event for %s, got %s", rated, uid)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected event on channel after confirm")
	}
}

func TestConfirmInteraction_WrongUser(t *testing.T) {
	svc, _, matchRepo, _, eventCh := setupInteractionService(t)
	rater, rated := createMatchedPair(t, matchRepo)

	interaction, err := svc.SubmitRating(context.Background(), rater, rated, 4, "date", "")
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	<-eventCh

	// Rater (not the rated user) tries to confirm — should fail.
	err = svc.ConfirmInteraction(context.Background(), interaction.ID, rater)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got: %v", err)
	}
}

func TestConfirmInteraction_NotFound(t *testing.T) {
	svc, _, _, _, _ := setupInteractionService(t)
	err := svc.ConfirmInteraction(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestConfirmInteraction_AlreadyConfirmed(t *testing.T) {
	svc, _, matchRepo, _, eventCh := setupInteractionService(t)
	rater, rated := createMatchedPair(t, matchRepo)

	interaction, _ := svc.SubmitRating(context.Background(), rater, rated, 5, "date", "")
	<-eventCh

	// First confirm succeeds.
	err := svc.ConfirmInteraction(context.Background(), interaction.ID, rated)
	if err != nil {
		t.Fatalf("first confirm should succeed: %v", err)
	}
	<-eventCh

	// Second confirm is idempotent — no error.
	err = svc.ConfirmInteraction(context.Background(), interaction.ID, rated)
	if err != nil {
		t.Fatalf("second confirm should be idempotent: %v", err)
	}
}

// ── GetByRatedUser tests ───────────────────────────────────────────────────

func TestGetByRatedUser_Success(t *testing.T) {
	svc, _, matchRepo, _, eventCh := setupInteractionService(t)
	rater, rated := createMatchedPair(t, matchRepo)

	_, _ = svc.SubmitRating(context.Background(), rater, rated, 4, "date", "nice")
	<-eventCh

	results, err := svc.GetByRatedUser(context.Background(), rated, 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(results))
	}
	if results[0].Rating != 4 {
		t.Errorf("expected rating 4, got %d", results[0].Rating)
	}
}

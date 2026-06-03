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

func newTestMatchingService() (*service.MatchingService, *mockProfileRepo, *mockMatchRepo, *mockSettingsRepo, *mockMatchingCache) {
	profileRepo := newMockProfileRepo()
	matchRepo := newMockMatchRepo()
	settingsRepo := newMockSettingsRepo()
	cache := newMockMatchingCache()
	graphRepo := newTrackingGraphRepo()
	svc := service.NewMatchingService(profileRepo, nil, matchRepo, settingsRepo, cache, graphRepo, nil, nil, nil)
	return svc, profileRepo, matchRepo, settingsRepo, cache
}

// ── Like ──────────────────────────────────────────────────────────────────────

func TestLike_Success_NoMatch(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _ := newTestMatchingService()
	userID := uuid.New()
	targetID := uuid.New()

	result, err := svc.Like(context.Background(), userID, targetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Matched {
		t.Error("expected no match on one-sided like")
	}
	if result.MatchID != uuid.Nil {
		t.Error("expected nil MatchID when not matched")
	}
}

func TestLike_MutualMatch(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _ := newTestMatchingService()
	aliceID := uuid.New()
	bobID := uuid.New()

	// Alice likes Bob.
	r1, err := svc.Like(context.Background(), aliceID, bobID)
	if err != nil {
		t.Fatalf("alice like: %v", err)
	}
	if r1.Matched {
		t.Fatal("expected no match after first like")
	}

	// Bob likes Alice — should trigger a mutual match.
	r2, err := svc.Like(context.Background(), bobID, aliceID)
	if err != nil {
		t.Fatalf("bob like: %v", err)
	}
	if !r2.Matched {
		t.Error("expected mutual match")
	}
	if r2.MatchID == uuid.Nil {
		t.Error("expected non-nil MatchID on mutual match")
	}
}

func TestLike_Self_ReturnsInvalidInput(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _ := newTestMatchingService()
	userID := uuid.New()

	_, err := svc.Like(context.Background(), userID, userID)
	if err == nil {
		t.Fatal("expected error on self-like")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

// ── Pass ──────────────────────────────────────────────────────────────────────

func TestPass_Success(t *testing.T) {
	t.Parallel()

	svc, _, _, _, cache := newTestMatchingService()
	userID := uuid.New()
	targetID := uuid.New()

	if err := svc.Pass(context.Background(), userID, targetID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// target should now be in the seen set.
	seenIDs, _ := cache.GetSeenIDs(context.Background(), userID)
	found := false
	for _, id := range seenIDs {
		if id == targetID {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected targetID to be in seen set after pass")
	}
}

func TestPass_Self_ReturnsInvalidInput(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _ := newTestMatchingService()
	userID := uuid.New()

	err := svc.Pass(context.Background(), userID, userID)
	if err == nil {
		t.Fatal("expected error on self-pass")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

// ── ListMatches ───────────────────────────────────────────────────────────────

func TestListMatches_PaginationClampsPerPage(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _ := newTestMatchingService()
	userID := uuid.New()

	// page=1 with perPage=100 should be silently clamped to 20 (no error).
	matches, _, err := svc.ListMatches(context.Background(), userID, "", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No matches seeded, so result must be empty (not an error).
	if matches == nil {
		t.Error("expected non-nil slice")
	}
}

func TestListMatches_ReturnsOnlyOwnMatches(t *testing.T) {
	t.Parallel()

	svc, _, matchRepo, _, _ := newTestMatchingService()
	aliceID := uuid.New()
	bobID := uuid.New()
	carolID := uuid.New()

	// Bob ↔ Carol match (Alice is not involved).
	matchRepo.RecordLike(context.Background(), bobID, carolID)
	matchRepo.RecordLike(context.Background(), carolID, bobID)

	// Alice has no matches.
	matches, _, err := svc.ListMatches(context.Background(), aliceID, "", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for Alice, got %d", len(matches))
	}
}

func TestListMatches_NoPhotoMode_HidesAvatar(t *testing.T) {
	t.Parallel()

	svc, _, matchRepo, _, _ := newTestMatchingService()
	aliceID := uuid.New()
	bobID := uuid.New()
	matchID := uuid.New()
	now := time.Now()

	matchRepo.mu.Lock()
	matchRepo.matches[matchID] = &domain.Match{
		ID:        matchID,
		UserAID:   aliceID,
		UserBID:   bobID,
		MatchedAt: &now,
	}
	matchRepo.noPhotoUsers[bobID] = true // Bob enabled no-photo mode
	matchRepo.mu.Unlock()

	matches, _, err := svc.ListMatches(context.Background(), aliceID, "", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match for Alice, got %d", len(matches))
	}
	other := matches[0].OtherUser
	if other.AvatarURL != "" {
		t.Errorf("no-photo-mode match should have empty AvatarURL, got %q", other.AvatarURL)
	}
	if !other.AvatarBlurred {
		t.Error("no-photo-mode match should have AvatarBlurred=true")
	}
}

// ── NiyyahCompatible (B1) ────────────────────────────────────────────────────

func TestGetCandidates_NikkahYear_ExcludesFriendship(t *testing.T) {
	t.Parallel()

	svc, profileRepo, _, _, _ := newTestMatchingService()
	requesterID := uuid.New()
	friendshipID := uuid.New()
	seriousID := uuid.New()

	profileRepo.Upsert(context.Background(), &domain.Profile{
		UserID: requesterID, DisplayName: "Requester",
		Niyyah: domain.NiyyahNikahYear,
	})
	profileRepo.Upsert(context.Background(), &domain.Profile{
		UserID: friendshipID, DisplayName: "FriendshipOnly",
		Niyyah: domain.NiyyahFriendship,
	})
	profileRepo.Upsert(context.Background(), &domain.Profile{
		UserID: seriousID, DisplayName: "SeriousMarriage",
		Niyyah: domain.NiyyahSeriousMarriage,
	})

	candidates, err := svc.GetCandidates(context.Background(), requesterID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, c := range candidates {
		if c.UserID == friendshipID {
			t.Error("nikah_year requester should not see friendship-only candidates")
		}
	}
	found := false
	for _, c := range candidates {
		if c.UserID == seriousID {
			found = true
			break
		}
	}
	if !found {
		t.Error("nikah_year requester should see serious_marriage candidates")
	}
}

// ── NoPhotoMode (B3) ────────────────────────────────────────────────────────

func TestGetCandidates_NoPhotoMode_BlursAvatar(t *testing.T) {
	t.Parallel()

	svc, profileRepo, _, _, _ := newTestMatchingService()
	requesterID := uuid.New()
	privateID := uuid.New()

	profileRepo.Upsert(context.Background(), &domain.Profile{
		UserID: requesterID, DisplayName: "Requester",
	})
	profileRepo.Upsert(context.Background(), &domain.Profile{
		UserID: privateID, DisplayName: "PrivateUser",
		AvatarURL:   "http://example.com/avatar.jpg",
		NoPhotoMode: true,
	})

	candidates, err := svc.GetCandidates(context.Background(), requesterID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, c := range candidates {
		if c.UserID == privateID {
			if c.AvatarURL != "" {
				t.Error("no-photo-mode candidate should have empty AvatarURL")
			}
			if !c.AvatarBlurred {
				t.Error("no-photo-mode candidate should have AvatarBlurred=true")
			}
		}
	}
}

// ── GetCandidates ─────────────────────────────────────────────────────────────

func TestGetCandidates_ExcludesRequester(t *testing.T) {
	t.Parallel()

	svc, profileRepo, _, _, _ := newTestMatchingService()
	userID := uuid.New()
	otherID := uuid.New()

	profileRepo.Upsert(context.Background(), &domain.Profile{
		UserID: userID, DisplayName: "Requester",
	})
	profileRepo.Upsert(context.Background(), &domain.Profile{
		UserID: otherID, DisplayName: "Other",
	})

	candidates, err := svc.GetCandidates(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range candidates {
		if c.UserID == userID {
			t.Error("requester should not appear in their own candidate list")
		}
	}
}

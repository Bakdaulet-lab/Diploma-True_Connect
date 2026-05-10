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

// ── helpers ───────────────────────────────────────────────────────────────────

func registerUserInRepo(repo *mockUserRepo, id uuid.UUID) {
	u := &domain.User{
		ID:                id,
		TrustScore:        42,
		VerificationLevel: domain.VerificationNone,
		TrustStatus:       domain.TrustStatusNormal,
	}
	repo.mu.Lock()
	repo.byID[id] = u
	repo.mu.Unlock()
}

func newTestProfileService() (*service.ProfileService, *mockProfileRepo, *mockMediaRepo, *mockUserRepo, *mockMediaStore, *mockMatchingCache) {
	profileRepo := newMockProfileRepo()
	mediaRepo := newMockMediaRepo()
	userRepo := newMockUserRepo()
	mediaStore := newMockMediaStore()
	matchingCache := newMockMatchingCache()
	svc := service.NewProfileService(profileRepo, mediaRepo, userRepo, mediaStore, matchingCache)
	return svc, profileRepo, mediaRepo, userRepo, mediaStore, matchingCache
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestUpsertProfile_Success(t *testing.T) {
	t.Parallel()

	svc, profileRepo, _, _, _, _ := newTestProfileService()
	userID := uuid.New()

	_, err := svc.UpsertProfile(context.Background(), userID, service.UpsertProfileInput{
		DisplayName: "Alice",
		Bio:         "Hello world",
		Gender:      domain.GenderFemale,
		City:        "Almaty",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profile, err := profileRepo.GetByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("profile not found after upsert: %v", err)
	}
	if profile.DisplayName != "Alice" {
		t.Errorf("expected DisplayName=Alice, got %q", profile.DisplayName)
	}
}

func TestGetProfile_UsesCachedTrustScore(t *testing.T) {
	t.Parallel()

	svc, profileRepo, _, userRepo, _, cache := newTestProfileService()
	userID := uuid.New()

	// Seed dependencies.
	registerUserInRepo(userRepo, userID)
	profileRepo.Upsert(context.Background(), &domain.Profile{UserID: userID, DisplayName: "Bob"})
	cache.CacheTrustScore(context.Background(), userID, 99, time.Hour)

	view, err := svc.GetProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if view.TrustScore != 99 {
		t.Errorf("expected TrustScore=99 from cache, got %d", view.TrustScore)
	}
}

func TestGetProfile_FallsBackToDBTrustScore(t *testing.T) {
	t.Parallel()

	svc, profileRepo, _, userRepo, _, _ := newTestProfileService()
	userID := uuid.New()

	registerUserInRepo(userRepo, userID)
	profileRepo.Upsert(context.Background(), &domain.Profile{UserID: userID, DisplayName: "Carol"})

	// Cache is empty → GetCachedTrustScore returns -1 → must use DB value.
	view, err := svc.GetProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if view.TrustScore != 42 {
		t.Errorf("expected DB TrustScore=42, got %d", view.TrustScore)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _, _ := newTestProfileService()

	_, err := svc.GetProfile(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUploadPhoto_FirstPhotoSetsAvatar(t *testing.T) {
	t.Parallel()

	svc, _, mediaRepo, _, _, _ := newTestProfileService()
	userID := uuid.New()

	// Fake a JPEG magic header so the mock wraps it correctly.
	jpegData := append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, make([]byte, 100)...)
	media, err := svc.UploadPhoto(context.Background(), userID, jpegData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if media.ID == uuid.Nil {
		t.Error("expected non-nil media ID")
	}

	// Avatar should have been set automatically.
	if mediaRepo.avatar[userID] == "" {
		t.Error("expected avatar to be set for first photo")
	}
}

func TestUploadPhoto_ExceedsLimit(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _, _ := newTestProfileService()
	userID := uuid.New()

	jpegData := append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, make([]byte, 100)...)

	// Upload 8 photos.
	for i := 0; i < 8; i++ {
		if _, err := svc.UploadPhoto(context.Background(), userID, jpegData); err != nil {
			t.Fatalf("upload %d: unexpected error: %v", i+1, err)
		}
	}

	// 9th upload must be rejected.
	_, err := svc.UploadPhoto(context.Background(), userID, jpegData)
	if err == nil {
		t.Fatal("expected error when exceeding photo limit")
	}
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestDeletePhoto_Success(t *testing.T) {
	t.Parallel()

	svc, _, mediaRepo, _, mediaStore, _ := newTestProfileService()
	userID := uuid.New()

	// Pre-create a media record so DeletePhoto can find it.
	objectKey := "users/" + userID.String() + "/photos/test.jpg"
	mediaID := uuid.New()
	mediaStore.objects[objectKey] = []byte("fake image")
	mediaRepo.Create(context.Background(), &domain.Media{
		ID:        mediaID,
		UserID:    userID,
		ObjectKey: objectKey,
	})

	if err := svc.DeletePhoto(context.Background(), userID, mediaID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Record should be gone.
	if _, err := mediaRepo.GetByID(context.Background(), mediaID); !errors.Is(err, domain.ErrNotFound) {
		t.Error("expected media record to be deleted")
	}
}

func TestDeletePhoto_WrongOwner(t *testing.T) {
	t.Parallel()

	svc, _, mediaRepo, _, _, _ := newTestProfileService()
	ownerID := uuid.New()
	otherID := uuid.New()

	mediaID := uuid.New()
	mediaRepo.Create(context.Background(), &domain.Media{
		ID:        mediaID,
		UserID:    ownerID,
		ObjectKey: "users/" + ownerID.String() + "/photos/test.jpg",
	})

	err := svc.DeletePhoto(context.Background(), otherID, mediaID)
	if err == nil {
		t.Fatal("expected error when non-owner deletes photo")
	}
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestListPhotos_ReturnsPresignedURLs(t *testing.T) {
	t.Parallel()

	svc, _, mediaRepo, _, _, _ := newTestProfileService()
	userID := uuid.New()

	for i := 0; i < 3; i++ {
		mediaRepo.Create(context.Background(), &domain.Media{
			ID:        uuid.New(),
			UserID:    userID,
			ObjectKey: "users/" + userID.String() + "/photos/photo" + string(rune('0'+i)) + ".jpg",
			SortOrder: i,
		})
	}

	views, err := svc.ListPhotos(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(views) != 3 {
		t.Errorf("expected 3 photos, got %d", len(views))
	}
	for _, v := range views {
		if v.URL == "" {
			t.Error("expected non-empty presigned URL")
		}
	}
}

func TestUpsertProfile_ValidNiyyah_Succeeds(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _, _ := newTestProfileService()
	userID := uuid.New()

	validCases := []domain.Niyyah{
		domain.NiyyahNikahYear,
		domain.NiyyahSeriousMarriage,
		domain.NiyyahFriendship,
	}
	for _, n := range validCases {
		_, err := svc.UpsertProfile(context.Background(), userID, service.UpsertProfileInput{
			DisplayName: "Test",
			Niyyah:      n,
		})
		if err != nil {
			t.Errorf("niyyah %q: unexpected error: %v", n, err)
		}
	}
}

func TestUpsertProfile_InvalidNiyyah_ReturnsError(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _, _ := newTestProfileService()
	userID := uuid.New()

	_, err := svc.UpsertProfile(context.Background(), userID, service.UpsertProfileInput{
		DisplayName: "Test",
		Niyyah:      domain.Niyyah("invalid_niyyah"),
	})
	if err == nil {
		t.Fatal("expected error for invalid niyyah, got nil")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUpsertProfile_ValidMadhab_Succeeds(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _, _ := newTestProfileService()
	userID := uuid.New()

	validCases := []domain.Madhab{
		domain.MadhabHanafi,
		domain.MadhabShafii,
		domain.MadhabMaliki,
		domain.MadhabHanbali,
		domain.MadhabNone,
	}
	for _, m := range validCases {
		_, err := svc.UpsertProfile(context.Background(), userID, service.UpsertProfileInput{
			DisplayName: "Test",
			Madhab:      m,
		})
		if err != nil {
			t.Errorf("madhab %q: unexpected error: %v", m, err)
		}
	}
}

func TestUpsertProfile_InvalidMadhab_ReturnsError(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _, _ := newTestProfileService()
	userID := uuid.New()

	_, err := svc.UpsertProfile(context.Background(), userID, service.UpsertProfileInput{
		DisplayName: "Test",
		Madhab:      domain.Madhab("not_a_madhab"),
	})
	if err == nil {
		t.Fatal("expected error for invalid madhab, got nil")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

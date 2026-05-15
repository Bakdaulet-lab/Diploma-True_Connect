package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

const maxPhotosPerUser = 8

// ProfileView is the full public view of a profile returned to API consumers.
type ProfileView struct {
	UserID            uuid.UUID             `json:"user_id"`
	DisplayName       string                `json:"display_name"`
	Bio               string                `json:"bio,omitempty"`
	Gender            string                `json:"gender,omitempty"`
	BirthDate         *time.Time            `json:"birth_date,omitempty"`
	Age               *int                  `json:"age,omitempty"`
	City              string                `json:"city,omitempty"`
	LookingFor        string                `json:"looking_for,omitempty"`
	AvatarURL         string                `json:"avatar_url,omitempty"`
	Photos            []PhotoView           `json:"photos"`
	TrustScore        int                   `json:"trust_score"`
	Badge             string                `json:"badge"`
	VerificationLevel string                `json:"verification_level"`
	Prompts           []domain.PromptAnswer `json:"prompts,omitempty"`
	Niyyah            string                `json:"niyyah,omitempty"`
	Madhab            string                `json:"madhab,omitempty"`
	Languages         []string              `json:"languages,omitempty"`
	NoPhotoMode       bool                  `json:"no_photo_mode"`
	MaritalStatus     string                `json:"marital_status"`
}

// PhotoView is a single photo with a time-limited presigned URL.
type PhotoView struct {
	ID        uuid.UUID `json:"id"`
	URL       string    `json:"url"`
	SortOrder int       `json:"sort_order"`
}

// UpsertProfileInput contains validated data for creating or updating a profile.
type UpsertProfileInput struct {
	DisplayName string
	Bio         string
	Gender      domain.Gender
	Prompts     []domain.PromptAnswer
	BirthDate   *time.Time
	City        string
	Latitude    *float64
	Longitude   *float64
	LookingFor  domain.Gender
	Niyyah      domain.Niyyah
	Madhab      domain.Madhab
	Languages   []string
	NoPhotoMode bool
}

// ProfileService handles profile management and photo uploads.
type ProfileService struct {
	profileRepo   repository.ProfileRepository
	mediaRepo     repository.MediaRepository
	userRepo      repository.UserRepository
	mediaStore    repository.MediaStore
	matchingCache repository.MatchingCache
}

// NewProfileService creates a new profile service.
func NewProfileService(
	profileRepo repository.ProfileRepository,
	mediaRepo repository.MediaRepository,
	userRepo repository.UserRepository,
	mediaStore repository.MediaStore,
	matchingCache repository.MatchingCache,
) *ProfileService {
	return &ProfileService{
		profileRepo:   profileRepo,
		mediaRepo:     mediaRepo,
		userRepo:      userRepo,
		mediaStore:    mediaStore,
		matchingCache: matchingCache,
	}
}

// GetProfile returns the full profile view for a given user ID.
func (s *ProfileService) GetProfile(ctx context.Context, targetUserID uuid.UUID) (*ProfileView, error) {
	profile, err := s.profileRepo.GetByUserID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("get profile user: %w", err)
	}

	// Try cache first, fallback to PG value.
	trustScore := user.TrustScore
	if cached, err := s.matchingCache.GetCachedTrustScore(ctx, targetUserID); err == nil && cached >= 0 {
		trustScore = cached
	}

	photos, err := s.mediaRepo.ListByUser(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("get profile photos: %w", err)
	}

	photoViews := make([]PhotoView, 0, len(photos))
	for _, p := range photos {
		url, err := s.mediaStore.PresignedURL(ctx, p.ObjectKey)
		if err != nil {
			// Non-fatal: return empty URL so the rest of the profile is still served.
			url = ""
		}
		photoViews = append(photoViews, PhotoView{
			ID:        p.ID,
			URL:       url,
			SortOrder: p.SortOrder,
		})
	}

	scoreModel := &domain.TrustScore{Score: trustScore}

	view := &ProfileView{
		UserID:            profile.UserID,
		DisplayName:       profile.DisplayName,
		Bio:               profile.Bio,
		Gender:            string(profile.Gender),
		BirthDate:         profile.BirthDate,
		City:              profile.City,
		LookingFor:        string(profile.LookingFor),
		AvatarURL:         fixAvatarURL(profile.AvatarURL),
		Photos:            photoViews,
		TrustScore:        trustScore,
		Badge:             scoreModel.GetBadge(),
		VerificationLevel: string(user.VerificationLevel),
		Prompts:           profile.Prompts,
		Niyyah:            string(profile.Niyyah),
		Madhab:            string(profile.Madhab),
		Languages:         profile.Languages,
		NoPhotoMode:       profile.NoPhotoMode,
		MaritalStatus:     profile.MaritalStatus,
	}

	if profile.BirthDate != nil {
		age := computeAge(*profile.BirthDate)
		view.Age = &age
	}

	return view, nil
}

// validNiyyahs is the set of accepted niyyah enum values.
var validNiyyahs = map[domain.Niyyah]bool{
	domain.NiyyahNikahYear:       true,
	domain.NiyyahSeriousMarriage: true,
	domain.NiyyahFriendship:      true,
}

// validMadhabs is the set of accepted madhab enum values.
var validMadhabs = map[domain.Madhab]bool{
	domain.MadhabHanafi:  true,
	domain.MadhabShafii:  true,
	domain.MadhabMaliki:  true,
	domain.MadhabHanbali: true,
	domain.MadhabNone:    true,
}

// UpsertProfile creates or updates the caller's own profile.
func (s *ProfileService) UpsertProfile(ctx context.Context, userID uuid.UUID, input UpsertProfileInput) (*domain.Profile, error) {
	if input.Niyyah != "" && !validNiyyahs[input.Niyyah] {
		return nil, fmt.Errorf("upsert profile: invalid niyyah %q: %w", input.Niyyah, domain.ErrInvalidInput)
	}
	if input.Madhab != "" && !validMadhabs[input.Madhab] {
		return nil, fmt.Errorf("upsert profile: invalid madhab %q: %w", input.Madhab, domain.ErrInvalidInput)
	}

	profile := &domain.Profile{
		UserID:      userID,
		DisplayName: input.DisplayName,
		Bio:         input.Bio,
		Gender:      input.Gender,
		Prompts:     input.Prompts,
		BirthDate:   input.BirthDate,
		City:        input.City,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		LookingFor:  input.LookingFor,
		Niyyah:      input.Niyyah,
		Madhab:      input.Madhab,
		Languages:   input.Languages,
		NoPhotoMode: input.NoPhotoMode,
	}

	if err := s.profileRepo.Upsert(ctx, profile); err != nil {
		return nil, fmt.Errorf("upsert profile: %w", err)
	}

	return profile, nil
}

// UploadPhoto stores a photo for the user and returns the new media record.
// If this is the user's first photo it is set as the profile avatar automatically.
func (s *ProfileService) UploadPhoto(ctx context.Context, userID uuid.UUID, data []byte) (*domain.Media, error) {
	count, err := s.mediaRepo.CountByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("upload photo: checking count: %w", err)
	}
	if count >= maxPhotosPerUser {
		return nil, fmt.Errorf("upload photo: %w", domain.ErrForbidden)
	}

	objectKey, err := s.mediaStore.UploadPhoto(ctx, userID, data)
	if err != nil {
		return nil, fmt.Errorf("upload photo: %w", err)
	}

	media := &domain.Media{
		ID:        uuid.New(),
		UserID:    userID,
		ObjectKey: objectKey,
		MediaType: "photo",
		SortOrder: count, // append after existing photos
	}

	if err := s.mediaRepo.Create(ctx, media); err != nil {
		// Best-effort cleanup: delete the uploaded object so it doesn't orphan.
		_ = s.mediaStore.DeletePhoto(ctx, objectKey)
		return nil, fmt.Errorf("upload photo: saving record: %w", err)
	}

	// Every uploaded photo updates the avatar so the most recent photo is always shown.
	if err := s.mediaRepo.UpdateAvatar(ctx, userID, objectKey); err != nil {
		// Non-fatal: existing avatar remains if this fails.
		_ = err
	}

	return media, nil
}

// DeletePhoto removes a photo for the authenticated user.
func (s *ProfileService) DeletePhoto(ctx context.Context, userID, photoID uuid.UUID) error {
	media, err := s.mediaRepo.GetByID(ctx, photoID)
	if err != nil {
		return fmt.Errorf("delete photo: %w", err)
	}
	if media.UserID != userID {
		return fmt.Errorf("delete photo: %w", domain.ErrForbidden)
	}

	// Remove from MinIO first, then DB. If MinIO fails, we keep the DB record
	// so the user can retry. If DB fails after MinIO succeeds, the object is
	// orphaned (acceptable at MVP scale — can be cleaned with a periodic sweep).
	if err := s.mediaStore.DeletePhoto(ctx, media.ObjectKey); err != nil {
		return fmt.Errorf("delete photo: removing object: %w", err)
	}

	if err := s.mediaRepo.Delete(ctx, photoID, userID); err != nil {
		return fmt.Errorf("delete photo: removing record: %w", err)
	}

	return nil
}

// ListPhotos returns all photos for a user with presigned download URLs.
func (s *ProfileService) ListPhotos(ctx context.Context, userID uuid.UUID) ([]PhotoView, error) {
	photos, err := s.mediaRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list photos: %w", err)
	}

	views := make([]PhotoView, 0, len(photos))
	for _, p := range photos {
		url, _ := s.mediaStore.PresignedURL(ctx, p.ObjectKey)
		views = append(views, PhotoView{
			ID:        p.ID,
			URL:       url,
			SortOrder: p.SortOrder,
		})
	}

	return views, nil
}

// computeAge calculates whole years elapsed since birthDate.
func computeAge(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()
	if now.Month() < birthDate.Month() ||
		(now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		age--
	}
	return age
}

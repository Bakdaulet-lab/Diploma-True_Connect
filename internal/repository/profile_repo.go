package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// ProfileRepository handles CRUD for public social profiles.
type ProfileRepository interface {
	// Upsert creates or updates a profile. Both insert and update are handled.
	Upsert(ctx context.Context, profile *domain.Profile) error

	// GetByUserID returns the profile for a user.
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Profile, error)

	// FindCandidates returns user IDs to show in the matching feed.
	// It queries within maxDistanceMeters of the given lat/lon,
	// filtering by looking_for gender and excluding excludeIDs.
	FindCandidates(ctx context.Context, opts FindCandidatesOpts) ([]*CandidateRow, error)

	// GetLeaderboard returns the list of users with the highest trust scores.
	GetLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardEntry, error)

	// SetMarriedViaApp sets marital_status = 'married_via_app' for the given user's profile.
	SetMarriedViaApp(ctx context.Context, userID uuid.UUID) error
}

// FindCandidatesOpts carries all parameters for the candidate search query.
type FindCandidatesOpts struct {
	RequesterID       uuid.UUID
	LookingFor        domain.Gender
	AgeRangeMin       int
	AgeRangeMax       int
	MaxDistanceMeters *int
	RequesterLat      *float64
	RequesterLon      *float64
	ExcludeIDs        []uuid.UUID // already-seen or already-matched user IDs
	Limit             int
	AllowedNiyyahs    []string // niyyah-compatibility filter; nil = no filter
	MadhabFilter      *string  // nil = no filter
	LanguageFilter    []string // nil = no filter; requires overlap with candidate
}

// CandidateRow is the minimal data returned per matching candidate.
type CandidateRow struct {
	UserID         uuid.UUID             `json:"user_id"`
	DisplayName    string                `json:"display_name"`
	AvatarURL      string                `json:"avatar_url"`
	City           string                `json:"city"`
	Prompts        []domain.PromptAnswer `json:"prompts"`
	TrustScore     int                   `json:"trust_score"`
	Niyyah         string                `json:"niyyah"`
	Madhab         string                `json:"madhab"`
	Languages      []string              `json:"languages"`
	NoPhotoMode    bool                  `json:"no_photo_mode"`
	IsKYCVerified  bool                  `json:"is_kyc_verified"`
}

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
}

// FindCandidatesOpts carries all parameters for the candidate search query.
type FindCandidatesOpts struct {
	RequesterID       uuid.UUID
	Lat               float64
	Lon               float64
	MaxDistanceMeters float64
	LookingFor        domain.Gender
	AgeRangeMin       int
	AgeRangeMax       int
	ExcludeIDs        []uuid.UUID // already-seen or already-matched user IDs
	Limit             int
}

// CandidateRow is the minimal data returned per matching candidate.
type CandidateRow struct {
	UserID      uuid.UUID
	DisplayName string
	AvatarURL   string
	City        string
	TrustScore  int
	DistanceKm  float64
}

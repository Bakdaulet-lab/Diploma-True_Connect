package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

const (
	candidateBatchSize = 20
	seenSetTTL         = 24 * time.Hour
	maxDistanceDefault = 50_000.0 // 50 km in metres
)

// CandidateView is a matching card shown in the swipe feed.
type CandidateView struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	City        string    `json:"city,omitempty"`
	TrustScore  int       `json:"trust_score"`
	DistanceKm  float64   `json:"distance_km"`
}

// LikeResult tells the caller whether a mutual match occurred.
type LikeResult struct {
	Matched bool      `json:"matched"`
	MatchID uuid.UUID `json:"match_id,omitempty"`
}

// MatchingService handles candidate retrieval, likes, and passes.
type MatchingService struct {
	profileRepo   repository.ProfileRepository
	matchRepo     repository.MatchRepository
	settingsRepo  repository.UserSettingsRepository
	matchingCache repository.MatchingCache
}

// NewMatchingService creates a new matching service.
func NewMatchingService(
	profileRepo repository.ProfileRepository,
	matchRepo repository.MatchRepository,
	settingsRepo repository.UserSettingsRepository,
	matchingCache repository.MatchingCache,
) *MatchingService {
	return &MatchingService{
		profileRepo:   profileRepo,
		matchRepo:     matchRepo,
		settingsRepo:  settingsRepo,
		matchingCache: matchingCache,
	}
}

// GetCandidates returns a batch of profiles for the swipe feed.
func (s *MatchingService) GetCandidates(ctx context.Context, userID uuid.UUID) ([]*CandidateView, error) {
	// Load requester's profile to get their current location.
	requesterProfile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get candidates: loading requester profile: %w", err)
	}

	// Load preferences for distance and age filters.
	settings, err := s.settingsRepo.Get(ctx, userID)
	if err != nil {
		settings = domain.DefaultSettings(userID.String())
	}

	maxDistanceM := float64(settings.MaxDistanceKm) * 1000.0
	if maxDistanceM <= 0 {
		maxDistanceM = maxDistanceDefault
	}

	// Retrieve already-seen IDs from Redis to exclude them.
	seenIDs, err := s.matchingCache.GetSeenIDs(ctx, userID)
	if err != nil {
		seenIDs = []uuid.UUID{}
	}

	// Also exclude users the requester has already matched with.
	matches, err := s.matchRepo.ListMatches(ctx, userID, 200, 0)
	if err == nil {
		for _, m := range matches {
			other := m.UserAID
			if other == userID {
				other = m.UserBID
			}
			seenIDs = append(seenIDs, other)
		}
	}

	opts := repository.FindCandidatesOpts{
		RequesterID:       userID,
		Lat:               requesterProfile.Latitude,
		Lon:               requesterProfile.Longitude,
		MaxDistanceMeters: maxDistanceM,
		LookingFor:        requesterProfile.LookingFor,
		AgeRangeMin:       settings.AgeRangeMin,
		AgeRangeMax:       settings.AgeRangeMax,
		ExcludeIDs:        seenIDs,
		Limit:             candidateBatchSize,
	}

	rows, err := s.profileRepo.FindCandidates(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("get candidates: querying: %w", err)
	}

	// Mark all returned candidates as seen so they won't appear again today.
	newSeenIDs := make([]uuid.UUID, 0, len(rows))
	views := make([]*CandidateView, 0, len(rows))
	for _, row := range rows {
		newSeenIDs = append(newSeenIDs, row.UserID)
		views = append(views, &CandidateView{
			UserID:      row.UserID,
			DisplayName: row.DisplayName,
			AvatarURL:   row.AvatarURL,
			City:        row.City,
			TrustScore:  row.TrustScore,
			DistanceKm:  row.DistanceKm,
		})
	}

	if len(newSeenIDs) > 0 {
		_ = s.matchingCache.AddSeen(ctx, userID, newSeenIDs, seenSetTTL)
	}

	return views, nil
}

// Like records that userID likes targetID.
func (s *MatchingService) Like(ctx context.Context, userID, targetID uuid.UUID) (*LikeResult, error) {
	if userID == targetID {
		return nil, fmt.Errorf("like: %w", domain.ErrInvalidInput)
	}

	matched, matchID, err := s.matchRepo.RecordLike(ctx, userID, targetID)
	if err != nil {
		return nil, fmt.Errorf("like: %w", err)
	}

	// Add to seen set regardless of outcome.
	_ = s.matchingCache.AddSeen(ctx, userID, []uuid.UUID{targetID}, seenSetTTL)

	return &LikeResult{Matched: matched, MatchID: matchID}, nil
}

// Pass records that userID passes on targetID (adds to seen, no DB row).
func (s *MatchingService) Pass(ctx context.Context, userID, targetID uuid.UUID) error {
	if userID == targetID {
		return fmt.Errorf("pass: %w", domain.ErrInvalidInput)
	}

	_ = s.matchRepo.RecordPass(ctx, userID, targetID)
	_ = s.matchingCache.AddSeen(ctx, userID, []uuid.UUID{targetID}, seenSetTTL)

	return nil
}

// GetMatchByID returns a match by ID, verifying the user is a participant.
func (s *MatchingService) GetMatchByID(ctx context.Context, matchID, userID uuid.UUID) (*domain.Match, error) {
	match, err := s.matchRepo.GetMatch(ctx, matchID, userID)
	if err != nil {
		return nil, fmt.Errorf("get match: %w", err)
	}
	return match, nil
}

// ListMatches returns paginated mutual matches for the authenticated user.
func (s *MatchingService) ListMatches(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*domain.Match, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	offset := (page - 1) * perPage
	matches, err := s.matchRepo.ListMatches(ctx, userID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("list matches: %w", err)
	}

	return matches, nil
}

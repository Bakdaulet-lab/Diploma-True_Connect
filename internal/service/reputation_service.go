package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

const trustScoreCacheTTL = 10 * time.Minute

// ReputationService handles trust score computation and caching.
type ReputationService struct {
	graphRepo     repository.TrustGraphRepository
	userRepo      repository.UserRepository
	profileRepo   repository.ProfileRepository
	matchingCache repository.MatchingCache
}

// NewReputationService creates a new reputation service.
func NewReputationService(
	graphRepo repository.TrustGraphRepository,
	userRepo repository.UserRepository,
	profileRepo repository.ProfileRepository,
	matchingCache repository.MatchingCache,
) *ReputationService {
	return &ReputationService{
		graphRepo:     graphRepo,
		userRepo:      userRepo,
		profileRepo:   profileRepo,
		matchingCache: matchingCache,
	}
}

// RecalculateScore computes a fresh trust score and persists it across all stores.
func (s *ReputationService) RecalculateScore(ctx context.Context, userID uuid.UUID) (int, error) {
	score, err := s.graphRepo.ComputeTrustScore(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("recalculate score: computing: %w", err)
	}

	if err := s.userRepo.UpdateTrustScore(ctx, userID, score); err != nil {
		return 0, fmt.Errorf("recalculate score: updating PG: %w", err)
	}

	if err := s.graphRepo.UpdateTrustScore(ctx, userID, score); err != nil {
		return 0, fmt.Errorf("recalculate score: updating Neo4j: %w", err)
	}

	_ = s.matchingCache.CacheTrustScore(ctx, userID, score, trustScoreCacheTTL)

	return score, nil
}

// GetScore returns a user's trust score, reading from Redis cache first.
func (s *ReputationService) GetScore(ctx context.Context, userID uuid.UUID) (*domain.TrustScore, error) {
	cached, err := s.matchingCache.GetCachedTrustScore(ctx, userID)
	if err == nil && cached >= 0 {
		return &domain.TrustScore{UserID: userID, Score: cached}, nil
	}

	score, err := s.graphRepo.ComputeTrustScore(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get score: computing: %w", err)
	}

	_ = s.matchingCache.CacheTrustScore(ctx, userID, score, trustScoreCacheTTL)

	return &domain.TrustScore{UserID: userID, Score: score}, nil
}

// GetLeaderboard returns the top N users with the highest reputation scores.
func (s *ReputationService) GetLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 100 // Cap the leaderboard limit
	}
	board, err := s.profileRepo.GetLeaderboard(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("get leaderboard: %w", err)
	}
	return board, nil
}

// RecalculateAllScores processes all users to decay older scores or promote active ones.
func (s *ReputationService) RecalculateAllScores(ctx context.Context) {
	// A naive implementation tracking through pagination
	// In a massive system, this would be highly optimized or handled directly via DB jobs.
	limit := 100
	offset := 0

	for {
		users, err := s.userRepo.ListByTrustStatus(ctx, domain.TrustStatusNormal, limit, offset)
		if err != nil {
			fmt.Printf("RecalculateAllScores error at offset %d: %v\n", offset, err)
			break
		}
		if len(users) == 0 {
			break
		}

		for _, u := range users {
			_, _ = s.RecalculateScore(ctx, u.ID)
		}

		if len(users) < limit {
			break
		}
		offset += limit
	}
}

// GetSybilClusters returns clusters of suspected bot/Sybil accounts.
func (s *ReputationService) GetSybilClusters(ctx context.Context) ([]repository.SybilCluster, error) {
	return s.graphRepo.DetectSybilClusters(ctx)
}

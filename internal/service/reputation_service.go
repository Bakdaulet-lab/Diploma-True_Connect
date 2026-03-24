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
	matchingCache repository.MatchingCache
}

// NewReputationService creates a new reputation service.
func NewReputationService(
	graphRepo repository.TrustGraphRepository,
	userRepo repository.UserRepository,
	matchingCache repository.MatchingCache,
) *ReputationService {
	return &ReputationService{
		graphRepo:     graphRepo,
		userRepo:      userRepo,
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

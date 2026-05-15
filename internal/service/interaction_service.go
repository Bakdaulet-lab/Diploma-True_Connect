package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

const ratingCooldown = 7 * 24 * time.Hour

// InteractionService handles rating submissions and confirmations.
type InteractionService struct {
	interactionRepo repository.InteractionRepository
	matchRepo       repository.MatchRepository
	graphRepo       repository.TrustGraphRepository
	uow             repository.UnitOfWork
	eventCh         chan<- uuid.UUID
}

// NewInteractionService creates a new interaction service.
func NewInteractionService(
	interactionRepo repository.InteractionRepository,
	matchRepo repository.MatchRepository,
	graphRepo repository.TrustGraphRepository,
	uow repository.UnitOfWork,
	eventCh chan<- uuid.UUID,
) *InteractionService {
	return &InteractionService{
		interactionRepo: interactionRepo,
		matchRepo:       matchRepo,
		graphRepo:       graphRepo,
		uow:             uow,
		eventCh:         eventCh,
	}
}

// SubmitRating records a rating from raterID to ratedID after a meeting.
func (s *InteractionService) SubmitRating(
	ctx context.Context,
	raterID, ratedID uuid.UUID,
	rating int,
	interactionContext string,
	comment string,
) (*domain.Interaction, error) {
	if raterID == ratedID {
		return nil, fmt.Errorf("submit rating: %w", domain.ErrInvalidInput)
	}

	if rating < 1 || rating > 5 {
		return nil, fmt.Errorf("submit rating: rating must be 1-5: %w", domain.ErrInvalidInput)
	}

	matched, err := s.matchRepo.IsMatched(ctx, raterID, ratedID)
	if err != nil {
		return nil, fmt.Errorf("submit rating: checking match: %w", err)
	}
	if !matched {
		return nil, fmt.Errorf("submit rating: users must be matched: %w", domain.ErrForbidden)
	}

	after := time.Now().Add(-ratingCooldown).Format(time.RFC3339)
	exists, err := s.interactionRepo.ExistsBetweenUsersAfter(ctx, raterID, ratedID, after)
	if err != nil {
		return nil, fmt.Errorf("submit rating: checking cooldown: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("submit rating: %w", domain.ErrRateLimitExceeded)
	}

	interaction := &domain.Interaction{
		RaterID:    raterID,
		RatedID:    ratedID,
		Rating:     rating,
		Context:    interactionContext,
		Comment:    comment,
		IsVerified: false,
	}

	err = s.uow.Do(ctx, func(txCtx context.Context) error {
		if err := s.interactionRepo.Create(txCtx, interaction); err != nil {
			return fmt.Errorf("creating interaction: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("submit rating: %w", err)
	}

	// Run Neo4j updates ONLY after successful PG commit (avoids split-brain).
	// A better approach would be the Outbox pattern, but this is a solid mitigation here.
	if err := s.graphRepo.AddRating(ctx, raterID, ratedID, rating, interactionContext, false); err != nil {
		slog.Default().Warn("failed to update neo4j graph rating",
			slog.String("rater_id", raterID.String()),
			slog.String("rated_id", ratedID.String()),
			slog.String("error", err.Error()),
		)
	}
	select {
	case s.eventCh <- ratedID:
	default:
	}

	return interaction, nil
}

// ConfirmInteraction allows the rated user to confirm a meeting happened.
func (s *InteractionService) ConfirmInteraction(ctx context.Context, interactionID, userID uuid.UUID) error {
	interaction, err := s.interactionRepo.GetByID(ctx, interactionID)
	if err != nil {
		return fmt.Errorf("confirm interaction: %w", err)
	}

	if interaction.RatedID != userID {
		return fmt.Errorf("confirm interaction: only the rated user can confirm: %w", domain.ErrForbidden)
	}

	if interaction.IsVerified {
		return nil // already confirmed — idempotent
	}

	err = s.uow.Do(ctx, func(txCtx context.Context) error {
		if err := s.interactionRepo.ConfirmInteraction(txCtx, interactionID); err != nil {
			return fmt.Errorf("updating PG: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("confirm interaction: %w", err)
	}

	// Add a verified rating edge in Neo4j (after PG commit).
	if err := s.graphRepo.AddRating(ctx, interaction.RaterID, interaction.RatedID, interaction.Rating, interaction.Context, true); err != nil {
		slog.Default().Warn("failed to add verified graph rating",
			slog.String("interaction_id", interactionID.String()),
			slog.String("error", err.Error()),
		)
	}

	// Record the confirmed meeting in Neo4j.
	if err := s.graphRepo.AddMeeting(ctx, interaction.RaterID, interaction.RatedID, true); err != nil {
		slog.Default().Warn("failed to add meeting edge",
			slog.String("interaction_id", interactionID.String()),
			slog.String("error", err.Error()),
		)
	}
	select {
	case s.eventCh <- interaction.RatedID:
	default:
	}

	return nil
}

// GetByRatedUser returns paginated interactions where the user was rated.
func (s *InteractionService) GetByRatedUser(ctx context.Context, ratedID uuid.UUID, limit, offset int) ([]domain.Interaction, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	interactions, err := s.interactionRepo.GetByRatedUser(ctx, ratedID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get rated interactions: %w", err)
	}

	return interactions, nil
}

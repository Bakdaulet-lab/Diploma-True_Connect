package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/pkg/imam"
	"github.com/trueconnect/backend/internal/repository"
)

// ImamService handles imam catalog lookup and nikah confirmation.
type ImamService struct {
	matchRepo   repository.MatchRepository
	profileRepo repository.ProfileRepository
	notifSvc    *NotificationService
	log         *slog.Logger
}

// NewImamService creates a new imam service.
func NewImamService(
	matchRepo repository.MatchRepository,
	profileRepo repository.ProfileRepository,
	notifSvc *NotificationService,
	log *slog.Logger,
) *ImamService {
	return &ImamService{
		matchRepo:   matchRepo,
		profileRepo: profileRepo,
		notifSvc:    notifSvc,
		log:         log,
	}
}

// ListImams returns imams in the given city from the embedded catalog.
func (s *ImamService) ListImams(_ context.Context, city string) ([]domain.Imam, error) {
	return imam.ListByCity(city), nil
}

// ConfirmNikah marks a match as imam-confirmed and sets both participants
// as married_via_app so they are removed from the discovery feed.
// callerID must be one of the match participants.
func (s *ImamService) ConfirmNikah(ctx context.Context, matchID, imamID, callerID uuid.UUID) error {
	if _, ok := imam.GetByID(imamID); !ok {
		return fmt.Errorf("confirm nikah: imam not found: %w", domain.ErrNotFound)
	}

	match, err := s.matchRepo.GetMatch(ctx, matchID, callerID)
	if err != nil {
		return fmt.Errorf("confirm nikah: %w", err)
	}
	if match.ImamConfirmed {
		return nil // idempotent
	}
	if !match.FamilyIntroDone {
		return fmt.Errorf("confirm nikah: family introduction required first: %w", domain.ErrForbidden)
	}

	if err := s.matchRepo.MarkImamConfirmed(ctx, matchID); err != nil {
		return fmt.Errorf("confirm nikah: %w", err)
	}

	for _, uid := range []uuid.UUID{match.UserAID, match.UserBID} {
		if err := s.profileRepo.SetMarriedViaApp(ctx, uid); err != nil {
			s.log.Warn("confirm nikah: failed to set marital status",
				slog.String("user_id", uid.String()),
				slog.String("error", err.Error()),
			)
		}
	}

	if s.notifSvc != nil {
		otherID := match.UserBID
		if callerID == match.UserBID {
			otherID = match.UserAID
		}
		matchIDCopy := matchID
		for _, uid := range []uuid.UUID{callerID, otherID} {
			if err := s.notifSvc.Create(ctx, &domain.Notification{
				UserID:   uid,
				Type:     domain.NotificationTypeSystem,
				EntityID: &matchIDCopy,
			}); err != nil {
				s.log.Warn("confirm nikah: notification failed",
					slog.String("user_id", uid.String()),
					slog.String("error", err.Error()),
				)
			}
		}
	}

	return nil
}

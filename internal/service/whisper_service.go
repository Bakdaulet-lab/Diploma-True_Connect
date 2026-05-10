package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/pkg/crypto"
	"github.com/trueconnect/backend/internal/repository"
)

const whisperStrikeThreshold = 3

// WhisperService manages anonymous whisper feedback submission.
type WhisperService struct {
	whisperRepo   repository.WhisperRepository
	matchRepo     repository.MatchRepository
	notifSvc      *NotificationService
	encryptionKey []byte
	log           *slog.Logger
}

// NewWhisperService creates a new whisper service.
func NewWhisperService(
	whisperRepo repository.WhisperRepository,
	matchRepo repository.MatchRepository,
	notifSvc *NotificationService,
	encryptionKey []byte,
	log *slog.Logger,
) *WhisperService {
	return &WhisperService{
		whisperRepo:   whisperRepo,
		matchRepo:     matchRepo,
		notifSvc:      notifSvc,
		encryptionKey: encryptionKey,
		log:           log,
	}
}

// SubmitWhisper records an anonymous feedback report from reporter against reported.
// The reporter must have an active mutual match with the reported user (matchID proves it).
// If the match had family_intro_done (physical meeting verified), the strike weight is 2.
// Three weighted strikes trigger admin_flagged on all uncounted reports.
func (s *WhisperService) SubmitWhisper(ctx context.Context, reporterID, reportedID, matchID uuid.UUID, feedback string) error {
	// GetMatch already enforces that reporterID is a participant in matchID.
	match, err := s.matchRepo.GetMatch(ctx, matchID, reporterID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("submit whisper: match not found or access denied: %w", domain.ErrNotFound)
		}
		return fmt.Errorf("submit whisper: %w", err)
	}

	// Confirm the reported user is the other party in the same match.
	if match.UserAID != reportedID && match.UserBID != reportedID {
		return fmt.Errorf("submit whisper: reported user is not in the given match: %w", domain.ErrForbidden)
	}

	strikeWeight := 1
	if match.FamilyIntroDone {
		// Physical meeting was verified — heavier weight for this report.
		strikeWeight = 2
	}

	feedbackEncrypted, err := crypto.Encrypt([]byte(feedback), s.encryptionKey)
	if err != nil {
		return fmt.Errorf("submit whisper: encrypting feedback: %w", err)
	}

	if err := s.whisperRepo.Create(ctx, reporterID, reportedID, matchID, feedbackEncrypted, strikeWeight); err != nil {
		return fmt.Errorf("submit whisper: storing report: %w", err)
	}

	total, err := s.whisperRepo.CountWeightedStrikes(ctx, reportedID)
	if err != nil {
		s.log.Warn("whisper: counting strikes failed",
			slog.String("reported_id", reportedID.String()),
			slog.String("error", err.Error()),
		)
		return nil
	}

	if total >= whisperStrikeThreshold {
		if flagErr := s.whisperRepo.FlagForAdmin(ctx, reportedID); flagErr != nil {
			s.log.Warn("whisper: flagging for admin failed",
				slog.String("reported_id", reportedID.String()),
				slog.String("error", flagErr.Error()),
			)
		} else {
			s.log.Info("whisper: user flagged for admin review",
				slog.String("reported_id", reportedID.String()),
				slog.Int("weighted_strikes", total),
			)
		}
	}

	return nil
}

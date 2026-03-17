package worker

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/service"
)

// TrustEngine listens for rating events and recalculates trust scores.
type TrustEngine struct {
	eventCh   <-chan uuid.UUID
	reputeSvc *service.ReputationService
	log       *slog.Logger
}

// NewTrustEngine creates a new trust engine worker.
func NewTrustEngine(eventCh <-chan uuid.UUID, reputeSvc *service.ReputationService, log *slog.Logger) *TrustEngine {
	return &TrustEngine{
		eventCh:   eventCh,
		reputeSvc: reputeSvc,
		log:       log,
	}
}

// Run processes rating events until the context is cancelled.
func (e *TrustEngine) Run(ctx context.Context) {
	e.log.Info("trust engine started")
	for {
		select {
		case <-ctx.Done():
			e.log.Info("trust engine stopped")
			return
		case userID := <-e.eventCh:
			score, err := e.reputeSvc.RecalculateScore(ctx, userID)
			if err != nil {
				e.log.Error("trust engine: recalculation failed",
					slog.String("user_id", userID.String()),
					slog.String("error", err.Error()),
				)
				continue
			}
			e.log.Info("trust engine: score updated",
				slog.String("user_id", userID.String()),
				slog.Int("new_score", score),
			)
		}
	}
}

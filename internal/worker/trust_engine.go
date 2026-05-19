package worker

import (
	"context"
	"log/slog"
	"time"

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

	// Feature B: Background Recalculation every 24 hours
	decayTicker := time.NewTicker(24 * time.Hour)
	defer decayTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			e.log.Info("trust engine stopped, draining remaining events")
			// Bound the drain so a slow/stuck recalculation can't hang shutdown.
			drainCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			for {
				select {
				case <-drainCtx.Done():
					e.log.Warn("trust engine drain timed out, dropping remaining events")
					cancel()
					return
				case userID := <-e.eventCh:
					if _, err := e.reputeSvc.RecalculateScore(drainCtx, userID); err != nil {
						e.log.Error("trust engine drain: recalculation failed",
							slog.String("user_id", userID.String()),
							slog.String("error", err.Error()),
						)
					}
				default:
					cancel()
					return
				}
			}
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
		case <-decayTicker.C:
			e.log.Info("trust engine: running periodic score background recalculation/decay")
			// Trigger the global recalculation logic via the service
			e.reputeSvc.RecalculateAllScores(ctx)
		}
	}
}

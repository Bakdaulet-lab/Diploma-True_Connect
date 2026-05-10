package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
	"github.com/trueconnect/backend/internal/service"
)

// NiyyahTimerWorker runs daily and notifies both match participants when their
// 90-day niyyah commitment window expires without a family intro milestone.
type NiyyahTimerWorker struct {
	matchRepo repository.MatchRepository
	notifSvc  *service.NotificationService
	log       *slog.Logger
	interval  time.Duration
}

// NewNiyyahTimerWorker creates a new niyyah timer worker.
func NewNiyyahTimerWorker(
	matchRepo repository.MatchRepository,
	notifSvc *service.NotificationService,
	log *slog.Logger,
) *NiyyahTimerWorker {
	return &NiyyahTimerWorker{
		matchRepo: matchRepo,
		notifSvc:  notifSvc,
		log:       log,
		interval:  24 * time.Hour,
	}
}

// Run processes expired niyyah timers until the context is cancelled.
func (w *NiyyahTimerWorker) Run(ctx context.Context) {
	w.log.Info("niyyah timer worker started")
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.process(ctx)

	for {
		select {
		case <-ctx.Done():
			w.log.Info("niyyah timer worker stopped")
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *NiyyahTimerWorker) process(ctx context.Context) {
	matches, err := w.matchRepo.FindExpiredNiyyahMatches(ctx)
	if err != nil {
		w.log.Error("niyyah timer: querying expired matches", slog.String("error", err.Error()))
		return
	}

	for _, m := range matches {
		w.log.Info("niyyah timer expired", slog.String("match_id", m.ID.String()))
		matchID := m.ID
		for _, uid := range []uuid.UUID{m.UserAID, m.UserBID} {
			if err := w.notifSvc.Create(ctx, &domain.Notification{
				UserID:   uid,
				Type:     domain.NotificationTypeNiyyahExpired,
				EntityID: &matchID,
			}); err != nil {
				w.log.Warn("niyyah timer: sending notification",
					slog.String("user_id", uid.String()),
					slog.String("error", err.Error()),
				)
			}
		}
	}

	if len(matches) > 0 {
		w.log.Info("niyyah timer processed expired matches", slog.Int("count", len(matches)))
	}
}

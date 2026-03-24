package worker

import (
	"context"
	"log/slog"

	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/provider"
	"github.com/trueconnect/backend/internal/repository"
)

type PushWorker struct {
	provider provider.PushProvider
	userRepo repository.UserRepository
	eventCh  <-chan domain.PushEvent
	log      *slog.Logger
}

func NewPushWorker(p provider.PushProvider, u repository.UserRepository, ch <-chan domain.PushEvent, log *slog.Logger) *PushWorker {
	return &PushWorker{provider: p, userRepo: u, eventCh: ch, log: log}
}

func (w *PushWorker) Run(ctx context.Context) {
	w.log.Info("push notification worker started")
	for {
		select {
		case <-ctx.Done():
			w.log.Info("push notification worker stopped")
			return
		case e := <-w.eventCh:
			user, err := w.userRepo.GetByID(ctx, e.UserID)
			if err != nil || user.FCMToken == nil || *user.FCMToken == "" {
				continue
			}
			err = w.provider.SendTargeted(ctx, *user.FCMToken, e.Title, e.Body, e.Data)
			if err != nil {
				w.log.Error("failed to send push", slog.String("user", e.UserID.String()), slog.String("err", err.Error()))
			}
		}
	}
}

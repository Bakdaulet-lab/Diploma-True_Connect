package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
	"github.com/trueconnect/backend/internal/service"
)

// SybilDetector runs periodic community detection to flag fake account clusters.
type SybilDetector struct {
	graphRepo repository.TrustGraphRepository
	userRepo  repository.UserRepository
	userSvc   *service.UserService
	log       *slog.Logger
	interval  time.Duration
}

// NewSybilDetector creates a new sybil detection worker.
func NewSybilDetector(
	graphRepo repository.TrustGraphRepository,
	userRepo repository.UserRepository,
	userSvc *service.UserService,
	log *slog.Logger,
	interval time.Duration,
) *SybilDetector {
	return &SybilDetector{
		graphRepo: graphRepo,
		userRepo:  userRepo,
		userSvc:   userSvc,
		log:       log,
		interval:  interval,
	}
}

// Run starts the periodic sybil detection loop.
func (d *SybilDetector) Run(ctx context.Context) {
	d.log.Info("sybil detector started", slog.Duration("interval", d.interval))

	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.log.Info("sybil detector stopped")
			return
		case <-ticker.C:
			d.detect(ctx)
		}
	}
}

func (d *SybilDetector) detect(ctx context.Context) {
	d.log.Info("sybil detector: running scan")

	clusters, err := d.graphRepo.DetectSybilClusters(ctx)
	if err != nil {
		d.log.Error("sybil detector: detection failed", slog.String("error", err.Error()))
		return
	}

	if len(clusters) == 0 {
		d.log.Info("sybil detector: no suspicious clusters found")
		return
	}

	d.log.Warn("sybil detector: suspicious clusters found", slog.Int("count", len(clusters)))

	for _, cluster := range clusters {
		d.log.Warn("sybil detector: flagging cluster",
			slog.Int("community_id", cluster.CommunityID),
			slog.Int("size", cluster.Size),
			slog.Int("external_connections", cluster.ExternalConnections),
		)

		for _, uid := range cluster.SuspectUIDs {
			// Issue 9: Sybil detection flags users and we now have a review workflow.
			if err := d.userRepo.UpdateTrustStatus(ctx, uid, domain.TrustStatusUnderReview); err != nil {
				d.log.Error("sybil detector: failed to flag user",
					slog.String("user_id", uid.String()),
					slog.String("error", err.Error()),
				)
			} else {
				d.log.Info("sybil detector: flagged user for review", slog.String("user_id", uid.String()))
			}
		}
	}
}

package service

import (
	"context"
	"log/slog"

	"github.com/trueconnect/backend/internal/pkg/halalfilter"
	"github.com/trueconnect/backend/internal/repository"
)

// ModerationService decides whether user text should be blocked or flagged.
//
// It prefers the ML provider (multilingual KZ/RU/EN classifier) and falls back
// to the keyword filter (internal/pkg/halalfilter) whenever the provider is not
// configured, errors, or times out. This keeps moderation working even if the
// ML microservice is down, and lets the app run with no ML service at all.
type ModerationService struct {
	provider repository.ModerationProvider // may be nil → always use keyword filter
	log      *slog.Logger
}

// NewModerationService creates the service. Pass provider=nil to run with the
// keyword filter only.
func NewModerationService(provider repository.ModerationProvider, log *slog.Logger) *ModerationService {
	return &ModerationService{provider: provider, log: log}
}

// Check returns (block, warn). It never returns an error: any provider failure
// degrades gracefully to the keyword filter so callers always get a verdict.
func (s *ModerationService) Check(ctx context.Context, text string) (block bool, warn bool) {
	if s.provider != nil {
		res, err := s.provider.Check(ctx, text)
		if err == nil && res != nil {
			return res.Block, res.Warn
		}
		if err != nil && s.log != nil {
			s.log.Warn("moderation: ML provider failed, falling back to keyword filter",
				slog.String("error", err.Error()))
		}
	}
	// Fallback: keyword filter.
	return halalfilter.CheckMessage(text)
}

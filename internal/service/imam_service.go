package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// ImamService handles imam catalog and nikah confirmation.
type ImamService struct {
	// Stub implementation for Sprint 11; real implementation deferred.
	// Will include imam catalog from embedded JSON and nikah confirmation logic.
}

// NewImamService creates a new imam service.
func NewImamService() *ImamService {
	return &ImamService{}
}

// ListImams returns imams in a given city (stub for Sprint 11).
func (s *ImamService) ListImams(ctx context.Context, city string) ([]domain.Imam, error) {
	// Stub: returns empty list. Full implementation in Sprint 11.
	return []domain.Imam{}, nil
}

// ConfirmNikah confirms a nikah via an imam (stub for Sprint 11).
func (s *ImamService) ConfirmNikah(ctx context.Context, matchID, imamID, callerID uuid.UUID) error {
	// Stub: returns success. Full implementation in Sprint 11.
	return nil
}

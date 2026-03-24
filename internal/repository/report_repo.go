package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// ReportRepository handles persistence of user reports.
type ReportRepository interface {
	// Create stores a new report.
	Create(ctx context.Context, report *domain.Report) error
	// ExistsBetween checks if a report from reporter against reported already exists (pending or reviewed).
	ExistsBetween(ctx context.Context, reporterID, reportedID uuid.UUID) (bool, error)
}

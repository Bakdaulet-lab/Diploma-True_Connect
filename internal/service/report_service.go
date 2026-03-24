package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// ReportService handles user report business logic.
type ReportService struct {
	reportRepo repository.ReportRepository
	graphRepo  repository.TrustGraphRepository
}

// NewReportService creates a new report service.
func NewReportService(reportRepo repository.ReportRepository, graphRepo repository.TrustGraphRepository) *ReportService {
	return &ReportService{
		reportRepo: reportRepo,
		graphRepo:  graphRepo,
	}
}

// CreateReportInput holds the data required to create a report.
type CreateReportInput struct {
	ReporterID  uuid.UUID
	ReportedID  uuid.UUID
	Reason      string
	Description string
}

// CreateReport validates and creates a new user report.
func (s *ReportService) CreateReport(ctx context.Context, input CreateReportInput) (*domain.Report, error) {
	// No self-reports.
	if input.ReporterID == input.ReportedID {
		return nil, fmt.Errorf("create report: cannot report yourself: %w", domain.ErrInvalidInput)
	}

	// Check for duplicate.
	exists, err := s.reportRepo.ExistsBetween(ctx, input.ReporterID, input.ReportedID)
	if err != nil {
		return nil, fmt.Errorf("create report: checking duplicate: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("create report: already reported this user: %w", domain.ErrAlreadyExists)
	}

	report := &domain.Report{
		ID:          uuid.New(),
		ReporterID:  input.ReporterID,
		ReportedID:  input.ReportedID,
		Reason:      input.Reason,
		Description: input.Description,
		Status:      domain.ReportStatusPending,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.reportRepo.Create(ctx, report); err != nil {
		return nil, fmt.Errorf("create report: %w", err)
	}

	// Also add edge to Neo4j trust graph for reputation impact.
	if err := s.graphRepo.AddReport(ctx, input.ReporterID, input.ReportedID, input.Reason); err != nil {
		// Log but don't fail — PG is the source of truth.
		// The trust engine will eventually reconcile.
	}

	return report, nil
}

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// Compile-time check: ReportRepo implements repository.ReportRepository.
var _ repository.ReportRepository = (*ReportRepo)(nil)

// ReportRepo is the PostgreSQL implementation of ReportRepository.
type ReportRepo struct {
	pool *pgxpool.Pool
}

// NewReportRepo creates a new report repository.
func NewReportRepo(pool *pgxpool.Pool) *ReportRepo {
	return &ReportRepo{pool: pool}
}

// Create stores a new report in the database.
func (r *ReportRepo) Create(ctx context.Context, report *domain.Report) error {
	const query = `
		INSERT INTO social.reports (id, reporter_id, reported_id, reason, description, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := runner(ctx, r.pool).Exec(ctx, query,
		report.ID,
		report.ReporterID,
		report.ReportedID,
		report.Reason,
		report.Description,
		report.Status,
		report.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create report: %w", err)
	}
	return nil
}

// ExistsBetween checks if a report from reporter against reported already exists.
func (r *ReportRepo) ExistsBetween(ctx context.Context, reporterID, reportedID uuid.UUID) (bool, error) {
	const query = `
		SELECT 1 FROM social.reports
		WHERE reporter_id = $1 AND reported_id = $2
		LIMIT 1`

	var dummy int
	err := runner(ctx, r.pool).QueryRow(ctx, query, reporterID, reportedID).Scan(&dummy)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("exists between: %w", err)
	}
	return true, nil
}

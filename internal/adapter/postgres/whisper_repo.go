package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trueconnect/backend/internal/repository"
)

// WhisperRepo implements repository.WhisperRepository using PostgreSQL.
type WhisperRepo struct {
	pool *pgxpool.Pool
}

var _ repository.WhisperRepository = (*WhisperRepo)(nil)

// NewWhisperRepo creates a new PostgreSQL-backed whisper repository.
func NewWhisperRepo(pool *pgxpool.Pool) *WhisperRepo {
	return &WhisperRepo{pool: pool}
}

func (r *WhisperRepo) Create(ctx context.Context, reporterID, reportedID, matchID uuid.UUID, feedbackEncrypted []byte, strikeWeight int) error {
	var meetingMatchID *uuid.UUID
	if matchID != uuid.Nil {
		meetingMatchID = &matchID
	}

	query := `
		INSERT INTO social.whisper_reports
			(reporter_id, reported_id, meeting_match_id, feedback_encrypted, strike_weight)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := runner(ctx, r.pool).Exec(ctx, query,
		reporterID, reportedID, meetingMatchID, feedbackEncrypted, strikeWeight,
	)
	if err != nil {
		return fmt.Errorf("creating whisper report: %w", err)
	}
	return nil
}

func (r *WhisperRepo) GetByReportedUser(ctx context.Context, reportedID uuid.UUID) ([]repository.WhisperReportRow, error) {
	query := `
		SELECT id, reporter_id, reported_id, COALESCE(meeting_match_id, $2), strike_weight, admin_flagged, created_at
		FROM social.whisper_reports
		WHERE reported_id = $1 AND strike_counted = false
		ORDER BY created_at DESC`

	rows, err := runner(ctx, r.pool).Query(ctx, query, reportedID, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("getting whisper reports: %w", err)
	}
	defer rows.Close()

	var result []repository.WhisperReportRow
	for rows.Next() {
		var row repository.WhisperReportRow
		if err := rows.Scan(
			&row.ID, &row.ReporterID, &row.ReportedID, &row.MatchID,
			&row.StrikeWeight, &row.AdminFlagged, &row.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning whisper report: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r *WhisperRepo) GetFlaggedReports(ctx context.Context) ([]repository.WhisperReportRow, error) {
	query := `
		SELECT id, reporter_id, reported_id, COALESCE(meeting_match_id, $1), strike_weight, admin_flagged, created_at
		FROM social.whisper_reports
		WHERE admin_flagged = true
		ORDER BY created_at DESC`

	rows, err := runner(ctx, r.pool).Query(ctx, query, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("getting flagged whisper reports: %w", err)
	}
	defer rows.Close()

	var result []repository.WhisperReportRow
	for rows.Next() {
		var row repository.WhisperReportRow
		if err := rows.Scan(
			&row.ID, &row.ReporterID, &row.ReportedID, &row.MatchID,
			&row.StrikeWeight, &row.AdminFlagged, &row.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning flagged whisper report: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r *WhisperRepo) CountWeightedStrikes(ctx context.Context, reportedID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(SUM(strike_weight), 0)
		FROM social.whisper_reports
		WHERE reported_id = $1 AND strike_counted = false`

	var total int
	err := runner(ctx, r.pool).QueryRow(ctx, query, reportedID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("counting weighted strikes: %w", err)
	}
	return total, nil
}

func (r *WhisperRepo) FlagForAdmin(ctx context.Context, reportedID uuid.UUID) error {
	query := `
		UPDATE social.whisper_reports
		SET admin_flagged = true, strike_counted = true
		WHERE reported_id = $1 AND strike_counted = false`

	_, err := runner(ctx, r.pool).Exec(ctx, query, reportedID)
	if err != nil {
		return fmt.Errorf("flagging for admin: %w", err)
	}
	return nil
}

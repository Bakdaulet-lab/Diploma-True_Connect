package repository

import (
	"context"

	"github.com/google/uuid"
)

// WhisperRepository manages anonymous whisper feedback reports.
type WhisperRepository interface {
	// Create inserts an encrypted whisper report.
	// feedbackEncrypted must be AES-256-GCM ciphertext. strikeWeight is 1 normally, 2 if PoI confirmed.
	Create(ctx context.Context, reporterID, reportedID, matchID uuid.UUID, feedbackEncrypted []byte, strikeWeight int) error

	// GetByReportedUser returns all unflagged reports for a given reported user.
	GetByReportedUser(ctx context.Context, reportedID uuid.UUID) ([]WhisperReportRow, error)

	// CountWeightedStrikes sums strike_weight for uncounted reports against a user.
	CountWeightedStrikes(ctx context.Context, reportedID uuid.UUID) (int, error)

	// FlagForAdmin marks all uncounted reports for reportedID as admin_flagged.
	FlagForAdmin(ctx context.Context, reportedID uuid.UUID) error
}

// WhisperReportRow is the minimal row returned for admin review.
type WhisperReportRow struct {
	ReporterID  uuid.UUID
	ReportedID  uuid.UUID
	MatchID     uuid.UUID
	StrikeWeight int
	AdminFlagged bool
}

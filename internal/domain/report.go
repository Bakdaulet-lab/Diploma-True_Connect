package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReportStatus represents the moderation state of a user report.
type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "pending"
	ReportStatusReviewed  ReportStatus = "reviewed"
	ReportStatusDismissed ReportStatus = "dismissed"
)

// Report represents a user-submitted report against another user.
type Report struct {
	ID          uuid.UUID
	ReporterID  uuid.UUID
	ReportedID  uuid.UUID
	Reason      string
	Description string
	Status      ReportStatus
	CreatedAt   time.Time
	ReviewedAt  *time.Time
}

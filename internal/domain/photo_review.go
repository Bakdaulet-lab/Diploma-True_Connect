package domain

import (
	"time"

	"github.com/google/uuid"
)

// Photo review statuses.
const (
	PhotoReviewPending  = "pending"
	PhotoReviewApproved = "approved"
	PhotoReviewRejected = "rejected"
)

// PhotoReview is a profile photo held for manual admin review after the CV
// verifier rejected it (no face / NSFW).
type PhotoReview struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	ObjectKey  string
	Reasons    []string
	NSFWScore  float64
	FaceCount  int
	Status     string
	CreatedAt  time.Time
	ReviewedAt *time.Time
	ReviewedBy *uuid.UUID
}

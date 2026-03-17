package domain

import "github.com/google/uuid"

// TrustScore represents a user's computed reputation score (0-100).
type TrustScore struct {
	UserID      uuid.UUID
	Score       int
	RatingCount int
}

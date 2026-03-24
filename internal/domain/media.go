package domain

import (
	"time"

	"github.com/google/uuid"
)

type Media struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	ObjectKey  string
	MediaType  string
	SortOrder  int
	IsVerified bool
	CreatedAt  time.Time
}

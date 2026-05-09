package domain

import (
	"time"

	"github.com/google/uuid"
)

type Interaction struct {
	ID         uuid.UUID
	RaterID    uuid.UUID
	RatedID    uuid.UUID
	Rating     int // 1-5
	Context    string
	Comment    string
	IsVerified bool
	CreatedAt  time.Time
}

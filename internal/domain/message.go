package domain

import (
	"time"

	"github.com/google/uuid"
)

type Match struct {
	ID                     uuid.UUID
	UserAID                uuid.UUID
	UserBID                uuid.UUID
	UserALiked             bool
	UserBLiked             bool
	MatchedAt              *time.Time
	NiyyahTimerEndsAt      *time.Time
	FamilyIntroDone        bool
	ImamConfirmed          bool
	CreatedAt              time.Time
}

type Message struct {
	ID               uuid.UUID
	MatchID          uuid.UUID
	SenderID         uuid.UUID
	ContentEncrypted []byte
	IsToxic          bool
	ReadAt           *time.Time
	CreatedAt        time.Time
}

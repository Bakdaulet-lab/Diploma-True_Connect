package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	MahramStatusPending  = "pending"
	MahramStatusVerified = "verified"
	MahramStatusRejected = "rejected"
)

type Mahram struct {
	ID                 uuid.UUID
	WomanUserID        uuid.UUID
	MahramPhoneHash    []byte
	TelegramChatID     *int64
	VerificationStatus string
	VerifiedAt         *time.Time
	CreatedAt          time.Time
}

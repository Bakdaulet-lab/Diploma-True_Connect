package domain

import (
	"time"

	"github.com/google/uuid"
)

type VerificationLevel string

const (
	VerificationNone          VerificationLevel = "none"
	VerificationPhoneVerified VerificationLevel = "phone_verified"
	VerificationIDVerified    VerificationLevel = "id_verified"
	VerificationPhotoVerified VerificationLevel = "photo_verified"
)

type TrustStatus string

const (
	TrustStatusNormal      TrustStatus = "normal"
	TrustStatusUnderReview TrustStatus = "under_review"
	TrustStatusSuspended   TrustStatus = "suspended"
	TrustStatusBanned      TrustStatus = "banned"
)

type User struct {
	ID                uuid.UUID
	PhoneHash         []byte
	PhoneEncrypted    []byte
	EmailEncrypted    []byte
	PasswordHash      string
	PublicKey         *string // X25519 Public Key for E2E encryption
	VerificationLevel VerificationLevel
	TrustStatus       TrustStatus
	TrustScore        int
	IsAdmin           bool
	IsActive          bool
	LastLoginAt       *time.Time
	FCMToken          *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

// MahramChatRoom is a 3-way chat room between a matched couple and the woman's mahram.
type MahramChatRoom struct {
	ID           uuid.UUID
	MatchID      uuid.UUID
	UserAID      uuid.UUID // populated via JOIN on social.matches
	UserBID      uuid.UUID // populated via JOIN on social.matches
	MahramUserID uuid.UUID
	CreatedAt    time.Time
}

// MahramMessage is an encrypted message in a mahram chat room.
type MahramMessage struct {
	ID               uuid.UUID
	RoomID           uuid.UUID
	SenderID         uuid.UUID
	ContentEncrypted []byte
	CreatedAt        time.Time
}

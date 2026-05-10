package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// MahramChatRepository manages 3-way mahram chat rooms and their messages.
type MahramChatRepository interface {
	// CreateRoom creates or upserts a mahram chat room for the given match.
	CreateRoom(ctx context.Context, matchID, mahramUserID uuid.UUID) (*domain.MahramChatRoom, error)

	// GetRoomByMatchID returns the mahram chat room associated with a match.
	GetRoomByMatchID(ctx context.Context, matchID uuid.UUID) (*domain.MahramChatRoom, error)

	// GetRoomByID returns the mahram chat room by its own ID.
	GetRoomByID(ctx context.Context, roomID uuid.UUID) (*domain.MahramChatRoom, error)

	// SaveMessage persists an already-encrypted mahram chat message.
	SaveMessage(ctx context.Context, roomID, senderID uuid.UUID, contentEncrypted []byte) (*domain.MahramMessage, error)

	// GetMessages returns paginated messages for a room, newest first.
	// cursor is the ID of the last-seen message ("" starts from the newest).
	GetMessages(ctx context.Context, roomID uuid.UUID, cursor string, limit int) ([]*domain.MahramMessage, string, error)
}

package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/pkg/crypto"
	"github.com/trueconnect/backend/internal/pkg/sanitize"
	"github.com/trueconnect/backend/internal/repository"
)

// DecryptedMessage is a message with plaintext content for API responses.
type DecryptedMessage struct {
	ID        uuid.UUID `json:"id"`
	MatchID   uuid.UUID `json:"match_id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Content   string    `json:"content"`
	IsToxic   bool      `json:"is_toxic"`
	ReadAt    *string   `json:"read_at,omitempty"`
	CreatedAt string    `json:"created_at"`
}

// ChatService handles chat message business logic.
type ChatService struct {
	messageRepo   repository.MessageRepository
	matchRepo     repository.MatchRepository
	encryptionKey []byte
	pushCh        chan<- domain.PushEvent
}

// NewChatService creates a new chat service.
func NewChatService(
	messageRepo repository.MessageRepository,
	matchRepo repository.MatchRepository,
	encryptionKey []byte,
	pushCh chan<- domain.PushEvent,
) *ChatService {
	return &ChatService{
		messageRepo:   messageRepo,
		matchRepo:     matchRepo,
		encryptionKey: encryptionKey,
		pushCh:        pushCh,
	}
}

// SendMessage encrypts and stores a chat message.
func (s *ChatService) SendMessage(ctx context.Context, senderID, matchID uuid.UUID, plaintext string, isToxic bool) (*DecryptedMessage, error) {
	plaintext = sanitize.StripHTML(plaintext)
	if len(plaintext) == 0 {
		return nil, fmt.Errorf("send message: content cannot be empty: %w", domain.ErrInvalidInput)
	}

	// Verify the sender is part of the match.
	match, err := s.matchRepo.GetMatch(ctx, matchID, senderID)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	if match.MatchedAt == nil {
		return nil, fmt.Errorf("send message: match not finalized: %w", domain.ErrForbidden)
	}

	encrypted, err := crypto.Encrypt([]byte(plaintext), s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("send message: encrypting: %w", err)
	}

	msg := &domain.Message{
		MatchID:          matchID,
		SenderID:         senderID,
		ContentEncrypted: encrypted,
		IsToxic:          isToxic,
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("send message: storing: %w", err)
	}

	// Send Push Notification
	recipientID := RecipientID(match, senderID)
	if s.pushCh != nil {
		select {
		case s.pushCh <- domain.PushEvent{
			UserID: recipientID,
			Title:  "New Message",
			Body:   "You have a new message from a match",
			Data:   map[string]string{"type": "chat", "match_id": matchID.String()},
		}:
		default:
			// channel full, skip
		}
	}

	return &DecryptedMessage{
		ID:        msg.ID,
		MatchID:   msg.MatchID,
		SenderID:  msg.SenderID,
		Content:   plaintext,
		IsToxic:   isToxic,
		CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// GetMessages returns decrypted messages for a match with cursor-based pagination.
func (s *ChatService) GetMessages(ctx context.Context, matchID, userID uuid.UUID, cursor string, limit int) ([]DecryptedMessage, string, error) {
	// Verify the user is part of the match.
	if _, err := s.matchRepo.GetMatch(ctx, matchID, userID); err != nil {
		return nil, "", fmt.Errorf("get messages: %w", err)
	}

	messages, nextCursor, err := s.messageRepo.ListByMatch(ctx, matchID, cursor, limit)
	if err != nil {
		return nil, "", fmt.Errorf("get messages: listing: %w", err)
	}

	result := make([]DecryptedMessage, 0, len(messages))
	for _, m := range messages {
		plaintext, err := crypto.Decrypt(m.ContentEncrypted, s.encryptionKey)
		if err != nil {
			return nil, "", fmt.Errorf("get messages: decrypting: %w", err)
		}

		dm := DecryptedMessage{
			ID:        m.ID,
			MatchID:   m.MatchID,
			SenderID:  m.SenderID,
			Content:   string(plaintext),
			IsToxic:   m.IsToxic,
			CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if m.ReadAt != nil {
			t := m.ReadAt.Format("2006-01-02T15:04:05Z07:00")
			dm.ReadAt = &t
		}
		result = append(result, dm)
	}

	return result, nextCursor, nil
}

// MarkRead marks all unread messages in a match as read for the given user.
func (s *ChatService) MarkRead(ctx context.Context, matchID, readerID uuid.UUID) error {
	if _, err := s.matchRepo.GetMatch(ctx, matchID, readerID); err != nil {
		return fmt.Errorf("mark read: %w", err)
	}

	if err := s.messageRepo.MarkRead(ctx, matchID, readerID); err != nil {
		return fmt.Errorf("mark read: %w", err)
	}

	return nil
}

// RecipientID returns the other user in a match.
func RecipientID(match *domain.Match, senderID uuid.UUID) uuid.UUID {
	if match.UserAID == senderID {
		return match.UserBID
	}
	return match.UserAID
}

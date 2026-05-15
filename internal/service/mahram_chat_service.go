package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/pkg/crypto"
	"github.com/trueconnect/backend/internal/repository"
)

// MahramChatRoomView is the API-visible representation of a mahram chat room.
type MahramChatRoomView struct {
	ID           uuid.UUID `json:"id"`
	MatchID      uuid.UUID `json:"match_id"`
	MahramUserID uuid.UUID `json:"mahram_user_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// MahramMessageView is a decrypted mahram chat message for API responses.
type MahramMessageView struct {
	ID        uuid.UUID `json:"id"`
	RoomID    uuid.UUID `json:"room_id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// MahramChatService manages 3-way mahram chat rooms.
type MahramChatService struct {
	repo          repository.MahramChatRepository
	matchRepo     repository.MatchRepository
	encryptionKey []byte
	rdb           *redis.Client
	log           *slog.Logger
}

// NewMahramChatService creates a new mahram chat service.
func NewMahramChatService(
	repo repository.MahramChatRepository,
	matchRepo repository.MatchRepository,
	encryptionKey []byte,
	rdb *redis.Client,
	log *slog.Logger,
) *MahramChatService {
	return &MahramChatService{
		repo:          repo,
		matchRepo:     matchRepo,
		encryptionKey: encryptionKey,
		rdb:           rdb,
		log:           log,
	}
}

// CreateRoom creates a mahram chat room for the given match.
// callerID must be one of the match participants (user_a or user_b).
func (s *MahramChatService) CreateRoom(ctx context.Context, matchID, callerID, mahramUserID uuid.UUID) (*MahramChatRoomView, error) {
	match, err := s.matchRepo.GetMatch(ctx, matchID, callerID)
	if err != nil {
		return nil, fmt.Errorf("create mahram room: %w", err)
	}
	if match.MatchedAt == nil {
		return nil, fmt.Errorf("create mahram room: match not finalized: %w", domain.ErrForbidden)
	}

	// Mahram must be a third party — not one of the two match participants.
	if mahramUserID == match.UserAID || mahramUserID == match.UserBID {
		return nil, fmt.Errorf("create mahram room: mahram must be a third party, not a match participant: %w", domain.ErrInvalidInput)
	}

	room, err := s.repo.CreateRoom(ctx, matchID, mahramUserID)
	if err != nil {
		return nil, fmt.Errorf("create mahram room: %w", err)
	}

	// Notify all 3 participants of the new room via WS pub/sub.
	s.publishToRoom(ctx, room, "mahram_room_invite", map[string]string{
		"room_id":  room.ID.String(),
		"match_id": matchID.String(),
	})

	return toMahramRoomView(room), nil
}

// SendMessage encrypts and persists a mahram message, then relays it to all room participants.
func (s *MahramChatService) SendMessage(ctx context.Context, roomID, senderID uuid.UUID, plaintext string) (*MahramMessageView, error) {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("mahram send: %w", err)
	}
	if !isRoomParticipant(room, senderID) {
		return nil, fmt.Errorf("mahram send: not a participant: %w", domain.ErrForbidden)
	}
	if len(plaintext) == 0 {
		return nil, fmt.Errorf("mahram send: content cannot be empty: %w", domain.ErrInvalidInput)
	}

	encrypted, err := crypto.Encrypt([]byte(plaintext), s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("mahram send: encrypting: %w", err)
	}

	msg, err := s.repo.SaveMessage(ctx, roomID, senderID, encrypted)
	if err != nil {
		return nil, fmt.Errorf("mahram send: saving: %w", err)
	}

	view := &MahramMessageView{
		ID:        msg.ID,
		RoomID:    msg.RoomID,
		SenderID:  msg.SenderID,
		Content:   plaintext,
		CreatedAt: msg.CreatedAt,
	}
	s.publishToRoom(ctx, room, "mahram_chat_msg", view)
	return view, nil
}

// GetMessages returns decrypted paginated messages for a room, newest first.
func (s *MahramChatService) GetMessages(ctx context.Context, roomID, callerID uuid.UUID, cursor string, limit int) ([]*MahramMessageView, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, "", fmt.Errorf("mahram get messages: %w", err)
	}
	if !isRoomParticipant(room, callerID) {
		return nil, "", fmt.Errorf("mahram get messages: not a participant: %w", domain.ErrForbidden)
	}

	msgs, nextCursor, err := s.repo.GetMessages(ctx, roomID, cursor, limit)
	if err != nil {
		return nil, "", fmt.Errorf("mahram get messages: %w", err)
	}

	views := make([]*MahramMessageView, 0, len(msgs))
	for _, msg := range msgs {
		plaintext, err := crypto.Decrypt(msg.ContentEncrypted, s.encryptionKey)
		if err != nil {
			s.log.Warn("failed to decrypt mahram message", slog.String("msg_id", msg.ID.String()))
			continue
		}
		views = append(views, &MahramMessageView{
			ID:        msg.ID,
			RoomID:    msg.RoomID,
			SenderID:  msg.SenderID,
			Content:   string(plaintext),
			CreatedAt: msg.CreatedAt,
		})
	}
	return views, nextCursor, nil
}

// GetRoom returns a mahram chat room domain object (used by the WS handler to verify membership).
func (s *MahramChatService) GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.MahramChatRoom, error) {
	return s.repo.GetRoomByID(ctx, roomID)
}

// publishToRoom broadcasts a WS-compatible JSON message to all 3 room participants via Redis.
func (s *MahramChatService) publishToRoom(ctx context.Context, room *domain.MahramChatRoom, msgType string, payload interface{}) {
	if s.rdb == nil {
		return
	}
	b, err := json.Marshal(map[string]interface{}{"type": msgType, "payload": payload})
	if err != nil {
		return
	}
	for _, uid := range []uuid.UUID{room.UserAID, room.UserBID, room.MahramUserID} {
		if uid == uuid.Nil {
			continue
		}
		s.rdb.Publish(ctx, "ws:user:"+uid.String(), b)
	}
}

func isRoomParticipant(room *domain.MahramChatRoom, userID uuid.UUID) bool {
	return userID == room.UserAID || userID == room.UserBID || userID == room.MahramUserID
}

func toMahramRoomView(r *domain.MahramChatRoom) *MahramChatRoomView {
	return &MahramChatRoomView{
		ID:           r.ID,
		MatchID:      r.MatchID,
		MahramUserID: r.MahramUserID,
		CreatedAt:    r.CreatedAt,
	}
}

package service_test

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// ── mockMahramChatRepo ────────────────────────────────────────────────────────

type mockMahramChatRepo struct {
	mu       sync.Mutex
	rooms    map[uuid.UUID]*domain.MahramChatRoom  // keyed by room ID
	byMatch  map[uuid.UUID]uuid.UUID               // match ID → room ID
	messages map[uuid.UUID][]*domain.MahramMessage // keyed by room ID
}

func newMockMahramChatRepo() *mockMahramChatRepo {
	return &mockMahramChatRepo{
		rooms:    make(map[uuid.UUID]*domain.MahramChatRoom),
		byMatch:  make(map[uuid.UUID]uuid.UUID),
		messages: make(map[uuid.UUID][]*domain.MahramMessage),
	}
}

func (m *mockMahramChatRepo) CreateRoom(_ context.Context, matchID, mahramUserID uuid.UUID) (*domain.MahramChatRoom, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Upsert: if room exists for match, update mahram.
	if roomID, ok := m.byMatch[matchID]; ok {
		rm := m.rooms[roomID]
		rm.MahramUserID = mahramUserID
		cp := *rm
		return &cp, nil
	}

	rm := &domain.MahramChatRoom{
		ID:           uuid.New(),
		MatchID:      matchID,
		MahramUserID: mahramUserID,
		// UserAID / UserBID left zero — tests set them via seedRoom if needed.
		CreatedAt: time.Now(),
	}
	m.rooms[rm.ID] = rm
	m.byMatch[matchID] = rm.ID
	cp := *rm
	return &cp, nil
}

func (m *mockMahramChatRepo) GetRoomByMatchID(_ context.Context, matchID uuid.UUID) (*domain.MahramChatRoom, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	roomID, ok := m.byMatch[matchID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *m.rooms[roomID]
	return &cp, nil
}

func (m *mockMahramChatRepo) GetRoomByID(_ context.Context, roomID uuid.UUID) (*domain.MahramChatRoom, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rm, ok := m.rooms[roomID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *rm
	return &cp, nil
}

func (m *mockMahramChatRepo) SaveMessage(_ context.Context, roomID, senderID uuid.UUID, contentEncrypted []byte) (*domain.MahramMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	msg := &domain.MahramMessage{
		ID:               uuid.New(),
		RoomID:           roomID,
		SenderID:         senderID,
		ContentEncrypted: contentEncrypted,
		CreatedAt:        time.Now(),
	}
	m.messages[roomID] = append(m.messages[roomID], msg)
	cp := *msg
	return &cp, nil
}

func (m *mockMahramChatRepo) GetMessages(_ context.Context, roomID uuid.UUID, _ string, limit int) ([]*domain.MahramMessage, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	msgs := m.messages[roomID]
	if len(msgs) > limit {
		msgs = msgs[len(msgs)-limit:]
	}
	result := make([]*domain.MahramMessage, len(msgs))
	for i, msg := range msgs {
		cp := *msg
		result[i] = &cp
	}
	return result, "", nil
}

// seedRoom populates a room with known participants for tests.
func (m *mockMahramChatRepo) seedRoom(room *domain.MahramChatRoom) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rooms[room.ID] = room
	m.byMatch[room.MatchID] = room.ID
}

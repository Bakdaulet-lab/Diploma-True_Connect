package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/service"
)

// newTestMahramChatService wires up MahramChatService with in-memory mocks.
// It also pre-seeds a mutual match between userA and userB so room-creation tests can run.
func newTestMahramChatService() (
	*service.MahramChatService,
	*mockMahramChatRepo,
	*mockMatchRepo,
	uuid.UUID, // userA
	uuid.UUID, // userB
	uuid.UUID, // matchID
) {
	chatRepo := newMockMahramChatRepo()
	matchRepo := newMockMatchRepo()

	svc := service.NewMahramChatService(chatRepo, matchRepo, make([]byte, 32), nil, nil)

	userA := uuid.New()
	userB := uuid.New()

	// Seed a mutual match between userA and userB.
	matchRepo.mu.Lock()
	matchID := uuid.New()
	now := time.Now()
	matchRepo.matches[matchID] = &domain.Match{
		ID:        matchID,
		UserAID:   userA,
		UserBID:   userB,
		MatchedAt: &now,
	}
	matchRepo.mu.Unlock()

	return svc, chatRepo, matchRepo, userA, userB, matchID
}

// ── CreateRoom ────────────────────────────────────────────────────────────────

func TestMahramChat_CreateRoom_Success(t *testing.T) {
	t.Parallel()

	svc, chatRepo, _, userA, _, matchID := newTestMahramChatService()
	mahramID := uuid.New()

	room, err := svc.CreateRoom(context.Background(), matchID, userA, mahramID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if room.MatchID != matchID {
		t.Errorf("expected match_id=%s, got %s", matchID, room.MatchID)
	}
	if room.MahramUserID != mahramID {
		t.Errorf("expected mahram_user_id=%s, got %s", mahramID, room.MahramUserID)
	}

	// Room must be retrievable by match ID.
	saved, err := chatRepo.GetRoomByMatchID(context.Background(), matchID)
	if err != nil {
		t.Fatalf("room not saved: %v", err)
	}
	if saved.ID != room.ID {
		t.Error("saved room ID mismatch")
	}
}

func TestMahramChat_CreateRoom_MatchNotFound(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _, _ := newTestMahramChatService()
	unknown := uuid.New()
	caller := uuid.New()

	_, err := svc.CreateRoom(context.Background(), unknown, caller, uuid.New())
	if err == nil {
		t.Fatal("expected error for unknown match")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMahramChat_CreateRoom_CallerNotInMatch(t *testing.T) {
	t.Parallel()

	svc, _, _, _, _, matchID := newTestMahramChatService()
	outsider := uuid.New()

	_, err := svc.CreateRoom(context.Background(), matchID, outsider, uuid.New())
	if err == nil {
		t.Fatal("expected error for non-participant caller")
	}
	// matchRepo.GetMatch returns ErrForbidden when the match exists but user is not a participant.
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// ── SendMessage ───────────────────────────────────────────────────────────────

func TestMahramChat_SendMessage_Success(t *testing.T) {
	t.Parallel()

	svc, chatRepo, _, userA, userB, matchID := newTestMahramChatService()
	mahramID := uuid.New()

	// Create room first.
	roomView, err := svc.CreateRoom(context.Background(), matchID, userA, mahramID)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	// Seed participants into mock repo so IsParticipant check works.
	chatRepo.seedRoom(&domain.MahramChatRoom{
		ID:           roomView.ID,
		MatchID:      matchID,
		UserAID:      userA,
		UserBID:      userB,
		MahramUserID: mahramID,
	})

	view, err := svc.SendMessage(context.Background(), roomView.ID, userA, "Ассалам алейкум")
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if view.Content != "Ассалам алейкум" {
		t.Errorf("expected decrypted content, got %q", view.Content)
	}
	if view.SenderID != userA {
		t.Errorf("expected sender=%s, got %s", userA, view.SenderID)
	}
}

func TestMahramChat_SendMessage_NotParticipant(t *testing.T) {
	t.Parallel()

	svc, chatRepo, _, userA, userB, matchID := newTestMahramChatService()
	mahramID := uuid.New()
	outsider := uuid.New()

	roomView, err := svc.CreateRoom(context.Background(), matchID, userA, mahramID)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	chatRepo.seedRoom(&domain.MahramChatRoom{
		ID:           roomView.ID,
		MatchID:      matchID,
		UserAID:      userA,
		UserBID:      userB,
		MahramUserID: mahramID,
	})

	_, err = svc.SendMessage(context.Background(), roomView.ID, outsider, "Hello")
	if err == nil {
		t.Fatal("expected error for non-participant sender")
	}
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestMahramChat_SendMessage_EmptyContent(t *testing.T) {
	t.Parallel()

	svc, chatRepo, _, userA, userB, matchID := newTestMahramChatService()
	mahramID := uuid.New()

	roomView, err := svc.CreateRoom(context.Background(), matchID, userA, mahramID)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	chatRepo.seedRoom(&domain.MahramChatRoom{
		ID:           roomView.ID,
		MatchID:      matchID,
		UserAID:      userA,
		UserBID:      userB,
		MahramUserID: mahramID,
	})

	_, err = svc.SendMessage(context.Background(), roomView.ID, userA, "")
	if err == nil {
		t.Fatal("expected error for empty content")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

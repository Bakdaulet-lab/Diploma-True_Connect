package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/service"
)

// helper to set up a match for chat tests.
func setupChatMatch(t *testing.T) (matchRepo *mockMatchRepo, userA, userB uuid.UUID, matchID uuid.UUID) {
	t.Helper()
	matchRepo = newMockMatchRepo()
	userA = uuid.New()
	userB = uuid.New()
	now := time.Now()
	matchID = uuid.New()
	matchRepo.mu.Lock()
	matchRepo.matches[matchID] = &domain.Match{
		ID:        matchID,
		UserAID:   userA,
		UserBID:   userB,
		MatchedAt: &now,
	}
	matchRepo.mu.Unlock()
	return
}

// testEncKey is a valid 32-byte AES-256 key for tests.
var testEncKey = []byte("0123456789abcdef0123456789abcdef")

func TestSendMessage_Success(t *testing.T) {
	matchRepo, userA, _, matchID := setupChatMatch(t)
	msgRepo := newMockMessageRepo()
	svc := service.NewChatService(msgRepo, matchRepo, testEncKey)

	dm, err := svc.SendMessage(context.Background(), userA, matchID, "hello!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dm.Content != "hello!" {
		t.Errorf("expected content %q, got %q", "hello!", dm.Content)
	}
	if dm.SenderID != userA {
		t.Errorf("expected sender %s, got %s", userA, dm.SenderID)
	}
	if dm.MatchID != matchID {
		t.Errorf("expected match %s, got %s", matchID, dm.MatchID)
	}
}

func TestSendMessage_EmptyContent(t *testing.T) {
	matchRepo, userA, _, matchID := setupChatMatch(t)
	msgRepo := newMockMessageRepo()
	svc := service.NewChatService(msgRepo, matchRepo, testEncKey)

	_, err := svc.SendMessage(context.Background(), userA, matchID, "")
	if err == nil {
		t.Fatal("expected error for empty message")
	}
}

func TestSendMessage_NotInMatch(t *testing.T) {
	matchRepo, _, _, matchID := setupChatMatch(t)
	msgRepo := newMockMessageRepo()
	svc := service.NewChatService(msgRepo, matchRepo, testEncKey)

	outsider := uuid.New()
	_, err := svc.SendMessage(context.Background(), outsider, matchID, "sneaky")
	if err == nil {
		t.Fatal("expected error for sender not in match")
	}
}

func TestSendMessage_MatchNotFinalized(t *testing.T) {
	matchRepo := newMockMatchRepo()
	userA := uuid.New()
	userB := uuid.New()
	matchID := uuid.New()
	matchRepo.mu.Lock()
	matchRepo.matches[matchID] = &domain.Match{
		ID:        matchID,
		UserAID:   userA,
		UserBID:   userB,
		MatchedAt: nil, // not finalized
	}
	matchRepo.mu.Unlock()

	msgRepo := newMockMessageRepo()
	svc := service.NewChatService(msgRepo, matchRepo, testEncKey)

	_, err := svc.SendMessage(context.Background(), userA, matchID, "hi")
	if err == nil {
		t.Fatal("expected error for non-finalized match")
	}
}

func TestGetMessages_Success(t *testing.T) {
	matchRepo, userA, _, matchID := setupChatMatch(t)
	msgRepo := newMockMessageRepo()
	svc := service.NewChatService(msgRepo, matchRepo, testEncKey)

	// Send some messages.
	svc.SendMessage(context.Background(), userA, matchID, "msg one")
	svc.SendMessage(context.Background(), userA, matchID, "msg two")

	msgs, err := svc.GetMessages(context.Background(), matchID, userA, 1, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Content != "msg one" {
		t.Errorf("expected first message %q, got %q", "msg one", msgs[0].Content)
	}
	if msgs[1].Content != "msg two" {
		t.Errorf("expected second message %q, got %q", "msg two", msgs[1].Content)
	}
}

func TestGetMessages_NotInMatch(t *testing.T) {
	matchRepo, _, _, matchID := setupChatMatch(t)
	msgRepo := newMockMessageRepo()
	svc := service.NewChatService(msgRepo, matchRepo, testEncKey)

	outsider := uuid.New()
	_, err := svc.GetMessages(context.Background(), matchID, outsider, 1, 50)
	if err == nil {
		t.Fatal("expected error for outsider reading messages")
	}
}

func TestGetMessages_EmptyList(t *testing.T) {
	matchRepo, userA, _, matchID := setupChatMatch(t)
	msgRepo := newMockMessageRepo()
	svc := service.NewChatService(msgRepo, matchRepo, testEncKey)

	msgs, err := svc.GetMessages(context.Background(), matchID, userA, 1, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages, got %d", len(msgs))
	}
}

func TestMarkRead_Success(t *testing.T) {
	matchRepo, userA, userB, matchID := setupChatMatch(t)
	msgRepo := newMockMessageRepo()
	svc := service.NewChatService(msgRepo, matchRepo, testEncKey)

	// User A sends a message.
	svc.SendMessage(context.Background(), userA, matchID, "read me")

	// User B marks read.
	if err := svc.MarkRead(context.Background(), matchID, userB); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msgRepo.readOps != 1 {
		t.Errorf("expected 1 read operation, got %d", msgRepo.readOps)
	}
}

func TestMarkRead_NotInMatch(t *testing.T) {
	matchRepo, _, _, matchID := setupChatMatch(t)
	msgRepo := newMockMessageRepo()
	svc := service.NewChatService(msgRepo, matchRepo, testEncKey)

	outsider := uuid.New()
	err := svc.MarkRead(context.Background(), matchID, outsider)
	if err == nil {
		t.Fatal("expected error for outsider marking read")
	}
}

func TestRecipientID(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	match := &domain.Match{UserAID: userA, UserBID: userB}

	if got := service.RecipientID(match, userA); got != userB {
		t.Errorf("expected recipient %s, got %s", userB, got)
	}
	if got := service.RecipientID(match, userB); got != userA {
		t.Errorf("expected recipient %s, got %s", userA, got)
	}
}

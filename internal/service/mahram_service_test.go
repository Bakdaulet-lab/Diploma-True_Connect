package service_test

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
	"github.com/trueconnect/backend/internal/service"
)

// ── mockMahramRepo ─────────────────────────────────────────────────────────────

type mockMahramRepo struct {
	mu      sync.Mutex
	mahrams map[uuid.UUID]*domain.Mahram
}

func newMockMahramRepo() *mockMahramRepo {
	return &mockMahramRepo{mahrams: make(map[uuid.UUID]*domain.Mahram)}
}

func (m *mockMahramRepo) Create(_ context.Context, womanUserID uuid.UUID, _, _ []byte) (*domain.Mahram, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	mahram := &domain.Mahram{
		ID:                 uuid.New(),
		WomanUserID:        womanUserID,
		VerificationStatus: domain.MahramStatusPending,
		CreatedAt:          time.Now(),
	}
	m.mahrams[mahram.ID] = mahram
	return mahram, nil
}

func (m *mockMahramRepo) GetByWomanID(_ context.Context, womanUserID uuid.UUID) ([]*domain.Mahram, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*domain.Mahram
	for _, mh := range m.mahrams {
		if mh.WomanUserID == womanUserID {
			cp := *mh
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (m *mockMahramRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Mahram, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	mh, ok := m.mahrams[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *mh
	return &cp, nil
}

func (m *mockMahramRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string, _ *interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	mh, ok := m.mahrams[id]
	if !ok {
		return domain.ErrNotFound
	}
	mh.VerificationStatus = status
	if status == domain.MahramStatusVerified {
		now := time.Now()
		mh.VerifiedAt = &now
	}
	return nil
}

func (m *mockMahramRepo) SetTelegramChatID(_ context.Context, id uuid.UUID, chatID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	mh, ok := m.mahrams[id]
	if !ok {
		return domain.ErrNotFound
	}
	mh.TelegramChatID = &chatID
	return nil
}

var _ repository.MahramRepository = (*mockMahramRepo)(nil)

// ── helpers ───────────────────────────────────────────────────────────────────

func newTestMahramService() (*service.MahramService, *mockMahramRepo) {
	repo := newMockMahramRepo()
	key := make([]byte, 32) // zero key for tests
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewMahramService(repo, key, logger)
	return svc, repo
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestRegisterMahram_CreatesRecord(t *testing.T) {
	t.Parallel()

	svc, repo := newTestMahramService()
	womanID := uuid.New()

	view, otp, err := svc.RegisterMahram(context.Background(), womanID, "+77001234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if view == nil {
		t.Fatal("expected non-nil view")
	}
	if otp == "" || len(otp) != 6 {
		t.Fatalf("expected 6-digit OTP, got %q", otp)
	}
	if view.VerificationStatus != domain.MahramStatusPending {
		t.Errorf("expected status pending, got %q", view.VerificationStatus)
	}

	mahrams, err := repo.GetByWomanID(context.Background(), womanID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mahrams) != 1 {
		t.Fatalf("expected 1 mahram, got %d", len(mahrams))
	}
}

func TestVerifyMahram_ValidCode_SetsVerified(t *testing.T) {
	t.Parallel()

	svc, _ := newTestMahramService()
	womanID := uuid.New()

	view, _, err := svc.RegisterMahram(context.Background(), womanID, "+77001234567")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if err := svc.VerifyMahram(context.Background(), view.ID, "123456"); err != nil {
		t.Fatalf("verify: %v", err)
	}

	mahrams, _ := svc.GetMahrams(context.Background(), womanID)
	if len(mahrams) == 0 {
		t.Fatal("no mahrams returned")
	}
	if mahrams[0].VerificationStatus != domain.MahramStatusVerified {
		t.Errorf("expected verified, got %q", mahrams[0].VerificationStatus)
	}
}

func TestVerifyMahram_InvalidCode_ReturnsError(t *testing.T) {
	t.Parallel()

	svc, _ := newTestMahramService()
	womanID := uuid.New()

	view, _, err := svc.RegisterMahram(context.Background(), womanID, "+77001234567")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	cases := []string{"", "12345", "12345X", "1234567"}
	for _, code := range cases {
		if err := svc.VerifyMahram(context.Background(), view.ID, code); err == nil {
			t.Errorf("expected error for code %q, got nil", code)
		}
	}
}

func TestVerifyMahram_AlreadyVerified_ReturnsError(t *testing.T) {
	t.Parallel()

	svc, _ := newTestMahramService()
	womanID := uuid.New()

	view, _, err := svc.RegisterMahram(context.Background(), womanID, "+77001234567")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := svc.VerifyMahram(context.Background(), view.ID, "123456"); err != nil {
		t.Fatalf("first verify: %v", err)
	}
	if err := svc.VerifyMahram(context.Background(), view.ID, "123456"); err == nil {
		t.Error("expected error on second verify, got nil")
	}
}

func TestGetMahrams_ReturnsAll(t *testing.T) {
	t.Parallel()

	svc, _ := newTestMahramService()
	womanID := uuid.New()

	if _, _, err := svc.RegisterMahram(context.Background(), womanID, "+77001111111"); err != nil {
		t.Fatalf("register 1: %v", err)
	}
	if _, _, err := svc.RegisterMahram(context.Background(), womanID, "+77002222222"); err != nil {
		t.Fatalf("register 2: %v", err)
	}

	views, err := svc.GetMahrams(context.Background(), womanID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(views) != 2 {
		t.Errorf("expected 2 mahrams, got %d", len(views))
	}
}

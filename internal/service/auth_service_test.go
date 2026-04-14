package service_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	tcjwt "github.com/trueconnect/backend/internal/pkg/jwt"
	"github.com/trueconnect/backend/internal/repository"
	"github.com/trueconnect/backend/internal/service"
)

// ── In-memory mocks ───────────────────────────────────────────────────────────

type mockUserRepo struct {
	mu    sync.Mutex
	users map[string]*domain.User // keyed by phone_hash hex
	byID  map[uuid.UUID]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[string]*domain.User),
		byID:  make(map[uuid.UUID]*domain.User),
	}
}

func (m *mockUserRepo) UpdateFCMToken(ctx context.Context, id uuid.UUID, token string) error {
	return nil
}

func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := string(user.PhoneHash)
	if _, exists := m.users[key]; exists {
		return domain.ErrAlreadyExists
	}
	copy := *user
	m.users[key] = &copy
	m.byID[user.ID] = &copy
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	copy := *u
	return &copy, nil
}

func (m *mockUserRepo) GetByPhoneHash(_ context.Context, hash []byte) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[string(hash)]
	if !ok {
		return nil, domain.ErrNotFound
	}
	copy := *u
	return &copy, nil
}

func (m *mockUserRepo) UpdateVerificationLevel(_ context.Context, id uuid.UUID, level domain.VerificationLevel) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[id]
	if !ok {
		return domain.ErrNotFound
	}
	u.VerificationLevel = level
	return nil
}

func (m *mockUserRepo) UpdateTrustScore(_ context.Context, id uuid.UUID, score int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[id]
	if !ok {
		return domain.ErrNotFound
	}
	u.TrustScore = score
	return nil
}

func (m *mockUserRepo) UpdateTrustStatus(_ context.Context, id uuid.UUID, status domain.TrustStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[id]
	if !ok {
		return domain.ErrNotFound
	}
	u.TrustStatus = status
	return nil
}

func (m *mockUserRepo) UpdateLastLogin(_ context.Context, _ uuid.UUID) error { return nil }

func (m *mockUserRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.byID, id)
	return nil
}

// ДОБАВЛЕНО: метод для реализации интерфейса UserRepository
func (m *mockUserRepo) ListByTrustStatus(ctx context.Context, status domain.TrustStatus, limit, offset int) ([]*domain.User, error) {
	return []*domain.User{}, nil
}

// ──────────────────────────────────────────────────────────────────────────────

type mockTokenRepo struct {
	mu     sync.Mutex
	tokens map[string]*repository.RefreshToken // keyed by string(tokenHash)
}

func newMockTokenRepo() *mockTokenRepo {
	return &mockTokenRepo{tokens: make(map[string]*repository.RefreshToken)}
}

func (m *mockTokenRepo) Create(_ context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) (uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := uuid.New()
	m.tokens[string(tokenHash)] = &repository.RefreshToken{
		ID:        id,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		Revoked:   false,
	}
	return id, nil
}

func (m *mockTokenRepo) GetByTokenHash(_ context.Context, tokenHash []byte) (*repository.RefreshToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tokens[string(tokenHash)]
	if !ok || t.ExpiresAt.Before(time.Now()) {
		return nil, domain.ErrNotFound
	}
	copy := *t
	return &copy, nil
}

func (m *mockTokenRepo) Revoke(_ context.Context, tokenHash []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.tokens[string(tokenHash)]; ok {
		t.Revoked = true
	}
	return nil
}

func (m *mockTokenRepo) RevokeAllForUser(_ context.Context, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.tokens {
		if t.UserID == userID {
			t.Revoked = true
		}
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────

type mockSessionStore struct {
	mu       sync.Mutex
	failures map[string]int64
}

func newMockSessionStore() *mockSessionStore {
	return &mockSessionStore{failures: make(map[string]int64)}
}

func (m *mockSessionStore) StoreRefreshToken(_ context.Context, _, _ string, _ time.Duration) error {
	return nil
}
func (m *mockSessionStore) ValidateRefreshToken(_ context.Context, _, _ string) (bool, error) {
	return true, nil
}
func (m *mockSessionStore) RemoveRefreshToken(_ context.Context, _, _ string) error  { return nil }
func (m *mockSessionStore) RemoveAllRefreshTokens(_ context.Context, _ string) error { return nil }

func (m *mockSessionStore) IncrementAuthFailure(_ context.Context, phoneHash string, _ time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failures[phoneHash]++
	return m.failures[phoneHash], nil
}

func (m *mockSessionStore) GetAuthFailureCount(_ context.Context, phoneHash string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.failures[phoneHash], nil
}

func (m *mockSessionStore) ClearAuthFailures(_ context.Context, phoneHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.failures, phoneHash)
	return nil
}

// ДОБАВЛЕНО: метод для реализации интерфейса SessionStore
func (m *mockSessionStore) PublishUserBanned(ctx context.Context, userID string) error {
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────

type mockGraphRepo struct{}

func (m *mockGraphRepo) CreateUserNode(_ context.Context, _ uuid.UUID, _ domain.VerificationLevel) error {
	return nil
}
func (m *mockGraphRepo) AddRating(_ context.Context, _, _ uuid.UUID, _ int, _ string, _ bool) error {
	return nil
}
func (m *mockGraphRepo) AddMeeting(_ context.Context, _, _ uuid.UUID, _ bool) error  { return nil }
func (m *mockGraphRepo) AddReport(_ context.Context, _, _ uuid.UUID, _ string) error { return nil }
func (m *mockGraphRepo) ComputeTrustScore(_ context.Context, _ uuid.UUID) (int, error) {
	return 50, nil
}
func (m *mockGraphRepo) UpdateTrustScore(_ context.Context, _ uuid.UUID, _ int) error { return nil }
func (m *mockGraphRepo) DetectSybilClusters(_ context.Context) ([]repository.SybilCluster, error) {
	return nil, nil
}

// ДОБАВЛЕНО: метод для реализации интерфейса TrustGraphRepository
func (m *mockGraphRepo) DeleteUserNode(ctx context.Context, uid uuid.UUID) error {
	return nil
}

// ── Test helpers ──────────────────────────────────────────────────────────────

func testEncryptionKey() []byte {
	return bytes.Repeat([]byte("k"), 32)
}

func newTestAuthService() *service.AuthService {
	jwtMgr := tcjwt.NewManager("test-secret-which-is-32-chars-longg", 15*time.Minute)
	return service.NewAuthService(
		newMockUserRepo(),
		newMockTokenRepo(),
		newMockSessionStore(),
		&mockGraphRepo{},
		jwtMgr,
		testEncryptionKey(),
		7*24*time.Hour,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	t.Parallel()

	svc := newTestAuthService()
	result, err := svc.Register(context.Background(), service.RegisterInput{
		Phone:    "+77001234567",
		Password: "secure-password-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if result.UserID == uuid.Nil {
		t.Error("expected non-nil user ID")
	}
	if result.ExpiresIn != 900 {
		t.Errorf("expected ExpiresIn=900, got %d", result.ExpiresIn)
	}
}

func TestRegister_DuplicatePhone(t *testing.T) {
	t.Parallel()

	svc := newTestAuthService()
	input := service.RegisterInput{Phone: "+77001234567", Password: "password123"}

	if _, err := svc.Register(context.Background(), input); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	_, err := svc.Register(context.Background(), input)
	if err == nil {
		t.Fatal("expected error on duplicate registration")
	}
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got: %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	t.Parallel()

	svc := newTestAuthService()
	phone, password := "+77009876543", "my-password-456"

	if _, err := svc.Register(context.Background(), service.RegisterInput{Phone: phone, Password: password}); err != nil {
		t.Fatalf("register: %v", err)
	}

	result, err := svc.Login(context.Background(), phone, password)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected access token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	t.Parallel()

	svc := newTestAuthService()
	if _, err := svc.Register(context.Background(), service.RegisterInput{
		Phone: "+77001111111", Password: "correct-password",
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err := svc.Login(context.Background(), "+77001111111", "wrong-password")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestLogin_UnknownPhone(t *testing.T) {
	t.Parallel()

	svc := newTestAuthService()
	_, err := svc.Login(context.Background(), "+77000000000", "any-password")
	if err == nil {
		t.Fatal("expected error for unknown phone")
	}
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestLogin_BruteForceProtection(t *testing.T) {
	t.Parallel()

	svc := newTestAuthService()
	phone := "+77002222222"

	if _, err := svc.Register(context.Background(), service.RegisterInput{
		Phone: phone, Password: "correct",
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	// Trigger 5 failures.
	for i := 0; i < 5; i++ {
		svc.Login(context.Background(), phone, "wrong") //nolint:errcheck
	}

	// 6th attempt — even with the correct password — must be blocked.
	_, err := svc.Login(context.Background(), phone, "correct")
	if err == nil {
		t.Fatal("expected brute-force lockout")
	}
	if !errors.Is(err, domain.ErrRateLimitExceeded) {
		t.Errorf("expected ErrRateLimitExceeded, got: %v", err)
	}
}

func TestRefresh_Success(t *testing.T) {
	t.Parallel()

	svc := newTestAuthService()

	reg, err := svc.Register(context.Background(), service.RegisterInput{
		Phone: "+77003333333", Password: "password",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	result, err := svc.Refresh(context.Background(), reg.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected new access token")
	}
	if result.RefreshToken == reg.RefreshToken {
		t.Error("new refresh token must differ from the old one (one-time use)")
	}
}

func TestRefresh_TokenReuse_IsRejected(t *testing.T) {
	t.Parallel()

	svc := newTestAuthService()

	reg, err := svc.Register(context.Background(), service.RegisterInput{
		Phone: "+77004444444", Password: "password",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// Use the token once.
	if _, err := svc.Refresh(context.Background(), reg.RefreshToken); err != nil {
		t.Fatalf("first refresh: %v", err)
	}

	// Reusing the same token must fail.
	_, err = svc.Refresh(context.Background(), reg.RefreshToken)
	if err == nil {
		t.Fatal("expected error on refresh token reuse")
	}
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestLogout_TokenBecomesInvalid(t *testing.T) {
	t.Parallel()

	svc := newTestAuthService()

	reg, err := svc.Register(context.Background(), service.RegisterInput{
		Phone: "+77005555555", Password: "password",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if err := svc.Logout(context.Background(), reg.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}

	// After logout the refresh token must be invalid.
	_, err = svc.Refresh(context.Background(), reg.RefreshToken)
	if err == nil {
		t.Fatal("expected error refreshing after logout")
	}
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestLogin_SuspendedAccount(t *testing.T) {
	t.Parallel()

	userRepo := newMockUserRepo()
	jwtMgr := tcjwt.NewManager("test-secret-which-is-32-chars-longg", 15*time.Minute)
	svc := service.NewAuthService(
		userRepo,
		newMockTokenRepo(),
		newMockSessionStore(),
		&mockGraphRepo{},
		jwtMgr,
		testEncryptionKey(),
		7*24*time.Hour,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	ctx := context.Background()
	phone, password := "+77006666666", "password"

	// Register and capture the user ID from the result.
	reg, err := svc.Register(ctx, service.RegisterInput{Phone: phone, Password: password})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// Suspend via repo using the user ID returned by Register.
	if err := userRepo.UpdateTrustStatus(ctx, reg.UserID, domain.TrustStatusSuspended); err != nil {
		t.Fatalf("suspending user: %v", err)
	}

	_, err = svc.Login(ctx, phone, password)
	if err == nil {
		t.Fatal("expected error for suspended account")
	}
	if !errors.Is(err, domain.ErrAccountSuspended) {
		t.Errorf("expected ErrAccountSuspended, got: %v", err)
	}
}

func TestRefresh_TokenReuse_RevokesAllTokens(t *testing.T) {
	t.Parallel()

	tokenRepo := newMockTokenRepo()
	jwtMgr := tcjwt.NewManager("test-secret-which-is-32-chars-longg", 15*time.Minute)
	svc := service.NewAuthService(
		newMockUserRepo(),
		tokenRepo,
		newMockSessionStore(),
		&mockGraphRepo{},
		jwtMgr,
		testEncryptionKey(),
		7*24*time.Hour,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	ctx := context.Background()

	// Register to get a refresh token.
	reg, err := svc.Register(ctx, service.RegisterInput{Phone: "+77007777777", Password: "password"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	originalToken := reg.RefreshToken

	// Rotate: use the token to get a new pair.
	rotated, err := svc.Refresh(ctx, originalToken)
	if err != nil {
		t.Fatalf("first refresh: %v", err)
	}

	// Attacker replays the original (now revoked) token — triggers family revocation.
	_, err = svc.Refresh(ctx, originalToken)
	if err == nil {
		t.Fatal("expected error on replay of revoked token")
	}
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got: %v", err)
	}

	// The legitimate new token should also be revoked (entire family wiped).
	_, err = svc.Refresh(ctx, rotated.RefreshToken)
	if err == nil {
		t.Fatal("expected error: rotated token should be revoked after family revocation")
	}
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got: %v", err)
	}
}

func (m *mockGraphRepo) GetRecommendations(ctx context.Context, uid uuid.UUID, limit int) ([]uuid.UUID, error) {
	return nil, nil
}

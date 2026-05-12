package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/pkg/crypto"
	tcjwt "github.com/trueconnect/backend/internal/pkg/jwt"
	"github.com/trueconnect/backend/internal/repository"
)

// AuthService handles registration, login, token refresh, and logout.
type AuthService struct {
	userRepo      repository.UserRepository
	profileRepo   repository.ProfileRepository
	tokenRepo     repository.RefreshTokenRepository
	sessionStore  repository.SessionStore
	graphRepo     repository.TrustGraphRepository
	jwt           *tcjwt.Manager
	encryptionKey []byte
	refreshExpiry time.Duration
	log           *slog.Logger
}

// NewAuthService creates a new auth service with all dependencies injected.
func NewAuthService(
	userRepo repository.UserRepository,
	profileRepo repository.ProfileRepository,
	tokenRepo repository.RefreshTokenRepository,
	sessionStore repository.SessionStore,
	graphRepo repository.TrustGraphRepository,
	jwtManager *tcjwt.Manager,
	encryptionKey []byte,
	refreshExpiry time.Duration,
	log *slog.Logger,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		profileRepo:   profileRepo,
		tokenRepo:     tokenRepo,
		sessionStore:  sessionStore,
		graphRepo:     graphRepo,
		jwt:           jwtManager,
		encryptionKey: encryptionKey,
		refreshExpiry: refreshExpiry,
		log:           log,
	}
}

// RegisterInput represents the data needed to register a new user.
type RegisterInput struct {
	Phone       string
	Password    string
	DisplayName string
	PublicKey   *string
}

// AuthResult contains the tokens returned after successful authentication.
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // seconds until the access token expires
	UserID       uuid.UUID
}

// Register creates a new user account and returns an initial token pair.
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	if strings.TrimSpace(input.DisplayName) == "" {
		return nil, fmt.Errorf("register: display name is required: %w", domain.ErrInvalidInput)
	}

	passwordHash, err := crypto.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	phoneHash := crypto.SHA256Hash([]byte(input.Phone))
	phoneEncrypted, err := crypto.Encrypt([]byte(input.Phone), s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("encrypting phone: %w", err)
	}

	user := &domain.User{
		ID:                uuid.New(),
		PhoneHash:         phoneHash,
		PhoneEncrypted:    phoneEncrypted,
		PasswordHash:      passwordHash,
		PublicKey:         input.PublicKey,
		VerificationLevel: domain.VerificationNone,
		TrustStatus:       domain.TrustStatusNormal,
		TrustScore:        50,
		IsActive:          true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return nil, fmt.Errorf("register: %w", domain.ErrAlreadyExists)
		}
		return nil, fmt.Errorf("register: creating user: %w", err)
	}

	profile := &domain.Profile{
		UserID:        user.ID,
		DisplayName:   input.DisplayName,
		MaritalStatus: domain.MaritalSingle,
	}
	if err := s.profileRepo.Upsert(ctx, profile); err != nil {
		return nil, fmt.Errorf("register: creating profile: %w", err)
	}

	// Create user node in Neo4j trust graph; non-fatal if it fails.
	if err := s.graphRepo.CreateUserNode(ctx, user.ID, user.VerificationLevel); err != nil {
		s.log.Error("failed to create Neo4j user node; trust graph may be incomplete",
			slog.String("user_id", user.ID.String()),
			slog.String("error", err.Error()),
		)
	}

	return s.issueTokens(ctx, user)
}

// Login authenticates a user with phone + password and returns a token pair.
func (s *AuthService) Login(ctx context.Context, phone, password string) (*AuthResult, error) {
	phoneHash := crypto.SHA256Hash([]byte(phone))
	phoneHashHex := hex.EncodeToString(phoneHash)

	// Brute-force protection: block after 5 failed attempts per 15-minute window.
	failCount, err := s.sessionStore.GetAuthFailureCount(ctx, phoneHashHex)
	if err != nil {
		return nil, fmt.Errorf("login: checking auth failures: %w", err)
	}
	if failCount >= 5 {
		return nil, fmt.Errorf("login: %w", domain.ErrRateLimitExceeded)
	}

	user, err := s.userRepo.GetByPhoneHash(ctx, phoneHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Increment even on unknown phone to prevent timing-based user enumeration.
			_, _ = s.sessionStore.IncrementAuthFailure(ctx, phoneHashHex, 15*time.Minute)
			return nil, fmt.Errorf("login: %w", domain.ErrInvalidCredentials)
		}
		return nil, fmt.Errorf("login: getting user: %w", err)
	}

	if user.TrustStatus == domain.TrustStatusSuspended || user.TrustStatus == domain.TrustStatusBanned {
		return nil, fmt.Errorf("login: %w", domain.ErrAccountSuspended)
	}

	match, err := crypto.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("login: verifying password: %w", err)
	}
	if !match {
		_, _ = s.sessionStore.IncrementAuthFailure(ctx, phoneHashHex, 15*time.Minute)
		return nil, fmt.Errorf("login: %w", domain.ErrInvalidCredentials)
	}

	_ = s.sessionStore.ClearAuthFailures(ctx, phoneHashHex)
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	return s.issueTokens(ctx, user)
}

// Refresh rotates a refresh token and issues a new access + refresh token pair.
// Reuse of a revoked token triggers full session revocation (compromise detection).
func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string) (*AuthResult, error) {
	tokenHash := crypto.SHA256Hash([]byte(rawRefreshToken))

	storedToken, err := s.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("refresh: %w", domain.ErrUnauthorized)
		}
		return nil, fmt.Errorf("refresh: getting token: %w", err)
	}

	// Compromise detection: if a revoked token is presented, an attacker is
	// replaying a previously-consumed token. Revoke the entire family.
	if storedToken.Revoked {
		_ = s.tokenRepo.RevokeAllForUser(ctx, storedToken.UserID)
		_ = s.sessionStore.RemoveAllRefreshTokens(ctx, storedToken.UserID.String())
		return nil, fmt.Errorf("refresh: token reuse detected: %w", domain.ErrUnauthorized)
	}

	// One-time use: revoke the consumed token immediately.
	if err := s.tokenRepo.Revoke(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("refresh: revoking old token: %w", err)
	}
	_ = s.sessionStore.RemoveRefreshToken(ctx, storedToken.UserID.String(), hex.EncodeToString(tokenHash))

	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("refresh: getting user: %w", err)
	}

	if user.TrustStatus == domain.TrustStatusSuspended || user.TrustStatus == domain.TrustStatusBanned {
		return nil, fmt.Errorf("refresh: %w", domain.ErrAccountSuspended)
	}

	return s.issueTokens(ctx, user)
}

// Logout revokes the provided refresh token.
func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	tokenHash := crypto.SHA256Hash([]byte(rawRefreshToken))

	// Look up first to get user ID for Redis cleanup (best-effort).
	storedToken, _ := s.tokenRepo.GetByTokenHash(ctx, tokenHash)

	if err := s.tokenRepo.Revoke(ctx, tokenHash); err != nil {
		return fmt.Errorf("logout: revoking token: %w", err)
	}

	if storedToken != nil {
		_ = s.sessionStore.RemoveRefreshToken(ctx, storedToken.UserID.String(), hex.EncodeToString(tokenHash))
	}

	return nil
}

// issueTokens creates a new JWT + refresh token pair and persists the refresh token.
func (s *AuthService) issueTokens(ctx context.Context, user *domain.User) (*AuthResult, error) {
	accessToken, err := s.jwt.Generate(user.ID, user.VerificationLevel, user.TrustStatus, user.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	rawRefreshToken, err := generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	tokenHash := crypto.SHA256Hash([]byte(rawRefreshToken))
	expiresAt := time.Now().Add(s.refreshExpiry)

	if _, err := s.tokenRepo.Create(ctx, user.ID, tokenHash, expiresAt); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	_ = s.sessionStore.StoreRefreshToken(ctx, user.ID.String(), hex.EncodeToString(tokenHash), s.refreshExpiry)

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresIn:    900, // 15 minutes in seconds
		UserID:       user.ID,
	}, nil
}

// generateSecureToken produces a hex-encoded cryptographically random token.
func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

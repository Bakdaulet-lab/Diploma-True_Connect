package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/pkg/crypto"
	"github.com/trueconnect/backend/internal/repository"
)

// UserService handles user account operations.
type UserService struct {
	userRepo      repository.UserRepository
	tokenRepo     repository.RefreshTokenRepository
	sessionStore  repository.SessionStore
	encryptionKey []byte
}

// NewUserService creates a new user service.
func NewUserService(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	sessionStore repository.SessionStore,
	encryptionKey []byte,
) *UserService {
	return &UserService{
		userRepo:      userRepo,
		tokenRepo:     tokenRepo,
		sessionStore:  sessionStore,
		encryptionKey: encryptionKey,
	}
}

// UserView is the response DTO for the authenticated user's account.
type UserView struct {
	ID                uuid.UUID                `json:"id"`
	Phone             string                   `json:"phone"`
	VerificationLevel domain.VerificationLevel `json:"verification_level"`
	TrustStatus       domain.TrustStatus       `json:"trust_status"`
	TrustScore        int                      `json:"trust_score"`
	CreatedAt         string                   `json:"created_at"`
}

// GetMe returns the authenticated user's account info with decrypted phone.
func (s *UserService) GetMe(ctx context.Context, userID uuid.UUID) (*UserView, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get me: %w", err)
	}

	phone, err := crypto.Decrypt(user.PhoneEncrypted, s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("get me: decrypting phone: %w", err)
	}

	return &UserView{
		ID:                user.ID,
		Phone:             string(phone),
		VerificationLevel: user.VerificationLevel,
		TrustStatus:       user.TrustStatus,
		TrustScore:        user.TrustScore,
		CreatedAt:         user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// DeleteMe soft-deletes the user and revokes all sessions.
func (s *UserService) DeleteMe(ctx context.Context, userID uuid.UUID) error {
	// Soft-delete the user record.
	if err := s.userRepo.SoftDelete(ctx, userID); err != nil {
		return fmt.Errorf("delete me: %w", err)
	}

	// Revoke all refresh tokens for this user.
	if err := s.tokenRepo.RevokeAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete me: revoking tokens: %w", err)
	}

	// Remove all session data from Redis.
	if err := s.sessionStore.RemoveAllRefreshTokens(ctx, userID.String()); err != nil {
		return fmt.Errorf("delete me: removing sessions: %w", err)
	}

	return nil
}

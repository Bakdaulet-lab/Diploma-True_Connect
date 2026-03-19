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
	graphRepo     repository.TrustGraphRepository
	encryptionKey []byte
}

// NewUserService creates a new user service.
func NewUserService(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	sessionStore repository.SessionStore,
	graphRepo repository.TrustGraphRepository,
	encryptionKey []byte,
) *UserService {
	return &UserService{
		userRepo:      userRepo,
		tokenRepo:     tokenRepo,
		sessionStore:  sessionStore,
		graphRepo:     graphRepo,
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
	// Soft-delete the user record (and cascades app-level data in repo).
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

	// Remove the user from the Trust Graph to prevent orphaned nodes
	_ = s.graphRepo.DeleteUserNode(ctx, userID)

	// Disconnect active websockets
	_ = s.sessionStore.PublishUserBanned(ctx, userID.String())

	return nil
}

// BanUser restricts a user from platform access instantly. Used by SybilDetector and Admins.
func (s *UserService) BanUser(ctx context.Context, userID uuid.UUID, reason string) error {
	if err := s.userRepo.UpdateTrustStatus(ctx, userID, domain.TrustStatusBanned); err != nil {
		return fmt.Errorf("ban user: %w", err)
	}

	// Remove from Neo4j so their edges don't break trust computations
	_ = s.graphRepo.DeleteUserNode(ctx, userID)

	// Flush tokens and sessions
	_ = s.tokenRepo.RevokeAllForUser(ctx, userID)
	_ = s.sessionStore.RemoveAllRefreshTokens(ctx, userID.String())

	// Push ban event to active instances so WebSockets are killed
	_ = s.sessionStore.PublishUserBanned(ctx, userID.String())

	return nil
}

func (s *UserService) ListUsersUnderReview(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	return s.userRepo.ListByTrustStatus(ctx, domain.TrustStatusUnderReview, limit, offset)
}

func (s *UserService) ReviewSybilVerdict(ctx context.Context, targetID uuid.UUID, isBan bool) error {
	if isBan {
		if err := s.BanUser(ctx, targetID, "manual review: sybil confirmed"); err != nil {
			return err
		}
	} else {
		// Restore to normal
		if err := s.userRepo.UpdateTrustStatus(ctx, targetID, domain.TrustStatusNormal); err != nil {
			return fmt.Errorf("restoring user: %w", err)
		}
	}
	return nil
}

func (s *UserService) UpdateFCMToken(ctx context.Context, userID uuid.UUID, token string) error {
	return s.userRepo.UpdateFCMToken(ctx, userID, token)
}

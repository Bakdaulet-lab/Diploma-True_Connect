package service

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/pkg/crypto"
	"github.com/trueconnect/backend/internal/repository"
)

// MahramService manages mahram (chaperone) registration and OTP verification.
type MahramService struct {
	mahramRepo    repository.MahramRepository
	encryptionKey []byte
	log           *slog.Logger
}

// NewMahramService creates a new mahram service.
func NewMahramService(
	mahramRepo repository.MahramRepository,
	encryptionKey []byte,
	log *slog.Logger,
) *MahramService {
	return &MahramService{
		mahramRepo:    mahramRepo,
		encryptionKey: encryptionKey,
		log:           log,
	}
}

// MahramView is the API-visible representation of a mahram record.
type MahramView struct {
	ID                 uuid.UUID  `json:"id"`
	WomanUserID        uuid.UUID  `json:"woman_user_id"`
	VerificationStatus string     `json:"verification_status"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// RegisterMahram registers a new mahram for the authenticated woman user.
// In dev mode the OTP is logged to stdout and returned in the response for testing.
// In production this would dispatch a Telegram Bot API message to @MahabbatVerifyBot.
func (s *MahramService) RegisterMahram(ctx context.Context, womanUserID uuid.UUID, phone string) (*MahramView, string, error) {
	phoneEncrypted, err := crypto.Encrypt([]byte(phone), s.encryptionKey)
	if err != nil {
		return nil, "", fmt.Errorf("register mahram: encrypting phone: %w", err)
	}
	phoneHash := crypto.SHA256Hash([]byte(phone))

	mahram, err := s.mahramRepo.Create(ctx, womanUserID, phoneEncrypted, phoneHash)
	if err != nil {
		return nil, "", fmt.Errorf("register mahram: %w", err)
	}

	otp := generateOTP()
	// Dev mode: log OTP. In production: send via Telegram Bot API.
	s.log.Info("mahram OTP generated (dev mode — send via @MahabbatVerifyBot in production)",
		slog.String("mahram_id", mahram.ID.String()),
		slog.String("otp", otp),
	)

	return toMahramView(mahram), otp, nil
}

// VerifyMahram accepts an OTP code for a mahram registration.
// Dev mode accepts any 6-digit numeric code.
func (s *MahramService) VerifyMahram(ctx context.Context, mahramID uuid.UUID, code string) error {
	mahram, err := s.mahramRepo.GetByID(ctx, mahramID)
	if err != nil {
		return fmt.Errorf("verify mahram: %w", err)
	}

	if mahram.VerificationStatus == domain.MahramStatusVerified {
		return fmt.Errorf("verify mahram: already verified: %w", domain.ErrAlreadyExists)
	}

	// Dev mode: accept any 6-digit code.
	if len(code) != 6 {
		return fmt.Errorf("verify mahram: invalid code format: %w", domain.ErrInvalidInput)
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return fmt.Errorf("verify mahram: code must be numeric: %w", domain.ErrInvalidInput)
		}
	}

	now := time.Now()
	var nowIface interface{} = now
	if err := s.mahramRepo.UpdateStatus(ctx, mahramID, domain.MahramStatusVerified, &nowIface); err != nil {
		return fmt.Errorf("verify mahram: updating status: %w", err)
	}

	return nil
}

// GetMahrams returns all mahrams registered by the given woman user.
func (s *MahramService) GetMahrams(ctx context.Context, womanUserID uuid.UUID) ([]*MahramView, error) {
	mahrams, err := s.mahramRepo.GetByWomanID(ctx, womanUserID)
	if err != nil {
		return nil, fmt.Errorf("get mahrams: %w", err)
	}

	views := make([]*MahramView, len(mahrams))
	for i, m := range mahrams {
		views[i] = toMahramView(m)
	}
	return views, nil
}

func toMahramView(m *domain.Mahram) *MahramView {
	return &MahramView{
		ID:                 m.ID,
		WomanUserID:        m.WomanUserID,
		VerificationStatus: m.VerificationStatus,
		VerifiedAt:         m.VerifiedAt,
		CreatedAt:          m.CreatedAt,
	}
}

func generateOTP() string {
	const digits = "0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}

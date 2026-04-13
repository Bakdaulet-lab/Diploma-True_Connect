package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AdminDashboardStats struct {
	TotalActiveUsers  int     `json:"total_active_users"`
	MatchesMadeToday  int     `json:"matches_made_today"`
	AverageTrustScore float64 `json:"average_trust_score"`
}

type AdminUserRow struct {
	ID                uuid.UUID `json:"id"`
	PhoneEncrypted    []byte    `json:"-"`
	EmailEncrypted    []byte    `json:"-"`
	Phone             string    `json:"phone"`
	Email             string    `json:"email"`
	DisplayName       *string   `json:"display_name"`
	AvatarURL         *string   `json:"avatar_url"`
	TrustScore        int       `json:"trust_score"`
	TrustStatus       string    `json:"trust_status"`
	VerificationLevel string    `json:"verification_level"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
}

type AdminRepository interface {
	GetDashboardStats(ctx context.Context) (*AdminDashboardStats, error)
	SearchUsers(ctx context.Context, phoneHash []byte, nameQuery string, limit, offset int) ([]*AdminUserRow, error)
}

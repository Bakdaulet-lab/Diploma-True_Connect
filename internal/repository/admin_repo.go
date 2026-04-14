package repository

import (
"context"
"time"

"github.com/google/uuid"
)

type PendingKYC struct {
ID             uuid.UUID `json:"id"`
UserID         uuid.UUID `json:"user_id"`
DisplayName    *string   `json:"display_name"`
PhoneEncrypted []byte    `json:"-"`
Phone          string    `json:"phone"`
DocumentURL    string    `json:"-"`
SubmittedAt    time.Time `json:"submitted_at"`
}

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


type SybilClusterRow struct {
ID                  uuid.UUID   `json:"id"`
CommunityID         int         `json:"community_id"`
Size                int         `json:"size"`
ExternalConnections int         `json:"external_connections"`
SuspectUIDs         []uuid.UUID `json:"suspect_uids"`
Status              string      `json:"status"`
CreatedAt           time.Time   `json:"created_at"`
ResolvedAt          *time.Time  `json:"resolved_at"`
}

type AdminRepository interface {
GetDashboardStats(ctx context.Context) (*AdminDashboardStats, error)
SearchUsers(ctx context.Context, phoneHash []byte, nameQuery string, limit, offset int) ([]*AdminUserRow, error)
GetPendingKYC(ctx context.Context, limit, offset int) ([]*PendingKYC, error)
GetKYCDocument(ctx context.Context, kycID uuid.UUID) (string, error)

UpdateKYCStatus(ctx context.Context, kycID uuid.UUID, status string) (uuid.UUID, error)
SaveSybilCluster(ctx context.Context, cluster *SybilClusterRow) error
GetPendingSybilClusters(ctx context.Context, limit, offset int) ([]*SybilClusterRow, error)
ResolveSybilCluster(ctx context.Context, clusterID uuid.UUID, status string) ([]uuid.UUID, error)
}

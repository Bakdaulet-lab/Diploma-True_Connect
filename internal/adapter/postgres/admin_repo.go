package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trueconnect/backend/internal/repository"
)

type adminRepo struct {
	db *pgxpool.Pool
}

func NewAdminRepo(db *pgxpool.Pool) repository.AdminRepository {
	return &adminRepo{db: db}
}

func (r *adminRepo) GetDashboardStats(ctx context.Context) (*repository.AdminDashboardStats, error) {
	var stats repository.AdminDashboardStats

	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM social.users 
		WHERE is_active = true
	`).Scan(&stats.TotalActiveUsers)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM social.matches 
		WHERE created_at >= current_date
	`).Scan(&stats.MatchesMadeToday)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(trust_score), 0)
		FROM social.users 
		WHERE is_active = true
	`).Scan(&stats.AverageTrustScore)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *adminRepo) SearchUsers(ctx context.Context, phoneHash []byte, nameQuery string, limit, offset int) ([]*repository.AdminUserRow, error) {
	query := `
		SELECT 
			u.id, u.phone_encrypted, u.email_encrypted, u.trust_score, 
			u.trust_status, u.verification_level, u.is_active, u.created_at,
			p.display_name, p.avatar_url
		FROM social.users u
		LEFT JOIN social.profiles p ON u.id = p.user_id
		WHERE ($1::bytea IS NULL OR u.phone_hash = $1)
		  AND ($2::text = '' OR p.display_name ILIKE '%' || $2 || '%')
		ORDER BY u.created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, phoneHash, nameQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*repository.AdminUserRow
	for rows.Next() {
		var u repository.AdminUserRow
		err := rows.Scan(
			&u.ID, &u.PhoneEncrypted, &u.EmailEncrypted, &u.TrustScore,
			&u.TrustStatus, &u.VerificationLevel, &u.IsActive, &u.CreatedAt,
			&u.DisplayName, &u.AvatarURL,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	return users, rows.Err()
}

func (r *adminRepo) GetPendingKYC(ctx context.Context, limit, offset int) ([]*repository.PendingKYC, error) {
	query := `SELECT v.id, v.user_id, p.display_name, u.phone_encrypted, v.document_url, v.created_at
FROM identity_vault.verifications v
JOIN social.users u ON v.user_id = u.id
LEFT JOIN social.profiles p ON u.id = p.user_id
WHERE v.status = 'pending'
ORDER BY v.created_at DESC
LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []*repository.PendingKYC
	for rows.Next() {
		var req repository.PendingKYC
		err := rows.Scan(&req.ID, &req.UserID, &req.DisplayName, &req.PhoneEncrypted, &req.DocumentURL, &req.SubmittedAt)
		if err != nil {
			return nil, err
		}
		reqs = append(reqs, &req)
	}
	return reqs, rows.Err()
}

func (r *adminRepo) GetKYCDocument(ctx context.Context, kycID uuid.UUID) (string, error) {
	query := `SELECT document_url FROM identity_vault.verifications WHERE id = $1`
	var url string
	err := r.db.QueryRow(ctx, query, kycID).Scan(&url)
	return url, err
}

func (r *adminRepo) UpdateKYCStatus(ctx context.Context, kycID uuid.UUID, status string) (uuid.UUID, error) {
	query := `UPDATE identity_vault.verifications SET status = $1, updated_at = NOW() WHERE id = $2 RETURNING user_id`
	var userID uuid.UUID
	err := r.db.QueryRow(ctx, query, status, kycID).Scan(&userID)
	return userID, err
}

func (r *adminRepo) SaveSybilCluster(ctx context.Context, cluster *repository.SybilClusterRow) error {
	query := `
INSERT INTO social.sybil_clusters (community_id, size, external_connections, suspect_uids)
VALUES ($1, $2, $3, $4)
`
	// Use jackc pgx directly instead of runner wrapper for admin queries.
	_, err := r.db.Exec(ctx, query, cluster.CommunityID, cluster.Size, cluster.ExternalConnections, cluster.SuspectUIDs)
	return err
}

func (r *adminRepo) GetPendingSybilClusters(ctx context.Context, limit, offset int) ([]*repository.SybilClusterRow, error) {
	query := `
SELECT id, community_id, size, external_connections, suspect_uids, status, created_at, resolved_at
FROM social.sybil_clusters
WHERE status = 'pending'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []*repository.SybilClusterRow
	for rows.Next() {
		var c repository.SybilClusterRow
		err := rows.Scan(&c.ID, &c.CommunityID, &c.Size, &c.ExternalConnections, &c.SuspectUIDs, &c.Status, &c.CreatedAt, &c.ResolvedAt)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, &c)
	}
	return clusters, rows.Err()
}

func (r *adminRepo) ResolveSybilCluster(ctx context.Context, clusterID uuid.UUID, status string) ([]uuid.UUID, error) {
	query := `
UPDATE social.sybil_clusters 
SET status = $1, resolved_at = NOW() 
WHERE id = $2 
RETURNING suspect_uids
`
	var suspects []uuid.UUID
	err := r.db.QueryRow(ctx, query, status, clusterID).Scan(&suspects)
	return suspects, err
}

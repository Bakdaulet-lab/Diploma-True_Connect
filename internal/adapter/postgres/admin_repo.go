package postgres

import (
	"context"

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

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trueconnect/backend/internal/repository"
)

type AuditRepo struct {
	pool *pgxpool.Pool
}

func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{pool: pool}
}

func (r *AuditRepo) LogAction(ctx context.Context, log *repository.AuditLog) error {
	query := `
		INSERT INTO identity_vault.access_log (user_id, accessed_by, action, ip_address)
		VALUES ($1, $2, $3, $4)
	`

	action := fmt.Sprintf("[%s] %s", log.Method, log.Path)

	// Ensure action length does not exceed VARCHAR(30)
	if len(action) > 30 {
		action = action[:30]
	}

	_, err := r.pool.Exec(ctx, query, log.UserID, log.UserID.String(), action, log.IPAddress)
	return err
}

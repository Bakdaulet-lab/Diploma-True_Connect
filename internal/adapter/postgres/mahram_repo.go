package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// MahramRepo implements repository.MahramRepository using PostgreSQL.
type MahramRepo struct {
	pool *pgxpool.Pool
}

var _ repository.MahramRepository = (*MahramRepo)(nil)

// NewMahramRepo creates a new PostgreSQL-backed mahram repository.
func NewMahramRepo(pool *pgxpool.Pool) *MahramRepo {
	return &MahramRepo{pool: pool}
}

func (r *MahramRepo) Create(ctx context.Context, womanUserID uuid.UUID, phoneEncrypted, phoneHash []byte) (*domain.Mahram, error) {
	query := `
		INSERT INTO social.mahrams (woman_user_id, mahram_phone_encrypted, mahram_phone_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	m := &domain.Mahram{
		WomanUserID:        womanUserID,
		VerificationStatus: domain.MahramStatusPending,
	}
	err := runner(ctx, r.pool).QueryRow(ctx, query, womanUserID, phoneEncrypted, phoneHash).
		Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating mahram: %w", err)
	}
	return m, nil
}

func (r *MahramRepo) GetByWomanID(ctx context.Context, womanUserID uuid.UUID) ([]*domain.Mahram, error) {
	query := `
		SELECT id, woman_user_id, mahram_phone_hash, telegram_chat_id,
		       verification_status, verified_at, created_at
		FROM social.mahrams
		WHERE woman_user_id = $1
		ORDER BY created_at DESC`

	rows, err := runner(ctx, r.pool).Query(ctx, query, womanUserID)
	if err != nil {
		return nil, fmt.Errorf("listing mahrams: %w", err)
	}
	defer rows.Close()

	var mahrams []*domain.Mahram
	for rows.Next() {
		m := &domain.Mahram{}
		if err := rows.Scan(
			&m.ID, &m.WomanUserID, &m.MahramPhoneHash, &m.TelegramChatID,
			&m.VerificationStatus, &m.VerifiedAt, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning mahram: %w", err)
		}
		mahrams = append(mahrams, m)
	}
	return mahrams, rows.Err()
}

func (r *MahramRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Mahram, error) {
	query := `
		SELECT id, woman_user_id, mahram_phone_hash, telegram_chat_id,
		       verification_status, verified_at, created_at
		FROM social.mahrams
		WHERE id = $1`

	m := &domain.Mahram{}
	err := runner(ctx, r.pool).QueryRow(ctx, query, id).Scan(
		&m.ID, &m.WomanUserID, &m.MahramPhoneHash, &m.TelegramChatID,
		&m.VerificationStatus, &m.VerifiedAt, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("getting mahram: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("getting mahram: %w", err)
	}
	return m, nil
}

func (r *MahramRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, verifiedAt *interface{}) error {
	var verifiedAtTime *time.Time
	if verifiedAt != nil {
		if t, ok := (*verifiedAt).(time.Time); ok {
			verifiedAtTime = &t
		}
	}

	query := `
		UPDATE social.mahrams
		SET verification_status = $2, verified_at = $3
		WHERE id = $1`

	_, err := runner(ctx, r.pool).Exec(ctx, query, id, status, verifiedAtTime)
	if err != nil {
		return fmt.Errorf("updating mahram status: %w", err)
	}
	return nil
}

func (r *MahramRepo) SetTelegramChatID(ctx context.Context, id uuid.UUID, chatID int64) error {
	query := `UPDATE social.mahrams SET telegram_chat_id = $2 WHERE id = $1`
	_, err := runner(ctx, r.pool).Exec(ctx, query, id, chatID)
	if err != nil {
		return fmt.Errorf("setting telegram chat id: %w", err)
	}
	return nil
}

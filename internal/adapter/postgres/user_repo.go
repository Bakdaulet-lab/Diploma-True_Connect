package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// UserRepo implements repository.UserRepository using PostgreSQL.
type UserRepo struct {
	pool *pgxpool.Pool
}

var _ repository.UserRepository = (*UserRepo)(nil)

// NewUserRepo creates a new PostgreSQL-backed user repository.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO social.users (
			id, phone_hash, phone_encrypted, email_encrypted,
			password_hash, public_key, verification_level, trust_status, trust_score, is_active, fcm_token
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`

	err := runner(ctx, r.pool).QueryRow(ctx, query,
		user.ID,
		user.PhoneHash,
		user.PhoneEncrypted,
		user.EmailEncrypted,
		user.PasswordHash,
		user.PublicKey,
		user.VerificationLevel,
		user.TrustStatus,
		user.TrustScore,
		user.IsActive,
		user.FCMToken,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if isDuplicateKey(err) {
			return fmt.Errorf("creating user: %w", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("creating user: %w", err)
	}

	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, phone_hash, phone_encrypted, email_encrypted,
			   password_hash, public_key, verification_level, trust_status, trust_score,
			   is_active, last_login_at, fcm_token, created_at, updated_at
		FROM social.users
		WHERE id = $1 AND is_active = true`

	user := &domain.User{}
	err := runner(ctx, r.pool).QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.PhoneHash,
		&user.PhoneEncrypted,
		&user.EmailEncrypted,
		&user.PasswordHash,
		&user.PublicKey,
		&user.VerificationLevel,
		&user.TrustStatus,
		&user.TrustScore,
		&user.IsActive,
		&user.LastLoginAt,
		&user.FCMToken,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("getting user by id: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("getting user by id: %w", err)
	}

	return user, nil
}

func (r *UserRepo) GetByPhoneHash(ctx context.Context, phoneHash []byte) (*domain.User, error) {
	query := `
		SELECT id, phone_hash, phone_encrypted, email_encrypted,
			   password_hash, public_key, verification_level, trust_status, trust_score,
			   is_active, last_login_at, fcm_token, created_at, updated_at
		FROM social.users
		WHERE phone_hash = $1 AND is_active = true`

	user := &domain.User{}
	err := runner(ctx, r.pool).QueryRow(ctx, query, phoneHash).Scan(
		&user.ID,
		&user.PhoneHash,
		&user.PhoneEncrypted,
		&user.EmailEncrypted,
		&user.PasswordHash,
		&user.PublicKey,
		&user.VerificationLevel,
		&user.TrustStatus,
		&user.TrustScore,
		&user.IsActive,
		&user.LastLoginAt,
		&user.FCMToken,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("getting user by phone: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("getting user by phone: %w", err)
	}

	return user, nil
}

func (r *UserRepo) UpdateVerificationLevel(ctx context.Context, id uuid.UUID, level domain.VerificationLevel) error {
	query := `
		UPDATE social.users
		SET verification_level = $2, updated_at = NOW()
		WHERE id = $1 AND is_active = true`

	tag, err := runner(ctx, r.pool).Exec(ctx, query, id, level)
	if err != nil {
		return fmt.Errorf("updating verification level: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("updating verification level: %w", domain.ErrNotFound)
	}

	return nil
}

func (r *UserRepo) UpdateTrustScore(ctx context.Context, id uuid.UUID, score int) error {
	query := `
		UPDATE social.users
		SET trust_score = $2, updated_at = NOW()
		WHERE id = $1 AND is_active = true`

	tag, err := runner(ctx, r.pool).Exec(ctx, query, id, score)
	if err != nil {
		return fmt.Errorf("updating trust score: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("updating trust score: %w", domain.ErrNotFound)
	}

	return nil
}

func (r *UserRepo) UpdateTrustStatus(ctx context.Context, id uuid.UUID, status domain.TrustStatus) error {
	query := `
		UPDATE social.users
		SET trust_status = $2, updated_at = NOW()
		WHERE id = $1 AND is_active = true`

	tag, err := runner(ctx, r.pool).Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("updating trust status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("updating trust status: %w", domain.ErrNotFound)
	}

	return nil
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE social.users
		SET last_login_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND is_active = true`

	tag, err := runner(ctx, r.pool).Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("updating last login: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("updating last login: %w", domain.ErrNotFound)
	}

	return nil
}

func (r *UserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("soft deleting user begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE social.users
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND is_active = true`

	tag, err := tx.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft deleting user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("soft deleting user: %w", domain.ErrNotFound)
	}

	// Cascade delete posts
	if _, err := tx.Exec(ctx, "DELETE FROM social.posts WHERE author_id = $1", id); err != nil {
		return fmt.Errorf("deleting user posts: %w", err)
	}

	// Cascade delete matches
	if _, err := tx.Exec(ctx, "DELETE FROM social.matches WHERE user_a_id = $1 OR user_b_id = $1", id); err != nil {
		return fmt.Errorf("deleting user matches: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit soft delete: %w", err)
	}

	return nil
}

// isDuplicateKey checks if a pgx error is a unique constraint violation (code 23505).
func isDuplicateKey(err error) bool {
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}

func (r *UserRepo) ListByTrustStatus(ctx context.Context, status domain.TrustStatus, limit, offset int) ([]*domain.User, error) {
	query := `
		SELECT id, phone_hash, phone_encrypted, email_encrypted, password_hash,
		       verification_level, trust_status, trust_score, is_active, last_login_at, fcm_token, created_at, updated_at
		FROM social.users
		WHERE trust_status = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3`

	rows, err := runner(ctx, r.pool).Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list by trust status: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.PhoneHash,
			&user.PhoneEncrypted,
			&user.EmailEncrypted,
			&user.PasswordHash,
			&user.VerificationLevel,
			&user.TrustStatus,
			&user.TrustScore,
			&user.IsActive,
			&user.LastLoginAt,
			&user.FCMToken,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *UserRepo) UpdateFCMToken(ctx context.Context, id uuid.UUID, token string) error {
	query := "UPDATE social.users SET fcm_token = $2, updated_at = NOW() WHERE id = $1"
	tag, err := runner(ctx, r.pool).Exec(ctx, query, id, token)
	if err != nil {
		return fmt.Errorf("updating fcm token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

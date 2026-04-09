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

// ProfileRepo implements repository.ProfileRepository using PostgreSQL.
type ProfileRepo struct {
	pool *pgxpool.Pool
}

var _ repository.ProfileRepository = (*ProfileRepo)(nil)

// NewProfileRepo creates a new PostgreSQL-backed profile repository.
func NewProfileRepo(pool *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{pool: pool}
}

func (r *ProfileRepo) Upsert(ctx context.Context, p *domain.Profile) error {
	query := `
		INSERT INTO social.profiles (
			user_id, display_name, bio, gender, birth_date,
			city, looking_for, avatar_url
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8
		)
		ON CONFLICT (user_id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			bio          = EXCLUDED.bio,
			gender       = EXCLUDED.gender,
			birth_date   = EXCLUDED.birth_date,
			city         = EXCLUDED.city,
			looking_for  = EXCLUDED.looking_for,
			avatar_url   = COALESCE(EXCLUDED.avatar_url, social.profiles.avatar_url),
			updated_at   = NOW()
		RETURNING created_at, updated_at`

	err := runner(ctx, r.pool).QueryRow(ctx, query,
		p.UserID,
		p.DisplayName,
		nullableString(p.Bio),
		nullableGender(p.Gender),
		p.BirthDate,
		nullableString(p.City),
		nullableGender(p.LookingFor), // $7
		nullableString(p.AvatarURL),  // $8
	).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upserting profile: %w", err)
	}

	return nil
}

func (r *ProfileRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	query := `
		SELECT
			p.user_id, p.display_name, p.bio, p.gender, p.birth_date,
			p.city,
			p.looking_for, p.avatar_url, p.created_at, p.updated_at
		FROM social.profiles p
		WHERE p.user_id = $1`

	p := &domain.Profile{}
	var bio, city, avatarURL *string
	var gender, lookingFor *string

	err := runner(ctx, r.pool).QueryRow(ctx, query, userID).Scan(
		&p.UserID, &p.DisplayName, &bio, &gender, &p.BirthDate,
		&city,
		&lookingFor, &avatarURL, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("getting profile: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("getting profile: %w", err)
	}

	if bio != nil {
		p.Bio = *bio
	}
	if city != nil {
		p.City = *city
	}
	if avatarURL != nil {
		p.AvatarURL = *avatarURL
	}
	if gender != nil {
		p.Gender = domain.Gender(*gender)
	}
	if lookingFor != nil {
		p.LookingFor = domain.Gender(*lookingFor)
	}

	return p, nil
}

func (r *ProfileRepo) FindCandidates(ctx context.Context, opts repository.FindCandidatesOpts) ([]*repository.CandidateRow, error) {
	// Convert exclude list to a string slice for ANY($n).
	excludeStrings := make([]string, len(opts.ExcludeIDs))
	for i, id := range opts.ExcludeIDs {
		excludeStrings[i] = id.String()
	}

	query := `
		SELECT
			p.user_id,
			p.display_name,
			COALESCE(p.avatar_url, '')   AS avatar_url,
			COALESCE(p.city, '')          AS city,
			u.trust_score
		FROM social.profiles p
		JOIN social.users u ON u.id = p.user_id
		WHERE
			u.is_active   = true
			AND u.trust_status = 'normal'
			AND ($1::text = '' OR p.gender::text = $1)
			AND p.user_id != $2
			AND NOT (p.user_id::text = ANY($3))
			AND (
				p.birth_date IS NULL
				OR EXTRACT(year FROM AGE(p.birth_date)) BETWEEN $4 AND $5
			)
		ORDER BY u.trust_score DESC, u.last_login_at DESC NULLS LAST
		LIMIT $6`

	// ---------------------------------------------------------
	// 🔥 НАШ РАДАР ДЛЯ ОТЛОВА БАГОВ (ВЫВОД В КОНСОЛЬ БЭКЕНДА)
	// ---------------------------------------------------------
	fmt.Println("==================================================")
	fmt.Println("🚀 ВЫЗОВ ФУНКЦИИ FindCandidates")
	fmt.Printf("LookingFor($1): '%v'\n", string(opts.LookingFor))
	fmt.Printf("RequesterID($2): %v\n", opts.RequesterID)
	fmt.Printf("ExcludeIDs($3): %v\n", excludeStrings)
	fmt.Printf("AgeRangeMin($4): %v | AgeRangeMax($5): %v\n", opts.AgeRangeMin, opts.AgeRangeMax)
	fmt.Printf("Limit($6): %v\n", opts.Limit)
	fmt.Println("==================================================")
	// ---------------------------------------------------------

	rows, err := runner(ctx, r.pool).Query(ctx, query,
		string(opts.LookingFor), // $1  ('' means any gender)
		opts.RequesterID,        // $2
		excludeStrings,          // $3
		opts.AgeRangeMin,        // $4
		opts.AgeRangeMax,        // $5
		opts.Limit,              // $6
	)
	if err != nil {
		return nil, fmt.Errorf("finding candidates: %w", err)
	}
	defer rows.Close()

	var candidates []*repository.CandidateRow
	for rows.Next() {
		c := &repository.CandidateRow{}
		if err := rows.Scan(
			&c.UserID, &c.DisplayName, &c.AvatarURL,
			&c.City, &c.TrustScore,
		); err != nil {
			return nil, fmt.Errorf("scanning candidate: %w", err)
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating candidates: %w", err)
	}

	return candidates, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullableGender(g domain.Gender) *string {
	if g == "" {
		return nil
	}
	s := string(g)
	return &s
}

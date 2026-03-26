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
			city, location, looking_for, avatar_url
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, 
			CASE WHEN $7::numeric = 0 AND $8::numeric = 0 THEN NULL ELSE ST_SetSRID(ST_MakePoint($8, $7), 4326)::geography END, 
			$9, $10
		)
		ON CONFLICT (user_id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			bio          = EXCLUDED.bio,
			gender       = EXCLUDED.gender,
			birth_date   = EXCLUDED.birth_date,
			city         = EXCLUDED.city,
			location     = COALESCE(EXCLUDED.location, social.profiles.location),
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
		p.Latitude,  // $7
		p.Longitude, // $8
		nullableGender(p.LookingFor),
		nullableString(p.AvatarURL),
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
			ST_Y(p.location::geometry) AS latitude,
			ST_X(p.location::geometry) AS longitude,
			p.looking_for, p.avatar_url, p.created_at, p.updated_at
		FROM social.profiles p
		WHERE p.user_id = $1`

	p := &domain.Profile{}
	var bio, city, avatarURL *string
	var gender, lookingFor *string
	var lat, lon *float64

	err := runner(ctx, r.pool).QueryRow(ctx, query, userID).Scan(
		&p.UserID, &p.DisplayName, &bio, &gender, &p.BirthDate,
		&city, &lat, &lon,
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

	if lat != nil {
		p.Latitude = *lat
	}
	if lon != nil {
		p.Longitude = *lon
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
			u.trust_score,
			ROUND(
				(ST_Distance(
					p.location,
					ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography
				) / 1000.0)::numeric, 1
			) AS distance_km
		FROM social.profiles p
		JOIN social.users u ON u.id = p.user_id
		WHERE
			u.is_active   = true
			AND u.trust_status = 'normal'
			AND p.location IS NOT NULL
			AND ST_DWithin(
				p.location,
				ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography,
				$3
			)
			AND ($4::text = '' OR p.gender::text = $4)
			AND p.user_id != $5
			AND NOT (p.user_id::text = ANY($6))
			AND (
				p.birth_date IS NULL
				OR EXTRACT(year FROM AGE(p.birth_date)) BETWEEN $7 AND $8
			)
		ORDER BY u.trust_score DESC, u.last_login_at DESC NULLS LAST
		LIMIT $9`

	// ---------------------------------------------------------
	// 🔥 НАШ РАДАР ДЛЯ ОТЛОВА БАГОВ (ВЫВОД В КОНСОЛЬ БЭКЕНДА)
	// ---------------------------------------------------------
	fmt.Println("==================================================")
	fmt.Println("🚀 ВЫЗОВ ФУНКЦИИ FindCandidates")
	fmt.Printf("Lat($1): %v | Lon($2): %v\n", opts.Lat, opts.Lon)
	fmt.Printf("MaxDistMeters($3): %v\n", opts.MaxDistanceMeters)
	fmt.Printf("LookingFor($4): '%v'\n", string(opts.LookingFor))
	fmt.Printf("RequesterID($5): %v\n", opts.RequesterID)
	fmt.Printf("ExcludeIDs($6): %v\n", excludeStrings)
	fmt.Printf("AgeRangeMin($7): %v | AgeRangeMax($8): %v\n", opts.AgeRangeMin, opts.AgeRangeMax)
	fmt.Printf("Limit($9): %v\n", opts.Limit)
	fmt.Println("==================================================")
	// ---------------------------------------------------------

	rows, err := runner(ctx, r.pool).Query(ctx, query,
		opts.Lat,                // $1
		opts.Lon,                // $2
		opts.MaxDistanceMeters,  // $3
		string(opts.LookingFor), // $4  ('' means any gender)
		opts.RequesterID,        // $5
		excludeStrings,          // $6
		opts.AgeRangeMin,        // $7
		opts.AgeRangeMax,        // $8
		opts.Limit,              // $9
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
			&c.City, &c.TrustScore, &c.DistanceKm,
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

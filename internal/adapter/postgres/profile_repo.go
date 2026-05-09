package postgres

import (
	"context"
	"encoding/json"
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

func locationPoint(lat, lon *float64) *string {
	if lat == nil || lon == nil {
		return nil
	}
	s := fmt.Sprintf("POINT(%f %f)", *lon, *lat)
	return &s
}

func (r *ProfileRepo) Upsert(ctx context.Context, p *domain.Profile) error {
	promptsJSON, _ := json.Marshal(p.Prompts)
	if p.Prompts == nil {
		promptsJSON = []byte("[]")
	}

	query := `
		INSERT INTO social.profiles (
			user_id, display_name, bio, gender, birth_date,
			city, location, looking_for, avatar_url, prompts
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, ST_GeographyFromText($7), $8, $9, $10
		)
		ON CONFLICT (user_id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			bio          = EXCLUDED.bio,
			gender       = EXCLUDED.gender,
			birth_date   = EXCLUDED.birth_date,
			city         = EXCLUDED.city,
			location     = EXCLUDED.location,
			looking_for  = EXCLUDED.looking_for,
			avatar_url   = COALESCE(EXCLUDED.avatar_url, social.profiles.avatar_url),
			prompts      = EXCLUDED.prompts,
			updated_at   = NOW()
		RETURNING created_at, updated_at`

	err := runner(ctx, r.pool).QueryRow(ctx, query,
		p.UserID,
		p.DisplayName,
		nullableString(p.Bio),
		nullableGender(p.Gender),
		p.BirthDate,
		nullableString(p.City),
		locationPoint(p.Latitude, p.Longitude), // $7
		nullableGender(p.LookingFor),           // $8
		nullableString(p.AvatarURL),            // $9
		promptsJSON,                            // $10
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
			p.city, ST_X(p.location::geometry) AS lon, ST_Y(p.location::geometry) AS lat,
			p.looking_for, p.avatar_url, p.prompts, p.created_at, p.updated_at
		FROM social.profiles p
		WHERE p.user_id = $1`

	p := &domain.Profile{}
	var bio, city, avatarURL *string
	var gender, lookingFor *string
	var lon, lat *float64
	var promptsJSON []byte

	err := runner(ctx, r.pool).QueryRow(ctx, query, userID).Scan(
		&p.UserID, &p.DisplayName, &bio, &gender, &p.BirthDate,
		&city, &lon, &lat,
		&lookingFor, &avatarURL, &promptsJSON, &p.CreatedAt, &p.UpdatedAt,
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
	if len(promptsJSON) > 0 {
		var prompts []domain.PromptAnswer
		_ = json.Unmarshal(promptsJSON, &prompts)
		p.Prompts = prompts
	}
	p.Longitude = lon
	p.Latitude = lat

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
			p.prompts,
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
			AND (
				$7::boolean = false OR 
				(p.location IS NULL OR ST_DWithin(p.location, ST_SetSRID(ST_MakePoint($8, $9), 4326)::geography, $10))
			)
		ORDER BY u.trust_score DESC, u.last_login_at DESC NULLS LAST
		LIMIT $6`

	useSpatial := false
	lon := 0.0
	lat := 0.0
	maxDist := 0

	if opts.MaxDistanceMeters != nil && opts.RequesterLon != nil && opts.RequesterLat != nil {
		useSpatial = true
		lon = *opts.RequesterLon
		lat = *opts.RequesterLat
		maxDist = *opts.MaxDistanceMeters
	}

	// ---------------------------------------------------------
	// 🔥 НАШ РАДАР ДЛЯ ОТЛОВА БАГОВ (ВЫВОД В КОНСОЛЬ БЭКЕНДА)
	// ---------------------------------------------------------
	fmt.Println("==================================================")
	fmt.Println("🚀 ВЫЗОВ ФУНКЦИИ FindCandidates")
	fmt.Printf("LookingFor($1): '%v'\n", string(opts.LookingFor))
	fmt.Printf("RequesterID($2): %v\n", opts.RequesterID)
	fmt.Printf("AgeRangeMin($4): %v | AgeRangeMax($5): %v\n", opts.AgeRangeMin, opts.AgeRangeMax)
	fmt.Printf("Spatial($7): %v, Lon($8): %v, Lat($9): %v, MaxDist($10): %v\n", useSpatial, lon, lat, maxDist)
	fmt.Println("==================================================")
	// ---------------------------------------------------------

	rows, err := runner(ctx, r.pool).Query(ctx, query,
		string(opts.LookingFor), // $1  ('' means any gender)
		opts.RequesterID,        // $2
		excludeStrings,          // $3
		opts.AgeRangeMin,        // $4
		opts.AgeRangeMax,        // $5
		opts.Limit,              // $6
		useSpatial,              // $7
		lon,                     // $8
		lat,                     // $9
		maxDist,                 // $10
	)
	if err != nil {
		return nil, fmt.Errorf("finding candidates: %w", err)
	}
	defer rows.Close()

	var candidates []*repository.CandidateRow
	for rows.Next() {
		c := &repository.CandidateRow{}
		var promptsJSON []byte
		if err := rows.Scan(
			&c.UserID, &c.DisplayName, &c.AvatarURL,
			&c.City, &promptsJSON, &c.TrustScore,
		); err != nil {
			return nil, fmt.Errorf("scanning candidate: %w", err)
		}
		if len(promptsJSON) > 0 {
			var prompts []domain.PromptAnswer
			_ = json.Unmarshal(promptsJSON, &prompts)
			c.Prompts = prompts
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating candidates: %w", err)
	}

	return candidates, nil
}

// GetLeaderboard returns the top users ordered by trust score.
func (r *ProfileRepo) GetLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardEntry, error) {
	query := `
		SELECT
			p.user_id,
			p.display_name,
			COALESCE(p.avatar_url, '') AS avatar_url,
			u.trust_score
		FROM social.profiles p
		JOIN social.users u ON u.id = p.user_id
		WHERE u.is_active = true
		ORDER BY u.trust_score DESC, u.last_login_at DESC NULLS LAST
		LIMIT $1`

	rows, err := runner(ctx, r.pool).Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("getting leaderboard: %w", err)
	}
	defer rows.Close()

	var board []domain.LeaderboardEntry
	for rows.Next() {
		var entry domain.LeaderboardEntry
		var score int
		if err := rows.Scan(&entry.UserID, &entry.DisplayName, &entry.AvatarURL, &score); err != nil {
			return nil, fmt.Errorf("scanning leaderboard row: %w", err)
		}
		entry.Score = score

		// Use the Score to compute Badge
		tScore := &domain.TrustScore{Score: score}
		entry.Badge = tScore.GetBadge()

		board = append(board, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating leaderboard: %w", err)
	}

	return board, nil
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

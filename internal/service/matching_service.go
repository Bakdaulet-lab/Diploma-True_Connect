package service

import (
	"context"
	"fmt"
	"strings" // Добавлено
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

const (
	candidateBatchSize = 20
	seenSetTTL         = 24 * time.Hour
)

// fixAvatarURL преобразует путь из БД в прямую ссылку на MinIO
func fixAvatarURL(url string) string {
	if url == "" {
		return ""
	}
	if strings.HasPrefix(url, "http") {
		return strings.ReplaceAll(url, "minio:9000", "localhost:9000")
	}
	// Если в базе лежит просто путь "users/...", превращаем в URL
	return fmt.Sprintf("http://localhost:9000/trueconnect/%s", strings.TrimPrefix(url, "/"))
}

// CandidateView is a matching card shown in the swipe feed.
type CandidateView struct {
	UserID        uuid.UUID `json:"user_id"`
	DisplayName   string    `json:"display_name"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	AvatarBlurred bool      `json:"avatar_blurred,omitempty"`
	City          string    `json:"city,omitempty"`
	TrustScore    int       `json:"trust_score"`
	Niyyah        string    `json:"niyyah,omitempty"`
	Madhab        string    `json:"madhab,omitempty"`
	IsKYCVerified bool      `json:"is_kyc_verified"`
}

// MatchView is a match card with user profile data for the frontend.
// Новая структура для вложенного пользователя
type MatchUserView struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	TrustScore  int       `json:"trust_score"`
	PublicKey   *string   `json:"public_key,omitempty"`
}

// Обновленная основная структура мэтча
type MatchView struct {
	ID                uuid.UUID      `json:"id"`
	OtherUser         *MatchUserView `json:"other_user"`
	NiyyahTimerEndsAt *time.Time     `json:"niyyah_timer_ends_at,omitempty"`
}

// LikeResult tells the caller whether a mutual match occurred.
type LikeResult struct {
	Matched bool      `json:"matched"`
	MatchID uuid.UUID `json:"match_id,omitempty"`
}

// MatchingService handles candidate retrieval, likes, and passes.
type MatchingService struct {
	profileRepo    repository.ProfileRepository
	userRepo       repository.UserRepository
	matchRepo      repository.MatchRepository
	settingsRepo   repository.UserSettingsRepository
	matchingCache  repository.MatchingCache
	trustGraphRepo repository.TrustGraphRepository
	notifSvc       *NotificationService
}

// NewMatchingService creates a new matching service.
func NewMatchingService(
	profileRepo repository.ProfileRepository,
	userRepo repository.UserRepository,
	matchRepo repository.MatchRepository,
	settingsRepo repository.UserSettingsRepository,
	matchingCache repository.MatchingCache,
	trustGraphRepo repository.TrustGraphRepository,
	notifSvc *NotificationService,
) *MatchingService {
	return &MatchingService{
		profileRepo:    profileRepo,
		userRepo:       userRepo,
		matchRepo:      matchRepo,
		settingsRepo:   settingsRepo,
		matchingCache:  matchingCache,
		trustGraphRepo: trustGraphRepo,
		notifSvc:       notifSvc,
	}
}

// niyyahCompatible returns the niyyah values that are compatible with the requester's
// own niyyah. A nikah_year requester only wants to see serious candidates; others
// are open to the full spectrum.
func niyyahCompatible(n domain.Niyyah) []string {
	switch n {
	case domain.NiyyahNikahYear:
		return []string{string(domain.NiyyahNikahYear), string(domain.NiyyahSeriousMarriage)}
	default:
		return nil // no filter — allow all niyyah values including unset
	}
}

// GetCandidates returns a batch of profiles for the swipe feed.
func (s *MatchingService) GetCandidates(ctx context.Context, userID uuid.UUID) ([]*CandidateView, error) {
	requesterProfile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get candidates: loading requester profile: %w", err)
	}

	settings, err := s.settingsRepo.Get(ctx, userID)
	if err != nil {
		settings = domain.DefaultSettings(userID.String())
	}

	seenIDs, err := s.matchingCache.GetSeenIDs(ctx, userID)
	if err != nil {
		seenIDs = []uuid.UUID{}
	}

	matchedIDs := []uuid.UUID{}
	matches, _, err := s.matchRepo.ListMatches(ctx, userID, "", 200)
	if err == nil {
		matchedIDs = make([]uuid.UUID, 0, len(matches))
		for _, m := range matches {
			other := m.UserAID
			if other == userID {
				other = m.UserBID
			}
			matchedIDs = append(matchedIDs, other)
		}
	}

	// Exclude users the requester has previously rejected (persistent dislikes).
	rejectedIDs, err := s.matchRepo.GetRejectedIDs(ctx, userID)
	if err != nil {
		rejectedIDs = []uuid.UUID{}
	}

	// Exclude users the requester has blocked.
	blockedIDs, err := s.matchRepo.GetBlockedIDs(ctx, userID)
	if err != nil {
		blockedIDs = []uuid.UUID{}
	}

	maxDistMeters := settings.MaxDistanceKm * 1000

	// B1: compute niyyah compatibility filter from requester's own niyyah.
	allowedNiyyahs := niyyahCompatible(requesterProfile.Niyyah)

	excludeIDs := append([]uuid.UUID{}, seenIDs...)
	excludeIDs = append(excludeIDs, matchedIDs...)
	excludeIDs = append(excludeIDs, rejectedIDs...)
	excludeIDs = append(excludeIDs, blockedIDs...)

	// Auto-infer LookingFor from gender when the user hasn't set it explicitly.
	// Male users see only females and vice versa — this is the Islamic halal default.
	lookingFor := requesterProfile.LookingFor
	if lookingFor == "" {
		switch requesterProfile.Gender {
		case domain.GenderMale:
			lookingFor = domain.GenderFemale
		case domain.GenderFemale:
			lookingFor = domain.GenderMale
		}
	}

	opts := repository.FindCandidatesOpts{
		RequesterID:       userID,
		LookingFor:        lookingFor,
		AgeRangeMin:       settings.AgeRangeMin,
		AgeRangeMax:       settings.AgeRangeMax,
		MaxDistanceMeters: &maxDistMeters,
		RequesterLat:      requesterProfile.Latitude,
		RequesterLon:      requesterProfile.Longitude,
		ExcludeIDs:        excludeIDs,
		Limit:             candidateBatchSize,
		AllowedNiyyahs:    allowedNiyyahs,
		MadhabFilter:      settings.MadhabFilter,
	}

	rows, err := s.profileRepo.FindCandidates(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("get candidates: querying: %w", err)
	}
	if len(rows) == 0 && len(seenIDs) > 0 {
		opts.ExcludeIDs = matchedIDs
		rows, err = s.profileRepo.FindCandidates(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("get candidates: retrying without seen cache: %w", err)
		}
	}
	if len(rows) == 0 && opts.MaxDistanceMeters != nil {
		// Fallback: widen discovery beyond distance when local pool is empty.
		opts.MaxDistanceMeters = nil
		opts.RequesterLat = nil
		opts.RequesterLon = nil
		rows, err = s.profileRepo.FindCandidates(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("get candidates: retrying without distance filter: %w", err)
		}
	}

	requesterMadhab := string(requesterProfile.Madhab)

	newSeenIDs := make([]uuid.UUID, 0, len(rows))
	views := make([]*CandidateView, 0, len(rows))
	for _, row := range rows {
		newSeenIDs = append(newSeenIDs, row.UserID)

		score := row.TrustScore
		// B2: boost score by +10 when candidate shares the requester's madhab.
		if requesterMadhab != "" && requesterMadhab != string(domain.MadhabNone) && row.Madhab == requesterMadhab {
			score = min(score+10, 100)
		}

		avatarURL := fixAvatarURL(row.AvatarURL)
		blurred := false
		// B3: candidates with no-photo mode enabled never expose their avatar
		// in the swipe feed (only visible after a mutual match in the chat view).
		if row.NoPhotoMode {
			avatarURL = ""
			blurred = true
		}

		views = append(views, &CandidateView{
			UserID:        row.UserID,
			DisplayName:   row.DisplayName,
			AvatarURL:     avatarURL,
			AvatarBlurred: blurred,
			City:          row.City,
			TrustScore:    score,
			Niyyah:        row.Niyyah,
			Madhab:        row.Madhab,
			IsKYCVerified: row.IsKYCVerified,
		})
	}

	if len(newSeenIDs) > 0 {
		_ = s.matchingCache.AddSeen(ctx, userID, newSeenIDs, seenSetTTL)
	}

	return views, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Like records that userID likes targetID.
func (s *MatchingService) Like(ctx context.Context, userID, targetID uuid.UUID) (*LikeResult, error) {
	if userID == targetID {
		return nil, fmt.Errorf("like: %w", domain.ErrInvalidInput)
	}

	matched, matchID, err := s.matchRepo.RecordLike(ctx, userID, targetID)
	if err != nil {
		return nil, fmt.Errorf("like: %w", err)
	}

	_ = s.matchingCache.AddSeen(ctx, userID, []uuid.UUID{targetID}, seenSetTTL)

	if s.notifSvc != nil {
		if matched {
			// Mutual match — notify the target that they have a new match.
			_ = s.notifSvc.Create(ctx, &domain.Notification{
				UserID:   targetID,
				ActorID:  &userID,
				Type:     domain.NotificationTypeMatch,
				EntityID: &matchID,
			})
		} else {
			// One-sided like — notify the target that someone liked them.
			_ = s.notifSvc.Create(ctx, &domain.Notification{
				UserID:  targetID,
				ActorID: &userID,
				Type:    domain.NotificationTypeLike,
			})
		}
	}

	return &LikeResult{Matched: matched, MatchID: matchID}, nil
}

// Pass records that userID passes on targetID (persisted to DB + Redis cache).
// Also removes any one-sided like row so the target disappears from pending likes immediately.
func (s *MatchingService) Pass(ctx context.Context, userID, targetID uuid.UUID) error {
	if userID == targetID {
		return fmt.Errorf("pass: %w", domain.ErrInvalidInput)
	}

	if err := s.matchRepo.RecordPass(ctx, userID, targetID); err != nil {
		return fmt.Errorf("pass: %w", err)
	}
	// Remove the one-sided match row so the passer no longer appears in pending likes.
	// Best-effort — don't fail the pass if there was no match row.
	_ = s.matchRepo.UnmatchByUsers(ctx, userID, targetID)
	_ = s.matchingCache.AddSeen(ctx, userID, []uuid.UUID{targetID}, seenSetTTL)

	return nil
}

// BlockUser records that callerID is blocking targetID and removes any existing match.
func (s *MatchingService) BlockUser(ctx context.Context, callerID, targetID uuid.UUID) error {
	if callerID == targetID {
		return fmt.Errorf("block: %w", domain.ErrInvalidInput)
	}
	// Remove any existing match so the blocked user immediately disappears from
	// the matches tab. Best-effort — don't fail the block if there's no match.
	_ = s.matchRepo.UnmatchByUsers(ctx, callerID, targetID)
	return s.matchRepo.BlockUser(ctx, callerID, targetID)
}

// Unmatch removes a mutual match between two users.
func (s *MatchingService) Unmatch(ctx context.Context, matchID, callerID uuid.UUID) error {
	return s.matchRepo.Unmatch(ctx, matchID, callerID)
}

// Остальные методы (GetMatchByID, ListMatches) остаются без изменений
func (s *MatchingService) GetMatchByID(ctx context.Context, matchID, userID uuid.UUID) (*domain.Match, error) {
	match, err := s.matchRepo.GetMatch(ctx, matchID, userID)
	if err != nil {
		return nil, fmt.Errorf("get match: %w", err)
	}
	return match, nil
}

func (s *MatchingService) ListMatches(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]*MatchView, string, error) {
	// ListMatchViews joins users+profiles in a single query, avoiding N+1 lookups.
	rows, nextCursor, err := s.matchRepo.ListMatchViews(ctx, userID, cursor, limit)
	if err != nil {
		return nil, "", err
	}

	views := make([]*MatchView, 0, len(rows))
	for _, row := range rows {
		views = append(views, &MatchView{
			ID: row.MatchID,
			OtherUser: &MatchUserView{
				UserID:      row.OtherUserID,
				DisplayName: row.DisplayName,
				AvatarURL:   fixAvatarURL(row.AvatarURL),
				TrustScore:  row.TrustScore,
				PublicKey:   row.PublicKey,
			},
			NiyyahTimerEndsAt: row.NiyyahTimerEndsAt,
		})
	}
	return views, nextCursor, nil
}

// FamilyIntroductionDone marks a match as having completed the family introduction milestone.
// callerID must be one of the match participants.
func (s *MatchingService) FamilyIntroductionDone(ctx context.Context, matchID, callerID uuid.UUID) error {
	match, err := s.matchRepo.GetMatch(ctx, matchID, callerID)
	if err != nil {
		return fmt.Errorf("family intro: %w", err)
	}
	if match.MatchedAt == nil {
		return fmt.Errorf("family intro: match not finalized: %w", domain.ErrForbidden)
	}

	if err := s.matchRepo.MarkFamilyIntroDone(ctx, matchID); err != nil {
		return fmt.Errorf("family intro: %w", err)
	}

	// Notify both participants.
	otherID := match.UserAID
	if otherID == callerID {
		otherID = match.UserBID
	}
	matchIDCopy := matchID
	if s.notifSvc != nil {
		for _, uid := range []uuid.UUID{callerID, otherID} {
			_ = s.notifSvc.Create(ctx, &domain.Notification{
				UserID:   uid,
				Type:     domain.NotificationTypeSystem,
				EntityID: &matchIDCopy,
			})
		}
	}
	return nil
}

// GetPendingLikes returns the profiles of users who liked the caller but haven't been liked back.
func (s *MatchingService) GetPendingLikes(ctx context.Context, userID uuid.UUID) ([]*CandidateView, error) {
	likerIDs, err := s.matchRepo.GetPendingLikes(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get pending likes: %w", err)
	}

	views := make([]*CandidateView, 0, len(likerIDs))
	for _, id := range likerIDs {
		prof, err := s.profileRepo.GetByUserID(ctx, id)
		if err != nil {
			continue
		}
		user, err := s.userRepo.GetByID(ctx, id)
		if err != nil {
			continue
		}
		avatarURL := fixAvatarURL(prof.AvatarURL)
		if prof.NoPhotoMode {
			avatarURL = ""
		}
		views = append(views, &CandidateView{
			UserID:        id,
			DisplayName:   prof.DisplayName,
			AvatarURL:     avatarURL,
			AvatarBlurred: prof.NoPhotoMode,
			City:          prof.City,
			TrustScore:    user.TrustScore,
			Niyyah:        string(prof.Niyyah),
			Madhab:        string(prof.Madhab),
			IsKYCVerified: user.VerificationLevel == domain.VerificationIDVerified ||
				user.VerificationLevel == domain.VerificationPhotoVerified,
		})
	}
	return views, nil
}

// GetGraphCandidates returns a batch of profiles from Neo4j (friends-of-friends).
func (s *MatchingService) GetGraphCandidates(ctx context.Context, userID uuid.UUID) ([]*CandidateView, error) {
	recIDs, err := s.trustGraphRepo.GetRecommendations(ctx, userID, candidateBatchSize)
	if err != nil {
		return nil, fmt.Errorf("get graph candidates: querying neo4j: %w", err)
	}

	var views []*CandidateView
	for _, targetID := range recIDs {
		prof, err := s.profileRepo.GetByUserID(ctx, targetID)
		if err != nil {
			continue
		}
		user, err := s.userRepo.GetByID(ctx, targetID)
		if err != nil {
			continue
		}
		views = append(views, &CandidateView{
			UserID:      prof.UserID,
			DisplayName: prof.DisplayName,
			AvatarURL:   fixAvatarURL(prof.AvatarURL),
			City:        prof.City,
			TrustScore:  user.TrustScore,
		})
	}

	if len(views) == 0 {
		return []*CandidateView{}, nil
	}
	return views, nil
}

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
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	City        string    `json:"city,omitempty"`
	TrustScore  int       `json:"trust_score"`
}

// MatchView is a match card with user profile data for the frontend.
// Новая структура для вложенного пользователя
type MatchUserView struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	PublicKey   *string   `json:"public_key,omitempty"`
}

// Обновленная основная структура мэтча
type MatchView struct {
	ID        uuid.UUID      `json:"id"`
	OtherUser *MatchUserView `json:"other_user"` // Тот самый вложенный объект!
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

	matches, _, err := s.matchRepo.ListMatches(ctx, userID, "", 200)
	if err == nil {
		for _, m := range matches {
			other := m.UserAID
			if other == userID {
				other = m.UserBID
			}
			seenIDs = append(seenIDs, other)
		}
	}

	maxDistMeters := settings.MaxDistanceKm * 1000
	opts := repository.FindCandidatesOpts{
		RequesterID:       userID,
		LookingFor:        requesterProfile.LookingFor,
		AgeRangeMin:       settings.AgeRangeMin,
		AgeRangeMax:       settings.AgeRangeMax,
		MaxDistanceMeters: &maxDistMeters,
		RequesterLat:      requesterProfile.Latitude,
		RequesterLon:      requesterProfile.Longitude,
		ExcludeIDs:        seenIDs,
		Limit:             candidateBatchSize,
	}

	rows, err := s.profileRepo.FindCandidates(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("get candidates: querying: %w", err)
	}

	newSeenIDs := make([]uuid.UUID, 0, len(rows))
	views := make([]*CandidateView, 0, len(rows))
	for _, row := range rows {
		newSeenIDs = append(newSeenIDs, row.UserID)
		views = append(views, &CandidateView{
			UserID:      row.UserID,
			DisplayName: row.DisplayName,
			AvatarURL:   fixAvatarURL(row.AvatarURL), // ПРИМЕНЯЕМ ФИКС ТУТ
			City:        row.City,
			TrustScore:  row.TrustScore,
		})
	}

	if len(newSeenIDs) > 0 {
		_ = s.matchingCache.AddSeen(ctx, userID, newSeenIDs, seenSetTTL)
	}

	return views, nil
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

	if matched {
		// New match! Notify the target user
		_ = s.notifSvc.Create(ctx, &domain.Notification{
			UserID:   targetID,
			ActorID:  &userID,
			Type:     domain.NotificationTypeMatch,
			EntityID: &matchID,
		})
		// We could optionally notify the current user too, but usually the current user knows since they just swiped "Like"
	}

	return &LikeResult{Matched: matched, MatchID: matchID}, nil
}

// Pass records that userID passes on targetID.
func (s *MatchingService) Pass(ctx context.Context, userID, targetID uuid.UUID) error {
	if userID == targetID {
		return fmt.Errorf("pass: %w", domain.ErrInvalidInput)
	}

	_ = s.matchRepo.RecordPass(ctx, userID, targetID)
	_ = s.matchingCache.AddSeen(ctx, userID, []uuid.UUID{targetID}, seenSetTTL)

	return nil
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
	matches, nextCursor, err := s.matchRepo.ListMatches(ctx, userID, cursor, limit)
	if err != nil {
		return nil, "", err
	}

	views := make([]*MatchView, 0, len(matches))
	for _, m := range matches {
		otherID := m.UserAID
		if otherID == userID {
			otherID = m.UserBID
		}

		profile, err := s.profileRepo.GetByUserID(ctx, otherID)
		if err != nil {
			continue
		}

		user, err := s.userRepo.GetByID(ctx, otherID)
		if err != nil {
			continue
		}

		// Формируем вложенную структуру, которую ждет фронтенд
		views = append(views, &MatchView{
			ID: m.ID,
			OtherUser: &MatchUserView{
				UserID:      otherID,
				DisplayName: profile.DisplayName,
				AvatarURL:   fixAvatarURL(profile.AvatarURL),
				PublicKey:   user.PublicKey,
			},
		})
	}
	return views, nextCursor, nil
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
		views = append(views, &CandidateView{
			UserID:      prof.UserID,
			DisplayName: prof.DisplayName,
			AvatarURL:   fixAvatarURL(prof.AvatarURL),
			City:        prof.City,
			TrustScore:  50, // Default for now
		})
	}

	if len(views) == 0 {
		return []*CandidateView{}, nil
	}
	return views, nil
}

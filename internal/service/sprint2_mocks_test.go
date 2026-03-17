package service_test

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// ── mockProfileRepo ───────────────────────────────────────────────────────────

type mockProfileRepo struct {
	mu       sync.Mutex
	profiles map[uuid.UUID]*domain.Profile
}

func newMockProfileRepo() *mockProfileRepo {
	return &mockProfileRepo{profiles: make(map[uuid.UUID]*domain.Profile)}
}

func (m *mockProfileRepo) Upsert(_ context.Context, p *domain.Profile) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *p
	m.profiles[p.UserID] = &cp
	return nil
}

func (m *mockProfileRepo) GetByUserID(_ context.Context, id uuid.UUID) (*domain.Profile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.profiles[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *p
	return &cp, nil
}

func (m *mockProfileRepo) FindCandidates(_ context.Context, opts repository.FindCandidatesOpts) ([]*repository.CandidateRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var rows []*repository.CandidateRow
	for _, p := range m.profiles {
		if p.UserID == opts.RequesterID {
			continue
		}
		for _, ex := range opts.ExcludeIDs {
			if p.UserID == ex {
				goto skip
			}
		}
		rows = append(rows, &repository.CandidateRow{
			UserID:      p.UserID,
			DisplayName: p.DisplayName,
			TrustScore:  50,
			DistanceKm:  1.0,
		})
		if len(rows) >= opts.Limit {
			break
		}
	skip:
	}
	return rows, nil
}

// ── mockMediaRepo ─────────────────────────────────────────────────────────────

type mockMediaRepo struct {
	mu     sync.Mutex
	media  map[uuid.UUID]*domain.Media // keyed by media ID
	avatar map[uuid.UUID]string        // keyed by userID
}

func newMockMediaRepo() *mockMediaRepo {
	return &mockMediaRepo{
		media:  make(map[uuid.UUID]*domain.Media),
		avatar: make(map[uuid.UUID]string),
	}
}

func (m *mockMediaRepo) Create(_ context.Context, med *domain.Media) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *med
	m.media[med.ID] = &cp
	return nil
}

func (m *mockMediaRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]*domain.Media, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []*domain.Media
	for _, med := range m.media {
		if med.UserID == userID {
			cp := *med
			list = append(list, &cp)
		}
	}
	return list, nil
}

func (m *mockMediaRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Media, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	med, ok := m.media[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *med
	return &cp, nil
}

func (m *mockMediaRepo) Delete(_ context.Context, id uuid.UUID, ownerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	med, ok := m.media[id]
	if !ok {
		return domain.ErrNotFound
	}
	if med.UserID != ownerID {
		return domain.ErrForbidden
	}
	delete(m.media, id)
	return nil
}

func (m *mockMediaRepo) CountByUser(_ context.Context, userID uuid.UUID) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, med := range m.media {
		if med.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *mockMediaRepo) UpdateAvatar(_ context.Context, userID uuid.UUID, objectKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.avatar[userID] = objectKey
	return nil
}

// ── mockMatchRepo ─────────────────────────────────────────────────────────────

type likeEntry struct {
	userID   uuid.UUID
	targetID uuid.UUID
}

type mockMatchRepo struct {
	mu      sync.Mutex
	likes   []likeEntry
	matches map[uuid.UUID]*domain.Match
}

func newMockMatchRepo() *mockMatchRepo {
	return &mockMatchRepo{matches: make(map[uuid.UUID]*domain.Match)}
}

func (m *mockMatchRepo) RecordLike(_ context.Context, userID, targetID uuid.UUID) (bool, uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if targetID already liked userID (mutual match).
	for _, l := range m.likes {
		if l.userID == targetID && l.targetID == userID {
			matchID := uuid.New()
			now := time.Now()
			m.matches[matchID] = &domain.Match{
				ID:        matchID,
				UserAID:   userID,
				UserBID:   targetID,
				MatchedAt: &now,
			}
			return true, matchID, nil
		}
	}

	m.likes = append(m.likes, likeEntry{userID: userID, targetID: targetID})
	return false, uuid.Nil, nil
}

func (m *mockMatchRepo) RecordPass(_ context.Context, _, _ uuid.UUID) error { return nil }

func (m *mockMatchRepo) ListMatches(_ context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Match, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []*domain.Match
	for _, match := range m.matches {
		if match.UserAID == userID || match.UserBID == userID {
			cp := *match
			list = append(list, &cp)
		}
	}
	start := offset
	if start >= len(list) {
		return []*domain.Match{}, nil
	}
	end := start + limit
	if end > len(list) {
		end = len(list)
	}
	return list[start:end], nil
}

func (m *mockMatchRepo) GetMatch(_ context.Context, matchID uuid.UUID, userID uuid.UUID) (*domain.Match, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	match, ok := m.matches[matchID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if match.UserAID != userID && match.UserBID != userID {
		return nil, domain.ErrForbidden
	}
	cp := *match
	return &cp, nil
}

func (m *mockMatchRepo) IsMatched(_ context.Context, userA, userB uuid.UUID) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, match := range m.matches {
		if (match.UserAID == userA && match.UserBID == userB) ||
			(match.UserAID == userB && match.UserBID == userA) {
			return true, nil
		}
	}
	return false, nil
}

// ── mockSettingsRepo ──────────────────────────────────────────────────────────

type mockSettingsRepo struct {
	mu       sync.Mutex
	settings map[uuid.UUID]*domain.UserSettings
}

func newMockSettingsRepo() *mockSettingsRepo {
	return &mockSettingsRepo{settings: make(map[uuid.UUID]*domain.UserSettings)}
}

func (m *mockSettingsRepo) Get(_ context.Context, userID uuid.UUID) (*domain.UserSettings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.settings[userID]
	if !ok {
		return domain.DefaultSettings(userID.String()), nil
	}
	cp := *s
	return &cp, nil
}

func (m *mockSettingsRepo) Upsert(_ context.Context, s *domain.UserSettings) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	userID, _ := uuid.Parse(s.UserID)
	cp := *s
	m.settings[userID] = &cp
	return nil
}

// ── mockMatchingCache ─────────────────────────────────────────────────────────

type mockMatchingCache struct {
	mu          sync.Mutex
	seen        map[uuid.UUID]map[uuid.UUID]struct{}
	trustScores map[uuid.UUID]int
}

func newMockMatchingCache() *mockMatchingCache {
	return &mockMatchingCache{
		seen:        make(map[uuid.UUID]map[uuid.UUID]struct{}),
		trustScores: make(map[uuid.UUID]int),
	}
}

func (m *mockMatchingCache) AddSeen(_ context.Context, requesterID uuid.UUID, seenIDs []uuid.UUID, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.seen[requesterID] == nil {
		m.seen[requesterID] = make(map[uuid.UUID]struct{})
	}
	for _, id := range seenIDs {
		m.seen[requesterID][id] = struct{}{}
	}
	return nil
}

func (m *mockMatchingCache) GetSeenIDs(_ context.Context, requesterID uuid.UUID) ([]uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	set := m.seen[requesterID]
	ids := make([]uuid.UUID, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	return ids, nil
}

func (m *mockMatchingCache) CacheTrustScore(_ context.Context, userID uuid.UUID, score int, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.trustScores[userID] = score
	return nil
}

func (m *mockMatchingCache) GetCachedTrustScore(_ context.Context, userID uuid.UUID) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	score, ok := m.trustScores[userID]
	if !ok {
		return -1, nil
	}
	return score, nil
}

// ── mockMediaStore ────────────────────────────────────────────────────────────

type mockMediaStore struct {
	mu      sync.Mutex
	objects map[string][]byte // objectKey → data
}

func newMockMediaStore() *mockMediaStore {
	return &mockMediaStore{objects: make(map[string][]byte)}
}

func (m *mockMediaStore) UploadPhoto(_ context.Context, userID uuid.UUID, data []byte) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := "users/" + userID.String() + "/photos/" + uuid.New().String() + ".jpg"
	cp := make([]byte, len(data))
	copy(cp, data)
	m.objects[key] = cp
	return key, nil
}

func (m *mockMediaStore) PresignedURL(_ context.Context, objectKey string) (string, error) {
	return "https://example.com/" + objectKey + "?sig=mock", nil
}

func (m *mockMediaStore) DeletePhoto(_ context.Context, objectKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objects, objectKey)
	return nil
}

func (m *mockMediaStore) UploadDocument(_ context.Context, userID uuid.UUID, data []byte, contentType string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := "kyc/" + userID.String() + "/" + uuid.New().String()
	cp := make([]byte, len(data))
	copy(cp, data)
	m.objects[key] = cp
	return key, nil
}

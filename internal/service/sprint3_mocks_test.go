package service_test

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// ── mockInteractionRepo ──────────────────────────────────────────────────────

type mockInteractionRepo struct {
	mu           sync.Mutex
	interactions map[uuid.UUID]*domain.Interaction
}

func newMockInteractionRepo() *mockInteractionRepo {
	return &mockInteractionRepo{interactions: make(map[uuid.UUID]*domain.Interaction)}
}

func (m *mockInteractionRepo) Create(_ context.Context, interaction *domain.Interaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	interaction.ID = uuid.New()
	interaction.CreatedAt = time.Now()
	cp := *interaction
	m.interactions[cp.ID] = &cp
	return nil
}

func (m *mockInteractionRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Interaction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, ok := m.interactions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *i
	return &cp, nil
}

func (m *mockInteractionRepo) ConfirmInteraction(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, ok := m.interactions[id]
	if !ok {
		return domain.ErrNotFound
	}
	i.IsVerified = true
	return nil
}

func (m *mockInteractionRepo) GetByRatedUser(_ context.Context, ratedID uuid.UUID, limit, offset int) ([]domain.Interaction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []domain.Interaction
	for _, i := range m.interactions {
		if i.RatedID == ratedID {
			list = append(list, *i)
		}
	}
	start := offset
	if start >= len(list) {
		return []domain.Interaction{}, nil
	}
	end := start + limit
	if end > len(list) {
		end = len(list)
	}
	return list[start:end], nil
}

func (m *mockInteractionRepo) ExistsBetweenUsersAfter(_ context.Context, raterID, ratedID uuid.UUID, after string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	afterTime, _ := time.Parse(time.RFC3339, after)
	for _, i := range m.interactions {
		if i.RaterID == raterID && i.RatedID == ratedID && i.CreatedAt.After(afterTime) {
			return true, nil
		}
	}
	return false, nil
}

// ── trackingGraphRepo ────────────────────────────────────────────────────────
// Extended graph repo mock that tracks all edges and supports configurable scores.

type trackingGraphRepo struct {
	mu            sync.Mutex
	nodes         map[uuid.UUID]bool
	ratings       []graphEdge
	meetings      []graphEdge
	reports       []graphEdge
	scoreToReturn int
	breakdown     *domain.TrustScoreBreakdown
	clusters      []repository.SybilCluster
}

type graphEdge struct {
	from     uuid.UUID
	to       uuid.UUID
	score    int
	context  string
	verified bool
}

func newTrackingGraphRepo() *trackingGraphRepo {
	return &trackingGraphRepo{
		nodes:         make(map[uuid.UUID]bool),
		scoreToReturn: 50,
	}
}

func (m *trackingGraphRepo) CreateUserNode(_ context.Context, uid uuid.UUID, _ domain.VerificationLevel) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nodes[uid] = true
	return nil
}

func (m *trackingGraphRepo) AddRating(_ context.Context, raterUID, ratedUID uuid.UUID, score int, interactionContext string, verified bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Upsert by (from, to) pair — mirrors Neo4j MERGE semantics.
	// ON MATCH: update only verified; score/context kept from first write.
	for i := range m.ratings {
		if m.ratings[i].from == raterUID && m.ratings[i].to == ratedUID {
			m.ratings[i].verified = verified
			return nil
		}
	}
	m.ratings = append(m.ratings, graphEdge{
		from: raterUID, to: ratedUID,
		score: score, context: interactionContext, verified: verified,
	})
	return nil
}

func (m *trackingGraphRepo) AddMeeting(_ context.Context, uidA, uidB uuid.UUID, _ bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.meetings = append(m.meetings, graphEdge{from: uidA, to: uidB})
	return nil
}

func (m *trackingGraphRepo) AddReport(_ context.Context, reporterUID, reportedUID uuid.UUID, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reports = append(m.reports, graphEdge{from: reporterUID, to: reportedUID, context: reason})
	return nil
}

func (m *trackingGraphRepo) ComputeTrustScore(_ context.Context, _ uuid.UUID) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.scoreToReturn, nil
}

func (m *trackingGraphRepo) ComputeTrustScoreBreakdown(_ context.Context, uid uuid.UUID) (*domain.TrustScoreBreakdown, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.breakdown != nil {
		cp := *m.breakdown
		cp.UserID = uid
		return &cp, nil
	}
	return &domain.TrustScoreBreakdown{
		UserID:         uid,
		Score:          m.scoreToReturn,
		SmoothedRating: 2.5,
		BaseScore:      50.0,
		RawScore:       float64(m.scoreToReturn),
	}, nil
}

func (m *trackingGraphRepo) UpdateTrustScore(_ context.Context, uid uuid.UUID, score int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

func (m *trackingGraphRepo) DetectSybilClusters(_ context.Context) ([]repository.SybilCluster, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.clusters, nil
}

// ИСПРАВЛЕНИЕ: Добавлен метод DeleteUserNode для соответствия новому интерфейсу TrustGraphRepository
func (m *trackingGraphRepo) DeleteUserNode(ctx context.Context, uid uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.nodes, uid)
	return nil
}

// ── mockUserRepoS3 ──────────────────────────────────────────────────────────
// Minimal user repo mock for Sprint 3 (trust score/status updates).

type mockUserRepoS3 struct {
	mu          sync.Mutex
	users       map[uuid.UUID]*domain.User
	trustScores map[uuid.UUID]int
	trustStatus map[uuid.UUID]domain.TrustStatus
}

func newMockUserRepoS3() *mockUserRepoS3 {
	return &mockUserRepoS3{
		users:       make(map[uuid.UUID]*domain.User),
		trustScores: make(map[uuid.UUID]int),
		trustStatus: make(map[uuid.UUID]domain.TrustStatus),
	}
}

func (m *mockUserRepoS3) UpdateFCMToken(ctx context.Context, id uuid.UUID, token string) error {
	return nil
}

func (m *mockUserRepoS3) Create(_ context.Context, user *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *user
	m.users[user.ID] = &cp
	return nil
}

func (m *mockUserRepoS3) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *mockUserRepoS3) GetByPhoneHash(_ context.Context, _ []byte) (*domain.User, error) {
	return nil, domain.ErrNotFound
}

func (m *mockUserRepoS3) UpdateVerificationLevel(_ context.Context, _ uuid.UUID, _ domain.VerificationLevel) error {
	return nil
}

func (m *mockUserRepoS3) UpdateTrustScore(_ context.Context, id uuid.UUID, score int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.trustScores[id] = score
	return nil
}

func (m *mockUserRepoS3) UpdateTrustStatus(_ context.Context, id uuid.UUID, status domain.TrustStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.trustStatus[id] = status
	return nil
}

func (m *mockUserRepoS3) UpdateLastLogin(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockUserRepoS3) SoftDelete(_ context.Context, _ uuid.UUID) error {
	return nil
}

// ИСПРАВЛЕНИЕ: Добавлен метод ListByTrustStatus для соответствия интерфейсу UserRepository
func (m *mockUserRepoS3) ListByTrustStatus(ctx context.Context, status domain.TrustStatus, limit, offset int) ([]*domain.User, error) {
	return []*domain.User{}, nil
}

func (m *trackingGraphRepo) GetRecommendations(ctx context.Context, uid uuid.UUID, limit int) ([]uuid.UUID, error) {
	return nil, nil
}
func (m *mockUserRepoS3) SubmitKYCRequest(ctx context.Context, userID uuid.UUID, documentURL string) error { return nil }

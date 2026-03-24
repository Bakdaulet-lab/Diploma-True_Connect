package service_test

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// ── mockPostRepo ─────────────────────────────────────────────────────────────

type mockPostRepo struct {
	mu       sync.Mutex
	posts    map[uuid.UUID]*domain.Post
	likes    map[uuid.UUID]map[uuid.UUID]struct{} // postID → userIDs
	comments map[uuid.UUID][]domain.PostComment   // postID → comments
}

func newMockPostRepo() *mockPostRepo {
	return &mockPostRepo{
		posts:    make(map[uuid.UUID]*domain.Post),
		likes:    make(map[uuid.UUID]map[uuid.UUID]struct{}),
		comments: make(map[uuid.UUID][]domain.PostComment),
	}
}

func (m *mockPostRepo) Create(_ context.Context, post *domain.Post) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	post.ID = uuid.New()
	post.CreatedAt = time.Now()
	post.UpdatedAt = post.CreatedAt
	cp := *post
	m.posts[cp.ID] = &cp
	return nil
}

func (m *mockPostRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Post, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.posts[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *p
	return &cp, nil
}

func (m *mockPostRepo) Delete(_ context.Context, id uuid.UUID, authorID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.posts[id]
	if !ok {
		return domain.ErrNotFound
	}
	if p.AuthorID != authorID {
		return domain.ErrNotFound
	}
	delete(m.posts, id)
	return nil
}

func (m *mockPostRepo) ListFeed(_ context.Context, cursor string, limit int) ([]domain.Post, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []domain.Post
	for _, p := range m.posts {
		all = append(all, *p)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.After(all[j].CreatedAt) })

	// Very simple mock cursor processing
	start := 0
	if cursor != "" {
		for i, p := range all {
			if p.CreatedAt.Format("2006-01-02T15:04:05.999999Z07:00") == cursor {
				start = i + 1
				break
			}
		}
	}
	if start >= len(all) {
		return []domain.Post{}, "", nil
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	nextCursor := ""
	if end == start+limit {
		nextCursor = all[end-1].CreatedAt.Format("2006-01-02T15:04:05.999999Z07:00")
	}
	return all[start:end], nextCursor, nil
}

func (m *mockPostRepo) ListByAuthor(_ context.Context, authorID uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []domain.Post
	for _, p := range m.posts {
		if p.AuthorID == authorID {
			all = append(all, *p)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.After(all[j].CreatedAt) })

	start := 0
	if cursor != "" {
		for i, p := range all {
			if p.CreatedAt.Format("2006-01-02T15:04:05.999999Z07:00") == cursor {
				start = i + 1
				break
			}
		}
	}
	if start >= len(all) {
		return []domain.Post{}, "", nil
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	nextCursor := ""
	if end == start+limit {
		nextCursor = all[end-1].CreatedAt.Format("2006-01-02T15:04:05.999999Z07:00")
	}
	return all[start:end], nextCursor, nil
}

func (m *mockPostRepo) IncrementLikeCount(_ context.Context, id uuid.UUID, delta int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.posts[id]
	if !ok {
		return domain.ErrNotFound
	}
	p.LikeCount += delta
	return nil
}

func (m *mockPostRepo) IncrementCommentCount(_ context.Context, id uuid.UUID, delta int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.posts[id]
	if !ok {
		return domain.ErrNotFound
	}
	p.CommentCount += delta
	return nil
}

func (m *mockPostRepo) LikePost(_ context.Context, postID, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.posts[postID]; !ok {
		return domain.ErrNotFound
	}
	if m.likes[postID] == nil {
		m.likes[postID] = make(map[uuid.UUID]struct{})
	}
	if _, alreadyLiked := m.likes[postID][userID]; alreadyLiked {
		return nil // idempotent
	}
	m.likes[postID][userID] = struct{}{}
	m.posts[postID].LikeCount++
	return nil
}

func (m *mockPostRepo) UnlikePost(_ context.Context, postID, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.posts[postID]; !ok {
		return domain.ErrNotFound
	}
	if m.likes[postID] == nil {
		return nil
	}
	if _, liked := m.likes[postID][userID]; !liked {
		return nil
	}
	delete(m.likes[postID], userID)
	m.posts[postID].LikeCount--
	return nil
}

func (m *mockPostRepo) IsLikedBy(_ context.Context, postID, userID uuid.UUID) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.likes[postID] == nil {
		return false, nil
	}
	_, liked := m.likes[postID][userID]
	return liked, nil
}

func (m *mockPostRepo) CreateComment(_ context.Context, comment *domain.PostComment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.posts[comment.PostID]; !ok {
		return domain.ErrNotFound
	}
	comment.ID = uuid.New()
	comment.CreatedAt = time.Now()
	m.comments[comment.PostID] = append(m.comments[comment.PostID], *comment)
	m.posts[comment.PostID].CommentCount++
	return nil
}

func (m *mockPostRepo) ListComments(_ context.Context, postID uuid.UUID, cursor string, limit int) ([]domain.PostComment, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	all := m.comments[postID]

	start := 0
	if cursor != "" {
		for i, c := range all {
			if c.CreatedAt.Format("2006-01-02T15:04:05.999999Z07:00") == cursor {
				start = i + 1
				break
			}
		}
	}
	if start >= len(all) {
		return []domain.PostComment{}, "", nil
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	nextCursor := ""
	if end == start+limit {
		nextCursor = all[end-1].CreatedAt.Format("2006-01-02T15:04:05.999999Z07:00")
	}
	return all[start:end], nextCursor, nil
}

// ── mockMessageRepo ──────────────────────────────────────────────────────────

type mockMessageRepo struct {
	mu       sync.Mutex
	messages []domain.Message
	readOps  int
}

func newMockMessageRepo() *mockMessageRepo {
	return &mockMessageRepo{}
}

func (m *mockMessageRepo) Create(_ context.Context, msg *domain.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	msg.ID = uuid.New()
	msg.CreatedAt = time.Now()
	m.messages = append(m.messages, *msg)
	return nil
}

func (m *mockMessageRepo) ListByMatch(_ context.Context, matchID uuid.UUID, before string, limit int) ([]domain.Message, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var filtered []domain.Message
	for _, msg := range m.messages {
		if msg.MatchID == matchID {
			filtered = append(filtered, msg)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].CreatedAt.Before(filtered[j].CreatedAt) })

	start := 0
	if before != "" {
		for i, msg := range filtered {
			if msg.CreatedAt.Format(time.RFC3339Nano) == before {
				start = i + 1
				break
			}
		}
	}

	if start >= len(filtered) {
		return []domain.Message{}, "", nil
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	nextCursor := ""
	if end == start+limit {
		nextCursor = filtered[end-1].CreatedAt.Format(time.RFC3339Nano)
	}

	return filtered[start:end], nextCursor, nil
}

// ИСПРАВЛЕНО: Метод снова называется MarkRead (а не MarkReadByMatch)
func (m *mockMessageRepo) MarkRead(_ context.Context, matchID uuid.UUID, readerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	m.readOps++
	for i := range m.messages {
		if m.messages[i].MatchID == matchID && m.messages[i].SenderID != readerID && m.messages[i].ReadAt == nil {
			m.messages[i].ReadAt = &now
		}
	}
	return nil
}

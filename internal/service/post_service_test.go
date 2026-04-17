package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/service"
	"github.com/trueconnect/backend/internal/domain"
)

// dummyMediaStore — простая заглушка хранилища для тестов
type dummyMediaStore struct{}

func (m *dummyMediaStore) UploadPhoto(ctx context.Context, userID uuid.UUID, data []byte) (string, error) {
	return "mock-media-key", nil
}
func (m *dummyMediaStore) DeletePhoto(ctx context.Context, objectKey string) error {
	return nil
}
func (m *dummyMediaStore) PresignedURL(ctx context.Context, objectKey string) (string, error) {
	if objectKey == "" {
		return "", nil
	}
	return "http://mock-url/" + objectKey, nil
}

// ИСПРАВЛЕНО: Добавлен недостающий параметр ext string
func (m *dummyMediaStore) UploadDocument(ctx context.Context, userID uuid.UUID, data []byte, ext string) (string, error) {
	return "mock-doc-key", nil
}

func TestCreatePost_Success(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	post, err := svc.CreatePost(context.Background(), uuid.New(), "Hello world", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post.Content != "Hello world" {
		t.Errorf("expected content %q, got %q", "Hello world", post.Content)
	}
	if post.ID == uuid.Nil {
		t.Error("expected non-nil post ID")
	}
}

func TestCreatePost_EmptyContent(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	_, err := svc.CreatePost(context.Background(), uuid.New(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestCreatePost_TooLong(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	long := strings.Repeat("a", 2001)
	_, err := svc.CreatePost(context.Background(), uuid.New(), long, nil)
	if err == nil {
		t.Fatal("expected error for content too long")
	}
}

func TestCreatePost_StripsHTML(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	post, err := svc.CreatePost(context.Background(), uuid.New(), "<b>hello</b>", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post.Content != "hello" {
		t.Errorf("expected stripped content %q, got %q", "hello", post.Content)
	}
}

func TestGetPost_Success(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	created, _ := svc.CreatePost(context.Background(), uuid.New(), "test post", nil)
	got, err := svc.GetPost(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected post ID %s, got %s", created.ID, got.ID)
	}
}

func TestGetPost_NotFound(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	_, err := svc.GetPost(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for non-existent post")
	}
}

func TestDeletePost_Success(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	authorID := uuid.New()
	post, _ := svc.CreatePost(context.Background(), authorID, "delete me", nil)

	if err := svc.DeletePost(context.Background(), post.ID, authorID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify gone.
	_, err := svc.GetPost(context.Background(), post.ID)
	if err == nil {
		t.Error("expected post to be deleted")
	}
}

func TestDeletePost_NotOwner(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	post, _ := svc.CreatePost(context.Background(), uuid.New(), "not yours", nil)

	err := svc.DeletePost(context.Background(), post.ID, uuid.New())
	if err == nil {
		t.Fatal("expected error for non-owner delete")
	}
}

func TestListFeed_Paginated(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	authorID := uuid.New()
	for i := 0; i < 5; i++ {
		svc.CreatePost(context.Background(), authorID, "post "+string(rune('A'+i)), nil)
	}

	posts, _, err := svc.ListFeed(context.Background(), "", 3, domain.PostFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 3 {
		t.Errorf("expected 3 posts, got %d", len(posts))
	}
}

func TestLikePost_Success(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	post, _ := svc.CreatePost(context.Background(), uuid.New(), "like me", nil)

	if err := svc.LikePost(context.Background(), post.ID, uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := svc.GetPost(context.Background(), post.ID)
	if updated.LikeCount != 1 {
		t.Errorf("expected like count 1, got %d", updated.LikeCount)
	}
}

func TestLikePost_Idempotent(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	post, _ := svc.CreatePost(context.Background(), uuid.New(), "like me twice", nil)
	userID := uuid.New()

	svc.LikePost(context.Background(), post.ID, userID)
	svc.LikePost(context.Background(), post.ID, userID)

	updated, _ := svc.GetPost(context.Background(), post.ID)
	if updated.LikeCount != 1 {
		t.Errorf("expected like count 1 (idempotent), got %d", updated.LikeCount)
	}
}

func TestUnlikePost_Success(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	post, _ := svc.CreatePost(context.Background(), uuid.New(), "unlike me", nil)
	userID := uuid.New()
	svc.LikePost(context.Background(), post.ID, userID)
	svc.UnlikePost(context.Background(), post.ID, userID)

	updated, _ := svc.GetPost(context.Background(), post.ID)
	if updated.LikeCount != 0 {
		t.Errorf("expected like count 0, got %d", updated.LikeCount)
	}
}

func TestCreateComment_Success(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	post, _ := svc.CreatePost(context.Background(), uuid.New(), "comment here", nil)

	comment, err := svc.CreateComment(context.Background(), post.ID, uuid.New(), "great post!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comment.Content != "great post!" {
		t.Errorf("expected comment content %q, got %q", "great post!", comment.Content)
	}
}

func TestCreateComment_TooLong(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	post, _ := svc.CreatePost(context.Background(), uuid.New(), "test", nil)
	long := strings.Repeat("a", 501)

	_, err := svc.CreateComment(context.Background(), post.ID, uuid.New(), long)
	if err == nil {
		t.Fatal("expected error for comment too long")
	}
}

func TestListComments_Paginated(t *testing.T) {
	repo := newMockPostRepo()
	svc := service.NewPostService(repo, &dummyMediaStore{}, nil)

	post, _ := svc.CreatePost(context.Background(), uuid.New(), "comments test", nil)
	for i := 0; i < 5; i++ {
		svc.CreateComment(context.Background(), post.ID, uuid.New(), "comment "+string(rune('A'+i)))
	}

	comments, _, err := svc.ListComments(context.Background(), post.ID, "", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(comments) != 3 {
		t.Errorf("expected 3 comments, got %d", len(comments))
	}
}

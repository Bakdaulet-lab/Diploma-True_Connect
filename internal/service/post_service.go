package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/pkg/sanitize"
	"github.com/trueconnect/backend/internal/repository"
)

// fixMediaURL преобразует ключ объекта в прямую публичную ссылку.
// Мы не используем PresignedURL, так как бакет PUBLIC, и подпись ломается при смене домена.
func fixMediaURL(mediaKey string) string {
	if mediaKey == "" {
		return ""
	}
	// Если ключ уже содержит протокол, значит это полная ссылка (редкий случай)
	if strings.HasPrefix(mediaKey, "http") {
		return strings.ReplaceAll(mediaKey, "minio:9000", "localhost:9000")
	}

	// Собираем прямую ссылку: http://localhost:9000/ + ИМЯ_БАКЕТА + / + КЛЮЧ
	// Убедись, что имя бакета в MinIO именно "trueconnect"
	return fmt.Sprintf("http://localhost:9000/trueconnect/%s", strings.TrimPrefix(mediaKey, "/"))
}

// PostService handles social feed business logic.
type PostService struct {
	postRepo   repository.PostRepository
	mediaStore repository.MediaStore
	notifSvc   *NotificationService
}

// NewPostService creates a new post service.
func NewPostService(postRepo repository.PostRepository, mediaStore repository.MediaStore, notifSvc *NotificationService) *PostService {
	return &PostService{
		postRepo:   postRepo,
		mediaStore: mediaStore,
		notifSvc:   notifSvc,
	}
}

// CreatePost creates a new social feed post.
func (s *PostService) CreatePost(ctx context.Context, authorID uuid.UUID, content string, mediaData []byte) (*domain.Post, error) {
	content = sanitize.StripHTML(content)
	if len(content) == 0 || len(content) > 2000 {
		return nil, fmt.Errorf("create post: content must be 1-2000 chars: %w", domain.ErrInvalidInput)
	}

	var mediaKey string
	if len(mediaData) > 0 {
		key, err := s.mediaStore.UploadPhoto(ctx, authorID, mediaData)
		if err != nil {
			return nil, fmt.Errorf("create post: upload media: %w", err)
		}
		mediaKey = key
	}

	post := &domain.Post{
		AuthorID: authorID,
		Content:  content,
		MediaURL: mediaKey,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		if mediaKey != "" {
			_ = s.mediaStore.DeletePhoto(ctx, mediaKey)
		}
		return nil, fmt.Errorf("create post: %w", err)
	}

	// Устанавливаем прямую ссылку для ответа
	post.MediaURL = fixMediaURL(post.MediaURL)

	return post, nil
}

// GetPost returns a single post by ID.
func (s *PostService) GetPost(ctx context.Context, postID uuid.UUID) (*domain.Post, error) {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}

	post.MediaURL = fixMediaURL(post.MediaURL)

	return post, nil
}

// DeletePost deletes a post owned by the given author.
func (s *PostService) DeletePost(ctx context.Context, postID, authorID uuid.UUID) error {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("delete post: get: %w", err)
	}

	if err := s.postRepo.Delete(ctx, postID, authorID); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}

	// В хранилище мы все еще используем оригинальный ключ для удаления
	if post.MediaURL != "" {
		_ = s.mediaStore.DeletePhoto(ctx, post.MediaURL)
	}

	return nil
}

// ListFeed returns a paginated feed of posts sorted by recency using cursor.
func (s *PostService) ListFeed(ctx context.Context, cursor string, limit int, filter domain.PostFilter) ([]domain.Post, string, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}

	posts, nextCursor, err := s.postRepo.ListFeed(ctx, cursor, limit, filter)
	if err != nil {
		return nil, "", fmt.Errorf("list feed: %w", err)
	}

	for i := range posts {
		posts[i].MediaURL = fixMediaURL(posts[i].MediaURL)
	}

	return posts, nextCursor, nil
}

// LikePost records a like on a post.
func (s *PostService) LikePost(ctx context.Context, postID, userID uuid.UUID) error {
	if err := s.postRepo.LikePost(ctx, postID, userID); err != nil {
		return fmt.Errorf("like post: %w", err)
	}

	post, err := s.postRepo.GetByID(ctx, postID)
	if err == nil && post.AuthorID != userID && s.notifSvc != nil {
		s.notifSvc.Create(ctx, &domain.Notification{
			UserID:   post.AuthorID,
			ActorID:  &userID,
			Type:     domain.NotificationTypeLike,
			EntityID: &postID,
		})
	}

	return nil
}

// UnlikePost removes a like from a post.
func (s *PostService) UnlikePost(ctx context.Context, postID, userID uuid.UUID) error {
	if err := s.postRepo.UnlikePost(ctx, postID, userID); err != nil {
		return fmt.Errorf("unlike post: %w", err)
	}
	return nil
}

// CreateComment adds a comment to a post.
func (s *PostService) CreateComment(ctx context.Context, postID, authorID uuid.UUID, content string) (*domain.PostComment, error) {
	content = sanitize.StripHTML(content)
	if len(content) == 0 || len(content) > 500 {
		return nil, fmt.Errorf("create comment: content must be 1-500 chars: %w", domain.ErrInvalidInput)
	}

	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}

	comment := &domain.PostComment{
		PostID:   postID,
		AuthorID: authorID,
		Content:  content,
	}

	if err := s.postRepo.CreateComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}

	if post.AuthorID != authorID && s.notifSvc != nil {
		s.notifSvc.Create(ctx, &domain.Notification{
			UserID:   post.AuthorID,
			ActorID:  &authorID,
			Type:     domain.NotificationTypeComment,
			EntityID: &postID,
		})
	}

	return comment, nil
}

// ListComments returns paginated comments for a post using cursor.
func (s *PostService) ListComments(ctx context.Context, postID uuid.UUID, cursor string, limit int) ([]domain.PostComment, string, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}

	comments, nextCursor, err := s.postRepo.ListComments(ctx, postID, cursor, limit)
	if err != nil {
		return nil, "", fmt.Errorf("list comments: %w", err)
	}

	return comments, nextCursor, nil
}

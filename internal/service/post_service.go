package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/pkg/sanitize"
	"github.com/trueconnect/backend/internal/repository"
)

// PostService handles social feed business logic.
type PostService struct {
	postRepo repository.PostRepository
}

// NewPostService creates a new post service.
func NewPostService(postRepo repository.PostRepository) *PostService {
	return &PostService{postRepo: postRepo}
}

// CreatePost creates a new social feed post.
func (s *PostService) CreatePost(ctx context.Context, authorID uuid.UUID, content, mediaURL string) (*domain.Post, error) {
	content = sanitize.StripHTML(content)
	if len(content) == 0 || len(content) > 2000 {
		return nil, fmt.Errorf("create post: content must be 1-2000 chars: %w", domain.ErrInvalidInput)
	}

	post := &domain.Post{
		AuthorID: authorID,
		Content:  content,
		MediaURL: mediaURL,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}

	return post, nil
}

// GetPost returns a single post by ID.
func (s *PostService) GetPost(ctx context.Context, postID uuid.UUID) (*domain.Post, error) {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}
	return post, nil
}

// DeletePost deletes a post owned by the given author.
func (s *PostService) DeletePost(ctx context.Context, postID, authorID uuid.UUID) error {
	if err := s.postRepo.Delete(ctx, postID, authorID); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	return nil
}

// ListFeed returns a paginated feed of posts sorted by recency.
func (s *PostService) ListFeed(ctx context.Context, page, perPage int) ([]domain.Post, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	posts, err := s.postRepo.ListFeed(ctx, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("list feed: %w", err)
	}

	return posts, nil
}

// LikePost records a like on a post.
func (s *PostService) LikePost(ctx context.Context, postID, userID uuid.UUID) error {
	if err := s.postRepo.LikePost(ctx, postID, userID); err != nil {
		return fmt.Errorf("like post: %w", err)
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

	// Verify post exists.
	if _, err := s.postRepo.GetByID(ctx, postID); err != nil {
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

	return comment, nil
}

// ListComments returns paginated comments for a post.
func (s *PostService) ListComments(ctx context.Context, postID uuid.UUID, page, perPage int) ([]domain.PostComment, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	comments, err := s.postRepo.ListComments(ctx, postID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}

	return comments, nil
}

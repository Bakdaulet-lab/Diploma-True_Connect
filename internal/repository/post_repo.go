package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error)
	Delete(ctx context.Context, id uuid.UUID, authorID uuid.UUID) error
	ListFeed(ctx context.Context, limit, offset int) ([]domain.Post, error)
	ListByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]domain.Post, error)
	IncrementLikeCount(ctx context.Context, id uuid.UUID, delta int) error
	IncrementCommentCount(ctx context.Context, id uuid.UUID, delta int) error
	LikePost(ctx context.Context, postID, userID uuid.UUID) error
	UnlikePost(ctx context.Context, postID, userID uuid.UUID) error
	IsLikedBy(ctx context.Context, postID, userID uuid.UUID) (bool, error)
	CreateComment(ctx context.Context, comment *domain.PostComment) error
	ListComments(ctx context.Context, postID uuid.UUID, limit, offset int) ([]domain.PostComment, error)
}

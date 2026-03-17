package domain

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID           uuid.UUID
	AuthorID     uuid.UUID
	Content      string
	MediaURL     string
	LikeCount    int
	CommentCount int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PostComment struct {
	ID        uuid.UUID
	PostID    uuid.UUID
	AuthorID  uuid.UUID
	Content   string
	CreatedAt time.Time
}

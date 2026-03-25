package domain

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID           uuid.UUID `json:"id"`
	AuthorID     uuid.UUID `json:"author_id"`
	Content      string    `json:"content"`
	MediaURL     string    `json:"media_url,omitempty"`
	LikeCount    int       `json:"like_count"`
	CommentCount int       `json:"comment_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Эти поля заполняются через JOIN в SQL-запросе
	AuthorName   string `json:"author_name,omitempty"`
	AuthorAvatar string `json:"author_avatar,omitempty"`
}

type PostComment struct {
	ID        uuid.UUID `json:"id"`
	PostID    uuid.UUID `json:"post_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`

	// ДОБАВЬ ЭТИ ПОЛЯ:
	AuthorName   string `json:"author_name"`   // Для отображения имени из профиля
	AuthorAvatar string `json:"author_avatar"` // Для отображения аватарки из профиля
}

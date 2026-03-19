package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

type MessageRepository interface {
	Create(ctx context.Context, msg *domain.Message) error
	ListByMatch(ctx context.Context, matchID uuid.UUID, cursor string, limit int) ([]domain.Message, string, error)
	MarkRead(ctx context.Context, matchID uuid.UUID, readerID uuid.UUID) error
}

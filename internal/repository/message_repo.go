package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

type MessageRepository interface {
	Create(ctx context.Context, msg *domain.Message) error
	ListByMatch(ctx context.Context, matchID uuid.UUID, limit, offset int) ([]domain.Message, error)
	MarkRead(ctx context.Context, matchID uuid.UUID, readerID uuid.UUID) error
}

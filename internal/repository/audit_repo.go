package repository

import (
	"context"

	"github.com/google/uuid"
)

type AuditLog struct {
	UserID    uuid.UUID
	Path      string
	Method    string
	IPAddress string
}

type AuditRepository interface {
	LogAction(ctx context.Context, log *AuditLog) error
}

package repository

import (
	"context"

	"github.com/google/uuid"
)

// MediaStore defines object-storage operations used by the profile service.
type MediaStore interface {
	UploadPhoto(ctx context.Context, userID uuid.UUID, data []byte) (string, error)
	UploadDocument(ctx context.Context, userID uuid.UUID, data []byte, contentType string) (string, error)
	PresignedURL(ctx context.Context, objectKey string) (string, error)
	DeletePhoto(ctx context.Context, objectKey string) error
}

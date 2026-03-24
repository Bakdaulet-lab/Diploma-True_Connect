package minioadapter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/trueconnect/backend/internal/repository"
)

// Compile-time check: MediaStore must implement repository.MediaStore.
var _ repository.MediaStore = (*MediaStore)(nil)

const (
	maxPhotoBytes = 10 * 1024 * 1024 // 10 MB
	presignExpiry = 1 * time.Hour
)

// MediaStore handles photo uploads and presigned URL generation via MinIO.
type MediaStore struct {
	client *minio.Client
	bucket string
}

// NewMediaStore creates a MediaStore backed by an existing MinIO client.
func NewMediaStore(client *minio.Client, bucket string) *MediaStore {
	return &MediaStore{client: client, bucket: bucket}
}

// UploadPhoto stores a user's photo in MinIO after validating type and size.
// Returns the object key (path within bucket) to be stored in the DB.
func (s *MediaStore) UploadPhoto(ctx context.Context, userID uuid.UUID, data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("upload photo: file is empty")
	}
	if len(data) > maxPhotoBytes {
		return "", fmt.Errorf("upload photo: file exceeds 10 MB limit")
	}

	// Detect MIME type from magic bytes (not file extension — prevents content spoofing).
	mimeType := http.DetectContentType(data)
	switch mimeType {
	case "image/jpeg", "image/png", "image/webp":
		// allowed
	default:
		return "", fmt.Errorf("upload photo: unsupported file type %q (allowed: jpeg, png, webp)", mimeType)
	}

	ext := mimeToExt(mimeType)
	objectKey := fmt.Sprintf("users/%s/photos/%s%s", userID.String(), uuid.New().String(), ext)

	_, err := s.client.PutObject(ctx, s.bucket, objectKey, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return "", fmt.Errorf("upload photo: storing object: %w", err)
	}

	return objectKey, nil
}

// PresignedURL generates a time-limited, pre-authenticated URL for downloading a private object.
func (s *MediaStore) PresignedURL(ctx context.Context, objectKey string) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, presignExpiry, nil)
	if err != nil {
		return "", fmt.Errorf("generating presigned URL: %w", err)
	}

	return u.String(), nil
}

// DeletePhoto removes an object from MinIO. Called when a user deletes a photo.
func (s *MediaStore) DeletePhoto(ctx context.Context, objectKey string) error {
	err := s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("deleting photo: %w", err)
	}

	return nil
}

// ReadAll reads the full contents of an uploaded file (used for validation in handler).
func ReadAll(r io.Reader, maxBytes int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading upload: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("file exceeds %d bytes", maxBytes)
	}
	return data, nil
}

func mimeToExt(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	}
	return ".bin"
}

// UploadDocument stores a KYC document (image or PDF) under a dedicated path.
func (s *MediaStore) UploadDocument(ctx context.Context, userID uuid.UUID, data []byte, contentType string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("upload document: file is empty")
	}
	if len(data) > maxPhotoBytes {
		return "", fmt.Errorf("upload document: file exceeds 10 MB limit")
	}

	switch contentType {
	case "image/jpeg", "image/png", "application/pdf":
		// allowed
	default:
		return "", fmt.Errorf("upload document: unsupported type %q (allowed: jpeg, png, pdf)", contentType)
	}

	ext := mimeToExt(contentType)
	objectKey := fmt.Sprintf("kyc/%s/%s%s", userID.String(), uuid.New().String(), ext)

	_, err := s.client.PutObject(ctx, s.bucket, objectKey, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload document: storing object: %w", err)
	}

	return objectKey, nil
}

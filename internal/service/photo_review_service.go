package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// PhotoReviewService handles the admin re-review queue for rejected photos.
type PhotoReviewService struct {
	reviewRepo repository.PhotoReviewRepository
	mediaRepo  repository.MediaRepository
	mediaStore repository.MediaStore
}

func NewPhotoReviewService(
	reviewRepo repository.PhotoReviewRepository,
	mediaRepo repository.MediaRepository,
	mediaStore repository.MediaStore,
) *PhotoReviewService {
	return &PhotoReviewService{reviewRepo: reviewRepo, mediaRepo: mediaRepo, mediaStore: mediaStore}
}

// PhotoReviewView is an admin-facing pending review with a viewable URL.
type PhotoReviewView struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	URL       string    `json:"url"`
	Reasons   []string  `json:"reasons"`
	NSFWScore float64   `json:"nsfw_score"`
	FaceCount int       `json:"face_count"`
	CreatedAt time.Time `json:"created_at"`
}

// ListPending returns pending reviews with presigned image URLs for the admin UI.
func (s *PhotoReviewService) ListPending(ctx context.Context, limit int) ([]*PhotoReviewView, error) {
	items, err := s.reviewRepo.ListPending(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list photo reviews: %w", err)
	}
	views := make([]*PhotoReviewView, 0, len(items))
	for _, it := range items {
		url, _ := s.mediaStore.PresignedURL(ctx, it.ObjectKey)
		views = append(views, &PhotoReviewView{
			ID: it.ID, UserID: it.UserID, URL: url, Reasons: it.Reasons,
			NSFWScore: it.NSFWScore, FaceCount: it.FaceCount, CreatedAt: it.CreatedAt,
		})
	}
	return views, nil
}

// Approve promotes the held photo into a real media record + avatar.
func (s *PhotoReviewService) Approve(ctx context.Context, reviewID, adminID uuid.UUID) error {
	r, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("approve review: %w", err)
	}
	if r.Status != domain.PhotoReviewPending {
		return fmt.Errorf("approve review: %w", domain.ErrInvalidInput)
	}

	count, _ := s.mediaRepo.CountByUser(ctx, r.UserID)
	media := &domain.Media{
		ID:        uuid.New(),
		UserID:    r.UserID,
		ObjectKey: r.ObjectKey,
		MediaType: "photo",
		SortOrder: count,
	}
	if err := s.mediaRepo.Create(ctx, media); err != nil {
		return fmt.Errorf("approve review: create media: %w", err)
	}
	_ = s.mediaRepo.UpdateAvatar(ctx, r.UserID, r.ObjectKey)

	return s.reviewRepo.UpdateStatus(ctx, reviewID, domain.PhotoReviewApproved, adminID)
}

// Reject deletes the held photo and marks the review rejected.
func (s *PhotoReviewService) Reject(ctx context.Context, reviewID, adminID uuid.UUID) error {
	r, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("reject review: %w", err)
	}
	if r.Status != domain.PhotoReviewPending {
		return fmt.Errorf("reject review: %w", domain.ErrInvalidInput)
	}
	_ = s.mediaStore.DeletePhoto(ctx, r.ObjectKey)
	return s.reviewRepo.UpdateStatus(ctx, reviewID, domain.PhotoReviewRejected, adminID)
}

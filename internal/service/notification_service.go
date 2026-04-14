package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

type wsOutgoing struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
	Error   string `json:"error,omitempty"`
}

type NotificationService struct {
	repo repository.NotificationRepository
	rdb  *redis.Client
	log  *slog.Logger
}

func NewNotificationService(repo repository.NotificationRepository, rdb *redis.Client, log *slog.Logger) *NotificationService {
	return &NotificationService{
		repo: repo,
		rdb:  rdb,
		log:  log,
	}
}

func (s *NotificationService) Create(ctx context.Context, notif *domain.Notification) error {
	if err := s.repo.Create(ctx, notif); err != nil {
		return fmt.Errorf("create notif: %w", err)
	}

	// Fetch full details (like actor name/avatar) after creation to broadcast
	fullNotif, err := s.repo.GetByID(ctx, notif.ID)
	if err == nil {
		s.broadcast(ctx, fullNotif)
	}

	return nil
}

func (s *NotificationService) List(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]domain.Notification, string, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.ListByUser(ctx, userID, cursor, limit)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, id, userID)
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

func (s *NotificationService) broadcast(ctx context.Context, n *domain.Notification) {
	msg := wsOutgoing{
		Type:    "NOTIFICATION_CREATED",
		Payload: n,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		s.log.Error("marshal notif broadcast error", slog.String("error", err.Error()))
		return
	}

	if err := s.rdb.Publish(ctx, "ws:user:"+n.UserID.String(), b).Err(); err != nil {
		s.log.Error("publish notif error", slog.String("error", err.Error()))
	}
}

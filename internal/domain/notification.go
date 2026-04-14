package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeLike    NotificationType = "like"
	NotificationTypeComment NotificationType = "comment"
	NotificationTypeMatch   NotificationType = "match"
	NotificationTypeSystem  NotificationType = "system"
)

type Notification struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	UserID    uuid.UUID        `json:"user_id" db:"user_id"`
	ActorID   *uuid.UUID       `json:"actor_id,omitempty" db:"actor_id"`
	Type      NotificationType `json:"type" db:"type"`
	EntityID  *uuid.UUID       `json:"entity_id,omitempty" db:"entity_id"`
	IsRead    bool             `json:"is_read" db:"is_read"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`

	// Joined fields
	ActorName   string `json:"actor_name,omitempty" db:"actor_name"`
	ActorAvatar string `json:"actor_avatar,omitempty" db:"actor_avatar"`
}

package domain

import "github.com/google/uuid"

type PushEvent struct {
	UserID uuid.UUID
	Title  string
	Body   string
	Data   map[string]string
}

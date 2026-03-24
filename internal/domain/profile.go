package domain

import (
	"time"

	"github.com/google/uuid"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type Profile struct {
	UserID      uuid.UUID
	DisplayName string
	Bio         string
	Gender      Gender
	BirthDate   *time.Time
	City        string
	Latitude    float64
	Longitude   float64
	LookingFor  Gender
	AvatarURL   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

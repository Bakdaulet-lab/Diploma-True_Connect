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

type PromptAnswer struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type Profile struct {
	UserID      uuid.UUID
	DisplayName string
	Bio         string
	Gender      Gender
	Prompts     []PromptAnswer
	BirthDate   *time.Time
	City        string
	Latitude    *float64 // PostGIS location
	Longitude   *float64 // PostGIS location
	LookingFor  Gender
	AvatarURL   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

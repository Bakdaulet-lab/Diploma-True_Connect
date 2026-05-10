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

// Niyyah represents the user's stated intention for marriage.
type Niyyah string

const (
	NiyyahNikahYear        Niyyah = "nikah_year"
	NiyyahSeriousMarriage  Niyyah = "serious_marriage"
	NiyyahFriendship       Niyyah = "friendship"
)

// Madhab represents the Islamic legal school the user follows.
type Madhab string

const (
	MadhabHanafi  Madhab = "hanafi"
	MadhabShafii  Madhab = "shafii"
	MadhabMaliki  Madhab = "maliki"
	MadhabHanbali Madhab = "hanbali"
	MadhabNone    Madhab = "none"
)

// MaritalStatus represents whether the user is available for discovery.
const (
	MaritalSingle        = "single"
	MaritalMarriedViaApp = "married_via_app"
	MaritalDivorced      = "divorced"
)

type Profile struct {
	UserID        uuid.UUID
	DisplayName   string
	Bio           string
	Gender        Gender
	Prompts       []PromptAnswer
	BirthDate     *time.Time
	City          string
	Latitude      *float64 // PostGIS location
	Longitude     *float64 // PostGIS location
	LookingFor    Gender
	AvatarURL     string
	Niyyah        Niyyah
	Madhab        Madhab
	Languages     []string
	NoPhotoMode   bool
	MaritalStatus string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

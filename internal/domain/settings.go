package domain

import "time"

type UserSettings struct {
	UserID            string
	PushNotifications bool
	ShowOnlineStatus  bool
	DistanceUnit      string
	MaxDistanceKm     int
	AgeRangeMin       int
	AgeRangeMax       int
	ModestyLevel      int
	NiyyahFilter      *string
	MadhabFilter      *string
	UpdatedAt         time.Time
}

// DefaultSettings returns sensible defaults for a new user.
func DefaultSettings(userID string) *UserSettings {
	return &UserSettings{
		UserID:            userID,
		PushNotifications: true,
		ShowOnlineStatus:  true,
		DistanceUnit:      "km",
		MaxDistanceKm:     50,
		AgeRangeMin:       18,
		AgeRangeMax:       60,
		UpdatedAt:         time.Now(),
	}
}

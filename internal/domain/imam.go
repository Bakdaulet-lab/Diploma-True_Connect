package domain

import "github.com/google/uuid"

type Imam struct {
	ID                uuid.UUID
	Name              string
	City              string
	Mosque            string
	Languages         []string
	AvailabilityNotes string
}

package imam

import (
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

//go:embed catalog.json
var catalogJSON []byte

type catalogEntry struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	City              string   `json:"city"`
	Mosque            string   `json:"mosque"`
	Languages         []string `json:"languages"`
	AvailabilityNotes string   `json:"availability_notes"`
}

var catalog []domain.Imam

func init() {
	var entries []catalogEntry
	if err := json.Unmarshal(catalogJSON, &entries); err != nil {
		panic("imam catalog: invalid JSON: " + err.Error())
	}
	for _, e := range entries {
		id, err := uuid.Parse(e.ID)
		if err != nil {
			panic("imam catalog: invalid UUID: " + e.ID)
		}
		catalog = append(catalog, domain.Imam{
			ID:                id,
			Name:              e.Name,
			City:              e.City,
			Mosque:            e.Mosque,
			Languages:         e.Languages,
			AvailabilityNotes: e.AvailabilityNotes,
		})
	}
}

// ListByCity returns all imams in the given city (case-insensitive).
func ListByCity(city string) []domain.Imam {
	city = strings.ToLower(city)
	var result []domain.Imam
	for _, i := range catalog {
		if strings.ToLower(i.City) == city {
			result = append(result, i)
		}
	}
	return result
}

// GetByID returns the imam with the given UUID, or false if not found.
func GetByID(id uuid.UUID) (domain.Imam, bool) {
	for _, i := range catalog {
		if i.ID == id {
			return i, true
		}
	}
	return domain.Imam{}, false
}

package venues

import (
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/trueconnect/backend/internal/domain"
)

//go:embed catalog.json
var catalogJSON []byte

var catalog []domain.Venue

func init() {
	if err := json.Unmarshal(catalogJSON, &catalog); err != nil {
		panic("venues catalog: invalid JSON: " + err.Error())
	}
}

// ListByCity returns all venues in the given city (case-insensitive).
func ListByCity(city string) []domain.Venue {
	city = strings.ToLower(city)
	var result []domain.Venue
	for _, v := range catalog {
		if strings.ToLower(v.City) == city {
			result = append(result, v)
		}
	}
	return result
}

// ListAll returns all venues.
func ListAll() []domain.Venue {
	return catalog
}

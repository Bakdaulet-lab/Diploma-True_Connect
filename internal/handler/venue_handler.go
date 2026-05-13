package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trueconnect/backend/internal/pkg/venues"
)

// VenueHandler serves the halal venue catalog.
type VenueHandler struct{}

// NewVenueHandler creates a new VenueHandler.
func NewVenueHandler() *VenueHandler {
	return &VenueHandler{}
}

// ListVenues handles GET /v1/venues?city=...
// Returns halal-friendly venues for the first meeting protocol.
func (h *VenueHandler) ListVenues(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		c.JSON(http.StatusOK, gin.H{"data": venues.ListAll()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": venues.ListByCity(city)})
}

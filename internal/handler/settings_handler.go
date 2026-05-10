package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/handler/middleware"
	"github.com/trueconnect/backend/internal/service"
)

// SettingsHandler holds HTTP handlers for user preferences.
type SettingsHandler struct {
	settingsSvc *service.SettingsService
	log         *slog.Logger
}

// NewSettingsHandler creates a new settings handler.
func NewSettingsHandler(settingsSvc *service.SettingsService, log *slog.Logger) *SettingsHandler {
	return &SettingsHandler{settingsSvc: settingsSvc, log: log}
}

// GetSettings handles GET /v1/settings
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	settings, err := h.settingsSvc.Get(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("get settings error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not load settings", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

// UpdateSettings handles PATCH /v1/settings (partial update — only provided fields are changed)
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	// Use a map to detect which fields were actually provided.
	var raw map[string]any
	if err := c.ShouldBindJSON(&raw); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	input := service.UpdateSettingsInput{}

	if v, ok := raw["push_notifications"]; ok {
		if b, ok := v.(bool); ok {
			input.PushNotifications = &b
		}
	}
	if v, ok := raw["show_online_status"]; ok {
		if b, ok := v.(bool); ok {
			input.ShowOnlineStatus = &b
		}
	}
	if v, ok := raw["distance_unit"]; ok {
		if s, ok := v.(string); ok {
			input.DistanceUnit = &s
		}
	}
	if v, ok := raw["max_distance_km"]; ok {
		if n, ok := toInt(v); ok {
			input.MaxDistanceKm = &n
		}
	}
	if v, ok := raw["age_range_min"]; ok {
		if n, ok := toInt(v); ok {
			input.AgeRangeMin = &n
		}
	}
	if v, ok := raw["age_range_max"]; ok {
		if n, ok := toInt(v); ok {
			input.AgeRangeMax = &n
		}
	}
	if v, ok := raw["modesty_level"]; ok {
		if n, ok := toInt(v); ok {
			input.ModestyLevel = &n
		}
	}
	if v, ok := raw["niyyah_filter"]; ok {
		if s, ok := v.(string); ok {
			input.NiyyahFilter = &s
		}
	}
	if v, ok := raw["madhab_filter"]; ok {
		if s, ok := v.(string); ok {
			input.MadhabFilter = &s
		}
	}

	updated, err := h.settingsSvc.Update(c.Request.Context(), userID, input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			errorResponse(c, http.StatusUnprocessableEntity, "INVALID_INPUT", err.Error(), nil)
			return
		}
		h.log.Error("update settings error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not update settings", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updated})
}

// toInt converts a JSON number (float64 from unmarshaling) to int.
func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	}
	return 0, false
}

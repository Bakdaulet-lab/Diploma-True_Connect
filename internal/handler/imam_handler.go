package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/handler/middleware"
	"github.com/trueconnect/backend/internal/service"
)

// ImamHandler holds HTTP handlers for imam connect.
type ImamHandler struct {
	imamSvc *service.ImamService
	log     *slog.Logger
}

// NewImamHandler creates a new imam handler.
func NewImamHandler(imamSvc *service.ImamService, log *slog.Logger) *ImamHandler {
	return &ImamHandler{imamSvc: imamSvc, log: log}
}

// ── GET /v1/imams?city=... ────────────────────────────────────────────────

// ListImams handles GET /v1/imams (returns imam catalog filtered by city).
func (h *ImamHandler) ListImams(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		errorResponse(c, http.StatusBadRequest, "MISSING_CITY", "city query parameter is required", nil)
		return
	}

	imams, err := h.imamSvc.ListImams(c.Request.Context(), city)
	if err != nil {
		h.log.Error("list imams error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not load imam catalog", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": imams})
}

// ── POST /v1/matches/:id/nikah-confirm ────────────────────────────────────

type confirmNikahRequest struct {
	ImamID string `json:"imam_id" validate:"required,uuid"`
}

// ConfirmNikah handles POST /v1/matches/:id/nikah-confirm (stub for Sprint 11).
func (h *ImamHandler) ConfirmNikah(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	matchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_MATCH_ID", "match ID must be a valid UUID", nil)
		return
	}

	var req confirmNikahRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	imamID, err := uuid.Parse(req.ImamID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_IMAM_ID", "imam ID must be a valid UUID", nil)
		return
	}

	if err := h.imamSvc.ConfirmNikah(c.Request.Context(), matchID, imamID, userID); err != nil {
		h.log.Error("confirm nikah error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not confirm nikah", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nikah confirmed successfully"})
}

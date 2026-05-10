package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/handler/middleware"
	"github.com/trueconnect/backend/internal/service"
)

// WhisperHandler holds HTTP handlers for anonymous whisper feedback.
type WhisperHandler struct {
	whisperSvc *service.WhisperService
	log        *slog.Logger
}

// NewWhisperHandler creates a new whisper handler.
func NewWhisperHandler(whisperSvc *service.WhisperService, log *slog.Logger) *WhisperHandler {
	return &WhisperHandler{whisperSvc: whisperSvc, log: log}
}

// ── POST /v1/whisper ──────────────────────────────────────────────────────

type submitWhisperRequest struct {
	ReportedUserID uuid.UUID `json:"reported_user_id" validate:"required"`
	MatchID        uuid.UUID `json:"match_id" validate:"required"`
	Feedback       string    `json:"feedback" validate:"required,max=1000"`
}

// SubmitWhisper handles POST /v1/whisper.
func (h *WhisperHandler) SubmitWhisper(c *gin.Context) {
	reporterID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	var req submitWhisperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	if err := h.whisperSvc.SubmitWhisper(c.Request.Context(), reporterID, req.ReportedUserID, req.MatchID, req.Feedback); err != nil {
		h.handleWhisperError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Feedback submitted anonymously"})
}

// handleWhisperError maps domain errors to HTTP responses.
func (h *WhisperHandler) handleWhisperError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "match or reported user not found", nil)
	case errors.Is(err, domain.ErrForbidden):
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "you do not have a mutual match with this user", nil)
	case errors.Is(err, domain.ErrInvalidInput):
		errorResponse(c, http.StatusUnprocessableEntity, "INVALID_INPUT", err.Error(), nil)
	default:
		h.log.Error("whisper error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred", nil)
	}
}

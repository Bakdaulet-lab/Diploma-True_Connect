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

// MahramChatHandler handles REST endpoints for mahram group chat rooms.
type MahramChatHandler struct {
	svc *service.MahramChatService
	log *slog.Logger
}

// NewMahramChatHandler creates a new mahram chat handler.
func NewMahramChatHandler(svc *service.MahramChatService, log *slog.Logger) *MahramChatHandler {
	return &MahramChatHandler{svc: svc, log: log}
}

// CreateRoom handles POST /v1/mahram-rooms
func (h *MahramChatHandler) CreateRoom(c *gin.Context) {
	callerID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	var req struct {
		MatchID      string `json:"match_id" binding:"required"`
		MahramUserID string `json:"mahram_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	matchID, err := uuid.Parse(req.MatchID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_MATCH_ID", "match_id must be a valid UUID", nil)
		return
	}

	mahramUserID, err := uuid.Parse(req.MahramUserID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_MAHRAM_USER_ID", "mahram_user_id must be a valid UUID", nil)
		return
	}

	room, err := h.svc.CreateRoom(c.Request.Context(), matchID, callerID, mahramUserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			errorResponse(c, http.StatusNotFound, "NOT_FOUND", "match not found", nil)
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			errorResponse(c, http.StatusForbidden, "FORBIDDEN", "not a match participant or match not finalized", nil)
			return
		}
		h.log.Error("create mahram room error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create room", nil)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": room})
}

// GetMessages handles GET /v1/mahram-rooms/:room_id/messages
func (h *MahramChatHandler) GetMessages(c *gin.Context) {
	callerID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	roomID, err := uuid.Parse(c.Param("room_id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "room_id must be a valid UUID", nil)
		return
	}

	cursor, limit := parseCursorPagination(c)

	messages, nextCursor, err := h.svc.GetMessages(c.Request.Context(), roomID, callerID, cursor, limit)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			errorResponse(c, http.StatusNotFound, "NOT_FOUND", "room not found", nil)
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			errorResponse(c, http.StatusForbidden, "FORBIDDEN", "not a room participant", nil)
			return
		}
		h.log.Error("get mahram messages error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not get messages", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": messages,
		"meta": gin.H{"has_more": nextCursor != "", "next_cursor": nextCursor},
	})
}

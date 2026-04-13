package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/handler/middleware"
	"github.com/trueconnect/backend/internal/service"
)

// MatchingHandler holds HTTP handlers for the matching feed and swipe actions.
type MatchingHandler struct {
	matchingSvc *service.MatchingService
	log         *slog.Logger
}

// NewMatchingHandler creates a new matching handler.
func NewMatchingHandler(matchingSvc *service.MatchingService, log *slog.Logger) *MatchingHandler {
	return &MatchingHandler{matchingSvc: matchingSvc, log: log}
}

// GetCandidates handles GET /v1/matching/candidates
func (h *MatchingHandler) GetCandidates(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	candidates, err := h.matchingSvc.GetCandidates(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Profile not set up yet.
			errorResponse(c, http.StatusUnprocessableEntity, "PROFILE_REQUIRED",
				"complete your profile (including location) before viewing matches", nil)
			return
		}
		h.log.Error("get candidates error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not load candidates", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": candidates})
}

// Like handles POST /v1/matching/like
func (h *MatchingHandler) Like(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	var req struct {
		TargetID string `json:"target_id" validate:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_TARGET_ID", "target_id must be a valid UUID", nil)
		return
	}

	result, err := h.matchingSvc.Like(c.Request.Context(), userID, targetID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			errorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "cannot like yourself", nil)
			return
		}
		h.log.Error("like error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not record like", nil)
		return
	}

	status := http.StatusOK
	if result.Matched {
		status = http.StatusCreated // 201 signals a new match to the client
	}
	c.JSON(status, gin.H{"data": result})
}

// Pass handles POST /v1/matching/pass
func (h *MatchingHandler) Pass(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	var req struct {
		TargetID string `json:"target_id" validate:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_TARGET_ID", "target_id must be a valid UUID", nil)
		return
	}

	if err := h.matchingSvc.Pass(c.Request.Context(), userID, targetID); err != nil {
		h.log.Error("pass error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not record pass", nil)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListMatches handles GET /v1/matches
func (h *MatchingHandler) ListMatches(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	cursor := c.Query("cursor")
	limit := 20 // default
	if c.Query("limit") != "" {
		// normally parse this safely, here just fallback
		fmt.Sscanf(c.Query("limit"), "%d", &limit)
	}

	matches, nextCursor, err := h.matchingSvc.ListMatches(c.Request.Context(), userID, cursor, limit)
	if err != nil {
		h.log.Error("list matches error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not list matches", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": matches,
		"meta": gin.H{"next_cursor": nextCursor, "limit": limit},
	})
}

// GetGraphCandidates handles GET /v1/matching/graph-candidates
func (h *MatchingHandler) GetGraphCandidates(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	candidates, err := h.matchingSvc.GetGraphCandidates(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("failed to get graph candidates", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get recommendations", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": candidates})
}

package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/handler/middleware"
	"github.com/trueconnect/backend/internal/service"
)

// InteractionHandler holds HTTP handlers for interactions and reputation.
type InteractionHandler struct {
	interactionSvc *service.InteractionService
	reputeSvc      *service.ReputationService
	log            *slog.Logger
}

// NewInteractionHandler creates a new interaction handler.
func NewInteractionHandler(
	interactionSvc *service.InteractionService,
	reputeSvc *service.ReputationService,
	log *slog.Logger,
) *InteractionHandler {
	return &InteractionHandler{
		interactionSvc: interactionSvc,
		reputeSvc:      reputeSvc,
		log:            log,
	}
}

// SubmitRating handles POST /v1/interactions
func (h *InteractionHandler) SubmitRating(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	var req struct {
		RatedID string `json:"rated_id" binding:"required"`
		Rating  int    `json:"rating" binding:"required,min=1,max=5"`
		Context string `json:"context" binding:"required,oneof=date meetup event"`
		Comment string `json:"comment" binding:"max=300"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	ratedID, err := uuid.Parse(req.RatedID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_RATED_ID", "rated_id must be a valid UUID", nil)
		return
	}

	interaction, err := h.interactionSvc.SubmitRating(
		c.Request.Context(), userID, ratedID,
		req.Rating, req.Context, req.Comment,
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			errorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "cannot rate yourself", nil)
		case errors.Is(err, domain.ErrForbidden):
			errorResponse(c, http.StatusForbidden, "NOT_MATCHED", "you must be matched with this user to rate them", nil)
		case errors.Is(err, domain.ErrRateLimitExceeded):
			errorResponse(c, http.StatusTooManyRequests, "RATE_LIMIT", "you can only rate this user once every 7 days", nil)
		default:
			h.log.Error("submit rating error", slog.String("error", err.Error()))
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not submit rating", nil)
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": interaction})
}

// ConfirmInteraction handles POST /v1/interactions/:id/confirm
func (h *InteractionHandler) ConfirmInteraction(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	interactionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID", nil)
		return
	}

	if err := h.interactionSvc.ConfirmInteraction(c.Request.Context(), interactionID, userID); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			errorResponse(c, http.StatusNotFound, "NOT_FOUND", "interaction not found", nil)
		case errors.Is(err, domain.ErrForbidden):
			errorResponse(c, http.StatusForbidden, "FORBIDDEN", "only the rated user can confirm an interaction", nil)
		default:
			h.log.Error("confirm interaction error", slog.String("error", err.Error()))
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not confirm interaction", nil)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"confirmed": true}})
}

// GetReputation handles GET /v1/users/:id/reputation
func (h *InteractionHandler) GetReputation(c *gin.Context) {
	_, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID", nil)
		return
	}

	trustScore, err := h.reputeSvc.GetScore(c.Request.Context(), targetID)
	if err != nil {
		h.log.Error("get reputation error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not get reputation", nil)
		return
	}

	// Also fetch recent ratings for the summary.
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	ratings, err := h.interactionSvc.GetByRatedUser(c.Request.Context(), targetID, limit, offset)
	if err != nil {
		h.log.Error("get ratings error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not get ratings", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"trust_score": trustScore.Score,
			"user_id":     trustScore.UserID,
			"ratings":     ratings,
		},
	})
}

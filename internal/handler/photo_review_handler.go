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

// PhotoReviewHandler exposes the admin re-review queue for rejected photos.
type PhotoReviewHandler struct {
	svc *service.PhotoReviewService
	log *slog.Logger
}

func NewPhotoReviewHandler(svc *service.PhotoReviewService, log *slog.Logger) *PhotoReviewHandler {
	return &PhotoReviewHandler{svc: svc, log: log}
}

// List handles GET /v1/admin/photo-reviews
func (h *PhotoReviewHandler) List(c *gin.Context) {
	_, limit := parseCursorPagination(c)
	items, err := h.svc.ListPending(c.Request.Context(), limit)
	if err != nil {
		h.log.Error("list photo reviews", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not list reviews", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// Approve handles POST /v1/admin/photo-reviews/:id/approve
func (h *PhotoReviewHandler) Approve(c *gin.Context) { h.act(c, true) }

// Reject handles POST /v1/admin/photo-reviews/:id/reject
func (h *PhotoReviewHandler) Reject(c *gin.Context) { h.act(c, false) }

func (h *PhotoReviewHandler) act(c *gin.Context, approve bool) {
	adminID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID", nil)
		return
	}

	if approve {
		err = h.svc.Approve(c.Request.Context(), id, adminID)
	} else {
		err = h.svc.Reject(c.Request.Context(), id, adminID)
	}
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			errorResponse(c, http.StatusNotFound, "NOT_FOUND", "review not found or already handled", nil)
			return
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			errorResponse(c, http.StatusConflict, "ALREADY_REVIEWED", "review already handled", nil)
			return
		}
		h.log.Error("photo review action", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not update review", nil)
		return
	}

	status := "rejected"
	if approve {
		status = "approved"
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": status}})
}

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

// MahramHandler holds HTTP handlers for mahram registration.
type MahramHandler struct {
	mahramSvc *service.MahramService
	log       *slog.Logger
}

// NewMahramHandler creates a new mahram handler.
func NewMahramHandler(mahramSvc *service.MahramService, log *slog.Logger) *MahramHandler {
	return &MahramHandler{mahramSvc: mahramSvc, log: log}
}

// ── POST /v1/mahram ───────────────────────────────────────────────────────

type registerMahramRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
}

// RegisterMahram handles POST /v1/mahram.
func (h *MahramHandler) RegisterMahram(c *gin.Context) {
	womanID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	var req registerMahramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	view, otp, err := h.mahramSvc.RegisterMahram(c.Request.Context(), womanID, req.PhoneNumber)
	if err != nil {
		h.handleMahramError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": view,
		"otp":  otp, // Dev mode only; in production this would be sent via SMS/Telegram
	})
}

// ── GET /v1/mahram ────────────────────────────────────────────────────────

// GetMahrams handles GET /v1/mahram.
func (h *MahramHandler) GetMahrams(c *gin.Context) {
	womanID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	mahrams, err := h.mahramSvc.GetMahrams(c.Request.Context(), womanID)
	if err != nil {
		h.log.Error("get mahrams error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not load mahrams", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mahrams})
}

// ── POST /v1/mahram/:id/verify ────────────────────────────────────────────

type verifyMahramRequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

// VerifyMahram handles POST /v1/mahram/:id/verify.
func (h *MahramHandler) VerifyMahram(c *gin.Context) {
	mahramID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_MAHRAM_ID", "mahram ID must be a valid UUID", nil)
		return
	}

	var req verifyMahramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	if err := h.mahramSvc.VerifyMahram(c.Request.Context(), mahramID, req.Code); err != nil {
		h.handleMahramError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mahram verified successfully"})
}

// handleMahramError maps domain errors to HTTP responses.
func (h *MahramHandler) handleMahramError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "mahram record not found", nil)
	case errors.Is(err, domain.ErrForbidden):
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "you do not have permission to perform this action", nil)
	case errors.Is(err, domain.ErrInvalidInput):
		errorResponse(c, http.StatusUnprocessableEntity, "INVALID_INPUT", err.Error(), nil)
	default:
		h.log.Error("mahram error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred", nil)
	}
}

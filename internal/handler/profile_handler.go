package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/handler/middleware"
	"github.com/trueconnect/backend/internal/pkg/validator"
	"github.com/trueconnect/backend/internal/service"
)

// ProfileHandler holds HTTP handlers for profile and photo endpoints.
type ProfileHandler struct {
	profileSvc *service.ProfileService
	log        *slog.Logger
}

// NewProfileHandler creates a new profile handler.
func NewProfileHandler(profileSvc *service.ProfileService, log *slog.Logger) *ProfileHandler {
	return &ProfileHandler{profileSvc: profileSvc, log: log}
}

// ── GET /v1/profiles/:id ─────────────────────────────────────────────────────

// GetProfile returns the public profile of any user by ID.
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "user ID must be a valid UUID", nil)
		return
	}

	view, err := h.profileSvc.GetProfile(c.Request.Context(), targetID)
	if err != nil {
		h.handleProfileError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": view})
}

// ── PUT /v1/profiles/me ───────────────────────────────────────────────────────

type upsertProfileRequest struct {
	DisplayName string  `json:"display_name" validate:"required,min=1,max=60"`
	Bio         string  `json:"bio"          validate:"max=500"`
	Gender      string  `json:"gender"       validate:"omitempty,oneof=male female other"`
	BirthDate   *string `json:"birth_date"` // ISO 8601 date: "1995-07-21"
	City        string  `json:"city"         validate:"max=100"`
	Latitude    float64 `json:"latitude"     validate:"min=-90,max=90"`
	Longitude   float64 `json:"longitude"    validate:"min=-180,max=180"`
	LookingFor  string  `json:"looking_for"  validate:"omitempty,oneof=male female other"`
}

// UpsertProfile handles PUT /v1/profiles/me.
func (h *ProfileHandler) UpsertProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	var req upsertProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}
	if err := validator.Validate.Struct(req); err != nil {
		errorResponse(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "input validation failed", validator.FormatErrors(err))
		return
	}

	input := service.UpsertProfileInput{
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
		Gender:      domain.Gender(req.Gender),
		City:        req.City,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		LookingFor:  domain.Gender(req.LookingFor),
	}

	if req.BirthDate != nil && *req.BirthDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			errorResponse(c, http.StatusUnprocessableEntity, "INVALID_BIRTH_DATE", "birth_date must be in YYYY-MM-DD format", nil)
			return
		}
		input.BirthDate = &parsed
	}

	profile, err := h.profileSvc.UpsertProfile(c.Request.Context(), userID, input)
	if err != nil {
		h.handleProfileError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": profile})
}

// ── GET /v1/profiles/me/photos ────────────────────────────────────────────────

// ListPhotos handles GET /v1/profiles/me/photos.
func (h *ProfileHandler) ListPhotos(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	photos, err := h.profileSvc.ListPhotos(c.Request.Context(), userID)
	if err != nil {
		h.handleProfileError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": photos})
}

// ── POST /v1/profiles/me/photos ──────────────────────────────────────────────

const maxUploadBytes = 10 * 1024 * 1024 // 10 MB

// UploadPhoto handles POST /v1/profiles/me/photos (multipart/form-data, field "photo").
func (h *ProfileHandler) UploadPhoto(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	file, _, err := c.Request.FormFile("photo")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "MISSING_FILE", "multipart field 'photo' is required", nil)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "READ_ERROR", "could not read uploaded file", nil)
		return
	}
	if int64(len(data)) > maxUploadBytes {
		errorResponse(c, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "file must be 10 MB or smaller", nil)
		return
	}

	media, err := h.profileSvc.UploadPhoto(c.Request.Context(), userID, data)
	if err != nil {
		h.handleProfileError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": media})
}

// ── DELETE /v1/profiles/me/photos/:id ────────────────────────────────────────

// DeletePhoto handles DELETE /v1/profiles/me/photos/:id.
func (h *ProfileHandler) DeletePhoto(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	photoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_PHOTO_ID", "photo ID must be a valid UUID", nil)
		return
	}

	if err := h.profileSvc.DeletePhoto(c.Request.Context(), userID, photoID); err != nil {
		h.handleProfileError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// handleProfileError maps domain errors to HTTP responses.
func (h *ProfileHandler) handleProfileError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "resource not found", nil)
	case errors.Is(err, domain.ErrForbidden):
		errorResponse(c, http.StatusForbidden, "FORBIDDEN", "you do not have permission to perform this action", nil)
	case errors.Is(err, domain.ErrInvalidInput):
		errorResponse(c, http.StatusUnprocessableEntity, "INVALID_INPUT", err.Error(), nil)
	default:
		h.log.Error("profile error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred", nil)
	}
}

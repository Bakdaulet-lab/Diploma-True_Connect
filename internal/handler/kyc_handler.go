package handler

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/handler/middleware"
	"github.com/trueconnect/backend/internal/repository"
)

const maxKYCDocBytes = 10 << 20 // 10 MB

// KYCHandler holds HTTP handlers for KYC document submission.
type KYCHandler struct {
	mediaStore repository.MediaStore
	userRepo   repository.UserRepository
	log        *slog.Logger
}

// NewKYCHandler creates a new KYC handler.
func NewKYCHandler(mediaStore repository.MediaStore, userRepo repository.UserRepository, log *slog.Logger) *KYCHandler {
	return &KYCHandler{mediaStore: mediaStore, userRepo: userRepo, log: log}
}

// SubmitKYC handles POST /v1/kyc/submit — accepts a document upload.
func (h *KYCHandler) SubmitKYC(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	file, header, err := c.Request.FormFile("document")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "MISSING_FILE", "document file is required", nil)
		return
	}
	defer file.Close()

	if header.Size > maxKYCDocBytes {
		errorResponse(c, http.StatusBadRequest, "FILE_TOO_LARGE", "document must be under 10 MB", nil)
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, maxKYCDocBytes+1))
	if err != nil {
		h.log.Error("kyc read file error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not read file", nil)
		return
	}
	if int64(len(data)) > maxKYCDocBytes {
		errorResponse(c, http.StatusBadRequest, "FILE_TOO_LARGE", "document must be under 10 MB", nil)
		return
	}

	mimeType := http.DetectContentType(data)
	switch mimeType {
	case "image/jpeg", "image/png", "application/pdf":
		// allowed
	default:
		errorResponse(c, http.StatusBadRequest, "UNSUPPORTED_FILE_TYPE",
			"document must be JPEG, PNG, or PDF", nil)
		return
	}

	objectKey, err := h.mediaStore.UploadDocument(c.Request.Context(), userID, data, mimeType)
	if err != nil {
		h.log.Error("kyc upload error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not upload document", nil)
		return
	}

	h.log.Info("kyc document submitted",
		slog.String("user_id", userID.String()),
		slog.String("object_key", objectKey),
		slog.String("filename", header.Filename),
	)

	c.JSON(http.StatusAccepted, gin.H{
		"data": gin.H{
			"status":  "pending",
			"message": "document received, verification in progress",
		},
	})
}

// GetStatus handles GET /v1/kyc/status — returns user's verification level.
func (h *KYCHandler) GetStatus(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("kyc get status error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not get status", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"verification_level": user.VerificationLevel,
			"is_verified":        user.VerificationLevel != domain.VerificationNone,
		},
	})
}

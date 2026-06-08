package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/handler/middleware"
	"github.com/trueconnect/backend/internal/provider"
	"github.com/trueconnect/backend/internal/repository"
)

const maxKYCDocBytes = 10 << 20 // 10 MB

// KYCHandler holds HTTP handlers for KYC document submission.
type KYCHandler struct {
	mediaStore  repository.MediaStore
	userRepo    repository.UserRepository
	kycProvider provider.KYCProvider
	log         *slog.Logger
	serverCtx   context.Context
}

// NewKYCHandler creates a new KYC handler.
func NewKYCHandler(mediaStore repository.MediaStore, userRepo repository.UserRepository, kycProvider provider.KYCProvider, log *slog.Logger, serverCtx context.Context) *KYCHandler {
	return &KYCHandler{mediaStore: mediaStore, userRepo: userRepo, kycProvider: kycProvider, log: log, serverCtx: serverCtx}
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
	case "image/jpeg", "image/png":
		// allowed
	default:
		errorResponse(c, http.StatusBadRequest, "UNSUPPORTED_FILE_TYPE",
			"document must be JPEG or PNG", nil)
		return
	}

	objectKey, err := h.mediaStore.UploadDocument(c.Request.Context(), userID, data, mimeType)
	if err != nil {
		h.log.Error("kyc upload error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not upload document", nil)
		return
	}

	// Insert into verifications table as pending
	if err := h.userRepo.SubmitKYCRequest(c.Request.Context(), userID, objectKey); err != nil {
		h.log.Error("failed to insert pending kyc request", slog.String("error", err.Error()))
		// don't fail the whole user response if they successfully uploaded
	}

	h.log.Info("kyc document submitted",
		slog.String("user_id", userID.String()),
		slog.String("object_key", objectKey),
		slog.String("filename", header.Filename),
	)

	// In a real system, this could be triggered via a worker task. For now, running in a goroutine.
	go func(uid uuid.UUID, key string, docData []byte, mime string) {
		ctx, cancel := context.WithTimeout(h.serverCtx, 2*time.Minute)
		defer cancel()
		level, verifyErr := h.kycProvider.VerifyDocument(ctx, uid, key, docData, mime)
		if verifyErr != nil {
			h.log.Error("external kyc verification failed", slog.String("error", verifyErr.Error()))
			return
		}

		if level != domain.VerificationNone {
			if updateErr := h.userRepo.UpdateVerificationLevel(ctx, uid, level); updateErr != nil {
				h.log.Error("failed to update user verification level", slog.String("error", updateErr.Error()))
			} else {
				h.log.Info("kyc verified synchronously", slog.String("user_id", uid.String()), slog.Any("level", level))
			}
		} else {
			h.log.Info("kyc verification async tracking started via provider", slog.String("user_id", uid.String()))
		}
	}(userID, objectKey, data, mimeType)

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


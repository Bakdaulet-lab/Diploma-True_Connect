package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/repository"
	"github.com/trueconnect/backend/internal/service"
)

type AdminHandler struct {
	userSvc    *service.UserService
	reputeSvc  *service.ReputationService
	adminSvc   *service.AdminService
	mediaStore repository.MediaStore
	log        *slog.Logger
}

func NewAdminHandler(userSvc *service.UserService, reputeSvc *service.ReputationService, adminSvc *service.AdminService, mediaStore repository.MediaStore, log *slog.Logger) *AdminHandler {
	return &AdminHandler{userSvc: userSvc, reputeSvc: reputeSvc, adminSvc: adminSvc, mediaStore: mediaStore, log: log}
}

func (h *AdminHandler) GetAnalytics(c *gin.Context) {
	stats, err := h.adminSvc.GetDashboardStats(c.Request.Context())
	if err != nil {
		h.log.Error("failed to get analytics", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch analytics", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

func (h *AdminHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, err := h.adminSvc.SearchUsers(c.Request.Context(), query, limit, offset)
	if err != nil {
		h.log.Error("failed to search users", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not search users", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *AdminHandler) ListUnderReview(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, err := h.userSvc.ListUsersUnderReview(c.Request.Context(), limit, offset)
	if err != nil {
		h.log.Error("list under review error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch users", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *AdminHandler) ReviewVerdict(c *gin.Context) {
	idStr := c.Param("id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "invalid user id", nil)
		return
	}

	var req struct {
		Action string `json:"action" binding:"required"` // 'ban' or 'restore'
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid action payload", nil)
		return
	}

	isBan := req.Action == "ban"
	if req.Action != "ban" && req.Action != "restore" {
		errorResponse(c, http.StatusBadRequest, "INVALID_ACTION", "action must be ban or restore", nil)
		return
	}

	if err := h.userSvc.ReviewSybilVerdict(c.Request.Context(), targetID, isBan); err != nil {
		h.log.Error("review verdict error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not apply verdict", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "verdict applied"}})
}

func (h *AdminHandler) GetSybilClusters(c *gin.Context) {
	clusters, err := h.reputeSvc.GetSybilClusters(c.Request.Context())
	if err != nil {
		h.log.Error("failed to detect sybil clusters", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not process graph clusters", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": clusters})
}

func (h *AdminHandler) GetPendingKYC(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	reqs, err := h.adminSvc.GetPendingKYC(c.Request.Context(), limit, offset)
	if err != nil {
		h.log.Error("failed to get pending KYC", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list KYC requests", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": reqs})
}

func (h *AdminHandler) GetKYCDocument(c *gin.Context) {
	kycID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "invalid kyc request ID", nil)
		return
	}
	urlKey, err := h.adminSvc.GetKYCDocument(c.Request.Context(), kycID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "kyc document not found", nil)
		return
	}

	presigned, err := h.mediaStore.PresignedURL(c.Request.Context(), urlKey)
	if err != nil {
		h.log.Error("failed to generate presigned document URL", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to secure document link", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]string{"url": presigned}})
}

func (h *AdminHandler) ReviewKYC(c *gin.Context) {
	kycID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "invalid kyc request ID", nil)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=approved rejected"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "status must be 'approved' or 'rejected'", nil)
		return
	}

	userID, err := h.adminSvc.UpdateKYCStatus(c.Request.Context(), kycID, req.Status)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update status", nil)
		return
	}

	if req.Status == "approved" {
		err = h.reputeSvc.RecalculateDepth1Targets(c.Request.Context(), userID)
		if err != nil {
			h.log.Error("failed to recalculate reputation for approved user", slog.String("userId", userID.String()), slog.String("error", err.Error()))
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]string{"status": "updated"}})
}

func (h *AdminHandler) GetPendingSybilClusters(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	clusters, err := h.adminSvc.GetPendingSybilClusters(c.Request.Context(), limit, offset)
	if err != nil {
		h.log.Error("failed to get pending sybil clusters", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list sybil clusters", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": clusters})
}

func (h *AdminHandler) ResolveSybilCluster(c *gin.Context) {
	clusterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "invalid cluster ID", nil)
		return
	}

	var req struct {
		Action string `json:"action" binding:"required,oneof=ban dismiss"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "action must be 'ban' or 'dismiss'", nil)
		return
	}

	err = h.adminSvc.ResolveSybilCluster(c.Request.Context(), clusterID, req.Action)
	if err != nil {
		h.log.Error("failed to resolve sybil cluster", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to resolve cluster", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]string{"status": "resolved"}})
}

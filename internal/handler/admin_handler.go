package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/service"
)

type AdminHandler struct {
	userSvc   *service.UserService
	reputeSvc *service.ReputationService
	adminSvc  *service.AdminService
	log       *slog.Logger
}

func NewAdminHandler(userSvc *service.UserService, reputeSvc *service.ReputationService, adminSvc *service.AdminService, log *slog.Logger) *AdminHandler {
	return &AdminHandler{userSvc: userSvc, reputeSvc: reputeSvc, adminSvc: adminSvc, log: log}
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

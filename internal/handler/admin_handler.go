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
	userSvc *service.UserService
	log     *slog.Logger
}

func NewAdminHandler(userSvc *service.UserService, log *slog.Logger) *AdminHandler {
	return &AdminHandler{userSvc: userSvc, log: log}
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

package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/handler/middleware"
	"github.com/trueconnect/backend/internal/service"
)

type NotificationHandler struct {
	notifSvc *service.NotificationService
	log      *slog.Logger
}

func NewNotificationHandler(notifSvc *service.NotificationService, log *slog.Logger) *NotificationHandler {
	return &NotificationHandler{
		notifSvc: notifSvc,
		log:      log,
	}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return
	}

	cursor, limit := parseCursorPagination(c)
	notifs, nextCursor, err := h.notifSvc.List(c.Request.Context(), userID, cursor, limit)
	if err != nil {
		h.log.Error("list notifications error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list notifications", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": notifs,
		"meta": gin.H{"next_cursor": nextCursor, "limit": limit},
	})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return
	}

	idParam := c.Param("id")
	notifID, err := uuid.Parse(idParam)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "invalid notification id", nil)
		return
	}

	if err := h.notifSvc.MarkAsRead(c.Request.Context(), notifID, userID); err != nil {
		h.log.Error("mark notif read error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to mark notification as read", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return
	}

	if err := h.notifSvc.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		h.log.Error("mark all notifs read error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to mark all notifications as read", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return
	}

	count, err := h.notifSvc.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("get unread notifs count error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get unread count", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

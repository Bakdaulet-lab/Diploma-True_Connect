package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trueconnect/backend/internal/handler/middleware"
	"github.com/trueconnect/backend/internal/service"
)

// UserHandler holds HTTP handlers for user account operations.
type UserHandler struct {
	userSvc *service.UserService
	log     *slog.Logger
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userSvc *service.UserService, log *slog.Logger) *UserHandler {
	return &UserHandler{userSvc: userSvc, log: log}
}

// GetMe handles GET /v1/users/me
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	user, err := h.userSvc.GetMe(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("get me error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not get user", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// DeleteMe handles DELETE /v1/users/me
func (h *UserHandler) DeleteMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	if err := h.userSvc.DeleteMe(c.Request.Context(), userID); err != nil {
		h.log.Error("delete me error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not delete account", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{"message": "account deleted"},
	})
}

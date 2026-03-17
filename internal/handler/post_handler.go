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

// PostHandler holds HTTP handlers for the social feed.
type PostHandler struct {
	postSvc *service.PostService
	log     *slog.Logger
}

// NewPostHandler creates a new post handler.
func NewPostHandler(postSvc *service.PostService, log *slog.Logger) *PostHandler {
	return &PostHandler{postSvc: postSvc, log: log}
}

// CreatePost handles POST /v1/posts
func (h *PostHandler) CreatePost(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	var req struct {
		Content  string `json:"content" binding:"required"`
		MediaURL string `json:"media_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	post, err := h.postSvc.CreatePost(c.Request.Context(), userID, req.Content, req.MediaURL)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			errorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "content must be 1-2000 characters", nil)
			return
		}
		h.log.Error("create post error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create post", nil)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": post})
}

// GetPost handles GET /v1/posts/:id
func (h *PostHandler) GetPost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID", nil)
		return
	}

	post, err := h.postSvc.GetPost(c.Request.Context(), postID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			errorResponse(c, http.StatusNotFound, "NOT_FOUND", "post not found", nil)
			return
		}
		h.log.Error("get post error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not get post", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": post})
}

// DeletePost handles DELETE /v1/posts/:id
func (h *PostHandler) DeletePost(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID", nil)
		return
	}

	if err := h.postSvc.DeletePost(c.Request.Context(), postID, userID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			errorResponse(c, http.StatusNotFound, "NOT_FOUND", "post not found or not owned by you", nil)
			return
		}
		h.log.Error("delete post error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not delete post", nil)
		return
	}

	c.Status(http.StatusNoContent)
}

// ListFeed handles GET /v1/posts
func (h *PostHandler) ListFeed(c *gin.Context) {
	page, perPage := parsePagination(c)

	posts, err := h.postSvc.ListFeed(c.Request.Context(), page, perPage)
	if err != nil {
		h.log.Error("list feed error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not list feed", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": posts,
		"meta": gin.H{"page": page, "per_page": perPage},
	})
}

// LikePost handles POST /v1/posts/:id/like
func (h *PostHandler) LikePost(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID", nil)
		return
	}

	if err := h.postSvc.LikePost(c.Request.Context(), postID, userID); err != nil {
		h.log.Error("like post error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not like post", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"liked": true}})
}

// UnlikePost handles DELETE /v1/posts/:id/like
func (h *PostHandler) UnlikePost(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID", nil)
		return
	}

	if err := h.postSvc.UnlikePost(c.Request.Context(), postID, userID); err != nil {
		h.log.Error("unlike post error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not unlike post", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"liked": false}})
}

// CreateComment handles POST /v1/posts/:id/comments
func (h *PostHandler) CreateComment(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID", nil)
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	comment, err := h.postSvc.CreateComment(c.Request.Context(), postID, userID, req.Content)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			errorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "content must be 1-500 characters", nil)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			errorResponse(c, http.StatusNotFound, "NOT_FOUND", "post not found", nil)
			return
		}
		h.log.Error("create comment error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create comment", nil)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": comment})
}

// ListComments handles GET /v1/posts/:id/comments
func (h *PostHandler) ListComments(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID", nil)
		return
	}

	page, perPage := parsePagination(c)

	comments, err := h.postSvc.ListComments(c.Request.Context(), postID, page, perPage)
	if err != nil {
		h.log.Error("list comments error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not list comments", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": comments,
		"meta": gin.H{"page": page, "per_page": perPage},
	})
}

package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/pkg/validator"
	"github.com/trueconnect/backend/internal/service"
)

// AuthHandler holds the HTTP handlers for authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
	log         *slog.Logger
	devMode     bool
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authService *service.AuthService, log *slog.Logger, devMode bool) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		log:         log,
		devMode:     devMode,
	}
}

type registerRequest struct {
	Phone     string  `json:"phone" validate:"required,kz_phone"`
	Password  string  `json:"password" validate:"required,min=8,max=128"`
	PublicKey *string `json:"public_key"` // Optional X25519 Public Key
}

type loginRequest struct {
	Phone    string `json:"phone" validate:"required,kz_phone"`
	Password string `json:"password" validate:"required"`
}

type authResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	UserID      string `json:"user_id"`
}

// Register handles POST /v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	if err := validator.Validate.Struct(req); err != nil {
		details := validator.FormatErrors(err)
		errorResponse(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "input validation failed", details)
		return
	}

	result, err := h.authService.Register(c.Request.Context(), service.RegisterInput{
		Phone:     req.Phone,
		Password:  req.Password,
		PublicKey: req.PublicKey,
	})
	if err != nil {
		h.handleAuthError(c, err, "register")
		return
	}

	h.setRefreshTokenCookie(c, result.RefreshToken)

	c.JSON(http.StatusCreated, gin.H{
		"data": authResponse{
			AccessToken: result.AccessToken,
			ExpiresIn:   result.ExpiresIn,
			UserID:      result.UserID.String(),
		},
	})
}

// Login handles POST /v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	if err := validator.Validate.Struct(req); err != nil {
		details := validator.FormatErrors(err)
		errorResponse(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "input validation failed", details)
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		h.handleAuthError(c, err, "login")
		return
	}

	h.setRefreshTokenCookie(c, result.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"data": authResponse{
			AccessToken: result.AccessToken,
			ExpiresIn:   result.ExpiresIn,
			UserID:      result.UserID.String(),
		},
	})
}

// Refresh handles POST /v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		errorResponse(c, http.StatusUnauthorized, "MISSING_REFRESH_TOKEN", "refresh token cookie is required", nil)
		return
	}

	result, err := h.authService.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		h.handleAuthError(c, err, "refresh")
		return
	}

	h.setRefreshTokenCookie(c, result.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"data": authResponse{
			AccessToken: result.AccessToken,
			ExpiresIn:   result.ExpiresIn,
			UserID:      result.UserID.String(),
		},
	})
}

// Logout handles POST /v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		// Already logged out or no cookie — just clear and return success
		h.clearRefreshTokenCookie(c)
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "logged out"}})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), refreshToken); err != nil {
		h.log.Warn("logout error", slog.String("error", err.Error()))
	}

	h.clearRefreshTokenCookie(c)

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "logged out"}})
}

// VerifyPhone handles POST /v1/auth/verify-phone (stub for development)
func (h *AuthHandler) VerifyPhone(c *gin.Context) {
	if !h.devMode {
		errorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED",
			"phone verification is not available in production", nil)
		return
	}

	var req struct {
		Code string `json:"code" validate:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	// DEV STUB: In development, any 6-digit code is accepted.
	if len(req.Code) != 6 {
		errorResponse(c, http.StatusUnprocessableEntity, "INVALID_CODE", "code must be 6 digits", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{"verified": true, "message": "phone verified (dev stub)"},
	})
}

// handleAuthError maps domain errors to appropriate HTTP responses.
func (h *AuthHandler) handleAuthError(c *gin.Context, err error, op string) {
	switch {
	case errors.Is(err, domain.ErrAlreadyExists):
		errorResponse(c, http.StatusConflict, "PHONE_ALREADY_REGISTERED", "this phone number is already registered", nil)
	case errors.Is(err, domain.ErrInvalidCredentials):
		errorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "phone or password is incorrect", nil)
	case errors.Is(err, domain.ErrRateLimitExceeded):
		errorResponse(c, http.StatusTooManyRequests, "TOO_MANY_ATTEMPTS", "too many failed attempts, try again later", nil)
	case errors.Is(err, domain.ErrAccountSuspended):
		errorResponse(c, http.StatusForbidden, "ACCOUNT_SUSPENDED", "your account has been suspended", nil)
	case errors.Is(err, domain.ErrUnauthorized):
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token", nil)
	default:
		h.log.Error("auth error", slog.String("op", op), slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred", nil)
	}
}

// setRefreshTokenCookie sets the HttpOnly refresh token cookie.
func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",
		token,
		7*24*3600,  // 7 days in seconds
		"/v1/auth", // only sent to auth endpoints
		"",         // domain (empty = current)
		!h.devMode, // secure=false in dev (http://localhost), true in production
		true,       // httpOnly
	)
}

// clearRefreshTokenCookie removes the refresh token cookie.
func (h *AuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/v1/auth",
		"",
		!h.devMode,
		true,
	)
}

// errorResponse writes a standardized error JSON response.
func errorResponse(c *gin.Context, status int, code, message string, details any) {
	resp := gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}
	if details != nil {
		resp["error"].(gin.H)["details"] = details
	}
	c.JSON(status, resp)
}

package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/handler/middleware"
	"github.com/trueconnect/backend/internal/pkg/validator"
	"github.com/trueconnect/backend/internal/service"
)

// ReportHandler holds HTTP handlers for user reports.
type ReportHandler struct {
	reportSvc *service.ReportService
	log       *slog.Logger
}

// NewReportHandler creates a new report handler.
func NewReportHandler(reportSvc *service.ReportService, log *slog.Logger) *ReportHandler {
	return &ReportHandler{reportSvc: reportSvc, log: log}
}

type createReportRequest struct {
	ReportedID  string `json:"reported_id" validate:"required,uuid"`
	Reason      string `json:"reason" validate:"required,max=50"`
	Description string `json:"description" validate:"max=500"`
}

// CreateReport handles POST /v1/reports.
func (h *ReportHandler) CreateReport(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	var req createReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	if err := validator.Validate.Struct(req); err != nil {
		details := validator.FormatErrors(err)
		errorResponse(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "input validation failed", details)
		return
	}

	reportedID, err := uuid.Parse(req.ReportedID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_UUID", "reported_id must be a valid UUID", nil)
		return
	}

	report, err := h.reportSvc.CreateReport(c.Request.Context(), service.CreateReportInput{
		ReporterID:  userID,
		ReportedID:  reportedID,
		Reason:      req.Reason,
		Description: req.Description,
	})
	if err != nil {
		h.handleReportError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"id":         report.ID.String(),
			"status":     report.Status,
			"created_at": report.CreatedAt,
		},
	})
}

func (h *ReportHandler) handleReportError(c *gin.Context, err error) {
	errStr := err.Error()
	switch {
	case contains(errStr, "cannot report yourself"):
		errorResponse(c, http.StatusBadRequest, "SELF_REPORT", "you cannot report yourself", nil)
	case contains(errStr, "already reported"):
		errorResponse(c, http.StatusConflict, "ALREADY_REPORTED", "you have already reported this user", nil)
	default:
		h.log.Error("report error", slog.String("error", err.Error()))
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create report", nil)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

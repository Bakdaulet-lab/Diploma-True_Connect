package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/service"
)

// ── Mock ReportRepository ────────────────────────────────────────────────────

type mockReportRepo struct {
	mu      sync.Mutex
	reports map[string]*domain.Report // keyed by "reporter_id:reported_id"
}

func newMockReportRepo() *mockReportRepo {
	return &mockReportRepo{reports: make(map[string]*domain.Report)}
}

func (m *mockReportRepo) key(reporterID, reportedID uuid.UUID) string {
	return reporterID.String() + ":" + reportedID.String()
}

func (m *mockReportRepo) Create(_ context.Context, report *domain.Report) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reports[m.key(report.ReporterID, report.ReportedID)] = report
	return nil
}

func (m *mockReportRepo) ExistsBetween(_ context.Context, reporterID, reportedID uuid.UUID) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.reports[m.key(reporterID, reportedID)]
	return exists, nil
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestCreateReport_Success(t *testing.T) {
	t.Parallel()

	reportRepo := newMockReportRepo()
	graphRepo := &mockGraphRepo{}
	svc := service.NewReportService(reportRepo, graphRepo)

	reporterID := uuid.New()
	reportedID := uuid.New()

	report, err := svc.CreateReport(context.Background(), service.CreateReportInput{
		ReporterID:  reporterID,
		ReportedID:  reportedID,
		Reason:      "spam",
		Description: "Sending unsolicited messages",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.ID == uuid.Nil {
		t.Error("expected non-nil report ID")
	}
	if report.ReporterID != reporterID {
		t.Errorf("expected reporter_id %s, got %s", reporterID, report.ReporterID)
	}
	if report.ReportedID != reportedID {
		t.Errorf("expected reported_id %s, got %s", reportedID, report.ReportedID)
	}
	if report.Status != domain.ReportStatusPending {
		t.Errorf("expected status pending, got %s", report.Status)
	}
}

func TestCreateReport_SelfReport_Fails(t *testing.T) {
	t.Parallel()

	svc := service.NewReportService(newMockReportRepo(), &mockGraphRepo{})
	userID := uuid.New()

	_, err := svc.CreateReport(context.Background(), service.CreateReportInput{
		ReporterID:  userID,
		ReportedID:  userID,
		Reason:      "spam",
		Description: "test",
	})

	if err == nil {
		t.Fatal("expected error for self-report")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestCreateReport_Duplicate_Fails(t *testing.T) {
	t.Parallel()

	reportRepo := newMockReportRepo()
	svc := service.NewReportService(reportRepo, &mockGraphRepo{})

	reporterID := uuid.New()
	reportedID := uuid.New()

	input := service.CreateReportInput{
		ReporterID:  reporterID,
		ReportedID:  reportedID,
		Reason:      "harassment",
		Description: "first report",
	}

	// First report should succeed.
	if _, err := svc.CreateReport(context.Background(), input); err != nil {
		t.Fatalf("first report failed: %v", err)
	}

	// Second report should fail.
	_, err := svc.CreateReport(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for duplicate report")
	}
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got: %v", err)
	}
}

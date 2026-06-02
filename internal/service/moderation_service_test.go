package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/trueconnect/backend/internal/repository"
	"github.com/trueconnect/backend/internal/service"
)

// mockModerationProvider is a configurable stub for repository.ModerationProvider.
type mockModerationProvider struct {
	result *repository.ModerationResult
	err    error
}

func (m *mockModerationProvider) Check(ctx context.Context, text string) (*repository.ModerationResult, error) {
	return m.result, m.err
}

func TestModeration_ProviderBlocks(t *testing.T) {
	svc := service.NewModerationService(
		&mockModerationProvider{result: &repository.ModerationResult{Block: true, Label: "block"}},
		nil,
	)
	block, warn := svc.Check(context.Background(), "anything")
	if !block || warn {
		t.Fatalf("expected block=true warn=false, got block=%v warn=%v", block, warn)
	}
}

func TestModeration_ProviderWarns(t *testing.T) {
	svc := service.NewModerationService(
		&mockModerationProvider{result: &repository.ModerationResult{Warn: true, Label: "warn"}},
		nil,
	)
	block, warn := svc.Check(context.Background(), "anything")
	if block || !warn {
		t.Fatalf("expected block=false warn=true, got block=%v warn=%v", block, warn)
	}
}

func TestModeration_ProviderClean(t *testing.T) {
	svc := service.NewModerationService(
		&mockModerationProvider{result: &repository.ModerationResult{Label: "clean"}},
		nil,
	)
	block, warn := svc.Check(context.Background(), "Сәлеметсіз бе, қалыңыз қалай?")
	if block || warn {
		t.Fatalf("expected clean (false,false), got block=%v warn=%v", block, warn)
	}
}

// When the ML provider errors, the service must fall back to the keyword filter.
func TestModeration_FallbackOnProviderError(t *testing.T) {
	svc := service.NewModerationService(
		&mockModerationProvider{err: errors.New("service unavailable")},
		nil,
	)

	// "секс" is in the keyword blocklist → should block via fallback.
	if block, _ := svc.Check(context.Background(), "давай про секс"); !block {
		t.Fatal("expected keyword fallback to block explicit content on provider error")
	}
	// Clean text → not blocked, not warned.
	if block, warn := svc.Check(context.Background(), "сәлем, таныса аламыз ба?"); block || warn {
		t.Fatalf("expected clean text to pass, got block=%v warn=%v", block, warn)
	}
}

// With no provider configured, the service uses the keyword filter only.
func TestModeration_NilProviderUsesKeyword(t *testing.T) {
	svc := service.NewModerationService(nil, nil)

	if block, _ := svc.Check(context.Background(), "porn here"); !block {
		t.Fatal("expected keyword filter to block 'porn'")
	}
	if _, warn := svc.Check(context.Background(), "this is a scam"); !warn {
		t.Fatal("expected keyword filter to warn on 'scam'")
	}
}

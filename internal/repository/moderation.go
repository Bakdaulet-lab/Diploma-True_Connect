package repository

import "context"

// ModerationResult is the verdict returned by a content-moderation provider.
type ModerationResult struct {
	// Block means the content must be rejected outright (explicit/sexual,
	// threats, solicitation).
	Block bool
	// Warn means the content is allowed but flagged toxic (insult/harassment),
	// which triggers a trust-score penalty.
	Warn bool
	// Label is the single collapsed class: "clean" | "warn" | "block".
	Label string
	// Scores holds the raw per-class model probabilities (clean/warn/block),
	// useful for logging, tuning thresholds, and the diploma evaluation.
	Scores map[string]float64
	// LangDetected is the language the model detected (e.g. "kk", "ru", "en").
	LangDetected string
}

// ModerationProvider is the port for an external ML content-moderation service.
// The concrete implementation lives in internal/adapter/moderation and talks to
// the Python (FastAPI) classifier over HTTP. Services depend only on this port.
type ModerationProvider interface {
	// Check classifies a piece of user text. An error means the provider was
	// unreachable/timed out; callers should fall back to the keyword filter.
	Check(ctx context.Context, text string) (*ModerationResult, error)
}

// Package moderation provides an HTTP client adapter for the external ML
// content-moderation microservice (a Python/FastAPI XLM-RoBERTa classifier).
// It implements repository.ModerationProvider.
package moderation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/trueconnect/backend/internal/repository"
)

// MLProvider talks to the moderation microservice over HTTP.
type MLProvider struct {
	baseURL string
	log     *slog.Logger
	client  *http.Client
}

var _ repository.ModerationProvider = (*MLProvider)(nil)

// NewMLProvider creates a moderation provider pointed at baseURL
// (e.g. "http://moderation:8000"). A short timeout keeps chat latency low; on
// timeout the calling service falls back to the keyword filter.
func NewMLProvider(baseURL string, log *slog.Logger) *MLProvider {
	return &MLProvider{
		baseURL: baseURL,
		log:     log,
		client:  &http.Client{Timeout: 1500 * time.Millisecond},
	}
}

type moderateRequest struct {
	Text string `json:"text"`
}

type moderateResponse struct {
	Label        string             `json:"label"` // "clean" | "warn" | "block"
	Scores       map[string]float64 `json:"scores"`
	LangDetected string             `json:"lang_detected"`
}

// Check sends the text to the ML service and maps the label to block/warn flags.
func (p *MLProvider) Check(ctx context.Context, text string) (*repository.ModerationResult, error) {
	body, err := json.Marshal(moderateRequest{Text: text})
	if err != nil {
		return nil, fmt.Errorf("moderation: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/moderate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("moderation: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("moderation: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("moderation: service status=%d body=%s", resp.StatusCode, string(b))
	}

	var out moderateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("moderation: decode response: %w", err)
	}

	return &repository.ModerationResult{
		Block:        out.Label == "block",
		Warn:         out.Label == "warn",
		Label:        out.Label,
		Scores:       out.Scores,
		LangDetected: out.LangDetected,
	}, nil
}

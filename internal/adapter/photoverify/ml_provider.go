// Package photoverify is an HTTP client adapter for the external photo
// verification microservice (face presence + NSFW). Implements
// repository.PhotoVerifier.
package photoverify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/trueconnect/backend/internal/repository"
)

// MLProvider talks to the photo-verifier service over HTTP.
type MLProvider struct {
	baseURL string
	log     *slog.Logger
	client  *http.Client
}

var _ repository.PhotoVerifier = (*MLProvider)(nil)

// NewMLProvider creates a verifier pointed at baseURL (e.g. "http://photo-verifier:8000").
func NewMLProvider(baseURL string, log *slog.Logger) *MLProvider {
	return &MLProvider{
		baseURL: baseURL,
		log:     log,
		client:  &http.Client{Timeout: 8 * time.Second},
	}
}

type verifyResponse struct {
	HasFace   bool     `json:"has_face"`
	FaceCount int      `json:"face_count"`
	NSFWScore float64  `json:"nsfw_score"`
	Decision  string   `json:"decision"` // "approve" | "reject"
	Reasons   []string `json:"reasons"`
}

// Verify uploads the image to the service as multipart field "image".
func (p *MLProvider) Verify(ctx context.Context, imageData []byte) (*repository.PhotoVerdict, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("image", "upload.jpg")
	if err != nil {
		return nil, fmt.Errorf("photoverify: form: %w", err)
	}
	if _, err := fw.Write(imageData); err != nil {
		return nil, fmt.Errorf("photoverify: write: %w", err)
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/verify", &body)
	if err != nil {
		return nil, fmt.Errorf("photoverify: request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("photoverify: call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("photoverify: status=%d body=%s", resp.StatusCode, string(b))
	}

	var out verifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("photoverify: decode: %w", err)
	}

	return &repository.PhotoVerdict{
		Approved:  out.Decision == "approve",
		HasFace:   out.HasFace,
		FaceCount: out.FaceCount,
		NSFWScore: out.NSFWScore,
		Reasons:   out.Reasons,
	}, nil
}

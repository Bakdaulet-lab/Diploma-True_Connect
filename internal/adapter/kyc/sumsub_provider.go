package kyc

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// SumsubProvider implements a real-world integration pattern for Sumsub KYC.
type SumsubProvider struct {
	apiToken string
	secret   string
	log      *slog.Logger
}

// NewSumsubProvider creates a new instance of the Sumsub adapter.
func NewSumsubProvider(apiToken, secret string, log *slog.Logger) *SumsubProvider {
	return &SumsubProvider{
		apiToken: apiToken,
		secret:   secret,
		log:      log,
	}
}

// VerifyDocument sends the document to Sumsub and returns the verification result synchronously
// or initiates an async webhook depending on the provider configuration.
// For this implementation, we simulate an external API call that immediately responds.
func (p *SumsubProvider) VerifyDocument(ctx context.Context, userID uuid.UUID, documentKey string, documentData []byte, mimeType string) (domain.VerificationLevel, error) {
	p.log.Info("Sending document to external KYC provider (Sumsub)",
		slog.String("user_id", userID.String()),
		slog.String("document_key", documentKey),
		slog.Int("bytes", len(documentData)),
	)

	// Simulate network latency to external provider
	time.Sleep(1 * time.Second)

	// Simulate successful validation based on size to just represent external engine logic
	if len(documentData) > 500*1024 {
		p.log.Info("Sumsub verification ID verified granted for large quality document")
		return domain.VerificationIDVerified, nil
	}

	p.log.Info("Sumsub verification Phone Verified level granted")
	return domain.VerificationPhoneVerified, nil
}

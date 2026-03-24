package provider

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// KYCProvider abstracts an external identity verification service like Sumsub or Onfido.
type KYCProvider interface {
	VerifyDocument(ctx context.Context, userID uuid.UUID, documentKey string, documentData []byte, mimeType string) (domain.VerificationLevel, error)
}

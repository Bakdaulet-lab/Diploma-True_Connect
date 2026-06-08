package kyc

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// PhotoProvider implements KYCProvider using the local photo-verification ML
// microservice. Verification is synchronous: a clear face in a non-NSFW image
// is sufficient to grant VerificationPhotoVerified.
type PhotoProvider struct {
	verifier repository.PhotoVerifier
	log      *slog.Logger
}

// NewPhotoProvider creates a KYC provider backed by the photo-verifier service.
func NewPhotoProvider(verifier repository.PhotoVerifier, log *slog.Logger) *PhotoProvider {
	return &PhotoProvider{verifier: verifier, log: log}
}

// VerifyDocument runs the photo-verification ML pipeline on documentData.
// Only JPEG and PNG are supported; PDF documents are rejected.
// Returns VerificationPhotoVerified on approval, VerificationNone + error on rejection.
// If the underlying verifier is nil (service URL not configured), returns an error.
func (p *PhotoProvider) VerifyDocument(ctx context.Context, userID uuid.UUID, _ string, documentData []byte, mimeType string) (domain.VerificationLevel, error) {
	if p.verifier == nil {
		return domain.VerificationNone, fmt.Errorf("kyc: photo verifier service is not configured (PHOTO_VERIFIER_URL not set)")
	}

	if mimeType != "image/jpeg" && mimeType != "image/png" {
		return domain.VerificationNone, fmt.Errorf("kyc: unsupported mime type %q — only JPEG and PNG are accepted", mimeType)
	}

	verdict, err := p.verifier.Verify(ctx, documentData)
	if err != nil {
		p.log.Error("kyc photo verification call failed",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return domain.VerificationNone, fmt.Errorf("kyc: photo verifier error: %w", err)
	}

	p.log.Info("kyc photo verification result",
		slog.String("user_id", userID.String()),
		slog.Bool("approved", verdict.Approved),
		slog.Bool("has_face", verdict.HasFace),
		slog.Float64("nsfw_score", verdict.NSFWScore),
		slog.Any("reasons", verdict.Reasons),
	)

	if !verdict.Approved {
		return domain.VerificationNone, fmt.Errorf("kyc: document rejected: %v", verdict.Reasons)
	}

	return domain.VerificationPhotoVerified, nil
}

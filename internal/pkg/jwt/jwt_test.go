package jwt_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
	tcjwt "github.com/trueconnect/backend/internal/pkg/jwt"
)

func newTestManager(expiry time.Duration) *tcjwt.Manager {
	return tcjwt.NewManager("test-secret-at-least-32-chars-long!", expiry)
}

func TestGenerate_ReturnsNonEmptyToken(t *testing.T) {
	t.Parallel()

	m := newTestManager(15 * time.Minute)
	token, err := m.Generate(uuid.New(), domain.VerificationNone, domain.TrustStatusNormal, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestGenerate_ProducesThreeParts(t *testing.T) {
	t.Parallel()

	m := newTestManager(15 * time.Minute)
	token, err := m.Generate(uuid.New(), domain.VerificationNone, domain.TrustStatusNormal, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("expected 3 JWT parts, got %d", len(parts))
	}
}

func TestVerify_ValidToken(t *testing.T) {
	t.Parallel()

	m := newTestManager(15 * time.Minute)
	userID := uuid.New()

	token, err := m.Generate(userID, domain.VerificationPhoneVerified, domain.TrustStatusNormal, false)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	claims, err := m.Verify(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if claims.Subject != userID.String() {
		t.Errorf("subject: got %q, want %q", claims.Subject, userID.String())
	}
	if claims.VerificationLevel != domain.VerificationPhoneVerified {
		t.Errorf("verification level: got %q, want %q", claims.VerificationLevel, domain.VerificationPhoneVerified)
	}
	if claims.TrustStatus != domain.TrustStatusNormal {
		t.Errorf("trust status: got %q, want %q", claims.TrustStatus, domain.TrustStatusNormal)
	}
}

func TestVerify_ExpiredToken(t *testing.T) {
	t.Parallel()

	// Issue a token that expired 1 second ago.
	m := newTestManager(-1 * time.Second)
	token, err := m.Generate(uuid.New(), domain.VerificationNone, domain.TrustStatusNormal, false)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = m.Verify(token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestVerify_TamperedToken(t *testing.T) {
	t.Parallel()

	m := newTestManager(15 * time.Minute)
	token, err := m.Generate(uuid.New(), domain.VerificationNone, domain.TrustStatusNormal, false)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	// Tamper with the signature part (last segment).
	parts := strings.Split(token, ".")
	parts[2] = parts[2] + "tampered"
	tampered := strings.Join(parts, ".")

	_, err = m.Verify(tampered)
	if err == nil {
		t.Error("expected error for tampered signature")
	}
}

func TestVerify_WrongSecret(t *testing.T) {
	t.Parallel()

	issuer := tcjwt.NewManager("secret-used-to-sign-xxxxxxxxxxxxxxxxx", 15*time.Minute)
	verifier := tcjwt.NewManager("different-secret-xxxxxxxxxxxxxxxxxxxxxxx", 15*time.Minute)

	token, err := issuer.Generate(uuid.New(), domain.VerificationNone, domain.TrustStatusNormal, false)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = verifier.Verify(token)
	if err == nil {
		t.Error("expected error when verifying with a different secret")
	}
}

func TestVerify_EmptyString(t *testing.T) {
	t.Parallel()

	m := newTestManager(15 * time.Minute)
	_, err := m.Verify("")
	if err == nil {
		t.Error("expected error for empty token string")
	}
}

func TestVerify_ClaimsContainCorrectExpiry(t *testing.T) {
	t.Parallel()

	expiry := 5 * time.Minute
	m := newTestManager(expiry)

	before := time.Now()
	token, err := m.Generate(uuid.New(), domain.VerificationNone, domain.TrustStatusNormal, false)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	claims, err := m.Verify(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	// ExpiresAt should be approximately now + expiry.
	expectedExpiry := before.Add(expiry)
	diff := claims.ExpiresAt.Time.Sub(expectedExpiry)
	if diff > 2*time.Second || diff < -2*time.Second {
		t.Errorf("expiry timestamp off by %v", diff)
	}
}

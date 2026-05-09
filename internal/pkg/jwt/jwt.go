package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// Claims contains the JWT payload for TrueConnect access tokens.
type Claims struct {
	jwt.RegisteredClaims
	VerificationLevel domain.VerificationLevel `json:"ver"`
	TrustStatus       domain.TrustStatus       `json:"tst"`
}

// Manager handles JWT creation and verification.
type Manager struct {
	secret []byte
	expiry time.Duration
}

// NewManager creates a new JWT manager.
func NewManager(secret string, expiry time.Duration) *Manager {
	return &Manager{
		secret: []byte(secret),
		expiry: expiry,
	}
}

// Generate creates a new signed JWT access token.
func (m *Manager) Generate(userID uuid.UUID, verLevel domain.VerificationLevel, trustStatus domain.TrustStatus) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiry)),
		},
		VerificationLevel: verLevel,
		TrustStatus:       trustStatus,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return signed, nil
}

// Verify parses and validates a JWT token string, returning the claims.
func (m *Manager) Verify(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

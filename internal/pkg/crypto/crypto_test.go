package crypto_test

import (
	"bytes"
	"testing"

	"github.com/trueconnect/backend/internal/pkg/crypto"
)

// ── Argon2id ─────────────────────────────────────────────────────────────────

func TestHashPassword_ProducesUniqueHashes(t *testing.T) {
	t.Parallel()

	h1, err := crypto.HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatalf("hash 1 error: %v", err)
	}
	h2, err := crypto.HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatalf("hash 2 error: %v", err)
	}

	// Same password must produce different hashes (different salts).
	if h1 == h2 {
		t.Error("expected different hashes for same password; salts may not be random")
	}
}

func TestHashPassword_ContainsArgon2idMarker(t *testing.T) {
	t.Parallel()

	h, err := crypto.HashPassword("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(h) < 10 || h[:9] != "$argon2id" {
		t.Errorf("expected hash to start with $argon2id, got: %s", h[:min(len(h), 20)])
	}
}

func TestVerifyPassword_CorrectPassword(t *testing.T) {
	t.Parallel()

	password := "my-secure-password-123"
	hash, err := crypto.HashPassword(password)
	if err != nil {
		t.Fatalf("hashing: %v", err)
	}

	ok, err := crypto.VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("verifying: %v", err)
	}
	if !ok {
		t.Error("expected verification to succeed for correct password")
	}
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	t.Parallel()

	hash, err := crypto.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hashing: %v", err)
	}

	ok, err := crypto.VerifyPassword("wrong-password", hash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected verification to fail for wrong password")
	}
}

func TestVerifyPassword_InvalidHashFormat(t *testing.T) {
	t.Parallel()

	_, err := crypto.VerifyPassword("any", "not-a-valid-hash")
	if err == nil {
		t.Error("expected error for invalid hash format")
	}
}

// ── AES-256-GCM ──────────────────────────────────────────────────────────────

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	t.Parallel()

	key := bytes.Repeat([]byte("k"), 32)
	plaintext := []byte("hello, Kazakhstan!")

	ciphertext, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	decoded, err := crypto.Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	if !bytes.Equal(plaintext, decoded) {
		t.Errorf("round-trip failed: got %q, want %q", decoded, plaintext)
	}
}

func TestEncrypt_ProducesUniqueCiphertexts(t *testing.T) {
	t.Parallel()

	key := bytes.Repeat([]byte("k"), 32)
	plaintext := []byte("same data")

	c1, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt 1: %v", err)
	}
	c2, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt 2: %v", err)
	}

	// Each call uses a fresh nonce so the ciphertexts must differ.
	if bytes.Equal(c1, c2) {
		t.Error("expected unique ciphertexts due to random nonces")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	t.Parallel()

	key1 := bytes.Repeat([]byte("a"), 32)
	key2 := bytes.Repeat([]byte("b"), 32)

	ciphertext, err := crypto.Encrypt([]byte("secret"), key1)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	_, err = crypto.Decrypt(ciphertext, key2)
	if err == nil {
		t.Error("expected decryption to fail with the wrong key")
	}
}

func TestEncrypt_RejectsShortKey(t *testing.T) {
	t.Parallel()

	_, err := crypto.Encrypt([]byte("data"), []byte("tooshort"))
	if err == nil {
		t.Error("expected error for key shorter than 32 bytes")
	}
}

func TestDecrypt_RejectsTamperedCiphertext(t *testing.T) {
	t.Parallel()

	key := bytes.Repeat([]byte("k"), 32)
	ciphertext, err := crypto.Encrypt([]byte("data"), key)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	// Flip a bit near the end to tamper with the authentication tag.
	ciphertext[len(ciphertext)-1] ^= 0xFF

	_, err = crypto.Decrypt(ciphertext, key)
	if err == nil {
		t.Error("expected decryption to fail on tampered ciphertext")
	}
}

// ── SHA256 ────────────────────────────────────────────────────────────────────

func TestSHA256Hash_Deterministic(t *testing.T) {
	t.Parallel()

	data := []byte("+77001234567")
	h1 := crypto.SHA256Hash(data)
	h2 := crypto.SHA256Hash(data)

	if !bytes.Equal(h1, h2) {
		t.Error("SHA256 must be deterministic — got different results")
	}
}

func TestSHA256Hash_DifferentInputs(t *testing.T) {
	t.Parallel()

	h1 := crypto.SHA256Hash([]byte("phone-a"))
	h2 := crypto.SHA256Hash([]byte("phone-b"))

	if bytes.Equal(h1, h2) {
		t.Error("different inputs must produce different hashes")
	}
}

func TestSHA256Hash_Length(t *testing.T) {
	t.Parallel()

	h := crypto.SHA256Hash([]byte("any"))
	if len(h) != 32 {
		t.Errorf("expected 32-byte SHA-256 hash, got %d bytes", len(h))
	}
}

// min is a copy of the built-in for Go < 1.21 compatibility in test helpers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

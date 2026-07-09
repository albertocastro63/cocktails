package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
)

// GenerateResetToken returns a URL-safe, high-entropy (256-bit) secret to embed
// in the reset link. The raw token is delivered only in the email; the server
// stores only its hash.
func GenerateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashResetToken returns the SHA-256 hex digest of a reset token, which is what
// gets persisted on the user record.
func HashResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// VerifyResetToken reports whether the presented token matches the stored hash,
// using a constant-time comparison. An empty stored hash never matches.
func VerifyResetToken(token, storedHash string) bool {
	if storedHash == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(HashResetToken(token)), []byte(storedHash)) == 1
}

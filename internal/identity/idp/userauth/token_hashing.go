package userauth

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// bcryptCost is the cost parameter for bcrypt hashing (2^12 iterations).
	// Higher cost = more secure but slower. Cost 12 balances security and performance.
	// NIST recommends at least cost 10; we use 12 for defense in depth.
	bcryptCost = 12
)

var (
	// ErrInvalidToken indicates the token is empty or invalid format.
	ErrInvalidToken = errors.New("token cannot be empty")

	// ErrHashGenerationFailed indicates bcrypt hash generation failed.
	ErrHashGenerationFailed = errors.New("failed to generate token hash")

	// ErrTokenMismatch indicates the plaintext token does not match the hash.
	ErrTokenMismatch = errors.New("token does not match hash")
)

// HashToken generates a bcrypt hash of the plaintext token.
// Returns the hash as a string suitable for database storage.
// Cost parameter is fixed at bcryptCost (12) for consistency.
//
// Security notes:
//   - Uses bcrypt with cost 12 (2^12 = 4096 iterations).
//   - Hash output is ~60 bytes (base64-encoded salt + hash).
//   - Safe for concurrent use (bcrypt is stateless).
//   - CRITICAL: Store hash in database, NEVER store plaintext token.
//   - NOTE: bcrypt has 72-byte input limit. Tokens >72 bytes are SHA256 pre-hashed.
//   - SHA256 pre-hash compresses token to 32 bytes (secure as long as SHA256 collision-resistant).
func HashToken(plaintext string) (string, error) {
	if plaintext == "" {
		return "", ErrInvalidToken
	}

	// For tokens >72 bytes, pre-hash with SHA256 to compress input.
	// bcrypt input limit: 72 bytes. SHA256 output: 32 bytes (always <72).
	input := []byte(plaintext)
	if len(input) > 72 {
		hash := sha256.Sum256(input)
		input = hash[:]
	}

	hash, err := bcrypt.GenerateFromPassword(input, bcryptCost)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrHashGenerationFailed, err)
	}

	return string(hash), nil
}

// VerifyToken compares a plaintext token against a bcrypt hash.
// Returns nil if token matches, ErrTokenMismatch if mismatch.
//
// Security notes:
//   - Constant-time comparison (bcrypt.CompareHashAndPassword uses subtle.ConstantTimeCompare internally).
//   - Safe against timing attacks.
//   - Validates hash format before comparison.
//   - Handles SHA256 pre-hashed tokens transparently.
func VerifyToken(plaintext, hash string) error {
	if plaintext == "" {
		return ErrInvalidToken
	}

	if hash == "" {
		return ErrTokenMismatch // Empty hash never matches
	}

	// For tokens >72 bytes, pre-hash with SHA256 to match HashToken behavior.
	input := []byte(plaintext)
	if len(input) > 72 {
		h := sha256.Sum256(input)
		input = h[:]
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), input)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrTokenMismatch
		}

		// Other errors (e.g., malformed hash) also return mismatch.
		return fmt.Errorf("%w: %w", ErrTokenMismatch, err)
	}

	return nil
}

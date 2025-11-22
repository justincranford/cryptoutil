package userauth

import (
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
func HashToken(plaintext string) (string, error) {
	if plaintext == "" {
		return "", ErrInvalidToken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcryptCost)
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
func VerifyToken(plaintext, hash string) error {
	if plaintext == "" {
		return ErrInvalidToken
	}

	if hash == "" {
		return ErrTokenMismatch // Empty hash never matches
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrTokenMismatch
		}

		// Other errors (e.g., malformed hash) also return mismatch.
		return fmt.Errorf("%w: %w", ErrTokenMismatch, err)
	}

	return nil
}

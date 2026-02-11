// Copyright (c) 2025 Justin Cranford

package userauth

import (
	"errors"
	"fmt"

	cryptoutilSharedCryptoDigests "cryptoutil/internal/shared/crypto/digests"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
)

var (
	// ErrInvalidToken indicates the token is empty or invalid format.
	ErrInvalidToken = errors.New("token cannot be empty")

	// ErrHashGenerationFailed indicates hash generation failed.
	ErrHashGenerationFailed = errors.New("failed to generate token hash")

	// ErrTokenMismatch indicates the plaintext token does not match the hash.
	ErrTokenMismatch = errors.New("token does not match hash")
)

// HashToken generates a FIPS-approved PBKDF2-HMAC-SHA256 hash of the plaintext token.
// Returns the hash as a string suitable for database storage.
//
// Security notes:
//   - Uses PBKDF2-HMAC-SHA256 with 210,000 iterations (FIPS-140-3 approved).
//   - Hash format: pbkdf2$iterations$base64(salt)$base64(dk).
//   - Safe for concurrent use (PBKDF2 is stateless).
//   - CRITICAL: Store hash in database, NEVER store plaintext token.
func HashToken(plaintext string) (string, error) {
	if plaintext == "" {
		return "", ErrInvalidToken
	}

	hash, err := cryptoutilSharedCryptoHash.HashLowEntropyNonDeterministic(plaintext)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrHashGenerationFailed, err)
	}

	return hash, nil
}

// VerifyToken compares a plaintext token against a PBKDF2 hash.
// Returns nil if token matches, ErrTokenMismatch if mismatch.
// Supports legacy hashes during migration (read-only).
//
// Security notes:
//   - Constant-time comparison via PBKDF2 or legacy hash verification.
//   - Safe against timing attacks.
//   - Validates hash format before comparison.
func VerifyToken(plaintext, hash string) error {
	if plaintext == "" {
		return ErrInvalidToken
	}

	if hash == "" {
		return ErrTokenMismatch // Empty hash never matches
	}

	match, err := cryptoutilSharedCryptoDigests.VerifySecret(hash, plaintext)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTokenMismatch, err)
	}

	if !match {
		return ErrTokenMismatch
	}

	return nil
}

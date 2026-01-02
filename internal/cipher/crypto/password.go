// Copyright (c) 2025 Justin Cranford

package crypto

import (
	crand "crypto/rand"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/pbkdf2"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// hashPasswordWithIterations hashes a password using PBKDF2-HMAC-SHA256 with custom iteration count.
// This internal function allows testing with reduced iterations while production uses OWASP recommendations.
func hashPasswordWithIterations(password string, iterations int) ([]byte, error) {
	// Generate random salt (32 bytes = 256 bits).
	salt := make([]byte, cryptoutilMagic.PBKDF2DefaultSaltBytes)
	if _, err := crand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// PBKDF2-HMAC-SHA256 with specified iterations.
	hash := pbkdf2.Key([]byte(password), salt, iterations, cryptoutilMagic.PBKDF2DefaultHashBytes, sha256.New)

	// Concatenate salt + hash for storage.
	result := make([]byte, 0, len(salt)+len(hash))
	result = append(result, salt...)
	result = append(result, hash...)

	return result, nil
}

// verifyPasswordWithIterations verifies a password against a stored hash with custom iteration count.
// This internal function allows testing with reduced iterations while production uses OWASP recommendations.
func verifyPasswordWithIterations(password string, storedHash []byte, iterations int) (bool, error) {
	const expectedHashLength = cryptoutilMagic.PBKDF2DefaultSaltBytes + cryptoutilMagic.PBKDF2DefaultHashBytes

	if len(storedHash) != expectedHashLength {
		return false, fmt.Errorf("invalid stored hash length: expected %d bytes, got %d", expectedHashLength, len(storedHash))
	}

	// Extract salt and hash.
	salt := storedHash[:cryptoutilMagic.PBKDF2DefaultSaltBytes]
	hash := storedHash[cryptoutilMagic.PBKDF2DefaultSaltBytes:]

	// Recompute hash with same salt and specified iterations.
	computedHash := pbkdf2.Key([]byte(password), salt, iterations, cryptoutilMagic.PBKDF2DefaultHashBytes, sha256.New)

	// Constant-time comparison.
	match := compareHashes(hash, computedHash)

	return match, nil
}

// HashPassword hashes a password using PBKDF2-HMAC-SHA256 (FIPS-compliant).
// Uses OWASP 2023 recommendation: 600,000 iterations.
// Returns: salt + hash (concatenated), error.
func HashPassword(password string) ([]byte, error) {
	// Generate random salt (32 bytes = 256 bits).
	salt := make([]byte, cryptoutilMagic.PBKDF2DefaultSaltBytes)
	if _, err := crand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// PBKDF2-HMAC-SHA256 with 600,000 iterations (OWASP 2023).
	hash := pbkdf2.Key([]byte(password), salt, cryptoutilMagic.PBKDF2DefaultIterations, cryptoutilMagic.PBKDF2DefaultHashBytes, sha256.New)

	// Concatenate salt + hash for storage.
	result := make([]byte, 0, len(salt)+len(hash))
	result = append(result, salt...)
	result = append(result, hash...)

	return result, nil
}

// VerifyPassword verifies a password against a stored hash.
// Expected format: salt (32 bytes) + hash (32 bytes) = 64 bytes total.
func VerifyPassword(password string, storedHash []byte) (bool, error) {
	const expectedHashLength = cryptoutilMagic.PBKDF2DefaultSaltBytes + cryptoutilMagic.PBKDF2DefaultHashBytes

	if len(storedHash) != expectedHashLength {
		return false, fmt.Errorf("invalid stored hash length: expected %d bytes, got %d", expectedHashLength, len(storedHash))
	}

	// Extract salt and hash.
	salt := storedHash[:cryptoutilMagic.PBKDF2DefaultSaltBytes]
	hash := storedHash[cryptoutilMagic.PBKDF2DefaultSaltBytes:]

	// Recompute hash with same salt.
	computedHash := pbkdf2.Key([]byte(password), salt, cryptoutilMagic.PBKDF2DefaultIterations, cryptoutilMagic.PBKDF2DefaultHashBytes, sha256.New)

	// Constant-time comparison.
	match := compareHashes(hash, computedHash)

	return match, nil
}

// compareHashes performs constant-time comparison of two byte slices.
func compareHashes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

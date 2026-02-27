// Copyright (c) 2025 Justin Cranford
//
//

package hash

import (
	crand "crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	cryptoutilSharedCryptoDigests "cryptoutil/internal/shared/crypto/digests"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	hashHighRandomCrandReadFn = func(b []byte) (int, error) { return crand.Read(b) }
	hashHighRandomHKDFFn      = cryptoutilSharedCryptoDigests.HKDF
)

// HashHighEntropyNonDeterministic hashes a high-entropy secret (e.g., API key, token) using a random salt.
// Each invocation produces a different hash for the same input (non-deterministic).
//
// This function uses HKDF-SHA256 with cryptographically random salt generation.
// Suitable for high-entropy secrets where unpredictability is guaranteed (>= 128 bits).
//
// For low-entropy secrets (passwords, PINs), use HashLowEntropyNonDeterministic.
// For deterministic hashing, use HashHighEntropyDeterministic or HashLowEntropyDeterministic.
//
// FIPS mode is ALWAYS enabled - no configurable algorithm selection.
func HashHighEntropyNonDeterministic(secret string) (string, error) {
	return HashSecretHKDFRandom(secret)
}

// HashSecretHKDFRandom hashes a high-entropy secret using HKDF-SHA256 with random salt.
// Format: hkdf-sha256$base64(salt)$base64(dk).
func HashSecretHKDFRandom(secret string) (string, error) {
	if secret == "" {
		return "", errors.New("secret is empty")
	}

	// Generate random salt.
	salt := make([]byte, cryptoutilSharedMagic.PBKDF2DefaultSaltBytes)
	if _, err := hashHighRandomCrandReadFn(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key using HKDF-SHA256.
	dk, err := hashHighRandomHKDFFn(cryptoutilSharedMagic.SHA256, []byte(secret), salt, nil, cryptoutilSharedMagic.PBKDF2DerivedKeyLength)
	if err != nil {
		return "", fmt.Errorf("HKDF failed: %w", err)
	}

	return fmt.Sprintf("%s%s%s%s%s",
		cryptoutilSharedMagic.HKDFHashName,
		cryptoutilSharedMagic.HKDFDelimiter,
		base64.RawStdEncoding.EncodeToString(salt),
		cryptoutilSharedMagic.HKDFDelimiter,
		base64.RawStdEncoding.EncodeToString(dk)), nil
}

// VerifySecretHKDFRandom verifies a stored HKDF hash against a provided secret.
// Format: hkdf-sha256$base64(salt)$base64(dk).
func VerifySecretHKDFRandom(stored, provided string) (bool, error) {
	if stored == "" {
		return false, errors.New("stored hash is empty")
	}

	if provided == "" {
		return false, errors.New("provided secret is empty")
	}

	// Parse stored hash format: hkdf-sha256$salt$dk.
	const expectedParts = 3

	parts := strings.Split(stored, "$")

	if len(parts) != expectedParts {
		return false, fmt.Errorf("invalid HKDF hash format (expected %d parts, got %d)", expectedParts, len(parts))
	}

	hashName := parts[0]
	if hashName != cryptoutilSharedMagic.HKDFHashName {
		return false, fmt.Errorf("unsupported hash algorithm: %s (expected: %s)", hashName, cryptoutilSharedMagic.HKDFHashName)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	storedDK, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false, fmt.Errorf("failed to decode derived key: %w", err)
	}

	// Derive key from provided secret using HKDF-SHA256.
	providedDK, err := hashHighRandomHKDFFn(cryptoutilSharedMagic.SHA256, []byte(provided), salt, nil, len(storedDK))
	if err != nil {
		return false, fmt.Errorf("HKDF failed: %w", err)
	}

	// Constant-time comparison using crypto/subtle.
	const equal = 1

	return subtle.ConstantTimeCompare(storedDK, providedDK) == equal, nil
}
